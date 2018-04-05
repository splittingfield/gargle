package gargle

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"syscall"
	"unsafe"

	"github.com/ckarenz/wordwrap"
)

// NewHelpFlag creates a standard help flag which invokes the an action when
// parsed. This flag should be attached to the root command.
func NewHelpFlag(writeHelp Action) *Flag {
	if writeHelp == nil {
		writeHelp = DefaultUsage()
	}

	return &Flag{
		Name: "help", Short: 'h',
		Help: "Show usage",
		PreAction: func(context *Command) error {
			writeHelp(context)
			os.Exit(0)
			return nil
		},
	}
}

// NewHelpCommand creates a standard help command, which prints help for a
// given subcommand. If no arguments are passed, it prints its parent's help.
// This should be added to each command group.
func NewHelpCommand(writeHelp Action) *Command {
	if writeHelp == nil {
		writeHelp = DefaultUsage()
	}

	args := new([]string)
	cmd := &Command{
		Name: "help",
		Help: "Show usage",
		PreAction: func(context *Command) error {
			writeCommandHelp(writeHelp, context.Parent(), *args)
			os.Exit(0)
			return nil
		},
	}
	cmd.AddArgs(&Arg{
		Name:  "command",
		Help:  "Show help for subcommand(s).",
		Value: StringsVar(args),
	})
	return cmd
}

func writeCommandHelp(writeHelp Action, context *Command, args []string) error {
nextArg:
	for _, arg := range args {
		for _, cmd := range context.Commands() {
			if cmd.Name == arg {
				context = cmd
				continue nextArg
			}
		}
		return fmt.Errorf("%q is not a valid command", context.FullName()+" "+arg)
	}
	return writeHelp(context)
}

var defaultUsage = &UsageWriter{Indent: "  ", Divider: "  ", MaxFirstColumn: 35}

// DefaultUsage returns the default usage writer as an action. This is usually
// attached to help flags and commands as a PreAction.
func DefaultUsage() Action {
	return defaultUsage.Format
}

// UsageWriter is a configurable usage formatter which can be used as a safe
// default for most applications.
type UsageWriter struct {
	// Indent is an indentation prefix for subsections.
	Indent string

	// Divider is a string to print between columns.
	Divider string

	// MaxFirstColumn is the maximum width of the left column in split sections.
	MaxFirstColumn int

	// MaxLineWidth overrides the maximum width of each line of text, defaults
	// to terminal width in TTY or 80 columns otherwise.
	MaxLineWidth int

	// Writer overrides the default writer, default os.Stdout.
	Writer io.Writer
}

