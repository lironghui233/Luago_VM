package ast

//Lua共有11种表达式，这些表达式又可以分为5类：字面量表达式、构造器表达式、运算符表达式、vararg表达式、前缀表达式
//字面量表达式：nil、布尔、数字、字符串
//构造器表达式：表构造器、函数构造器
//运算符表达式：一元运算符表达式、二元运算符表达式

/*
exp ::=  nil | false | true | Numeral | LiteralString | ‘...’ | functiondef |
	 prefixexp | tableconstructor | exp binop exp | unop exp

prefixexp ::= var | functioncall | ‘(’ exp ‘)’

var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name

functioncall ::=  prefixexp args | prefixexp ‘:’ Name args
*/
type Exp interface{}

type NilExp struct{ Line int }    // nil
type TrueExp struct{ Line int }   // true
type FalseExp struct{ Line int }  // false
type VarargExp struct{ Line int } // ...

// Numeral
type IntegerExp struct {
	Line int
	Val  int64
}
type FloatExp struct {
	Line int
	Val  float64
}

// LiteralString
type StringExp struct {
	Line int
	Str  string
}

/*
prefixexp ::= Name |
              ‘(’ exp ‘)’ |
              prefixexp ‘[’ exp ‘]’ |
              prefixexp ‘.’ Name |
              prefixexp ‘:’ Name args |
              prefixexp args
*/
type NameExp struct {
	Line int
	Name string
}

// unop exp
//一元运算符表达式
type UnopExp struct {
	Line int // line of operator
	Op   int // operator
	Exp  Exp
}

// exp1 op exp2
//二元运算符表达式
type BinopExp struct {
	Line int // line of operator
	Op   int // operator
	Exp1 Exp
	Exp2 Exp
}

//拼接运算符表达式
type ConcatExp struct {
	Line int // line of last ..
	Exps []Exp
}

// tableconstructor ::= ‘{’ [fieldlist] ‘}’
// fieldlist ::= field {fieldsep field} [fieldsep]
// field ::= ‘[’ exp ‘]’ ‘=’ exp | Name ‘=’ exp | exp
// fieldsep ::= ‘,’ | ‘;’
//表构造表达式
type TableConstructorExp struct {
	Line     int // line of `{` ?
	LastLine int // line of `}`
	KeyExps  []Exp
	ValExps  []Exp
}

// functiondef ::= function funcbody
// funcbody ::= ‘(’ [parlist] ‘)’ block end
// parlist ::= namelist [‘,’ ‘...’] | ‘...’
// namelist ::= Name {‘,’ Name}
//函数定义表达式
type FuncDefExp struct {
	Line     int
	LastLine int // line of `end`
	ParList  []string
	IsVararg bool
	Block    *Block
}

//圆括号表达式
type ParensExp struct {
	Exp Exp
}

//表访问表达式
type TableAccessExp struct {
	LastLine  int // line of `]` ?
	PrefixExp Exp
	KeyExp    Exp
}

//函数调用表达式
type FuncCallExp struct {
	Line      int // line of `(` ?
	LastLine  int // line of ')'
	PrefixExp Exp
	NameExp   *StringExp
	Args      []Exp
}
