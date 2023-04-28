package ast

//抽象语法树（AST，Abstract Syntax Tree）

// chunk ::= block
// type Chunk *Block

// block ::= {stat} [retstat]
// retstat ::= return [explist] [‘;’]
// explist ::= exp {‘,’ exp}
type Block struct {
	LastLine int
	Stats    []Stat		//语句
	RetExps  []Exp		//表达式
}
