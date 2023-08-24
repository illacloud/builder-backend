package parser_sql

/**
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

func IsSelectSQL(lexer *Lexer) (bool, error) {

	var token int
	var nextToken string
	var err error

	for {
		token, err = lexer.LookAhead()
		if err != nil {
			return false, err
		}
		if isSQLEnd(token) {
			break
		}

		_, _, nextToken, err = lexer.GetNextToken()

		if err != nil {
			return false, err
		}
		switch nextToken {
		case tokenNameMap[TOKEN_SELECT]:
			return true, nil
		case tokenNameMap[TOKEN_INSERT]:
			fallthrough
		case tokenNameMap[TOKEN_UPDATE]:
			fallthrough
		case tokenNameMap[TOKEN_DELETE]:
			return false, nil
		}
	}
	// not a select query
	return false, nil
}

func isSQLEnd(tokenType int) bool {
	if tokenType == TOKEN_EOF {
		return true
	}
	return false
}
