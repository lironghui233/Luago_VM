package parser

import . "luago/compiler/ast"
import . "luago/compiler/lexer"

/* recursive descent parser */

func Parse(chunk, chunkName string) *Block {
	lexer := NewLexer(chunk, chunkName)	//词法分析器
	block := parseBlock(lexer)			//语法分析
	lexer.NextTokenOfKind(TOKEN_EOF)
	return block
}
