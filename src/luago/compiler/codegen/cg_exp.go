package codegen

import . "luago/compiler/ast"
import . "luago/compiler/lexer"
import . "luago/vm"

//Lua表达式可以大致分为字面量表达式、构造器表达式、运算符表达式、前缀表达式和vararg表达式
//编译表达式
func cgExp(fi *funcInfo, node Exp, a, n int) {
	switch exp := node.(type) {
	case *NilExp:
		fi.emitLoadNil(a, n)
	case *FalseExp:
		fi.emitLoadBool(a, 0, 0)
	case *TrueExp:
		fi.emitLoadBool(a, 1, 0)
	case *IntegerExp:
		fi.emitLoadK(a, exp.Val)
	case *FloatExp:
		fi.emitLoadK(a, exp.Val)
	case *StringExp:
		fi.emitLoadK(a, exp.Str)
	case *ParensExp:
		cgExp(fi, exp.Exp, a, 1)
	case *VarargExp:
		cgVarargExp(fi, exp, a, n)
	case *FuncDefExp:
		cgFuncDefExp(fi, exp, a)
	case *TableConstructorExp:
		cgTableConstructorExp(fi, exp, a)
	case *UnopExp:
		cgUnopExp(fi, exp, a)
	case *BinopExp:
		cgBinopExp(fi, exp, a)
	case *ConcatExp:
		cgConcatExp(fi, exp, a)
	case *NameExp:
		cgNameExp(fi, exp, a)
	case *TableAccessExp:
		cgTableAccessExp(fi, exp, a)
	case *FuncCallExp:
		cgFuncCallExp(fi, exp, a, n)
	}
}

func cgVarargExp(fi *funcInfo, node *VarargExp, a, n int) {
	//确保只有varagrg函数内才能出现vararg表达式
	if !fi.isVararg {
		panic("cannot use '...' outside a vararg function")
	}
	fi.emitVararg(a, n)
}

// f[a] := function(args) body end
//函数定义表达式
func cgFuncDefExp(fi *funcInfo, node *FuncDefExp, a int) {
	subFI := newFuncInfo(fi, node)
	fi.subFuncs = append(fi.subFuncs, subFI)

	for _, param := range node.ParList {
		subFI.addLocVar(param)
	}

	cgBlock(subFI, node.Block)
	subFI.exitScope()
	subFI.emitReturn(0, 0)

	//Lua编译器给每个函数都追加了一条RETURN指令
	bx := len(fi.subFuncs) - 1
	fi.emitClosure(a, bx)
}

//表构造表达式
func cgTableConstructorExp(fi *funcInfo, node *TableConstructorExp, a int) {
	nArr := 0
	for _, keyExp := range node.KeyExps {
		if keyExp == nil {
			nArr++
		}
	}
	nExps := len(node.KeyExps)
	multRet := nExps > 0 &&
		isVarargOrFuncCall(node.ValExps[nExps-1])
	//生成NEWTABLE指令
	fi.emitNewTable(a, nArr, nExps-nArr)

	arrIdx := 0
	for i, keyExp := range node.KeyExps {
		valExp := node.ValExps[i]

		if keyExp == nil {	//数组
			arrIdx++
			tmp := fi.allocReg()
			if i == nExps-1 && multRet {
				cgExp(fi, valExp, tmp, -1)
			} else {
				cgExp(fi, valExp, tmp, 1)
			}

			if arrIdx%50 == 0 || arrIdx == nArr { // LFIELDS_PER_FLUSH
				n := arrIdx % 50
				if n == 0 {
					n = 50
				}
				fi.freeRegs(n)
				c := (arrIdx-1)/50 + 1 // todo: c > 0xFF
				if i == nExps-1 && multRet {
					fi.emitSetList(a, 0, c)
				} else {
					fi.emitSetList(a, n, c)
				}
			}

			continue
		}

		//关联表
		//对每一个键值对，分别给键和值表达式分配局部变量，求值，并生成SETTABLE指令
		b := fi.allocReg()
		cgExp(fi, keyExp, b, 1)
		c := fi.allocReg()
		cgExp(fi, valExp, c, 1)
		fi.freeRegs(2)

		fi.emitSetTable(a, b, c)
	}
}

// r[a] := op exp
//一元运算符表达式
func cgUnopExp(fi *funcInfo, node *UnopExp, a int) {
	//先分配一个临时变量，然后对表达式求值，最后释放临时变量并生成相应的一元运算符指令
	b := fi.allocReg()
	cgExp(fi, node.Exp, b, 1)
	fi.emitUnaryOp(node.Op, a, b)
	fi.freeReg()
}

