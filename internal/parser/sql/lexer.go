package parser_sql

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// token const
const (
	TOKEN_EOF              = iota // end-of-file
	TOKEN_NOT                     // !
	TOKEN_LEFT_PAREN              // (
	TOKEN_RIGHT_PAREN             // )
	TOKEN_LEFT_BRACKET            // [
	TOKEN_RIGHT_BRACKET           // ]
	TOKEN_LEFT_BRACE              // {
	TOKEN_RIGHT_BRACE             // }
	TOKEN_LT                      // <
	TOKEN_GT                      // >
	TOKEN_COLON                   // :
	TOKEN_SEMICOLON               // ;
	TOKEN_DOT                     // .
	TOKEN_COMMA                   // ,
	TOKEN_EQUAL                   // =
	TOKEN_AT                      // @
	TOKEN_AND                     // &
	TOKEN_VERTICAL_BAR            // |
	TOKEN_QUOTE                   // "
	TOKEN_DUOQUOTE                // ""
	TOKEN_SINGLE_QUOTE            // '
	TOKEN_DUO_SINGLE_QUOTE        // ''
	TOKEN_BACKQUOTE               // `
	TOKEN_ESCAPE_CHARACTER        // \

	// comment
	TOKEN_MULTI_LINE_COMMENT_START // /*
	TOKEN_MULTI_LINE_COMMENT_END   // */
	TOKEN_COMMENT_SHARP            // #
	TOKEN_SINGLE_LINE_COMMENT      // --

	// literal
	TOKEN_NUMBER // number literal

	// keywords
	TOKEN_SELECT // select
	TOKEN_UPDATE // update
	TOKEN_DELETE // delete
	TOKEN_CREATE // create
	TOKEN_INSERT // insert

	TOKEN_OTHER_TOKEN // SourceCharacter - Words
)

var tokenNameMap = map[int]string{
	TOKEN_EOF:              "EOF",
	TOKEN_NOT:              "!",
	TOKEN_LEFT_PAREN:       "(",
	TOKEN_RIGHT_PAREN:      ")",
	TOKEN_LEFT_BRACKET:     "[",
	TOKEN_RIGHT_BRACKET:    "]",
	TOKEN_LEFT_BRACE:       "{",
	TOKEN_RIGHT_BRACE:      "}",
	TOKEN_LT:               "<",
	TOKEN_GT:               ">",
	TOKEN_COLON:            ":",
	TOKEN_SEMICOLON:        ";",
	TOKEN_DOT:              ".",
	TOKEN_COMMA:            ",",
	TOKEN_EQUAL:            "=",
	TOKEN_AT:               "@",
	TOKEN_AND:              "&",
	TOKEN_VERTICAL_BAR:     "|",
	TOKEN_QUOTE:            "\"",
	TOKEN_DUOQUOTE:         "\"\"",
	TOKEN_SINGLE_QUOTE:     "'",
	TOKEN_DUO_SINGLE_QUOTE: "''",
	TOKEN_BACKQUOTE:        "`",
	TOKEN_ESCAPE_CHARACTER: "\\",

	TOKEN_MULTI_LINE_COMMENT_START: "/*",
	TOKEN_MULTI_LINE_COMMENT_END:   "*/",
	TOKEN_COMMENT_SHARP:            "#",
	TOKEN_SINGLE_LINE_COMMENT:      "--",

	TOKEN_NUMBER: "number",

	TOKEN_SELECT: "select",
	TOKEN_UPDATE: "update",
	TOKEN_DELETE: "delete",
	TOKEN_CREATE: "create",
	TOKEN_INSERT: "insert",

	TOKEN_OTHER_TOKEN: "other_token",
}

var keywords = map[string]int{
	"select": TOKEN_SELECT,
	"update": TOKEN_UPDATE,
	"delete": TOKEN_DELETE,
	"create": TOKEN_CREATE,
	"insert": TOKEN_INSERT,
	"SELECT": TOKEN_SELECT,
	"UPDATE": TOKEN_UPDATE,
	"DELETE": TOKEN_DELETE,
	"CREATE": TOKEN_CREATE,
	"INSERT": TOKEN_INSERT,
}

