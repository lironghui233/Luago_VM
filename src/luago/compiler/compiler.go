package compiler

import "luago/binchunk"
import "luago/compiler/codegen"
import "luago/compiler/parser"

//把语法分析和代码生成阶段合二为一
func Compile(chunk, chunkName string) *binchunk.Prototype {
	ast := parser.Parse(chunk, chunkName)	//词法分析、语法分析（Lua代码转成AST）
	return codegen.GenProto(ast)	//代码生成（AST转成*Prototype）
}
