package gargle

// Arg represents a positional argument attached to a command.
type Arg struct {
	// Name is an optional unprefixed long form of the argument.
	// For example, "help" would match the argument "--help".
	Name string

	// Help is text describing the argument. It may be a single line or an an
	// arbitrarily long description. Usage writers generally assume the first
	// line can serve independently as a short-form description.
	Help string

	// Hidden sets whether the argument should be omitted from usage text.
	Hidden bool

	// Required sets the argument to generate an error when absent.
	Required bool

	// PreAction is a function invoked after parsing, but before values are set.
	// Each pre-action will be executed unconditionally in the order encountered
	// during parsing.
	PreAction Action

	// Value is the backing value for the argument.
	Value Value
}

func (a *Arg) invokePre(c *Command) error {
	if a.PreAction != nil {
		return a.PreAction(c)
	}
	return nil
}

func (a *Arg) setValue(s string) error {
	if a.Value != nil {
		return a.Value.Set(s)
	}
	return nil
}