// Format writes a given command's usage using the writer's configuration.
func (u *UsageWriter) Format(command *Command) error {
	w := u.Writer
	if w == nil {
		w = os.Stdout
	}

	var subs commandSlice
	for _, cmd := range command.Commands() {
		if !cmd.Hidden {
			subs = append(subs, cmd)
		}
	}
	sort.Sort(subs)

	var flags flagSlice
	for _, flag := range command.FullFlags() {
		if !flag.Hidden {
			flags = append(flags, flag)
		}
	}
	sort.Sort(flags)

	args := command.Args() // These must be given in order, so don't sort them.

	// Print the one-line usage summary. line: "some command [<command>] [<flags>] [<arg>...]"
	fmt.Fprint(w, "Usage: ", command.FullName())
	if len(flags) != 0 {
		fmt.Fprint(w, " [<flags>]")
	}
	if len(subs) != 0 {
		fmt.Fprint(w, " "+brackets("<command>", command.Action != nil))
	} else if len(args) != 0 {
		// Since positional args are ordered, everything prior to a required arg
		// must also be required. Find the last one.
		lastRequired := len(args)
		for i := len(args) - 1; i >= 0; i-- {
			if args[i].Required {
				lastRequired = i
				break
			}
		}

		for i, arg := range args {
			name := "<" + arg.Name + ">"
			if IsAggregate(arg.Value) {
				name += "..."
			}
			fmt.Fprint(w, " "+brackets(name, i > lastRequired))
		}
	}
	fmt.Fprintln(w)

	maxWidth := u.MaxLineWidth
	if maxWidth == 0 {
		if width, err := ttyWidth(); err != nil {
			maxWidth = 80
		} else {
			maxWidth = width
		}
	}

	// Show the command's help.
	if command.Help != "" {
		fmt.Fprintln(w)
		wordwrap.NewScanner(strings.NewReader(command.Help), maxWidth).WriteTo(w)
		fmt.Fprintln(w)
	}

	// Print commands/args first since we want them near the summary. We omit
	// args if commands are present since we know they'll be ignored on parse.
	if len(subs) != 0 {
		fmt.Fprintln(w, "\nCommands:")
		rows := make([][2]string, 0, len(subs))
		for _, cmd := range subs {
			// TODO: Should help be trimmed to the first line?
			rows = append(rows, [2]string{u.Indent + cmd.Name, cmd.Help})
		}
		u.formatTwoColumns(w, rows, maxWidth)
	} else if len(args) != 0 {
		fmt.Fprintln(w, "\nArguments:")
		rows := make([][2]string, 0, len(args))
		for _, arg := range args {
			// TODO: Should help be trimmed to the first line?
			rows = append(rows, [2]string{u.Indent + arg.Name, arg.Help})
		}
		u.formatTwoColumns(w, rows, maxWidth)
	}

	if len(flags) != 0 {
		fmt.Fprintln(w, "\nOptions:")

		// Pre-scan for short and long strings so we know whether to add a separator.
		haveShorts := false
		for _, flag := range flags {
			if flag.Short != rune(0) {
				haveShorts = true
				break
			}
		}

		// Print each flag with short and long flags vertically aligned.
		rows := make([][2]string, 0, len(flags))
		for _, flag := range flags {
			var flagStr string
			if haveShorts {
				if flag.Short == rune(0) {
					flagStr = "  "
				} else {
					flagStr = "-" + string(flag.Short)
				}
			}
			if haveShorts && flag.Name != "" {
				if flag.Short == rune(0) {
					flagStr += "  "
				} else {
					flagStr += ", "
				}
			}
			if flag.Name != "" {
				flagStr += "--" + flag.Name
			}

			// Now add the argument's placeholder if it has one.
			if flag.Value != nil && !IsBoolean(flag.Value) {
				flagStr += " "
				if flag.Placeholder == "" {
					flagStr += "VALUE"
				} else {
					flagStr += flag.Placeholder
				}
				if IsAggregate(flag.Value) {
					flagStr += "..."
				}
			}

			// TODO: Should help be trimmed to the first line?
			rows = append(rows, [2]string{u.Indent + flagStr, flag.Help})

			// Add the negative boolean version
		}
		u.formatTwoColumns(w, rows, maxWidth)
	}

	return nil
}

func brackets(s string, optional bool) string {
	if optional {
		return "[" + s + "]"
	}
	return s
}

func ttyWidth() (int, error) {
	type windowSize struct {
		Rows    uint16
		Columns uint16
		Width   uint16
		Height  uint16
	}

	ws := &windowSize{}
	retCode, _, _ := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))
	if int(retCode) == -1 {
		return 0, errors.New("no TTY enabled")
	}
	return int(ws.Columns), nil
}

func (u *UsageWriter) formatTwoColumns(w io.Writer, rows [][2]string, width int) {
	// Dynamically size the first column.
	var leftWidth int
	for _, row := range rows {
		if size := len(row[0]); size > leftWidth {
			leftWidth = size
		}
	}
	if u.MaxFirstColumn != 0 && leftWidth > u.MaxFirstColumn {
		leftWidth = u.MaxFirstColumn
	}

	// Print each row, wrapped to its column size
	for _, row := range rows {
		leftScan := wordwrap.NewScanner(strings.NewReader(row[0]), leftWidth)
		rightScan := wordwrap.NewScanner(strings.NewReader(row[1]), width-leftWidth-len(u.Divider))
		for {
			left, leftErr := leftScan.ReadLine()
			right, rightErr := rightScan.ReadLine()
			if leftErr == io.EOF && rightErr == io.EOF {
				break
			}
			fmt.Fprintf(w, "%-*s%s%s\n", leftWidth, left, u.Divider, right)
		}
	}
}

type commandSlice []*Command

func (s commandSlice) Len() int           { return len(s) }
func (s commandSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s commandSlice) Less(i, j int) bool { return s[i].Name < s[j].Name }

type flagSlice []*Flag

func (s flagSlice) Len() int      { return len(s) }
func (s flagSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s flagSlice) Less(i, j int) bool {
	nameOrRune := func(f *Flag) string {
		if f.Name == "" {
			return string(f.Short)
		}
		return f.Name
	}
	return nameOrRune(s[i]) < nameOrRune(s[j])
}
