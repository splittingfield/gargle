package gargle

// Arg represents a positional argument attached to a command.
type Arg struct {
	// An optional name to display in help and errors.
	Name string

	// Text describing the argument. It may be a single line or an arbitrarily
	// long description. Usage writers may assume the first line can serve
	// independently as a short-form description.
	Help string

	// Required sets the argument to generate an error when absent.
	Required bool

	// PreAction is invoked after parsing, but before values are set. All pre-actions
	// are executed unconditionally in the order encountered during parsing.
	PreAction Action

	// Underlying value for the argument, set during parsing.
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
