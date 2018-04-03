package gargle

import (
	"strings"
	"unicode/utf8"
)

// tokenType enumerates the possible kinds of tokens.
type tokenType int

const (
	tokenEOF tokenType = iota

	tokenVerbatim // Verbatim symbol "--"
	tokenLong     // Long flag with "--" prefix
	tokenShort    // Short flag with "-" prefix
	tokenValue    // Naked value
	tokenAssigned // Value explicitly assigned to the previous flag.

	tokenRemainder // Remainder from previous short flag; never returned from Next().
)

// token represents a single parsed token.
type token struct {
	Type  tokenType
	Value string
}

func (t token) String() string {
	switch t.Type {
	case tokenLong:
		return "--" + t.Value
	case tokenShort:
		return "-" + t.Value
	default:
		return t.Value
	}
}

// tokenizer is a scanning tokenizer for command-line arguments.
type tokenizer struct {
	args []string
	next *token // Next token on deck
}

func newTokenizer(args []string) *tokenizer {
	return &tokenizer{args: args}
}

// Peek returns the next token without consuming it, if there is one.
func (t *tokenizer) Peek() *token {
	// Never return a remainder, because it's not a known tokon yet.
	if t.next != nil && t.next.Type == tokenRemainder {
		return nil
	}
	return t.next
}

// Next returns the next token. If no tokens are available, it returns EOF.
// When forceVerbatim is set the token, if any, will be returned as an argument.
func (t *tokenizer) Next(verbatim bool) token {
	if next := t.next; next != nil {
		t.next = nil
		if next.Type == tokenRemainder {
			if verbatim {
				return token{tokenAssigned, next.Value}
			}
			return t.decodeShort(next.Value)
		}
		return *next
	}

	if len(t.args) == 0 {
		return token{Type: tokenEOF}
	}
	arg := t.args[0]
	t.args = t.args[1:]
	if verbatim {
		return token{tokenValue, arg}
	}

	if strings.HasPrefix(arg, "--") {
		// Special case: -- means return everything following verbatim.
		if arg == "--" {
			return token{tokenVerbatim, "--"}
		}

		parts := strings.SplitN(arg[2:], "=", 2)
		if len(parts) > 1 {
			t.next = &token{tokenAssigned, parts[1]}
		}
		return token{tokenLong, parts[0]}
	}

	if strings.HasPrefix(arg, "-") {
		// Special case: - is often a placeholder for STDIN, and is a valid arg.
		if arg == "-" {
			return token{tokenValue, arg}
		}
		return t.decodeShort(arg[1:])
	}

	// Anything else is an argument.
	return token{tokenValue, arg}
}

func (t *tokenizer) decodeShort(arg string) token {
	flag, size := utf8.DecodeRuneInString(arg)
	if remainder := arg[size:]; remainder != "" {
		t.next = &token{tokenRemainder, arg[size:]}
	}
	return token{tokenShort, string(flag)}
}
