package gargle

// Arg represents a positional argument attached to a command.
type Arg struct {
	name      string
	help      string
	hidden    bool
	required  bool
	preAction Action
	value     Value
}

// Name returns an argument's name.
func (a *Arg) Name() string { return a.name }

// Help returns an argument's description.
func (a *Arg) Help() string { return a.help }

// Hidden configures an argument to be omitted from help text.
func (a *Arg) Hidden() *Arg {
	a.hidden = true
	return a
}

// IsHidden returns whether an argument should be omitted from help text.
func (a *Arg) IsHidden() bool { return a.hidden }

// Required configures an argument to produce an error when not present.
func (a *Arg) Required() *Arg {
	a.required = true
	return a
}

// IsRequired returns whether an argument must be present.
func (a *Arg) IsRequired() bool { return a.required }

// PreAction sets a function to invoke when an argument is encountered. The action is
// run after parsing, but before values are set.
func (a *Arg) PreAction(action Action) *Arg {
	a.preAction = action
	return a
}

// AsValue configures an argument with a custom backing value.
func (a *Arg) AsValue(v ValueSetter) *Value {
	a.value = Value{setter: v}
	return &a.value
}

// Value returns an argument's backing value.
func (a *Arg) Value() *Value {
	return &a.value
}
