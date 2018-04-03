package gargle

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddCommand(t *testing.T) {
	parent := &Command{Name: "root"}
	child1 := &Command{Name: "first"}
	child2 := &Command{Name: "second"}
	parent.AddCommand(child1)
	child1.AddCommand(child2)

	assert.Equal(t, []*Command{child1}, parent.Commands())
	assert.Equal(t, []*Command{child2}, child1.Commands())
	assert.Equal(t, "root", parent.FullName())
	assert.Equal(t, "root first", child1.FullName())
	assert.Equal(t, "root first second", child2.FullName())

	assert.Panics(t, func() { parent.AddCommand(parent) }, "A command can't be added to itself.")
	assert.Panics(t, func() { parent.AddCommand(child2) }, "A command can't have multiple parents.")
}

// TODO: Test required flags/args.

func TestParseMinimal(t *testing.T) {
	action := &testAction{}
	command := &Command{Action: action.Invoke}
	assert.NoError(t, command.Parse(nil))
	assert.Equal(t, command, action.Result)
}

func TestParseFlags(t *testing.T) {
	action := &testAction{}
	var i int
	var s string
	var b bool
	var a []int

	command := &Command{Action: action.Invoke}
	command.AddFlag(
		&Flag{Name: "int", Short: 'i', Value: IntVar(&i)},
		&Flag{Name: "string", Short: 's', Value: WithDefault(StringVar(&s), "default")},
		&Flag{Name: "bool", Short: 'b', Value: BoolVar(&b)},
		&Flag{Name: "array", Short: 'a', Value: IntsVar(&a)},
	)

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
			err:  "--int requires a value",
		},
		"ShortFlagWithValue": {
			args: []string{"-i", "42"},
			i:    42, s: "default", b: false, a: nil,
		},
		"ShortFlagWithoutValue": {
			args: []string{"-i"},
			err:  "-i requires a value",
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
			args: []string{"-i-27", "--bool", "--string", "seven"},
			i:    -27, s: "seven", b: true, a: nil,
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
		"LongNoValue": {
			args: []string{"--int="},
			err:  `invalid value "" for --int: strconv.ParseInt: parsing "": invalid syntax`,
		},
		"LongInvalidType": {
			args: []string{"--int=foo"},
			err:  `invalid value "foo" for --int: strconv.ParseInt: parsing "foo": invalid syntax`,
		},
		"ShortInvalidType": {
			args: []string{"-ifoo"},
			err:  `invalid value "foo" for -i: strconv.ParseInt: parsing "foo": invalid syntax`,
		},
		"UnknownLongFlag": {
			args: []string{"--not-here"},
			err:  "unknown flag: not-here",
		},
		"UnknownShortFlag": {
			args: []string{"-n"},
			err:  "unknown flag: n",
		},
	}

	for name, c := range cases {
		action.Reset()
		i = 0
		s = ""
		b = false
		a = nil

		t.Run(name, func(t *testing.T) {
			err := command.Parse(c.args)
			if c.err == "" {
				require.NoError(t, err)
				assert.Equal(t, command, action.Result)
			} else {
				assert.Nil(t, action.Result)
				assert.EqualError(t, err, c.err)
				return
			}

			assert.Equal(t, c.i, i)
			assert.Equal(t, c.s, s)
			assert.Equal(t, c.b, b)
			assert.Equal(t, c.a, a)
		})
	}
}

func TestParseArgs(t *testing.T) {
	action := &testAction{}
	var i int
	var s string
	var b bool
	var a []int
	var ignored *int // This should never be set; crash if it is.

	command := &Command{Action: action.Invoke}
	command.AddArg(
		&Arg{Name: "int", Value: IntVar(&i)},
		&Arg{Name: "string", Value: StringVar(&s)},
		&Arg{Name: "bool", Value: BoolVar(&b)},
		&Arg{Name: "array", Value: IntsVar(&a)},
		&Arg{Name: "ignored", Value: IntVar(ignored)},
	)

	cases := map[string]struct {
		args []string
		err  string
		i    int
		s    string
		b    bool
		a    []int
	}{
		"NoArgs": {},
		"OneValue": {
			args: []string{"42"},
			i:    42, s: "", b: false, a: nil,
		},
		"AllValues": {
			args: []string{"1", "two", "true", "3", "4", "5"},
			i:    1, s: "two", b: true, a: []int{3, 4, 5},
		},
		"InvalidType": {
			args: []string{"foo"},
			err:  `invalid value "foo" for int: strconv.ParseInt: parsing "foo": invalid syntax`,
		},
		"ForceVerbatim": {
			args: []string{"81", "--", "--foo"},
			i:    81, s: "--foo", b: false, a: nil,
		},
		// "Too many args" case is covered in TestParseCommands.
	}

	for name, c := range cases {
		action.Reset()
		i = 0
		s = ""
		b = false
		a = nil

		t.Run(name, func(t *testing.T) {
			err := command.Parse(c.args)
			if c.err == "" {
				require.NoError(t, err)
				assert.Equal(t, command, action.Result)
			} else {
				assert.Nil(t, action.Result)
				assert.EqualError(t, err, c.err)
				return
			}

			assert.Equal(t, c.i, i)
			assert.Equal(t, c.s, s)
			assert.Equal(t, c.b, b)
			assert.Equal(t, c.a, a)
		})
	}
}

