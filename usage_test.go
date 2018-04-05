package gargle

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteCommandHelp(t *testing.T) {
	root := &Command{Name: "root"}
	sub1 := &Command{Name: "sub1"}
	sub2 := &Command{Name: "sub2"}
	root.AddCommands(sub1)
	sub1.AddCommands(sub2)

	cases := map[string]struct {
		args     []string
		expected string
		err      string
	}{
		"Parent": {expected: "root"},
		"Child": {
			args:     []string{"sub1"},
			expected: "root sub1",
		},
		"GrandChild": {
			args:     []string{"sub1", "sub2"},
			expected: "root sub1 sub2",
		},
		"ChildNotFound": {
			args: []string{"nothere"},
			err:  `"root nothere" is not a valid command`,
		},
		"GrandChildNotFound": {
			args: []string{"sub1", "nothere"},
			err:  `"root sub1 nothere" is not a valid command`,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			b := &strings.Builder{}
			write := func(context *Command) error {
				b.WriteString(context.FullName())
				return nil
			}

			err := writeCommandHelp(write, root, c.args)
			if c.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, c.expected, b.String())
			} else {
				assert.EqualError(t, err, c.err)
			}
		})
	}
}

func ExampleDefaultUsage() {
	var b bool
	var i int
	var s []string

	root := &Command{Name: "root", Help: "A root command which does something."}
	root.AddCommands(
		NewHelpCommand(nil),
		&Command{Name: "subcommand1", Help: "The first subcommand does some things too."},
		&Command{Name: "sub2", Help: "The second command has long explanatory text. Perhaps there is a complex edge case which is very important for a user to know."},
	)
	root.AddFlags(
		NewHelpFlag(nil),
		&Flag{Short: 'v', Help: "Show version information"},
		&Flag{Short: 'a', Help: "Short flag with no long form", Value: IntVar(&i)},
		&Flag{Name: "long-only", Help: "Long flag with no short form", Value: IntVar(&i)},
		&Flag{Name: "bool", Short: 'b', Help: "Boolean flag", Value: BoolVar(&b)},
		&Flag{Name: "hidden", Help: "A hidden flag", Hidden: true},
		&Flag{
			Name: "slice", Short: 's',
			Placeholder: "STR",
			Help:        "Aggregate value with custom placeholder",
			Value:       StringsVar(&s),
		},
	)
	root.Parse([]string{"help"})

	// Output:
	// Usage: root [<flags>] <command>
	//
	// A root command which does something.
	//
	// Commands:
	//   help         Show usage
	//   sub2         The second command has long explanatory text. Perhaps there is a
	//                complex edge case which is very important for a user to know.
	//   subcommand1  The first subcommand does some things too.
	//
	// Options:
	//   -a VALUE               Short flag with no long form
	//   -b, --bool             Boolean flag
	//   -h, --help             Show usage
	//       --long-only VALUE  Long flag with no short form
	//   -s, --slice STR...     Aggregate value with custom placeholder
	//   -v                     Show version information
}

func TestUsageWriter(t *testing.T) {
	b := &strings.Builder{} // Must be reset before each test.
	writer := UsageWriter{
		Indent:         "++",
		Divider:        "||",
		MaxFirstColumn: 35,
		MaxLineWidth:   80,
		Writer:         b,
	}

	b.Reset()
	t.Run("Minimal", func(t *testing.T) {
		assert.NoError(t, writer.Format(&Command{Name: "name"}))
		assert.Equal(t, "Usage: name\n", b.String())
	})

	b.Reset()
	t.Run("InheritedFlags", func(t *testing.T) {
		root := &Command{Name: "root"}
		sub := &Command{Name: "sub"}
		root.AddCommands(sub)
		root.AddFlags(NewHelpFlag(nil))
		sub.AddFlags(&Flag{Name: "flag", Short: 'f', Help: "A subcommand flag"})

		expected := strings.Join([]string{
			"Usage: root sub [<flags>]",
			"",
			"Options:",
			"++-f, --flag||A subcommand flag",
			"++-h, --help||Show usage",
			"",
		}, "\n")

		assert.NoError(t, writer.Format(sub))
		assert.Equal(t, expected, b.String())
	})

	b.Reset()
	t.Run("RequiredArgs", func(t *testing.T) {
		cmd := &Command{Name: "command"}
		cmd.AddArgs(
			&Arg{Name: "arg1", Help: "First arg"},
			&Arg{Name: "arg2", Help: "Second arg", Required: true},
			&Arg{Name: "arg3", Help: "Third arg"},
		)

		expected := strings.Join([]string{
			"Usage: command <arg1> <arg2> [<arg3>]",
			"",
			"Arguments:",
			"++arg1||First arg",
			"++arg2||Second arg",
			"++arg3||Third arg",
			"",
		}, "\n")

		assert.NoError(t, writer.Format(cmd))
		assert.Equal(t, expected, b.String())
	})
}

func TestSortCommands(t *testing.T) {
	expected := []*Command{
		{Name: "a"},
		{Name: "apple"},
		{Name: "one"},
		{Name: "two"},
	}

	sorted := []*Command{
		{Name: "apple"},
		{Name: "two"},
		{Name: "a"},
		{Name: "one"},
	}
	sort.Sort(commandSlice(sorted))
	assert.Equal(t, expected, sorted)
}

func TestSortFlags(t *testing.T) {
	expected := []*Flag{
		{Short: '3'},
		{Short: 'a'},
		{Name: "apple"},
		{Name: "one", Short: '1'},
		{Name: "two"},
	}

	sorted := []*Flag{
		{Name: "apple"},
		{Name: "two"},
		{Short: 'a'},
		{Name: "one", Short: '1'},
		{Short: '3'},
	}
	sort.Sort(flagSlice(sorted))
	assert.Equal(t, expected, sorted)
}
