package codegen

import . "luago/compiler/ast"

//函数体由任意条语句和一条可选的返回语句构成，所以我们先循环调用cgStat()函数处理每一条语句。如果有返回语句，则调用cgRetStat()函数进行处理
func cgBlock(fi *funcInfo, node *Block) {
	for _, stat := range node.Stats {
		cgStat(fi, stat)
	}

	if node.RetExps != nil {
		cgRetStat(fi, node.RetExps)
	}
}

func cgRetStat(fi *funcInfo, exps []Exp) {
	nExps := len(exps)
	//如果返回语句后面没有任何表达式，那么只要生成一条RETURN指令即可
	if nExps == 0 {
		fi.emitReturn(0, 0)
		return
	}

	if nExps == 1 {
		if nameExp, ok := exps[0].(*NameExp); ok {
			if r := fi.slotOfLocVar(nameExp.Name); r >= 0 {
				fi.emitReturn(r, 1)
				return
			}
		}
		if fcExp, ok := exps[0].(*FuncCallExp); ok {
			r := fi.allocReg()
			cgTailCallExp(fi, fcExp, r)
			fi.freeReg()
			fi.emitReturn(r, -1)
			return
		}
	}

	//最后一个表达式为vararg或者函数调用的情况进行处理
	multRet := isVarargOrFuncCall(exps[nExps-1])
	for i, exp := range exps {
		r := fi.allocReg()
		if i == nExps-1 && multRet {	//多参数
			cgExp(fi, exp, r, -1)
		} else {
			cgExp(fi, exp, r, 1)
		}
	}
	fi.freeRegs(nExps)

	a := fi.usedRegs // correct?
	if multRet {
		fi.emitReturn(a, -1)
	} else {
		fi.emitReturn(a, nExps)
	}
}
