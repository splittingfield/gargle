package gargle

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"syscall"
	"unsafe"

	"github.com/ckarenz/wordwrap"
)

// UsageWriter is a configurable usage formatter which can be used as a safe
// default for most applications.
type UsageWriter struct {
	// Indent is an indentation prefix for subsections.
	Indent string

	// Divider is a string to print between columns.
	Divider string

	// MaxFirstColumn is the maximum width of the left column in split sections.
	MaxFirstColumn int

	// MaxLineWidth overrides the maximum width of each line of text.
	MaxLineWidth int
}

// Format writes a given command's usage using the writer's configuration.
func (u *UsageWriter) Format(w io.Writer, command *Command) error {
	var subs commandSlice
	for _, cmd := range command.Commands() {
		if !cmd.Hidden {
			subs = append(subs, cmd)
		}
	}
	sort.Sort(subs)

	var flags flagSlice
	for _, flag := range command.Flags() {
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
			fmt.Fprint(w, " "+brackets(name, i <= lastRequired))
		}
	}
	fmt.Fprintln(w)

	maxWidth, err := ttyWidth()
	if err != nil {
		maxWidth = 80
	}
	if maxWidth > u.MaxLineWidth {
		maxWidth = u.MaxLineWidth
	}

	// Show the command's help.
	if command.Help != "" {
		fmt.Fprintln(w)
		wordwrap.NewScanner(strings.NewReader(command.Help), maxWidth).WriteTo(w)
		fmt.Fprintln(w)
	}

	// Print args first since the list is likely to be short and we want them near the summary.
	if len(args) != 0 {
		fmt.Fprintln(w, "\nArguments:")
		rows := make([][2]string, 0, len(args))
		for _, arg := range args {
			// TODO: Should help be trimmed to the first line?
			rows = append(rows, [2]string{"  " + arg.Name, arg.Help})
		}
		u.formatTwoColumns(w, rows, maxWidth)
	}

	if len(subs) != 0 {
		fmt.Fprintln(w, "\nCommands:")
		rows := make([][2]string, 0, len(subs))
		for _, cmd := range subs {
			// TODO: Should help be trimmed to the first line?
			rows = append(rows, [2]string{u.Indent + cmd.Name, cmd.Help})
		}
		u.formatTwoColumns(w, rows, maxWidth)
	}

	// TODO: Check if we have any short flags before adding their stuff.
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
			// TODO: Should we also print the negated version of boolean flags?
			if flag.Value != nil && !IsBoolean(flag.Value) {
				flagStr += " "
				if flag.Placeholder == "" {
					flagStr += "VALUE"
				} else {
					flagStr += flag.Placeholder
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
