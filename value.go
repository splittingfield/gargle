package gargle

// Value is an interface implemented by all value types.
type Value interface {
	String() string
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

// IsBoolean returns whether a value is of boolean type.
func IsBoolean(v Value) bool {
	b, ok := v.(BooleanValue)
	return ok && b.IsBoolean()
}

// AggregateValue is an optional interface which may be implemented by aggregate
// types, such as slices and maps. If present, the value's Set function will be
// called once for each instance of an argument parsed.
type AggregateValue interface {
	IsAggregate() bool
}

// IsAggregate returns whether a value can be set multiple times.
func IsAggregate(v Value) bool {
	agg, ok := v.(AggregateValue)
	return ok && agg.IsAggregate()
}

type defaultValue struct {
	value    Value
	defaults []string
}

// WithDefault wraps a value with string default value(s) which will be applied
// after parsing if (and only if) a value is left unset.
func WithDefault(v Value, s ...string) Value {
	if len(s) > 1 && !IsAggregate(v) {
		panic("only aggregate values may have multiple defaults")
	}
	return defaultValue{v, s}
}

func (v defaultValue) String() string     { return v.value.String() }
func (v defaultValue) Set(s string) error { return v.value.Set(s) }
func applyDefault(v Value) error {
	def, ok := v.(defaultValue)
	if ok {
		for _, d := range def.defaults {
			if err := def.value.Set(d); err != nil {
				return err
			}
		}
	}
	return nil
}
