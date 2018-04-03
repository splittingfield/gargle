package gargle

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultUsageFull(t *testing.T) {
	var b *bool
	var i *int

	root := &Command{Name: "root", Help: "A root command which does something."}
	root.AddCommand(
		&Command{Name: "subcommand1", Help: "The first command does some things too."},
		&Command{Name: "sub2", Help: "The second command has long explanatory text. Perhaps there is a complex edge case which is very important for a user to know."},
	)
	root.AddFlag(
		&Flag{Name: "help", Short: 'h', Help: "Show usage"},
		&Flag{Short: 'v', Help: "Show version information"},
		&Flag{Short: 'a', Help: "Short flag with no long form", Value: IntVar(i)},
		&Flag{Name: "long-only", Help: "Long flag with no short form", Value: IntVar(i)},
		&Flag{Name: "bool", Short: 'b', Help: "Boolean flag", Value: BoolVar(b)},
		&Flag{Name: "hidden", Help: "A hidden flag", Hidden: true},
	)
	root.AddArg(
		&Arg{Name: "optional", Help: "An argument"},
		&Arg{Name: "required", Help: "A required argument", Required: true},
	)

	expected := `Usage: root [<flags>] <command>

A root command which does something.

Arguments:
  optional  An argument
  required  A required argument

Commands:
  sub2         The second command has long explanatory text. Perhaps there is a
               complex edge case which is very important for a user to know.
  subcommand1  The first command does some things too.

Options:
  -a VALUE               Short flag with no long form
  -b, --bool             Boolean flag
  -h, --help             Show usage
      --long-only VALUE  Long flag with no short form
  -v                     Show version information
`

	actual := &strings.Builder{}
	writer := UsageWriter{
		Indent:       "  ",
		Divider:      "  ",
		MaxLineWidth: 80,
		Writer:       actual,
	}

	require.NoError(t, writer.Format(root))
	assert.Equal(t, expected, actual.String())
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
