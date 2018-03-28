package gargle

import (
	"strconv"
	"time"
)

type stringValue struct{ v *string }

func (v stringValue) Set(s string) error {
	*v.v = s
	return nil
}

func (f *Flag) AsString(v *string) *Value { return f.AsValue(stringValue{v}) }
func (a *Arg) AsString(v *string) *Value  { return a.AsValue(stringValue{v}) }

type stringSliceValue struct{ v *[]string }

func (v stringSliceValue) IsAggregate() bool { return true }
func (v stringSliceValue) Set(s string) error {
	*v.v = append(*v.v, s)
	return nil
}

func (f *Flag) AsStrings(v *[]string) *Value { return f.AsValue(stringSliceValue{v}) }
func (a *Arg) AsStrings(v *[]string) *Value  { return a.AsValue(stringSliceValue{v}) }

type boolValue struct{ v *bool }

func (v boolValue) IsBoolean() bool { return true }
func (v boolValue) Set(s string) error {
	val, err := strconv.ParseBool(s)
	if err == nil {
		*v.v = val
	}
	return err
}

func (f *Flag) AsBool(v *bool) *Value { return f.AsValue(boolValue{v}) }
func (a *Arg) AsBool(v *bool) *Value  { return a.AsValue(boolValue{v}) }

type intValue struct{ v *int }

func (v intValue) Set(s string) error {
	val, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err == nil {
		*v.v = int(val)
	}
	return err
}

func (f *Flag) AsInt(v *int) *Value { return f.AsValue(intValue{v}) }
func (a *Arg) AsInt(v *int) *Value  { return a.AsValue(intValue{v}) }

type intSliceValue struct{ v *[]int }

func (v intSliceValue) IsAggregate() bool { return true }
func (v intSliceValue) Set(s string) error {
	val, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err == nil {
		*v.v = append(*v.v, int(val))
	}
	return nil
}

func (f *Flag) AsInts(v *[]int) *Value { return f.AsValue(intSliceValue{v}) }
func (a *Arg) AsInts(v *[]int) *Value  { return a.AsValue(intSliceValue{v}) }

type int8Value struct{ v *int8 }

func (v int8Value) Set(s string) error {
	val, err := strconv.ParseInt(s, 0, 8)
	if err == nil {
		*v.v = int8(val)
	}
	return err
}

func (f *Flag) AsInt8(v *int8) *Value { return f.AsValue(int8Value{v}) }
func (a *Arg) AsInt8(v *int8) *Value  { return a.AsValue(int8Value{v}) }

type int16Value struct{ v *int16 }

func (v int16Value) Set(s string) error {
	val, err := strconv.ParseInt(s, 0, 16)
	if err == nil {
		*v.v = int16(val)
	}
	return err
}

func (f *Flag) AsInt16(v *int16) *Value { return f.AsValue(int16Value{v}) }
func (a *Arg) AsInt16(v *int16) *Value  { return a.AsValue(int16Value{v}) }

type int32Value struct{ v *int32 }

func (v int32Value) Set(s string) error {
	val, err := strconv.ParseInt(s, 0, 32)
	if err == nil {
		*v.v = int32(val)
	}
	return err
}

func (f *Flag) AsInt32(v *int32) *Value { return f.AsValue(int32Value{v}) }
func (a *Arg) AsInt32(v *int32) *Value  { return a.AsValue(int32Value{v}) }

type int64Value struct{ v *int64 }

func (v int64Value) Set(s string) error {
	val, err := strconv.ParseInt(s, 0, 64)
	if err == nil {
		*v.v = val
	}
	return err
}

func (f *Flag) AsInt64(v *int64) *Value { return f.AsValue(int64Value{v}) }
func (a *Arg) AsInt64(v *int64) *Value  { return a.AsValue(int64Value{v}) }

type uintValue struct{ v *uint }

func (v uintValue) Set(s string) error {
	val, err := strconv.ParseUint(s, 0, strconv.IntSize)
	if err == nil {
		*v.v = uint(val)
	}
	return err
}

func (f *Flag) AsUInt(v *uint) *Value { return f.AsValue(uintValue{v}) }
func (a *Arg) AsUInt(v *uint) *Value  { return a.AsValue(uintValue{v}) }

type uint8Value struct{ v *uint8 }

func (v uint8Value) Set(s string) error {
	val, err := strconv.ParseUint(s, 0, 8)
	if err == nil {
		*v.v = uint8(val)
	}
	return err
}

func (f *Flag) AsUInt8(v *uint8) *Value { return f.AsValue(uint8Value{v}) }
func (a *Arg) AsUInt8(v *uint8) *Value  { return a.AsValue(uint8Value{v}) }

type uint16Value struct{ v *uint16 }

func (v uint16Value) Set(s string) error {
	val, err := strconv.ParseUint(s, 0, 16)
	if err == nil {
		*v.v = uint16(val)
	}
	return err
}

func (f *Flag) AsUInt16(v *uint16) *Value { return f.AsValue(uint16Value{v}) }
func (a *Arg) AsUInt16(v *uint16) *Value  { return a.AsValue(uint16Value{v}) }

type uint32Value struct{ v *uint32 }

func (v uint32Value) Set(s string) error {
	val, err := strconv.ParseUint(s, 0, 32)
	if err == nil {
		*v.v = uint32(val)
	}
	return err
}

func (f *Flag) AsUInt32(v *uint32) *Value { return f.AsValue(uint32Value{v}) }
func (a *Arg) AsUInt32(v *uint32) *Value  { return a.AsValue(uint32Value{v}) }

type uint64Value struct{ v *uint64 }

func (v uint64Value) Set(s string) error {
	val, err := strconv.ParseUint(s, 0, 64)
	if err == nil {
		*v.v = val
	}
	return err
}

func (f *Flag) AsUInt64(v *uint64) *Value { return f.AsValue(uint64Value{v}) }
func (a *Arg) AsUInt64(v *uint64) *Value  { return a.AsValue(uint64Value{v}) }

type float32Value struct{ v *float32 }

func (v float32Value) Set(s string) error {
	val, err := strconv.ParseFloat(s, 32)
	if err == nil {
		*v.v = float32(val)
	}
	return err
}

func (f *Flag) AsFloat32(v *float32) *Value { return f.AsValue(float32Value{v}) }
func (a *Arg) AsFloat32(v *float32) *Value  { return a.AsValue(float32Value{v}) }

type float64Value struct{ v *float64 }

func (v float64Value) Set(s string) error {
	val, err := strconv.ParseFloat(s, 64)
	if err == nil {
		*v.v = val
	}
	return err
}

func (f *Flag) AsFloat64(v *float64) *Value { return f.AsValue(float64Value{v}) }
func (a *Arg) AsFloat64(v *float64) *Value  { return a.AsValue(float64Value{v}) }

type durationValue struct{ v *time.Duration }

func (v durationValue) Set(s string) error {
	val, err := time.ParseDuration(s)
	if err == nil {
		*v.v = val
	}
	return err
}

func (f *Flag) AsDuration(v *time.Duration) *Value { return f.AsValue(durationValue{v}) }
func (a *Arg) AsDuration(v *time.Duration) *Value  { return a.AsValue(durationValue{v}) }
