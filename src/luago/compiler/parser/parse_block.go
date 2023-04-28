package parser

import . "luago/compiler/ast"
import . "luago/compiler/lexer"

//语法分析器的作用是按照某种语言的BNF（巴科斯范式）描述将这种语言的源代码转换成抽象语法树（AST）以供后续阶段使用

// block ::= {stat} [retstat]
func parseBlock(lexer *Lexer) *Block {
	return &Block{
		Stats:    parseStats(lexer),
		RetExps:  parseRetExps(lexer),
		LastLine: lexer.Line(),
	}
}

func parseStats(lexer *Lexer) []Stat {
	stats := make([]Stat, 0, 8)
	//循环调用parseStat()函数解析语句，直到通过前瞻看到关键字return或者发现块已经结束为止
	for !_isReturnOrBlockEnd(lexer.LookAhead()) {
		stat := parseStat(lexer)
		if _, ok := stat.(*EmptyStat); !ok {
			stats = append(stats, stat)
		}
	}
	return stats
}

func _isReturnOrBlockEnd(tokenKind int) bool {
	switch tokenKind {
	case TOKEN_KW_RETURN, TOKEN_EOF, TOKEN_KW_END,
		TOKEN_KW_ELSE, TOKEN_KW_ELSEIF, TOKEN_KW_UNTIL:
		return true
	}
	return false
}

// retstat ::= return [explist] [‘;’]
// explist ::= exp {‘,’ exp}
func parseRetExps(lexer *Lexer) []Exp {
	//如果不是关键字return，说明没有返回语句，直接返回nil即可
	if lexer.LookAhead() != TOKEN_KW_RETURN {
		return nil
	}
	//跳过关键字return
	lexer.NextToken()
	switch lexer.LookAhead() {
		//块已经结束，那么返回语句没有任何表达式
	case TOKEN_EOF, TOKEN_KW_END,
		TOKEN_KW_ELSE, TOKEN_KW_ELSEIF, TOKEN_KW_UNTIL:
		return []Exp{}
		//分号，那么返回语句没有任何表达式
	case TOKEN_SEP_SEMI:
		lexer.NextToken()
		return []Exp{}
	default:
		//解析表达式序列
		exps := parseExpList(lexer)
		//跳过可选的分号
		if lexer.LookAhead() == TOKEN_SEP_SEMI {
			lexer.NextToken()
		}
		return exps
	}
}
