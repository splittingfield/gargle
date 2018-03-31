package gargle

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

	// Help is arbitrary text describing the command. It may be a single line or
	// an arbitrarily long description. Usage writers generally assume the first
	// line can serve independently as a short-form description.
	Help string

	// Hidden specifies whether the command should be omitted from usage text.
	Hidden bool

	// PreAction is a function invoked after parsing but before values are set.
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
func (c *Command) AddFlag(name, help string) *Flag {
	if name == "" {
		panic("flags must have a name")
	}
	flag := &Flag{name: name, help: help}
	c.flags = append(c.flags, flag)
	return flag
}

// Flags returns a command's flags, not including those of its parents.
func (c *Command) Flags() []*Flag {
	return c.flags[:]
}

// AddArg creates a new positional argument under a command. The arg is
// automatically applied to all subcommands.
func (c *Command) AddArg(name, help string) *Arg {
	arg := &Arg{name: name, help: help}
	c.args = append(c.args, arg)
	return arg
}

// Args returns a command's positional arguments, not including those of its parents.
func (c *Command) Args() []*Arg {
	return c.args[:]
}

// Parse reads arguments and executes a command or one of its subcommands.
func (c *Command) Parse(args []string) error {
	parser := newParser(c, args)
	if err := parser.Parse(); err != nil {
		return err
	}

	// TODO:
	// Run pre-actions before setting values. This is useful for commands and
	// flags that short-circuit parsing, like '--help', '-h', '-v', 'help'...

	if err := parser.setValues(); err != nil {
		return err
	}

	if c.Action == nil {
		// TODO: Print help.
		return nil
	}
	return c.Action(parser.Context())
}