// r[a] := exp1 .. exp2
//拼接表达式
func cgConcatExp(fi *funcInfo, node *ConcatExp, a int) {
	for _, subExp := range node.Exps {
		a := fi.allocReg()
		cgExp(fi, subExp, a, 1)
	}

	c := fi.usedRegs - 1
	b := c - len(node.Exps) + 1
	fi.freeRegs(c - b + 1)
	fi.emitABC(OP_CONCAT, a, b, c)
}

// r[a] := exp1 op exp2
//二元运算符表达式
func cgBinopExp(fi *funcInfo, node *BinopExp, a int) {
	switch node.Op {
	case TOKEN_OP_AND, TOKEN_OP_OR: //逻辑与和逻辑或表达式
		b := fi.allocReg()
		cgExp(fi, node.Exp1, b, 1)
		fi.freeReg()
		if node.Op == TOKEN_OP_AND {
			fi.emitTestSet(a, b, 0)
		} else {
			fi.emitTestSet(a, b, 1)
		}
		pcOfJmp := fi.emitJmp(0, 0)

		b = fi.allocReg()
		cgExp(fi, node.Exp2, b, 1)
		fi.freeReg()
		fi.emitMove(a, b)
		fi.fixSbx(pcOfJmp, fi.pc()-pcOfJmp)
	default:	//其他二元运算符表达式
		b := fi.allocReg()	
		cgExp(fi, node.Exp1, b, 1)
		c := fi.allocReg()
		cgExp(fi, node.Exp2, c, 1)
		fi.emitBinaryOp(node.Op, a, b, c)
		fi.freeRegs(2)
	}
}

// r[a] := name
//名字和表访问表达式
//名字表达式的求值结果可能是局部变量、Upvalue或者全局变量
func cgNameExp(fi *funcInfo, node *NameExp, a int) {
	if r := fi.slotOfLocVar(node.Name); r >= 0 {	//局部变量，生成MOVE指令
		fi.emitMove(a, r)
	} else if idx := fi.indexOfUpval(node.Name); idx >= 0 {	//Upvalue，生成GETUPVAL指令
		fi.emitGetUpval(a, idx)
	} else { // x => _ENV['x']	全局变量，转换成表访问表达式
		taExp := &TableAccessExp{
			PrefixExp: &NameExp{0, "_ENV"},
			KeyExp:    &StringExp{0, node.Name},
		}
		cgTableAccessExp(fi, taExp, a)
	}
}

// r[a] := prefix[key]
func cgTableAccessExp(fi *funcInfo, node *TableAccessExp, a int) {
	//先后给表和键分配临时变量并对表达式求值，然后生成GETTABLE指令并释放临时变量
	b := fi.allocReg()
	cgExp(fi, node.PrefixExp, b, 1)
	c := fi.allocReg()
	cgExp(fi, node.KeyExp, c, 1)
	fi.emitGetTable(a, b, c)
	fi.freeRegs(2)
}

// r[a] := f(args)
//函数调用表达式
func cgFuncCallExp(fi *funcInfo, node *FuncCallExp, a, n int) {
	nArgs := prepFuncCall(fi, node, a)
	//生成CALL指令
	fi.emitCall(a, nArgs, n)
}

// return f(args)
func cgTailCallExp(fi *funcInfo, node *FuncCallExp, a int) {
	nArgs := prepFuncCall(fi, node, a)
	fi.emitTailCall(a, nArgs)
}

func prepFuncCall(fi *funcInfo, node *FuncCallExp, a int) int {
	nArgs := len(node.Args)
	lastArgIsVarargOrFuncCall := false
	//处理前缀表达式
	cgExp(fi, node.PrefixExp, a, 1)
	if node.NameExp != nil {
		c := 0x100 + fi.indexOfConstant(node.NameExp.Str)
		fi.emitSelf(a, a, c)
	}
	//依次处理每一个参数表达式
	for i, arg := range node.Args {
		tmp := fi.allocReg()
		if i == nArgs-1 && isVarargOrFuncCall(arg) {
			lastArgIsVarargOrFuncCall = true
			cgExp(fi, arg, tmp, -1)
		} else {
			cgExp(fi, arg, tmp, 1)
		}
	}
	fi.freeRegs(nArgs)

	if node.NameExp != nil {
		nArgs++
	}
	if lastArgIsVarargOrFuncCall {
		nArgs = -1
	}

	return nArgs
}