var avaliableNumberParts = map[byte]bool{
	'0': true,
	'1': true,
	'2': true,
	'3': true,
	'4': true,
	'5': true,
	'6': true,
	'7': true,
	'8': true,
	'9': true,
	'-': true,
	'x': true,
	'X': true,
	'.': true,
	'+': true,
	'p': true,
	'P': true,
	'a': true,
	'b': true,
	'c': true,
	'd': true,
	'e': true,
	'f': true,
	'A': true,
	'B': true,
	'C': true,
	'D': true,
	'E': true,
	'F': true,
}

// regex match patterns
var regexNumber = regexp.MustCompile(`^0[xX][0-9a-fA-F]*(\.[0-9a-fA-F]*)?([pP][+\-]?[0-9]+)?|^[-]?[0-9]*(\.[0-9]*)?([eE][+\-]?[0-9]+)?`)
var reDecEscapeSeq = regexp.MustCompile(`^\\[0-9]{1,3}`)
var reHexEscapeSeq = regexp.MustCompile(`^\\x[0-9a-fA-F]{2}`)
var reUnicodeEscapeSeq = regexp.MustCompile(`^\\u\{[0-9a-fA-F]+\}`)

// lexer struct
type Lexer struct {
	sql              string
	lineNum          int
	nextToken        string
	nextTokenType    int
	nextTokenLineNum int
	pos              int
}

func NewLexer(sql string) *Lexer {
	return &Lexer{sql, 1, "", 0, 0, 0} // start at line 1 in default.
}

func (lexer *Lexer) GetLineNum() int {
	return lexer.lineNum
}

func (lexer *Lexer) GetPos() int {
	return lexer.pos
}

func (lexer *Lexer) NextTokenIs(tokenType int) (lineNum int, token string, err error) {
	nowLineNum, nowTokenType, nowToken, err := lexer.GetNextToken()
	if err != nil {
		return nowLineNum, nowToken, err
	}
	// syntax error
	if tokenType != nowTokenType {
		err := errors.New(fmt.Sprintf("line %d: syntax error near '%s'.", lexer.GetLineNum(), nowToken))
		return nowLineNum, nowToken, err
	}
	return nowLineNum, nowToken, nil
}

func (lexer *Lexer) LookAhead() (int, error) {
	// lexer.nextToken* already setted
	if lexer.nextTokenLineNum > 0 {
		return lexer.nextTokenType, nil
	}
	// set it
	nowLineNum := lexer.lineNum
	lineNum, tokenType, token, err := lexer.GetNextToken()
	if err != nil {
		return 0, err
	}

	lexer.lineNum = nowLineNum
	lexer.nextTokenLineNum = lineNum
	lexer.nextTokenType = tokenType
	lexer.nextToken = token
	return tokenType, nil
}

func (lexer *Lexer) nextSQLIs(s string) bool {
	return len(lexer.sql) >= len(s) && lexer.sql[0:len(s)] == s
}

func (lexer *Lexer) skipSQL(n int) {
	lexer.pos += n
	lexer.sql = lexer.sql[n:]
}

// target pattern
func isNewLine(c byte) bool {
	return c == '\r' || c == '\n'
}

func isWhiteSpace(c byte) bool {
	switch c {
	case '\t', '\n', '\v', '\f', '\r', ' ':
		return true
	}
	return false
}

func (lexer *Lexer) skipIgnored() {
	// matching
	for len(lexer.sql) > 0 {
		if lexer.nextSQLIs("\r\n") || lexer.nextSQLIs("\n\r") {
			lexer.skipSQL(2)
			lexer.lineNum += 1
		} else if isNewLine(lexer.sql[0]) {
			lexer.skipSQL(1)
			lexer.lineNum += 1
		} else if isWhiteSpace(lexer.sql[0]) {
			lexer.skipSQL(1)
			// check comment
		} else if lexer.nextSQLIs(tokenNameMap[TOKEN_COMMENT_SHARP]) || lexer.nextSQLIs(tokenNameMap[TOKEN_SINGLE_LINE_COMMENT]) {
			lexer.skipSQL(1)
			for !isNewLine(lexer.sql[0]) {
				lexer.skipSQL(1)
			}
		} else if lexer.nextSQLIs(tokenNameMap[TOKEN_MULTI_LINE_COMMENT_START]) {
			for !lexer.nextSQLIs(tokenNameMap[TOKEN_MULTI_LINE_COMMENT_END]) {
				lexer.skipSQL(1)
			}
			lexer.skipSQL(2) // skip "*/"
		} else {
			break
		}
	}
}

