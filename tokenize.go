package gargle

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// tokenType enumerates the possible kinds of tokens.
type tokenType int

const (
	tokenEOF tokenType = iota
	tokenLong
	tokenShort
	tokenArg
)

// token represents a single parsed token.
type token struct {
	Type  tokenType
	Value string
}

// tokenizer is a scanning tokenizer for command-line arguments.
type tokenizer struct {
	args           []string
	verbatim       bool
	shortRemainder string
	stack          tokenStack
}

func newTokenizer(args []string) *tokenizer {
	return &tokenizer{args: args}
}

// Peek returns the next token without consuming it.
func (t *tokenizer) Peek() token {
	result := t.Next()
	t.stack.Push(result)
	return result
}

// Next returns the next token. If no tokens are available, it returns EOF.
func (t *tokenizer) Next() token {
	if t.stack.Len() > 0 {
		return t.stack.Pop()
	}

	var arg string
	if t.shortRemainder == "" {
		if len(t.args) == 0 {
			return token{Type: tokenEOF}
		}
		arg = t.args[0]
		t.args = t.args[1:]
	} else {
		arg = "-" + t.shortRemainder
		t.shortRemainder = ""
	}

	if t.verbatim {
		return token{tokenArg, arg}
	}

	if strings.HasPrefix(arg, "--") {
		// Special case: -- means return everything following verbatim.
		if arg == "--" {
			t.verbatim = true
			return t.Next()
		}

		parts := strings.SplitN(arg[2:], "=", 2)
		if len(parts) > 1 {
			t.stack.Push(token{tokenArg, parts[1]})
		}
		return token{tokenLong, parts[0]}
	}

	if strings.HasPrefix(arg, "-") {
		// Special case: - is often a placeholder for STDIN, and is a valid arg.
		if arg == "-" {
			return token{tokenArg, arg}
		}

		arg = arg[1:]
		flag, size := utf8.DecodeRuneInString(arg)
		t.shortRemainder = arg[size:]
		return token{tokenShort, string(flag)}
	}

	// Anything else is an argument.
	return token{tokenArg, arg}
}

// NextArg is similar to Next except that it requires the next token to be an argument.
func (t *tokenizer) NextArg() (token, error) {
	if t.shortRemainder != "" {
		result := t.shortRemainder
		t.shortRemainder = ""
		return token{tokenArg, result}, nil
	}

	next := t.Peek()
	if next.Type != tokenArg {
		return token{}, errors.New("expected argument")
	}
	return t.Next(), nil
}

type tokenStack []token

func (s *tokenStack) Len() int {
	return len(*s)
}

func (s *tokenStack) Push(t token) {
	*s = append(*s, t)
}

func (s *tokenStack) Pop() token {
	t := (*s)[0]
	*s = (*s)[:len(*s)-1]
	return t
}
