// Package gargle implements a library for command-line parsing.
package gargle

import "fmt"

// Action is a function which is invoked during or after parsing. The passed
// context is actively parsed command, i.e. the last encountered during parsing.
type Action func(context *Command) error

// Command is a hierarchical structured argument. It can serve as an application
// entry point, a command group, or both. For exmple, in the command line
// "go test .", "go" is a root command and "test" is a subcommand of "go".
type Command struct {
	// Name of the command.
	Name string

	// Text describing the command. It may be a single line or an arbitrarily
	// long description. Usage writers may assume the first line can serve
	// independently as a short-form description.
	Help string

	// Hidden sets whether the command should be omitted from usage text.
	Hidden bool

	// PreAction is invoked after parsing, but before values are set. All pre-actions
	// are executed unconditionally in the order encountered during parsing.
	PreAction Action

	// Action invoked after parsing and argument validation. Only the active
	// context, i.e. the last command parsed, is invoked.
	Action Action

	// Client-defined labels for grouping and processing commands.
	Labels map[string]string

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

// AddCommands adds any number of child commands. It is an error to add the same
// child command to multiple parents, or to add a command to itself.
func (c *Command) AddCommands(commands ...*Command) {
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

// AddFlags creates a new flag under a command. Flags are inherited by subcommands
// unless overridden by a flag with the same name.
func (c *Command) AddFlags(flags ...*Flag) {
	for _, flag := range flags {
		if flag.Name == "" && flag.Short == rune(0) {
			panic("flags may not be anonymous")
		}
		c.flags = append(c.flags, flag)
	}
}

// Flags returns a command's flags, not including those of its parents.
func (c *Command) Flags() []*Flag {
	return c.flags[:]
}

// FullFlags returns a command's flags along with those of its parents. Flags
// are ordered parent to child.
func (c *Command) FullFlags() []*Flag {
	if c.parent == nil {
		return c.flags[:]
	}
	return append(c.parent.FullFlags(), c.flags...)
}

// AddArgs creates a new positional argument under a command.
func (c *Command) AddArgs(args ...*Arg) { c.args = append(c.args, args...) }

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

	if err := setValues(context, parsed); err != nil {
		return err
	}

	if c.Action != nil {
		return context.Action(context)
	}
	return nil
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
		val, ok := e.Option.(invocable)
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
		val, ok := e.Option.(setter)
		if !ok {
			continue
		}

		seen[e.Option] = true
		if err := val.setValue(e.Value); err != nil {
			return fmt.Errorf("invalid value %q for %s: %s", e.Value, e.Name, err.Error())
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
