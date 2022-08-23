package parser_sql

import (
	"errors"
	"fmt"
)

/*
 * SourceCharacter Expression
 * SourceCharacter ::=  #x0009 | #x000A | #x000D | [#x0020-#xFFFF] // /[\u0009\u000A\u000D\u0020-\uFFFF]/
 *
 *
 *
 * Ignored Tokens Expression
 * Ignored            ::= UnicodeBOM | WhiteSpace | LineTerminator | SingleLineComment | MultiLineComment
 * UnicodeBOM         ::= #xFEFF  // Byte Order Mark (U+FEFF)
 * WhiteSpace         ::= #x0009 | #x0020 // ASCII: \t | Space, Horizontal Tab (U+0009), Space (U+0020)
 * LineTerminator     ::= #x000A | #x000D | #x000D#x000A   // ASCII: \n | \r\n | \r, New Line (U+000A) | Carriage Return (U+000D) [Lookahead != New Line (U+000A)] | Carriage Return (U+000D)New Line (U+000A)
 * SingleLineComment  ::= "#" CommentChar* | "--" CommentChar*
 * MultiLineComment   ::= "\/*" SourceCharacter "*\/"
 * CommentChar        ::= SourceCharacter - LineTerminator
 *
 * Lexical Tokens Expression
 * Token                ::= Words | OtherToken | StringValue
 * Words                ::= [_A-Za-z][_0-9A-Za-z]*
 * OtherToken           ::= SourceCharacter - Words
 * StringValue          ::= '"' '"' | '"' StringCharacter* '"' | '"""' BlockStringCharacter* '"""'
 * StringCharacter      ::= SourceCharacter - '"' | SourceCharacter - "\" | SourceCharacter - LineTerminator | "\u" EscapedUnicode | "\" EscapedCharacter // SourceCharacter but not " or \ or LineTerminator | \uEscapedUnicode | \EscapedCharacter
 * EscapedUnicode       ::= [#x0000-#xFFFF]
 * EscapedCharacter     ::= '"' | '\' | '/' | 'b' | 'f' | 'n' | 'r' | 't'
 * BlockStringCharacter ::= SourceCharacter - '"""' | SourceCharacter - '\"""' | '\"""'
 *
 * SQL             ::= Ignored Statement+ Ignored
 * Statement       ::= Ignored QueryType Ignored Query Ignored
 * QueryType       ::= "select" | "update" | "delete" | "create"
 * Query           ::= Token
 *
 */

func parseWords(lexer *Lexer) (*Words, error) {
	lineNum, _, token := lexer.GetNextToken()
	for _, b := range []rune(token) {
		if b == '_' ||
			b >= 'a' && b <= 'z' ||
			b >= 'A' && b <= 'Z' ||
			b >= '0' && b <= '9' {
			continue
		} else {
			err := fmt.Sprintf("parseWords(): line %d: unexpected symbol near '%v', it is not a words expression", lineNum, token)
			return nil, errors.New(err)
		}
	}
	return &Words{lineNum, token}, nil
}

// output golang built-in string type value
func parseStringValueSimple(lexer *Lexer) (string, error) {
	var str string
	var err error
	// quotes
	if lexer.LookAhead() == TOKEN_DUOQUOTE {
		lexer.NextTokenIs(TOKEN_DUOQUOTE)
		return "", nil
	}
	if lexer.LookAhead() == TOKEN_QUOTE {
		lexer.NextTokenIs(TOKEN_QUOTE)
		quoteRune := []byte(tokenNameMap[TOKEN_QUOTE])
		if str, err = lexer.scanBeforeByte(quoteRune[0]); err != nil {
			return "", err
		}
		lexer.NextTokenIs(TOKEN_QUOTE)
		return str, nil
	}
	err = errors.New("not a StringValue")
	return "", err
}

func parseSQL(lexer *Lexer) (*SQL, error) {
	var document SQL
	var err error

	// LastLineNum
	document.LastLineNum = lexer.GetLineNum()
	// Statement+
	if document.Statements, err = parseStatements(lexer); err != nil {
		return nil, err
	}
	return &document, nil
}

func isSQLEnd(tokenType int) bool {
	if tokenType == TOKEN_EOF {
		return true
	}
	return false
}

func parseStatements(lexer *Lexer) ([]Statement, error) {
	var definitions []Statement
	for !isDocumentEnd(lexer.LookAhead()) {
		var definition Statement
		var err error
		if definition, err = parseStatement(lexer); err != nil {
			return nil, err
		}
		definitions = append(definitions, definition)
	}
	return definitions, nil
}

func parseStatement(lexer *Lexer) (Statement, error) {
	switch lexer.LookAhead() {
	}

}
