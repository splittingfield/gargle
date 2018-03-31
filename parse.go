package gargle

import (
	"fmt"
	"strconv"
	"strings"
)

// element is a parsed entity with an optional associated value.
type element struct {
	entity interface{}
	token  token
	value  string
}

// parser is a multi-phase command-line argument parser. Parsers are stateful
// and should not be reused.
type parser struct {
	tokenizer *tokenizer
	context   *Command
	parsed    []element

	// Representations including all parseable entities.
	commands   map[string]*Command
	flags      map[string]*Flag
	shortFlags map[string]*Flag
	args       []*Arg
}

// newParser creates a new parser with the a given command as its initial context.
func newParser(rootCommand *Command, args []string) *parser {
	p := &parser{
		tokenizer:  newTokenizer(args),
		flags:      map[string]*Flag{},
		shortFlags: map[string]*Flag{},
	}
	p.setContext(rootCommand)
	return p
}

// Context returns the most recent command seen by the parser. This is never nil.
func (p *parser) Context() *Command {
	return p.context
}

// setContext sets the parser's context and includes its subcommands, flags, and
// arguments as parseable entities.
func (p *parser) setContext(context *Command) {
	p.commands = map[string]*Command{}
	for _, command := range context.Commands() {
		p.commands[command.Name] = command
	}

	for _, flag := range context.Flags() {
		p.flags[flag.Name()] = flag
		if short := flag.Short(); short != 0 {
			s := []rune{short}
			p.shortFlags[string(s)] = flag
		}
	}

	p.args = append(p.args, context.Args()...)
	p.context = context
}

func (p *parser) Parse() error {
	verbatim := false
	for {
		switch token := p.tokenizer.Next(verbatim); token.Type {
		case tokenEOF:
			return nil

		case tokenVerbatim:
			verbatim = true

		case tokenLong:
			negate := false
			flag, ok := p.flags[token.Value]
			if !ok && strings.HasPrefix(token.Value, "no-") {
				// This may be a boolean flag; try its negated form.
				flag, ok = p.flags[token.Value[3:]]
				negate = true
			}
			if !ok || negate && !flag.Value().IsBoolean() {
				return fmt.Errorf("unknown flag: %s", token.Value)
			}

			if flag.Value().IsBoolean() {
				value := strconv.FormatBool(!negate)
				p.parsed = append(p.parsed, element{flag, token, value})
				break
			}

			argToken := p.tokenizer.Next(true)
			if argToken.Type == tokenEOF {
				return fmt.Errorf("%s must have a value", token)
			}
			p.parsed = append(p.parsed, element{flag, token, argToken.Value})

		case tokenShort:
			flag, ok := p.shortFlags[token.Value]
			if !ok {
				return fmt.Errorf("unknown flag: %s", token.Value)
			}
			if flag.Value().IsBoolean() {
				p.parsed = append(p.parsed, element{flag, token, "true"})
				break
			}

			argToken := p.tokenizer.Next(true)
			if argToken.Type == tokenEOF {
				return fmt.Errorf("%s must have a value", token)
			}
			p.parsed = append(p.parsed, element{flag, token, argToken.Value})

		case tokenArg:
			// Commands take precedence over positional arguments. Any remaining
			// unparsed args are preserved for the next context.
			if len(p.commands) != 0 {
				command, ok := p.commands[token.Value]
				if !ok {
					fullName := p.context.FullName() + " " + token.Value
					return fmt.Errorf("%q is not a valid command", fullName)
				}
				p.setContext(command)
				break
			}

			if len(p.args) == 0 {
				return fmt.Errorf("unexpected argument: %q", token.Value)
			}

			arg := p.args[0]
			if !arg.Value().IsAggregate() {
				p.args = p.args[1:]
			}
			p.parsed = append(p.parsed, element{arg, token, token.Value})
		}
	}
}

func (p *parser) setValues() error {
	type settable interface {
		Value() *Value
	}

	// Set all values we saw during parsing.
	seen := map[interface{}]bool{}
	for _, element := range p.parsed {
		val, ok := element.entity.(settable)
		if !ok {
			continue
		}

		seen[element.entity] = true
		if err := val.Value().set(element.value); err != nil {
			return fmt.Errorf("invalid argument for %s: %s", element.token, err.Error())
		}
	}

	var stack []*Command
	for command := p.context; command != nil; command = command.Parent() {
		stack = append(stack, command)
	}

	// Validate unset arguments/flags and apply defaults.
	for i := len(stack) - 1; i >= 0; i-- {
		command := stack[i]
		for _, flag := range command.Flags() {
			if seen[flag] {
				continue
			}
			if flag.IsRequired() {
				return fmt.Errorf("missing required flag --%s", flag.Name())
			}
			if err := flag.Value().applyDefault(); err != nil {
				return err
			}
		}

		for _, arg := range command.Args() {
			if seen[arg] {
				continue
			}
			if arg.IsRequired() {
				return fmt.Errorf("missing required argument %s", arg.Name())
			}
			if err := arg.Value().applyDefault(); err != nil {
				return err
			}
		}
	}
	return nil
}
