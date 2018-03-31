package gargle

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: Test valuless types.
// TODO: Test hierarchy.
// TODO: Test args (incl. verbatim)
// TODO: Test required flags/args.

func TestParseMinimal(t *testing.T) {
	action := testAction{}
	command := Command{Action: action.invoke}
	assert.NoError(t, command.Parse(nil))
	assert.Equal(t, command, action.result)
}

func TestParseFlags(t *testing.T) {
	action := testAction{}
	var i int
	var s string
	var b bool
	var a []int

	command := Command{Action: action.invoke}
	command.AddFlag("int", "").WithShort('i').AsInt(&i)
	command.AddFlag("string", "").WithShort('s').AsString(&s).Default("default")
	command.AddFlag("bool", "").WithShort('b').AsBool(&b)
	command.AddFlag("array", "").WithShort('a').AsInts(&a)

	cases := map[string]struct {
		args []string
		err  string
		i    int
		s    string
		b    bool
		a    []int
	}{
		"NoArgs": {s: "default"},
		"LongFlagWithValue": {
			args: []string{"--int", "42"},
			i:    42, s: "default", b: false, a: nil,
		},
		"LongFlagWithoutValue": {
			args: []string{"--int"},
			err:  "--int must have a value",
		},
		"ShortFlagWithValue": {
			args: []string{"-i", "42"},
			i:    42, s: "default", b: false, a: nil,
		},
		"ShortFlagWithoutValue": {
			args: []string{"-i"},
			err:  "-i must have a value",
		},
		"BooleanLongFlag": {
			args: []string{"--bool"},
			i:    0, s: "default", b: true, a: nil,
		},
		"BooleanShortFlag": {
			args: []string{"-b"},
			i:    0, s: "default", b: true, a: nil,
		},
		"SeveralFlags": {
			args: []string{"-i27", "--bool", "--string", "seven"},
			i:    27, s: "seven", b: true, a: nil,
		},
		"RepeatedSingleValue": {
			args: []string{"-i", "39", "--int", "-12"},
			i:    -12, s: "default", b: false, a: nil,
		},
		"RepeatedAggregate": {
			args: []string{"-a", "39", "--array", "-12"},
			i:    0, s: "default", b: false, a: []int{39, -12},
		},
		"RepeatedBoolean": {
			args: []string{"-b", "--bool", "--bool"},
			i:    0, s: "default", b: true, a: nil,
		},
		"LongInvalidType": {
			args: []string{"--int=foo"},
			err:  `invalid argument for --int: strconv.ParseInt: parsing "foo": invalid syntax`,
		},
		"ShortInvalidType": {
			args: []string{"-ifoo"},
			err:  `invalid argument for -i: strconv.ParseInt: parsing "foo": invalid syntax`,
		},
	}

	for name, c := range cases {
		action.reset()
		i = 0
		s = ""
		b = false
		a = nil

		t.Run(name, func(t *testing.T) {
			err := command.Parse(c.args)
			if c.err != "" {
				assert.EqualError(t, err, c.err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, c.i, i)
			assert.Equal(t, c.s, s)
			assert.Equal(t, c.b, b)
			assert.Equal(t, c.a, a)
		})
	}
}

type testAction struct {
	result *Command
}

func (a *testAction) reset() { a.result = nil }
func (a *testAction) invoke(context *Command) error {
	a.result = context
	return nil
}