// use regex scan for number, identifier
func (lexer *Lexer) scan(regexp *regexp.Regexp) (string, error) {
	if token := regexp.FindString(lexer.sql); token != "" {
		lexer.skipSQL(len(token))
		return token, nil
	}
	err := errors.New("unreachable!")
	return "", err
}

// return content before token
func (lexer *Lexer) scanBeforeToken(token string) (string, error) {
	s := strings.Split(lexer.sql, token)
	if len(s) < 2 {
		err := errors.New("unreachable!")
		return "", err
	}
	lexer.skipSQL(len(s[0]))
	return s[0], nil
}

// NOTE: this method can skip escape character
func (lexer *Lexer) scanBeforeByte(b byte) (string, error) {
	docLen := len(lexer.sql)
	var r string
	var err error
	i := 0
	for ; i < docLen; i++ {
		// hit target
		if lexer.sql[i] == b {
			r = lexer.sql[:i]
			// convert escape character
			if r, err = lexer.escape(r); err != nil {
				return "", err
			}
			lexer.skipSQL(i)
			return r, nil
		}
		// skip escape character
		if lexer.sql[i] == '\\' {
			i += 2
		}
	}
	return "", errors.New("Can not find target byte.")
}

func (lexer *Lexer) scanNumber() (string, error) {
	docLen := len(lexer.sql)
	for i := 0; i < docLen; i++ {
		c := lexer.sql[i]
		if _, ok := avaliableNumberParts[c]; ok {
			continue
		} else {
			target := lexer.sql[:i]
			lexer.skipSQL(i)
			return target, nil
		}
	}
	err := errors.New("unreachable!")
	return "", err
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func (lexer *Lexer) scanWord() (string, error) {
	docLen := len(lexer.sql)
	for i := 0; i < docLen; i++ {
		c := lexer.sql[i]
		if c == '_' || isLetter(c) || isDigit(c) {
			continue
		} else {
			target := lexer.sql[:i]
			lexer.skipSQL(i)
			return target, nil
		}
	}
	err := errors.New("unreachable!")
	return "", err
}

func (lexer *Lexer) GetNextToken() (lineNum int, tokenType int, token string, err error) {
	// next token already loaded
	if lexer.nextTokenLineNum > 0 {
		lineNum = lexer.nextTokenLineNum
		tokenType = lexer.nextTokenType
		token = lexer.nextToken
		lexer.lineNum = lexer.nextTokenLineNum
		lexer.nextTokenLineNum = 0
		return
	}
	return lexer.MatchToken()

}

func (lexer *Lexer) MatchToken() (lineNum int, tokenType int, token string, err error) {
	lexer.skipIgnored()
	// finish
	if len(lexer.sql) == 0 {
		return lexer.lineNum, TOKEN_EOF, tokenNameMap[TOKEN_EOF], nil
	}
	// check token
	switch lexer.sql[0] {
	case '!':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_NOT, "!", nil
	case '(':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_LEFT_PAREN, "(", nil
	case ')':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_RIGHT_PAREN, ")", nil
	case '[':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_LEFT_BRACKET, "[", nil
	case ']':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_RIGHT_BRACKET, "]", nil
	case '{':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_LEFT_BRACE, "{", nil
	case '}':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_RIGHT_BRACE, "}", nil
	case '<':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_LT, "<", nil
	case '>':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_GT, ">", nil
	case ':':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_COLON, ":", nil
	case ';':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_SEMICOLON, ";", nil
	case '.':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_DOT, ".", nil
	case ',':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_COMMA, ",", nil
	case '=':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_EQUAL, "=", nil
	case '@':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_AT, "@", nil
	case '&':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_AND, "&", nil
	case '|':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_VERTICAL_BAR, "|", nil
	case '`':
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_BACKQUOTE, "`", nil
	case '"':
		if lexer.nextSQLIs("\"\"") {
			lexer.skipSQL(2)
			return lexer.lineNum, TOKEN_DUOQUOTE, "\"\"", nil
		}
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_QUOTE, "\"", nil
	case '\'':
		if lexer.nextSQLIs("''") {
			lexer.skipSQL(2)
			return lexer.lineNum, TOKEN_DUO_SINGLE_QUOTE, "''", nil
		}
		lexer.skipSQL(1)
		return lexer.lineNum, TOKEN_SINGLE_QUOTE, "'", nil
	}

	// check multiple character token
	if lexer.sql[0] == '_' || isLetter(lexer.sql[0]) {
		token, err := lexer.scanWord()
		if err != nil {
			return lexer.lineNum, 0, "", err
		}
		// to lowercase, SQL is not case sensitive
		token = strings.ToLower(token)
		if tokenType, isMatch := keywords[token]; isMatch {
			return lexer.lineNum, tokenType, token, nil
		} else {
			return lexer.lineNum, TOKEN_OTHER_TOKEN, token, nil
		}
	}
	if lexer.sql[0] == '.' || lexer.sql[0] == '-' || isDigit(lexer.sql[0]) {
		token, err := lexer.scanNumber()
		if err != nil {
			return lexer.lineNum, 0, "", err
		}
		return lexer.GetLineNum(), TOKEN_NUMBER, token, nil
	}

	// unexpected symbol
	err = errors.New(fmt.Sprintf("line %d: unexpected symbol near '%q'.", lexer.lineNum, lexer.sql[0]))
	return lexer.lineNum, 0, "", err
}

func (lexer *Lexer) escape(str string) (string, error) {
	var buf bytes.Buffer

	for len(str) > 0 {
		if str[0] != '\\' {
			buf.WriteByte(str[0])
			str = str[1:]
			continue
		}

		if len(str) == 1 {
			return "", errors.New("unfinished string")
		}

		switch str[1] {
		case 'a':
			buf.WriteByte('\a')
			str = str[2:]
			continue
		case 'b':
			buf.WriteByte('\b')
			str = str[2:]
			continue
		case 'f':
			buf.WriteByte('\f')
			str = str[2:]
			continue
		case 'n', '\n':
			buf.WriteByte('\n')
			str = str[2:]
			continue
		case 'r':
			buf.WriteByte('\r')
			str = str[2:]
			continue
		case 't':
			buf.WriteByte('\t')
			str = str[2:]
			continue
		case 'v':
			buf.WriteByte('\v')
			str = str[2:]
			continue
		case '"':
			buf.WriteByte('"')
			str = str[2:]
			continue
		case '\'':
			buf.WriteByte('\'')
			str = str[2:]
			continue
		case '\\':
			buf.WriteByte('\\')
			str = str[2:]
			continue
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // \ddd
			if found := reDecEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[1:], 10, 32)
				if d <= 0xFF {
					buf.WriteByte(byte(d))
					str = str[len(found):]
					continue
				}
				return "", errors.New(fmt.Sprintf("decimal escape too large near '%s'", found))
			}
		case 'x': // \xXX
			if found := reHexEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[2:], 16, 32)
				buf.WriteByte(byte(d))
				str = str[len(found):]
				continue
			}
		case 'u': // \u{XXX}
			if found := reUnicodeEscapeSeq.FindString(str); found != "" {
				d, err := strconv.ParseInt(found[3:len(found)-1], 16, 32)
				if err == nil && d <= 0x10FFFF {
					buf.WriteRune(rune(d))
					str = str[len(found):]
					continue
				}
				return "", errors.New(fmt.Sprintf("UTF-8 value too large near '%s'", found))
			}
		case 'z':
			str = str[2:]
			for len(str) > 0 && isWhiteSpace(str[0]) { // todo
				str = str[1:]
			}
			continue
		}
		return "", errors.New(fmt.Sprintf("invalid escape sequence near '\\%c'", str[1]))
	}

	return buf.String(), nil
}
