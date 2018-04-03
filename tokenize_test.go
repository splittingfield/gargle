package gargle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenize(t *testing.T) {
	cases := map[string]struct {
		args     []string
		expected []token
	}{
		"NoArgs": {},
		"SingleLong": {
			[]string{"--foo"},
			[]token{{tokenLong, "foo"}},
		},
		"MultipleLongs": {
			[]string{"--one", "--two", "--", "--three"},
			[]token{
				{tokenLong, "one"},
				{tokenLong, "two"},
				{tokenVerbatim, "--"},
				{tokenLong, "three"},
			},
		},
		"SingleShort": {
			[]string{"-1"},
			[]token{{tokenShort, "1"}},
		},
		"MultipleShort": {
			[]string{"-ab", "-", "-c"},
			[]token{
				{tokenShort, "a"},
				{tokenShort, "b"},
				{tokenValue, "-"},
				{tokenShort, "c"},
			},
		},
		"SeparateValues": {
			[]string{"--one", "arg1", "-2", "arg2"},
			[]token{
				{tokenLong, "one"},
				{tokenValue, "arg1"},
				{tokenShort, "2"},
				{tokenValue, "arg2"},
			},
		},
		"LongWithJoinedValue": {
			[]string{"--one=arg", "--two=--"},
			[]token{
				{tokenLong, "one"},
				{tokenAssigned, "arg"},
				{tokenLong, "two"},
				{tokenAssigned, "--"},
			},
		},
		/*
			"EmptyValue": {
				[]string{"--flag="},
				[]token{
					{tokenLong, "flag"},
					{tokenAssigned, ""},
				},
			},
		*/
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			tokenizer := newTokenizer(c.args)
			var actual []token
			for tok := tokenizer.Next(false); tok.Type != tokenEOF; tok = tokenizer.Next(false) {
				actual = append(actual, tok)
			}

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestTokenizeVerbatim(t *testing.T) {
	cases := map[string]struct {
		args      []string
		skipFirst bool
		expected  []string
	}{
		"NoArgs": {},
		"Longs": {
			args:     []string{"--one", "--two", "--", "--three"},
			expected: []string{"--one", "--two", "--", "--three"},
		},
		"Shorts": {
			args:     []string{"-1", "-", "-two"},
			expected: []string{"-1", "-", "-two"},
		},
		"SplitShort": {
			args:      []string{"-123"},
			skipFirst: true,
			expected:  []string{"23"},
		},
		"SplitLong": {
			args:      []string{"--one=arg"},
			skipFirst: true,
			expected:  []string{"arg"},
		},
		"SplitLongNullValue": {
			args:      []string{"--flag="},
			skipFirst: true,
			expected:  []string{""},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			tokenizer := newTokenizer(c.args)
			if c.skipFirst {
				tokenizer.Next(false)
			}

			var actual []string
			for tok := tokenizer.Next(true); tok.Type != tokenEOF; tok = tokenizer.Next(true) {
				actual = append(actual, tok.Value)
			}

			assert.Equal(t, c.expected, actual)
		})
	}
}
