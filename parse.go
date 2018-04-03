package gargle

import (
	"fmt"
)

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

	p.args = context.Args()
	p.context = context
}

// entity is a parsed object with an optional associated value.
type entity struct {
	Option interface{}
	Name   string
	Value  string
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
			flag, ok := p.flags[token.Value]
			if !ok {
				return parsed, fmt.Errorf("unknown flag: %s", token.Value)
			}

			value, err := p.parseFlagValue(flag, token)
			if err != nil {
				return parsed, err
			}
			parsed = append(parsed, entity{flag, token.String(), value})

		case tokenShort:
			flag, ok := p.shortFlags[token.Value]
			if !ok {
				return parsed, fmt.Errorf("unknown flag: %s", token.Value)
			}

			value, err := p.parseFlagValue(flag, token)
			if err != nil {
				return parsed, err
			}
			parsed = append(parsed, entity{flag, token.String(), value})

		case tokenValue:
			// Commands take precedence over positional arguments. Any remaining
			// unparsed args are discarded for the next context.
			if len(p.commands) != 0 {
				command, ok := p.commands[token.Value]
				if !ok {
					fullName := p.context.FullName() + " " + token.Value
					return parsed, fmt.Errorf("%q is not a valid command", fullName)
				}
				parsed = append(parsed, entity{command, command.Name, token.Value})
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
			parsed = append(parsed, entity{arg, arg.Name, token.Value})
		}
	}
}

func (p *parser) parseFlagValue(flag *Flag, flagToken token) (string, error) {
	if tok := p.tokenizer.Peek(); tok != nil && tok.Type == tokenAssigned {
		if flag.Value == nil {
			return "", fmt.Errorf("invalid value %q for %s: flag does not accept a value", tok.Value, flagToken)
		}
		return tok.Value, nil
	}

	// Valueless flags don't consume an argument.
	if flag.Value == nil {
		return "", nil
	}

	// Boolean values are optional.
	if IsBoolean(flag.Value) {
		return "true", nil
	}

	tok := p.tokenizer.Next(true)
	if tok.Type == tokenEOF {
		return "", fmt.Errorf("%s requires a value", flagToken)
	}
	return tok.Value, nil
}
