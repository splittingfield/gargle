package gargle

import "fmt"

/*
Future:
- Default help command and flags.
- Aliased commands
- Enable formatting brief help even when the help strings are long. First line?
- Prevent collisions on flags and commands
- Display long-format help (e.g. man page)
- Completion script generation
- Parse tests
- Examples
- Emit errors when a non-aggregate flag is specified multiple times?
- Improve help for boolean flags (attach "no-" prefix if the flag's default is true)
*/

// Action is a function which is invoked during or after parsing. The passed
// context is actively parsed command, i.e. the last encountered during parsing.
type Action func(context *Command) error

// Command is a hierarchical structured argument. It can serve as an application
// entry point, a command group, or both.
type Command struct {
	// Name of the command.
	Name string

	// Help is text describing the command. It may be a single line or an
	// arbitrarily long description. Usage writers generally assume the first
	// line can serve independently as a short-form description.
	Help string

	// Hidden sets whether the command should be omitted from usage text.
	Hidden bool

	// PreAction is a function invoked after parsing, but before values are set.
	// Each pre-action will be executed unconditionally in the order encountered
	// during parsing.
	PreAction Action

	// Action is a function invoked after parsing and argument validation. Only
	// the active context, i.e. the latest command parsed, is invoked.
	Action Action

	parent   *Command
	commands []*Command
	flags    []*Flag
	args     []*Arg
}

// FullName returns a command's fully qualified name.
func (c *Command) FullName() string {
	if c.parent == nil {
		return c.Name
	}
	return c.parent.FullName() + " " + c.Name
}

// Parent returns a command's parent command, if any.
func (c *Command) Parent() *Command { return c.parent }

// AddCommand adds any number of child commands. It is an error to add the same
// child command to multiple parents, or to add a command to itself.
func (c *Command) AddCommand(commands ...*Command) {
	for _, cmd := range commands {
		if cmd == c {
			panic("cannot add a command to itself")
		}
		if cmd.parent != nil {
			panic("commands may only be added to one parent")
		}
		cmd.parent = c
		c.commands = append(c.commands, cmd)
	}
}

// Commands returns a command's immediate children.
func (c *Command) Commands() []*Command {
	return c.commands[:]
}

// AddFlag creates a new flag under a command. The flag is automatically applied
// to all subcommands unless overridden by a flag with the same name.
func (c *Command) AddFlag(flags ...*Flag) { c.flags = append(c.flags, flags...) }

// Flags returns a command's flags, not including those of its parents.
func (c *Command) Flags() []*Flag {
	return c.flags[:]
}

// AddArg creates a new positional argument under a command. The arg is
// automatically applied to all subcommands.
func (c *Command) AddArg(args ...*Arg) { c.args = append(c.args, args...) }

// Args returns a command's positional arguments, not including those of its parents.
func (c *Command) Args() []*Arg {
	return c.args[:]
}

// Parse reads arguments and executes a command or one of its subcommands.
func (c *Command) Parse(args []string) error {
	parser := newParser(c, args)
	parsed, parseErr := parser.Parse()
	context := parser.Context()

	// We invoke before returning to ensure commands like "help" can run even in
	// the presence of bad flags. Invoke errors supersede parse errors.
	if err := invokePreActions(context, parsed); err != nil {
		return err
	}
	if parseErr != nil {
		return parseErr
	}

	if err := setValues(parser.Context(), parsed); err != nil {
		return err
	}

	if c.Action == nil {
		// TODO: Print help.
		return nil
	}
	return c.Action(parser.Context())
}

func (c *Command) invokePre(context *Command) error {
	if c.PreAction != nil {
		return c.PreAction(context)
	}
	return nil
}

func invokePreActions(context *Command, parsed []entity) error {
	type invocable interface{ invokePre(context *Command) error }

	for _, e := range parsed {
		val, ok := e.object.(invocable)
		if !ok {
			continue
		}

		if err := val.invokePre(context); err != nil {
			return err
		}
	}
	return nil
}

func setValues(context *Command, parsed []entity) error {
	type setter interface{ setValue(s string) error }

	// Set all values we saw during parsing.
	seen := map[interface{}]bool{}
	for _, e := range parsed {
		val, ok := e.object.(setter)
		if !ok {
			continue
		}

		seen[e.object] = true
		if err := val.setValue(e.value); err != nil {
			return fmt.Errorf("invalid argument for %s: %s", e.token, err.Error())
		}
	}

	var stack []*Command
	for c := context; c != nil; c = c.Parent() {
		stack = append(stack, c)
	}

	// Validate unset arguments/flags and apply defaults.
	for i := len(stack) - 1; i >= 0; i-- {
		command := stack[i]
		for _, flag := range command.Flags() {
			if seen[flag] {
				continue
			}
			if flag.Required {
				return fmt.Errorf("missing required flag --%s", flag.Name)
			}
			if err := applyDefault(flag.Value); err != nil {
				return err
			}
		}

		for _, arg := range command.Args() {
			if seen[arg] {
				continue
			}
			if arg.Required {
				return fmt.Errorf("missing required argument %s", arg.Name)
			}
			if err := applyDefault(arg.Value); err != nil {
				return err
			}
		}
	}
	return nil
}
