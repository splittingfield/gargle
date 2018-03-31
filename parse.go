package gargle

import (
	"fmt"
	"strconv"
	"strings"
)

// entity is a parsed object with an optional associated value.
type entity struct {
	object interface{}
	token  token
	value  string
}

// parser is a multi-phase command-line argument parser. Parsers are stateful
// and should not be reused.
type parser struct {
	tokenizer *tokenizer
	context   *Command

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
		if f := flag.Name; f != "" {
			p.flags[f] = flag
		}
		if s := flag.Short; s != 0 {
			p.shortFlags[string(s)] = flag
		}
	}

	p.args = append(p.args, context.Args()...)
	p.context = context
}

func (p *parser) Parse() ([]entity, error) {
	var parsed []entity
	verbatim := false
	for {
		switch token := p.tokenizer.Next(verbatim); token.Type {
		case tokenEOF:
			return parsed, nil

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
			if !ok || negate && !IsBoolean(flag.Value) {
				return parsed, fmt.Errorf("unknown flag: %s", token.Value)
			}

			if IsBoolean(flag.Value) {
				value := strconv.FormatBool(!negate)
				parsed = append(parsed, entity{flag, token, value})
				break
			}

			argToken := p.tokenizer.Next(true)
			if argToken.Type == tokenEOF {
				return parsed, fmt.Errorf("%s must have a value", token)
			}
			parsed = append(parsed, entity{flag, token, argToken.Value})

		case tokenShort:
			flag, ok := p.shortFlags[token.Value]
			if !ok {
				return parsed, fmt.Errorf("unknown flag: %s", token.Value)
			}
			if IsBoolean(flag.Value) {
				parsed = append(parsed, entity{flag, token, "true"})
				break
			}

			argToken := p.tokenizer.Next(true)
			if argToken.Type == tokenEOF {
				return parsed, fmt.Errorf("%s must have a value", token)
			}
			parsed = append(parsed, entity{flag, token, argToken.Value})

		case tokenArg:
			// Commands take precedence over positional arguments. Any remaining
			// unparsed args are preserved for the next context.
			if len(p.commands) != 0 {
				command, ok := p.commands[token.Value]
				if !ok {
					fullName := p.context.FullName() + " " + token.Value
					return parsed, fmt.Errorf("%q is not a valid command", fullName)
				}
				p.setContext(command)
				break
			}

			if len(p.args) == 0 {
				return parsed, fmt.Errorf("unexpected argument: %q", token.Value)
			}

			arg := p.args[0]
			if !IsAggregate(arg.Value) {
				p.args = p.args[1:]
			}
			parsed = append(parsed, entity{arg, token, token.Value})
		}
	}
}
