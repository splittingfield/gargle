package gargle

import (
	"fmt"
	"strconv"
	"time"
)

type boolValue bool

// BoolVar wraps a single boolean value.
func BoolVar(v *bool) Value { return (*boolValue)(v) }

func (v *boolValue) IsBoolean() bool { return true }
func (v *boolValue) String() string  { return strconv.FormatBool(bool(*v)) }
func (v *boolValue) Set(s string) error {
	val, err := strconv.ParseBool(s)
	if err == nil {
		*v = boolValue(val)
	}
	return err
}

type negatedValue bool

// NegatedBoolVar wraps a single boolean value. It sets v to the opposite of a
// parsed string. It's most useful when paired with a BoolVar.
func NegatedBoolVar(v *bool) Value { return (*negatedValue)(v) }

func (v *negatedValue) IsBoolean() bool { return true }
func (v *negatedValue) String() string  { return strconv.FormatBool(bool(*v)) }
func (v *negatedValue) Set(s string) error {
	val, err := strconv.ParseBool(s)
	if err == nil {
		*v = !negatedValue(val)
	}
	return err
}

type stringValue string

// StringVar wraps a string.
func StringVar(v *string) Value { return (*stringValue)(v) }

func (v *stringValue) String() string { return string(*v) }
func (v *stringValue) Set(s string) error {
	*v = stringValue(s)
	return nil
}

type stringSliceValue []string

// StringsVar wraps a slice of strings.
func StringsVar(v *[]string) Value { return (*stringSliceValue)(v) }

func (v *stringSliceValue) IsAggregate() bool { return true }
func (v *stringSliceValue) String() string    { return fmt.Sprintf("%v", *v) }
func (v *stringSliceValue) Set(s string) error {
	*v = append(*v, s)
	return nil
}

type intValue int

// IntVar wraps a signed integer with machine-dependent bit width.
func IntVar(v *int) Value { return (*intValue)(v) }

func (v *intValue) String() string { return strconv.FormatInt(int64(*v), 10) }
func (v *intValue) Set(s string) error {
	val, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err == nil {
		*v = intValue(val)
	}
	return err
}

type intSliceValue []int

// IntsVar wraps a slice of integers.
func IntsVar(v *[]int) Value { return (*intSliceValue)(v) }

func (v *intSliceValue) IsAggregate() bool { return true }
func (v *intSliceValue) String() string    { return fmt.Sprintf("%v", *v) }
func (v *intSliceValue) Set(s string) error {
	val, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err == nil {
		*v = append(*v, int(val))
	}
	return nil
}

type int64Value int64

// Int64Var wraps a 64-bit signed integer.
func Int64Var(v *int64) Value { return (*int64Value)(v) }

func (v *int64Value) String() string { return strconv.FormatInt(int64(*v), 10) }
func (v *int64Value) Set(s string) error {
	val, err := strconv.ParseInt(s, 0, 64)
	if err == nil {
		*v = int64Value(val)
	}
	return err
}

type uintValue uint

// UintVar wraps an unsigned integer with machine-dependent bit width.
func UintVar(v *uint) Value { return (*uintValue)(v) }

func (v *uintValue) String() string { return strconv.FormatUint(uint64(*v), 10) }
func (v *uintValue) Set(s string) error {
	val, err := strconv.ParseUint(s, 0, strconv.IntSize)
	if err == nil {
		*v = uintValue(val)
	}
	return err
}

type uint64Value uint64

// Uint64Var wraps a 64-bit unsigned integer.
func Uint64Var(v *uint64) Value { return (*uint64Value)(v) }

func (v *uint64Value) String() string { return strconv.FormatUint(uint64(*v), 10) }
func (v *uint64Value) Set(s string) error {
	val, err := strconv.ParseUint(s, 0, 64)
	if err == nil {
		*v = uint64Value(val)
	}
	return err
}

type float64Value float64

// Float64Var wraps a double-precision floating point.
func Float64Var(v *float64) Value { return (*float64Value)(v) }

func (v *float64Value) String() string { return strconv.FormatFloat(float64(*v), 'g', -1, 64) }
func (v *float64Value) Set(s string) error {
	val, err := strconv.ParseFloat(s, 64)
	if err == nil {
		*v = float64Value(val)
	}
	return err
}

type durationValue time.Duration

// DurationVar wraps a time duration, including units.
func DurationVar(v *time.Duration) Value { return (*durationValue)(v) }

func (v *durationValue) String() string { return time.Duration(*v).String() }
func (v *durationValue) Set(s string) error {
	val, err := time.ParseDuration(s)
	if err == nil {
		*v = durationValue(val)
	}
	return err
}
