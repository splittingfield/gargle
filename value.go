package gargle

import "errors"

// ValueSetter is an interface which signals that a value can be parsed from a
// string. It must be implemented by all value types.
type ValueSetter interface {
	Set(s string) error
}

// BooleanValue is an optional interface which may be implemented by bool types.
//
// Flags backed by boolean values are parsed differently from other flags. Such
// flags are true if provided, false if not, without requiring an explict value.
//
// Additionally, boolean flags also generate their negated counterparts, e.g.
// "--warn" and "--no-warn". This does not apply to positional arguments.
type BooleanValue interface {
	IsBoolean() bool
}

// AggregateValue is an optional interface which may be implemented by aggregate
// types, such as slices and maps. If present, the value's Set function will be
// called once for each instance of an argument parsed.
type AggregateValue interface {
	IsAggregate() bool
}

// Value encapsulates a the backing value of a flag or argument and its defaults.
type Value struct {
	setter   ValueSetter
	defaults []string
}

// Default sets default value(s) which are to be applied if a value is left unset.
func (v *Value) Default(s ...string) *Value {
	if len(s) > 1 && !v.IsAggregate() {
		panic("only aggregate values may have multiple defaults")
	}
	v.defaults = s
	return v
}

func (v *Value) applyDefault() error {
	for _, d := range v.defaults {
		if err := v.set(d); err != nil {
			return err
		}
	}
	return nil
}

// IsBoolean returns whether a value is backed by a boolean type.
func (v *Value) IsBoolean() bool {
	b, ok := v.setter.(BooleanValue)
	return ok && b.IsBoolean()
}

// IsAggregate returns whether a value can be set multiple times.
func (v *Value) IsAggregate() bool {
	agg, ok := v.setter.(AggregateValue)
	return ok && agg.IsAggregate()
}

func (v *Value) set(s string) error {
	if v.setter == nil {
		return errors.New("value has no type and cannot be set")
	}
	return v.setter.Set(s)
}
