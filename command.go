package gargle

import "errors"

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

// Command is a parseable argument which executes an action when invoked. It may
// optionally include any number of subcommands, and may itself be a subcommand.
type Command struct {
	name      string
	help      string
	hidden    bool
	preAction Action
	action    Action
	parent    *Command

	commands []*Command
	flags    []*Flag
	args     []*Arg
}

// NewCommand creates a new root-level command. This is often an entry point to an application.
func NewCommand(name, help string) *Command {
	if name == "" {
		panic("commands must have a name")
	}
	return &Command{name: name, help: help}
}

// Name returns a command's name.
func (c *Command) Name() string { return c.name }

// FullName returns a command's fully qualified name.
func (c *Command) FullName() string {
	if c.parent == nil {
		return c.name
	}
	return c.parent.FullName() + " " + c.name
}

// Help returns a command's description.
func (c *Command) Help() string { return c.help }

// Hidden configures a command to be omitted from help text.
func (c *Command) Hidden() *Command {
	c.hidden = true
	return c
}

// IsHidden returns whether a command should be omitted from help text.
func (c *Command) IsHidden() bool { return c.hidden }

// PreAction sets a function to be run after parsing, but before values are set.
// Pre-actions are executed parent-to-child, in order of declaration.
func (c *Command) PreAction(action Action) *Command {
	c.preAction = action
	return c
}

// Action assigns a function to call when a command is invoked. Only the active
// command's action is invoked; parent actions are ignored.
func (c *Command) Action(action Action) *Command {
	c.action = action
	return c
}

// Parent returns a command's parent command, if any.
func (c *Command) Parent() *Command { return c.parent }

// AddSubcommand creates and returns a new subcommand.
func (c *Command) AddSubcommand(name, help string) *Command {
	child := NewCommand(name, help)
	child.parent = c
	c.commands = append(c.commands, child)
	return child
}

// Subcommands returns a command's subcommnads.
func (c *Command) Subcommands() []*Command {
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
	if lastArg := len(c.args) - 1; lastArg >= 0 && c.args[lastArg].Value().IsAggregate() {
		panic("no positional argument may follow an aggregate arg")
	}
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
	return errors.New("not implemented")
}
