package ast

//在命令式编程语言里，语句（Statement）是最基本的执行单位，表达式（Expression）则是构成语句的要素之一。
//语句和表达式的主要区别在于：语句只能执行不能用于求值，而表达式只能用于求值不能单独执行。

//Lua的十五种语句：
/*
stat ::=  ‘;’ |
	 varlist ‘=’ explist |
	 functioncall |
	 label |
	 break |
	 goto Name |
	 do block end |
	 while exp do block end |
	 repeat block until exp |
	 if exp then block {elseif exp then block} [else block] end |
	 for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end |
	 for namelist in explist do block end |
	 function funcname funcbody |
	 local function Name funcbody |
	 local namelist [‘=’ explist]
*/
type Stat interface{}

type EmptyStat struct{}              // ‘;’
type BreakStat struct{ Line int }    // break
type LabelStat struct{ Name string } // ‘::’ Name ‘::’
type GotoStat struct{ Name string }  // goto Name
type DoStat struct{ Block *Block }   // do block end
type FuncCallStat = FuncCallExp      // functioncall

// while exp do block end
//while语句
type WhileStat struct {
	Exp   Exp
	Block *Block
}

// repeat block until exp
//repeat语句
type RepeatStat struct {
	Block *Block
	Exp   Exp
}

// if exp then block {elseif exp then block} [else block] end
//if语句
type IfStat struct {
	//表达式和语句块按索引一一对应，索引0处是if-then表达式和块，其他索引处是elseif-then表达式和块
	Exps   []Exp
	Blocks []*Block
}

// for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
//数值for循环语句
type ForNumStat struct {
	LineOfFor int			//关键字for所在的行号
	LineOfDo  int			//关键字do所在的行号
	VarName   string
	InitExp   Exp
	LimitExp  Exp
	StepExp   Exp
	Block     *Block
}

// for namelist in explist do block end
// namelist ::= Name {‘,’ Name}
// explist ::= exp {‘,’ exp}
//通用for循环语句
type ForInStat struct {
	LineOfDo int			//关键字do所在的行号
	NameList []string
	ExpList  []Exp
	Block    *Block
}

// varlist ‘=’ explist
// varlist ::= var {‘,’ var}
// var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
//赋值语句
type AssignStat struct {
	LastLine int			//末尾行号
	VarList  []Exp
	ExpList  []Exp
}

// local namelist [‘=’ explist]
// namelist ::= Name {‘,’ Name}
// explist ::= exp {‘,’ exp}
//局部变量声明语句
type LocalVarDeclStat struct {
	LastLine int			//末尾行号
	NameList []string
	ExpList  []Exp
}

// local function Name funcbody
//局部函数定义语句
type LocalFuncDefStat struct {
	Name string				//函数名
	Exp  *FuncDefExp		//函数定义表达式
}