func TestParseCommands(t *testing.T) {
	var rootArg *string // Intentionally null; test should fail if this is ever set.
	var flag int
	var arg string

	rootAction := &testAction{}
	root := &Command{Name: "root", Action: rootAction.Invoke}
	sub1Action := &testAction{}
	sub1 := &Command{Name: "sub1", Action: sub1Action.Invoke}
	sub2Action := &testAction{}
	sub2 := &Command{Name: "sub2", Action: sub2Action.Invoke}
	subSubAction := &testAction{}
	subSub := &Command{Name: "sub-sub", Action: subSubAction.Invoke}

	root.AddArg(&Arg{Name: "root-arg", Value: StringVar(rootArg)})
	root.AddCommand(sub1, sub2)
	sub1.AddFlag(&Flag{Name: "flag", Value: IntVar(&flag)})
	sub1.AddCommand(subSub)
	subSub.AddArg(&Arg{Name: "arg", Value: StringVar(&arg)})

	cases := map[string]struct {
		args    []string
		err     string
		invoked *testAction
		context *Command
		flag    int
		arg     string
	}{
		// Command ordering...
		"NoArgs": {
			invoked: rootAction,
			context: root,
		},
		"ChildWithChildren": {
			args:    []string{"sub1"},
			invoked: sub1Action,
			context: sub1,
		},
		"LeafCommand": {
			args:    []string{"sub2"},
			invoked: sub2Action,
			context: sub2,
		},
		"MissingParentCommand": {
			args: []string{"sub-sub"},
			err:  `"root sub-sub" is not a valid command`,
		},
		"NestedLeafCommand": {
			args:    []string{"sub1", "sub-sub"},
			invoked: subSubAction,
			context: subSub,
		},

		// Extra arguments...
		"UnknownChild": {
			args: []string{"missing"},
			err:  `"root missing" is not a valid command`,
		},
		"UnknownChildSubcommand": {
			args: []string{"sub1", "missing"},
			err:  `"root sub1 missing" is not a valid command`,
		},
		"LeafWithExtraArg": {
			args: []string{"sub2", "missing"},
			err:  `unexpected argument: "missing"`,
		},
		"NestedLeafWithExtraArg": {
			args: []string{"sub1", "sub-sub", "first", "second"},
			err:  `unexpected argument: "second"`,
		},

		// Flag inheritance...
		"UnknownChildFlag": {
			args: []string{"--flag", "42"},
			err:  "unknown flag: flag",
		},
		"ChildWithFlag": {
			args:    []string{"sub1", "--flag=27"},
			invoked: sub1Action,
			context: sub1,
			flag:    27,
		},
		"ChildWithoutFlag": {
			args: []string{"sub2", "--flag=42"},
			err:  "unknown flag: flag",
		},
		"MisorderedChildFlag": {
			args: []string{"--flag=42", "sub1"},
			err:  "unknown flag: flag",
		},
		"FlagBetweenCommands": {
			args:    []string{"sub1", "--flag", "137", "sub-sub"},
			invoked: subSubAction,
			context: subSub,
			flag:    137,
		},
		"FlagAfterChild": {
			args:    []string{"sub1", "sub-sub", "some arg", "--flag=-10"},
			invoked: subSubAction,
			context: subSub,
			flag:    -10,
			arg:     "some arg",
		},
	}

	for name, c := range cases {
		actions := []*testAction{rootAction, sub1Action, sub2Action, subSubAction}
		for _, action := range actions {
			action.Reset()
		}
		flag = 0
		arg = ""

		t.Run(name, func(t *testing.T) {
			for _, action := range actions {
				if c.invoked != action {
					assert.Nil(t, action.Result, "Unexpected action invocation.")
				}
			}

			err := root.Parse(c.args)
			if c.err == "" {
				require.NoError(t, err)
				require.NotNil(t, c.invoked.Result, "Expected invocation for: "+c.context.FullName())
				assert.Equal(t, c.context, c.invoked.Result)
			} else {
				assert.EqualError(t, err, c.err)
				return
			}

			assert.Equal(t, c.flag, flag)
			assert.Equal(t, c.arg, arg)
		})
	}
}

func TestParseNilValue(t *testing.T) {
	command := &Command{}
	command.AddFlag(&Flag{Name: "flag", Short: 'f'})
	command.AddArg(&Arg{Name: "arg"})

	cases := map[string]struct {
		args []string
		err  string
	}{
		"NoArgs": {},
		"NoValue": {
			args: []string{"--flag", "-f"},
		},
		"LongValue": {
			args: []string{"--flag=test"},
			err:  `invalid value "test" for --flag: flag does not accept a value`,
		},
		"ShortValue": {
			args: []string{"-ftest"},
			err:  `unknown flag: t`,
		},
		"ArgValue": {
			args: []string{"test"},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			err := command.Parse(c.args)
			if c.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, c.err)
			}
		})
	}
}

func TestParseBoolValue(t *testing.T) {
	// This test covers explicit flag overlap as one might find in negated bools.
	var value bool
	command := &Command{}
	command.AddFlag(&Flag{Name: "flag", Short: 'f', Value: BoolVar(&value)})
	command.AddFlag(&Flag{Name: "no-flag", Value: NegatedBoolVar(&value)})

	cases := map[string]struct {
		args     []string
		expected bool
	}{
		"NoArgs":           {},
		"Positive":         {[]string{"--flag"}, true},
		"PositiveShort":    {[]string{"-f"}, true},
		"Negative":         {[]string{"--no-flag"}, false},
		"OverridePositive": {[]string{"--no-flag", "--flag"}, true},
		"OverrideNegative": {[]string{"--flag", "--no-flag"}, false},
		"ExplicitPositive": {[]string{"--flag=true"}, true},
		"ExplicitNegative": {[]string{"--flag=false"}, false},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, command.Parse(c.args))
			assert.Equal(t, c.expected, value)
		})
	}
}

type testAction struct {
	Result *Command
}

func (a *testAction) Reset() { a.Result = nil }
func (a *testAction) Invoke(context *Command) error {
	a.Result = context
	return nil
}
