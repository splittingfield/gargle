package gargle

import (
	"strings"
	"unicode/utf8"
)

// tokenType enumerates the possible kinds of tokens.
type tokenType int

const (
	tokenEOF tokenType = iota
	tokenVerbatim
	tokenLong
	tokenShort
	tokenArg
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
	args         []string
	prevWasShort bool
	remainder    string
}

func newTokenizer(args []string) *tokenizer {
	return &tokenizer{args: args}
}

// Next returns the next token. If no tokens are available, it returns EOF.
// When forceVerbatim is set the token, if any, will be returned as an argument.
func (t *tokenizer) Next(verbatim bool) token {
	if t.remainder != "" {
		arg := t.remainder
		t.remainder = ""

		tok := token{tokenArg, arg}
		if !verbatim && t.prevWasShort {
			tok = t.decodeShort(arg)
		}

		t.prevWasShort = false
		return tok
	}

	if len(t.args) == 0 {
		return token{Type: tokenEOF}
	}
	arg := t.args[0]
	t.args = t.args[1:]
	if verbatim {
		return token{tokenArg, arg}
	}

	if strings.HasPrefix(arg, "--") {
		// Special case: -- means return everything following verbatim.
		if arg == "--" {
			return token{tokenVerbatim, "--"}
		}

		parts := strings.SplitN(arg[2:], "=", 2)
		if len(parts) > 1 {
			t.remainder = parts[1]
		}
		return token{tokenLong, parts[0]}
	}

	if strings.HasPrefix(arg, "-") {
		// Special case: - is often a placeholder for STDIN, and is a valid arg.
		if arg == "-" {
			return token{tokenArg, arg}
		}
		return t.decodeShort(arg[1:])
	}

	// Anything else is an argument.
	return token{tokenArg, arg}
}

func (t *tokenizer) decodeShort(arg string) token {
	flag, size := utf8.DecodeRuneInString(arg)
	t.remainder = arg[size:]
	t.prevWasShort = true
	return token{tokenShort, string(flag)}
}
