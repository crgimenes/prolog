package engine

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexer_Token(t *testing.T) {
	tests := []struct {
		input           string
		charConversions map[rune]rune
		token           Token
		err             error
	}{
		{input: ``, err: io.EOF},
		{input: `🙈`, err: errMonkey}, // In this test, we use a see-no-evil monkey emoji to denote a non-EOF error.

		{input: ".", token: Token{kind: tokenEnd, val: "."}},
		{input: ";", token: Token{kind: tokenSemicolon, val: ";"}},
		{input: "!", token: Token{kind: tokenCut, val: "!"}},
		{input: "(", token: Token{kind: tokenOpenCT, val: "("}},
		{input: " (", token: Token{kind: tokenOpen, val: "("}},
		{input: ")", token: Token{kind: tokenClose, val: ")"}},
		{input: "[", token: Token{kind: tokenOpenList, val: "["}},
		{input: "]", token: Token{kind: tokenCloseList, val: "]"}},
		{input: "{", token: Token{kind: tokenOpenCurly, val: "{"}},
		{input: "}", token: Token{kind: tokenCloseCurly, val: "}"}},
		{input: "|", token: Token{kind: tokenBar, val: "|"}},
		{input: ",", token: Token{kind: tokenComma, val: ","}},

		{input: "% comment\nfoo", token: Token{kind: tokenLetterDigit, val: "foo"}},
		{input: "% comment", err: io.EOF},
		{input: "/* comment \n * also comment \n */foo", token: Token{kind: tokenLetterDigit, val: "foo"}},
		{input: "/* comment ", err: io.EOF},
		{input: `/`, token: Token{kind: tokenGraphic, val: `/`}},
		{input: `/ *`, token: Token{kind: tokenGraphic, val: `/`}},
		{input: "/* comment *", err: io.EOF},
		{input: `/🙈`, err: errMonkey},
		{input: `/* **/foo`, token: Token{kind: tokenLetterDigit, val: "foo"}}, // https://github.com/crgimenes/prolog/issues/326

		{input: `改善`, token: Token{kind: tokenLetterDigit, val: `改善`}},
		{input: `プロログ`, token: Token{kind: tokenLetterDigit, val: `プロログ`}},
		{input: `ぷろろぐ`, token: Token{kind: tokenLetterDigit, val: `ぷろろぐ`}},
		{input: `프롤로그`, token: Token{kind: tokenLetterDigit, val: `프롤로그`}},
		{input: `برولوغ`, token: Token{kind: tokenLetterDigit, val: `برولوغ`}},
		{input: `פרולוג`, token: Token{kind: tokenLetterDigit, val: `פרולוג`}},
		{input: `ゴー`, token: Token{kind: tokenLetterDigit, val: `ゴー`}},
		{input: `prolog.`, token: Token{kind: tokenLetterDigit, val: `prolog`}},
		{input: `prolog🙈`, err: errMonkey},

		{input: `..`, token: Token{kind: tokenGraphic, val: `..`}},
		{input: `#`, token: Token{kind: tokenGraphic, val: `#`}},
		{input: `\`, token: Token{kind: tokenGraphic, val: `\`}},
		{input: `∀`, token: Token{kind: tokenGraphic, val: `∀`}},
		{input: `⨀`, token: Token{kind: tokenGraphic, val: `⨀`}},
		{input: `+🙈`, err: errMonkey},

		{input: `'abc'`, token: Token{kind: tokenQuoted, val: "'abc'"}},
		{input: `'abc'.`, token: Token{kind: tokenQuoted, val: "'abc'"}},
		{input: `'don''t panic'`, token: Token{kind: tokenQuoted, val: "'don''t panic'"}},
		{input: `'this is \
a quoted ident'`, token: Token{kind: tokenQuoted, val: "'this is \\\na quoted ident'"}},
		{input: `'\a'`, token: Token{kind: tokenQuoted, val: "'\\a'"}},
		{input: `'\b'`, token: Token{kind: tokenQuoted, val: "'\\b'"}},
		{input: `'\f'`, token: Token{kind: tokenQuoted, val: "'\\f'"}},
		{input: `'\n'`, token: Token{kind: tokenQuoted, val: "'\\n'"}},
		{input: `'\r'`, token: Token{kind: tokenQuoted, val: "'\\r'"}},
		{input: `'\t'`, token: Token{kind: tokenQuoted, val: "'\\t'"}},
		{input: `'\v'`, token: Token{kind: tokenQuoted, val: "'\\v'"}},
		{input: `'\xa3\'`, token: Token{kind: tokenQuoted, val: "'\\xa3\\'"}},
		{input: `'\xa333333333\'`, token: Token{kind: tokenInvalid, val: `'\xa333333333\'`}},
		{input: `'\xa333333333\'.`, token: Token{kind: tokenInvalid, val: `'\xa333333333\'`}},
		{input: `'\43333333\'`, token: Token{kind: tokenInvalid, val: `'\43333333\'`}},
		{input: `'\\'`, token: Token{kind: tokenQuoted, val: `'\\'`}},
		{input: `'\''`, token: Token{kind: tokenQuoted, val: `'\''`}},
		{input: `'\"'`, token: Token{kind: tokenQuoted, val: `'\"'`}},
		{input: "'`'", token: Token{kind: tokenQuoted, val: "'`'"}},
		{input: "'\\`'", token: Token{kind: tokenQuoted, val: "'\\`'"}},
		{input: `'`, err: io.EOF},
		{input: `'\`, err: io.EOF},
		{input: `'\x`, err: io.EOF},
		{input: `'\xG`, token: Token{kind: tokenInvalid, val: `'\xG`}},
		{input: `'\0`, err: io.EOF},
		{input: `'\08`, token: Token{kind: tokenInvalid, val: `'\08`}},
		{input: "'\x01'", token: Token{kind: tokenInvalid, val: "'\x01"}},
		{input: `'abc'🙈`, err: errMonkey},
		{input: `'this is \🙈'`, err: errMonkey},

		{input: `X`, token: Token{kind: tokenVariable, val: `X`}},
		{input: `X.`, token: Token{kind: tokenVariable, val: `X`}},
		{input: `_123`, token: Token{kind: tokenVariable, val: `_123`}},
		{input: `X🙈`, err: errMonkey},

		{input: `012345`, token: Token{kind: tokenInteger, val: "012345"}},
		{input: `012345,`, token: Token{kind: tokenInteger, val: "012345"}},
		{input: `012345..`, token: Token{kind: tokenInteger, val: "012345"}},
		{input: `0b10110101`, token: Token{kind: tokenInteger, val: "0b10110101"}},
		{input: `0b10110101.`, token: Token{kind: tokenInteger, val: "0b10110101"}},
		{input: `0b`, token: Token{kind: tokenInteger, val: "0"}},
		{input: `0b.`, token: Token{kind: tokenInteger, val: "0"}},
		{input: `0o567`, token: Token{kind: tokenInteger, val: "0o567"}},
		{input: `0o567.`, token: Token{kind: tokenInteger, val: "0o567"}},
		{input: `0o`, token: Token{kind: tokenInteger, val: "0"}},
		{input: `0o.`, token: Token{kind: tokenInteger, val: "0"}},
		{input: `0x89ABC`, token: Token{kind: tokenInteger, val: "0x89ABC"}},
		{input: `0x89ABC.`, token: Token{kind: tokenInteger, val: "0x89ABC"}},
		{input: `0x`, token: Token{kind: tokenInteger, val: "0"}},
		{input: `0x.`, token: Token{kind: tokenInteger, val: "0"}},
		{input: `0'a`, token: Token{kind: tokenInteger, val: "0'a"}},
		{input: `0'''`, token: Token{kind: tokenInteger, val: "0'''"}},
		{input: `0''`, token: Token{kind: tokenInteger, val: "0"}},
		{input: `0''.`, token: Token{kind: tokenInteger, val: "0"}},
		{input: `0'\n`, token: Token{kind: tokenInteger, val: `0'\n`}},
		{input: `0'\
`, token: Token{kind: tokenInteger, val: `0`}},
		{input: `0'\`, err: io.EOF},
		{input: `0'\q`, token: Token{kind: tokenInvalid, val: `0'\q`}},
		{input: `0'\😀`, token: Token{kind: tokenInvalid, val: `0'\😀`}},
		{input: `0'`, err: io.EOF},
		{input: "0'\x01", token: Token{kind: tokenInvalid, val: "0'\x01"}},
		{input: `0`, token: Token{kind: tokenInteger, val: "0"}},
		{input: `0.`, token: Token{kind: tokenInteger, val: "0"}},
		{input: `0🙈`, err: errMonkey},
		{input: `0'🙈`, err: errMonkey},
		{input: `0''🙈`, err: errMonkey},
		{input: `0'\🙈`, err: errMonkey},
		{input: `0b🙈`, err: errMonkey},
		{input: `0o🙈`, err: errMonkey},
		{input: `0x🙈`, err: errMonkey},
		{input: `012345🙈`, err: errMonkey},
		{input: `012345.🙈`, err: errMonkey},
		{input: `0b10110101🙈`, err: errMonkey},
		{input: `0o567🙈`, err: errMonkey},
		{input: `0x89ABC🙈`, err: errMonkey},

		{input: `2.34`, token: Token{kind: tokenFloatNumber, val: "2.34"}},
		{input: `2.34.`, token: Token{kind: tokenFloatNumber, val: "2.34"}},
		{input: `2.34E5`, token: Token{kind: tokenFloatNumber, val: "2.34E5"}},
		{input: `2.34E5.`, token: Token{kind: tokenFloatNumber, val: "2.34E5"}},
		{input: `2.34E`, token: Token{kind: tokenFloatNumber, val: "2.34"}},
		{input: `2.34E.`, token: Token{kind: tokenFloatNumber, val: "2.34"}},
		{input: `2.34E+5`, token: Token{kind: tokenFloatNumber, val: "2.34E+5"}},
		{input: `2.34E+5.`, token: Token{kind: tokenFloatNumber, val: "2.34E+5"}},
		{input: `2.34E+`, token: Token{kind: tokenFloatNumber, val: "2.34"}},
		{input: `2.34E+.`, token: Token{kind: tokenFloatNumber, val: "2.34"}},
		{input: `2.34E-10`, token: Token{kind: tokenFloatNumber, val: "2.34E-10"}},
		{input: `2.34E-10.`, token: Token{kind: tokenFloatNumber, val: "2.34E-10"}},
		{input: `2.34E-`, token: Token{kind: tokenFloatNumber, val: "2.34"}},
		{input: `2.34E-.`, token: Token{kind: tokenFloatNumber, val: "2.34"}},
		{input: `0.333`, token: Token{kind: tokenFloatNumber, val: "0.333"}},
		{input: `2.34🙈`, err: errMonkey},
		{input: `2.34E🙈`, err: errMonkey},
		{input: `2.34E+🙈`, err: errMonkey},
		{input: `2.34E-🙈`, err: errMonkey},
		{input: `2.34E5🙈`, err: errMonkey},
		{input: `2.34E+5🙈`, err: errMonkey},
		{input: `2.34E-10🙈`, err: errMonkey},

		{input: `"abc"`, token: Token{kind: tokenDoubleQuotedList, val: `"abc"`}},
		{input: `"abc".`, token: Token{kind: tokenDoubleQuotedList, val: `"abc"`}},
		{input: `"don""t panic"`, token: Token{kind: tokenDoubleQuotedList, val: `"don""t panic"`}},
		{input: `"this is \
a quoted ident"`, token: Token{kind: tokenDoubleQuotedList, val: `"this is \
a quoted ident"`}},
		{input: `"\a"`, token: Token{kind: tokenDoubleQuotedList, val: `"\a"`}},
		{input: `"\b"`, token: Token{kind: tokenDoubleQuotedList, val: `"\b"`}},
		{input: `"\f"`, token: Token{kind: tokenDoubleQuotedList, val: `"\f"`}},
		{input: `"\n"`, token: Token{kind: tokenDoubleQuotedList, val: `"\n"`}},
		{input: `"\r"`, token: Token{kind: tokenDoubleQuotedList, val: `"\r"`}},
		{input: `"\t"`, token: Token{kind: tokenDoubleQuotedList, val: `"\t"`}},
		{input: `"\v"`, token: Token{kind: tokenDoubleQuotedList, val: `"\v"`}},
		{input: `"\xa3\"`, token: Token{kind: tokenDoubleQuotedList, val: `"\xa3\"`}},
		{input: `"\xa3`, err: io.EOF},
		{input: `"\xa3g`, token: Token{kind: tokenInvalid, val: `"\xa3g`}},
		{input: `"\43\"`, token: Token{kind: tokenDoubleQuotedList, val: `"\43\"`}},
		{input: `"\43`, err: io.EOF},
		{input: `"\438`, token: Token{kind: tokenInvalid, val: `"\438`}},
		{input: `"\\"`, token: Token{kind: tokenDoubleQuotedList, val: `"\\"`}},
		{input: `"\'"`, token: Token{kind: tokenDoubleQuotedList, val: `"\'"`}},
		{input: `"\""`, token: Token{kind: tokenDoubleQuotedList, val: `"\""`}},
		{input: "\"\\`\"", token: Token{kind: tokenDoubleQuotedList, val: "\"\\`\""}},
		{input: `"`, err: io.EOF},
		{input: `"\`, err: io.EOF},
		{input: `"abc"🙈`, err: errMonkey},

		{input: "\x01", token: Token{kind: tokenInvalid, val: "\x01"}},

		{input: `abc`, charConversions: map[rune]rune{'b': 'a'}, token: Token{kind: tokenLetterDigit, val: "aac"}},
		{input: `'abc'`, charConversions: map[rune]rune{'b': 'a'}, token: Token{kind: tokenQuoted, val: "'abc'"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := Lexer{input: newRuneRingBuffer(noMonkeyReader{strings.NewReader(tt.input)}), charConversions: tt.charConversions}

			token, err := l.Token()
			assert.Equal(t, tt.token, token)
			assert.Equal(t, tt.err, err)
		})
	}
}

var errMonkey = errors.New("monkey")

type noMonkeyReader struct {
	io.RuneReader
}

func (n noMonkeyReader) ReadRune() (rune, int, error) {
	r, size, err := n.RuneReader.ReadRune()
	if r == '🙈' {
		return 0, 0, errMonkey
	}
	return r, size, err
}

func TestTokenKind_GoString(t *testing.T) {
	assert.Equal(t, "invalid", tokenInvalid.GoString())
}
