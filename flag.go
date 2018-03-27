package gargle

// Flag is a single named/value argument pair
type Flag struct {
	name        string
	help        string
	placeholder string
	short       rune
	hidden      bool
	required    bool
	preAction   Action
	value       Value
}

// Name returns a flag's name.
func (f *Flag) Name() string { return f.name }

// Help returns a flag's description.
func (f *Flag) Help() string { return f.help }

// WithPlaceholder overrides the placeholder name for a flag's value.
func (f *Flag) WithPlaceholder(name string) *Flag {
	f.placeholder = name
	return f
}

// Placeholder returns the name of a flag's value. By default this is the flag's
// name in all caps.
func (f *Flag) Placeholder() string { return f.placeholder }

// WithShort configures a flag with a single-character short form.
func (f *Flag) WithShort(s rune) *Flag {
	f.short = s
	return f
}

// Short returns a flag's single-character short form.
func (f *Flag) Short() rune { return f.short }

// Hidden configures a flag to be omitted from help text.
func (f *Flag) Hidden() *Flag {
	f.hidden = true
	return f
}

// IsHidden returns whether a flag should be omitted from help text.
func (f *Flag) IsHidden() bool { return f.hidden }

// Required configures a flag to produce an error when not present.
func (f *Flag) Required() *Flag {
	f.required = true
	return f
}

// IsRequired returns whether a flag must be present.
func (f *Flag) IsRequired() bool { return f.required }

// PreAction sets a function to invoke when a flag is encountered. The action is
// run after parsing, but before values are set.
func (f *Flag) PreAction(action Action) *Flag {
	f.preAction = action
	return f
}

// AsValue configures a flag with a custom backing value.
func (f *Flag) AsValue(v ValueSetter) *Value {
	f.value = Value{setter: v}
	return &f.value
}

// Value returns a flag's backing value.
func (f *Flag) Value() *Value {
	return &f.value
}
