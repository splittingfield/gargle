package gargle

// Flag is a single named/value argument pair
type Flag struct {
	// Name is an optional unprefixed long form of the flag.
	// For example, "help" would match the argument "--help".
	Name string

	// Text describing the command. It may be a single line or an arbitrarily
	// long description. Usage writers may assume the first line can serve
	// independently as a short-form description.
	Help string

	// Placeholder is an optional override for the name of a flag's value.
	Placeholder string

	// Short is an optional single-character short form for the flag.
	// For example, 'h' would match the argument "-h".
	Short rune

	// Hidden sets whether the flag should be omitted from usage text.
	Hidden bool

	// Required sets the flag to generate an error when absent.
	Required bool

	// PreAction is invoked after parsing, but before values are set. All pre-actions
	// are executed unconditionally in the order encountered during parsing.
	PreAction Action

	// Underlying value for the flag, set during parsing.
	Value Value
}

func (f *Flag) invokePre(c *Command) error {
	if f.PreAction != nil {
		return f.PreAction(c)
	}
	return nil
}

func (f *Flag) setValue(s string) error {
	if f.Value != nil {
		return f.Value.Set(s)
	}
	return nil
}
