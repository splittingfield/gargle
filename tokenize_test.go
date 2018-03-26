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
			[]string{"--one", "--two", "--three"},
			[]token{
				{tokenLong, "one"},
				{tokenLong, "two"},
				{tokenLong, "three"},
			},
		},
		"SingleShort": {
			[]string{"-1"},
			[]token{{tokenShort, "1"}},
		},
		"LongWithSeparateValue": {
			[]string{"--one", "arg"},
			[]token{
				{tokenLong, "one"},
				{tokenArg, "arg"},
			},
		},
		"LongWithJoinedValue": {
			[]string{"--one=arg"},
			[]token{
				{tokenLong, "one"},
				{tokenArg, "arg"},
			},
		},
		"ShortWithValue": {
			[]string{"-1", "arg"},
			[]token{
				{tokenShort, "1"},
				{tokenArg, "arg"},
			},
		},
		"SingleDash": {
			[]string{"-a", "-", "-b"},
			[]token{
				{tokenShort, "a"},
				{tokenArg, "-"},
				{tokenShort, "b"},
			},
		},
		"DoubleDash": {
			[]string{"--foo", "--", "--bar"},
			[]token{
				{tokenLong, "foo"},
				{tokenArg, "--bar"},
			},
		},
		"MixedComplex": { // sanity check for non-trivial args.
			[]string{"--foo=bar", "-ab", "arg", "-", "-c", "--", "-d", "--baz"},
			[]token{
				{tokenLong, "foo"},
				{tokenArg, "bar"},
				{tokenShort, "a"},
				{tokenShort, "b"},
				{tokenArg, "arg"},
				{tokenArg, "-"},
				{tokenShort, "c"},
				{tokenArg, "-d"},
				{tokenArg, "--baz"},
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			tokenizer := newTokenizer(c.args)
			var actual []token
			for tok := tokenizer.Next(); tok.Type != tokenEOF; tok = tokenizer.Next() {
				actual = append(actual, tok)
			}

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestTokenizeNextArg(t *testing.T) {
	cases := map[string]struct {
		args      []string
		skip      int
		expected  token
		expectErr bool
	}{
		"NoArgs": {expectErr: true},
		"SingleLong": {
			args:      []string{"--foo"},
			expectErr: true,
		},
		"SingleShort": {
			args:      []string{"-1"},
			expectErr: true,
		},
		"LongWithSeparateValue": {
			args:     []string{"--one", "arg"},
			skip:     1,
			expected: token{tokenArg, "arg"},
		},
		"LongWithJoinedValue": {
			args:     []string{"--one=arg"},
			skip:     1,
			expected: token{tokenArg, "arg"},
		},
		"ShortWithJoinedValue": {
			args:     []string{"-1arg"},
			skip:     1,
			expected: token{tokenArg, "arg"},
		},
		"SingleDash": {
			args:     []string{"-ab", "-"},
			skip:     2,
			expected: token{tokenArg, "-"},
		},
		"DoubleDash": {
			args:     []string{"--foo", "--", "--bar"},
			skip:     1,
			expected: token{tokenArg, "--bar"},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			tokenizer := newTokenizer(c.args)
			for i := 0; i < c.skip; i++ {
				tokenizer.Next()
			}

			actual, err := tokenizer.NextArg()
			assert.Equal(t, c.expected, actual)
			if c.expectErr {
				assert.EqualError(t, err, "expected argument")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
