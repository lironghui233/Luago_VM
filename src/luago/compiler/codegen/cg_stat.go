package codegen

import . "luago/compiler/ast"

//编译语句
func cgStat(fi *funcInfo, node Stat) {
	switch stat := node.(type) {
	case *FuncCallStat:
		cgFuncCallStat(fi, stat)
	case *BreakStat:
		cgBreakStat(fi, stat)
	case *DoStat:
		cgDoStat(fi, stat)
	case *WhileStat:
		cgWhileStat(fi, stat)
	case *RepeatStat:
		cgRepeatStat(fi, stat)
	case *IfStat:
		cgIfStat(fi, stat)
	case *ForNumStat:
		cgForNumStat(fi, stat)
	case *ForInStat:
		cgForInStat(fi, stat)
	case *AssignStat:
		cgAssignStat(fi, stat)
	case *LocalVarDeclStat:
		cgLocalVarDeclStat(fi, stat)
	case *LocalFuncDefStat:
		cgLocalFuncDefStat(fi, stat)
	case *LabelStat, *GotoStat:
		panic("label and goto statements are not supported!")
	}
}

func cgLocalFuncDefStat(fi *funcInfo, node *LocalFuncDefStat) {
	r := fi.addLocVar(node.Name)
	cgFuncDefExp(fi, node.Exp, r)
}

func cgFuncCallStat(fi *funcInfo, node *FuncCallStat) {
	r := fi.allocReg()
	cgFuncCallExp(fi, node, r, 0)
	fi.freeReg()
}

func cgBreakStat(fi *funcInfo, node *BreakStat) {
	pc := fi.emitJmp(0, 0)
	fi.addBreakJmp(pc)
}

func cgDoStat(fi *funcInfo, node *DoStat) {
	fi.enterScope(false)
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals()	//将处于开启状态的Upvalue闭合
	fi.exitScope()
}

/*
           ______________
          /  false? jmp  |
         /               |
while exp do block end <-'
      ^           \
      |___________/
           jmp
*/
func cgWhileStat(fi *funcInfo, node *WhileStat) {
	//第一步，记住当前pc，因为后面计算跳转偏移量时要用到
	pcBeforeExp := fi.pc()
	//第二步，分配临时变量，对表达式进行求值，然后是否临时变量
	r := fi.allocReg()
	cgExp(fi, node.Exp, r, 1)
	fi.freeReg()
	//第三步，生成Test和JMP指令，实现条件跳转
	fi.emitTest(r, 0)
	pcJmpToEnd := fi.emitJmp(0, 0)
	//第四步，对块进行处理，生成一条JMP指令跳转到最开始
	fi.enterScope(true)
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals()
	fi.emitJmp(0, pcBeforeExp-fi.pc()-1)
	fi.exitScope()
	//第五步，修复第一条JMP指令的偏移量
	fi.fixSbx(pcJmpToEnd, fi.pc()-pcJmpToEnd)
}

/*
        ______________
       |  false? jmp  |
       V              /
repeat block until exp
*/
func cgRepeatStat(fi *funcInfo, node *RepeatStat) {
	fi.enterScope(true)

	pcBeforeBlock := fi.pc()
	cgBlock(fi, node.Block)

	r := fi.allocReg()
	cgExp(fi, node.Exp, r, 1)
	fi.freeReg()

	fi.emitTest(r, 0)
	fi.emitJmp(fi.getJmpArgA(), pcBeforeBlock-fi.pc()-1)
	fi.closeOpenUpvals()

	fi.exitScope()
}

/*
         _________________       _________________       _____________
        / false? jmp      |     / false? jmp      |     / false? jmp  |
       /                  V    /                  V    /              V
if exp1 then block1 elseif exp2 then block2 elseif true then block3 end <-.
                   \                       \                       \      |
                    \_______________________\_______________________\_____|
                    jmp                     jmp                     jmp
*/
func cgIfStat(fi *funcInfo, node *IfStat) {
	pcJmpToEnds := make([]int, len(node.Exps))
	pcJmpToNextExp := -1

	for i, exp := range node.Exps {
		if pcJmpToNextExp >= 0 {
			fi.fixSbx(pcJmpToNextExp, fi.pc()-pcJmpToNextExp)
		}

		r := fi.allocReg()
		cgExp(fi, exp, r, 1)
		fi.freeReg()

		fi.emitTest(r, 0)
		pcJmpToNextExp = fi.emitJmp(0, 0)

		fi.enterScope(false)
		cgBlock(fi, node.Blocks[i])
		fi.closeOpenUpvals()
		fi.exitScope()
		if i < len(node.Exps)-1 {
			pcJmpToEnds[i] = fi.emitJmp(0, 0)
		} else {
			pcJmpToEnds[i] = pcJmpToNextExp
		}
	}

	for _, pc := range pcJmpToEnds {
		fi.fixSbx(pc, fi.pc()-pc)
	}
}

func cgForNumStat(fi *funcInfo, node *ForNumStat) {
	fi.enterScope(true)
	//第一步，声明三个特殊的局部变量，分别存储索引、限制和步长
	cgLocalVarDeclStat(fi, &LocalVarDeclStat{
		NameList: []string{"(for index)", "(for limit)", "(for step)"},
		ExpList:  []Exp{node.InitExp, node.LimitExp, node.StepExp},
	})
	fi.addLocVar(node.VarName)
	//第二步，生成FORPREP指令，处理快，然后生成FORLOOP指令
	a := fi.usedRegs - 4
	pcForPrep := fi.emitForPrep(a, 0)
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals()
	pcForLoop := fi.emitForLoop(a, 0)
	//第三步，把指令里的调整偏移量修复
	fi.fixSbx(pcForPrep, pcForLoop-pcForPrep-1)
	fi.fixSbx(pcForLoop, pcForPrep-pcForLoop)

	fi.exitScope()
}

func cgForInStat(fi *funcInfo, node *ForInStat) {
	fi.enterScope(true)
	//第一步
	cgLocalVarDeclStat(fi, &LocalVarDeclStat{
		NameList: []string{"(for generator)", "(for state)", "(for control)"},
		ExpList:  node.ExpList,
	})
	for _, name := range node.NameList {
		fi.addLocVar(name)
	}
	//第二步
	pcJmpToTFC := fi.emitJmp(0, 0)
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals()
	fi.fixSbx(pcJmpToTFC, fi.pc()-pcJmpToTFC)
	//第三步
	rGenerator := fi.slotOfLocVar("(for generator)")
	fi.emitTForCall(rGenerator, len(node.NameList))
	fi.emitTForLoop(rGenerator+2, pcJmpToTFC-fi.pc()-1)

	fi.exitScope()
}

//局部变量声明语句
func cgLocalVarDeclStat(fi *funcInfo, node *LocalVarDeclStat) {
	exps := removeTailNils(node.ExpList)
	nExps := len(exps)
	nNames := len(node.NameList)

	oldRegs := fi.usedRegs
	if nExps == nNames {	//等号左侧声明的局部变量和等号右侧提供的表达式一样多
		for _, exp := range exps {
			a := fi.allocReg()
			cgExp(fi, exp, a, 1)
		}
	} else if nExps > nNames {
		for i, exp := range exps {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {	//如果最后一个表达式是vararg或者函数调用
				cgExp(fi, exp, a, 0)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}
	} else { // nNames > nExps
		multRet := false
		for i, exp := range exps {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {	//如果最后一个表达式是vararg或者函数调用
				multRet = true
				n := nNames - nExps + 1
				cgExp(fi, exp, a, n)
				fi.allocRegs(n - 1)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}
		if !multRet {
			n := nNames - nExps
			a := fi.allocRegs(n)
			fi.emitLoadNil(a, n)
		}
	}
	//释放临时变量，声明局部变量即可
	fi.usedRegs = oldRegs
	for _, name := range node.NameList {
		fi.addLocVar(name)
	}
}

//赋值语句
func cgAssignStat(fi *funcInfo, node *AssignStat) {
	//由于赋值语句等号左边可以出现t[k]这样的表达式，等号右边可以出现任意表达式，所以我们需要先分配临时变量，对这些表达式进行求值，然后再统一生成赋值指令
	exps := removeTailNils(node.ExpList)
	nExps := len(exps)
	nVars := len(node.VarList)

	tRegs := make([]int, nVars)		//为表分配的临时变量
	kRegs := make([]int, nVars)		//为键分配的临时变量
	vRegs := make([]int, nVars)		//为值分配的临时变量
	oldRegs := fi.usedRegs

	for i, exp := range node.VarList {
		if taExp, ok := exp.(*TableAccessExp); ok {	//先处理等号左侧的索引表达式，分配临时变量，并对表和键求值
			tRegs[i] = fi.allocReg()
			cgExp(fi, taExp.PrefixExp, tRegs[i], 1)
			kRegs[i] = fi.allocReg()
			cgExp(fi, taExp.KeyExp, kRegs[i], 1)
		} else {
			name := exp.(*NameExp).Name
			if fi.slotOfLocVar(name) < 0 && fi.indexOfUpval(name) < 0 {
				// global var
				kRegs[i] = -1
				if fi.indexOfConstant(name) > 0xFF {
					kRegs[i] = fi.allocReg()
				}
			}
		}
	}
	for i := 0; i < nVars; i++ {	//为等号右侧的表达式计算寄存器索引
		vRegs[i] = fi.usedRegs + i
	}

	//考虑等号左侧声明的局部变量和等号右侧提供的表达式的数量，考虑多重赋值
	if nExps >= nVars {
		for i, exp := range exps {
			a := fi.allocReg()
			if i >= nVars && i == nExps-1 && isVarargOrFuncCall(exp) {
				cgExp(fi, exp, a, 0)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}
	} else { // nVars > nExps
		multRet := false
		for i, exp := range exps {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				multRet = true
				n := nVars - nExps + 1
				cgExp(fi, exp, a, n)
				fi.allocRegs(n - 1)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}
		if !multRet {
			n := nVars - nExps
			a := fi.allocRegs(n)
			fi.emitLoadNil(a, n)
		}
	}

	for i, exp := range node.VarList {
		if nameExp, ok := exp.(*NameExp); ok {
			varName := nameExp.Name
			if a := fi.slotOfLocVar(varName); a >= 0 {	//给局部变量赋值，生成MOVE指令
				fi.emitMove(a, vRegs[i])
			} else if b := fi.indexOfUpval(varName); b >= 0 {	//给Upvalue赋值，生成SETUPVAL指令
				fi.emitSetUpval(vRegs[i], b)
			} else if a := fi.slotOfLocVar("_ENV"); a >= 0 {	//按索引给表赋值，生成SETTABLE指令
				if kRegs[i] < 0 {
					b := 0x100 + fi.indexOfConstant(varName)
					fi.emitSetTable(a, b, vRegs[i])
				} else {
					fi.emitSetTable(a, kRegs[i], vRegs[i])
				}
			} else { // global var	给全局变量赋值，生成SETTABUP指令
				a := fi.indexOfUpval("_ENV")
				if kRegs[i] < 0 {
					b := 0x100 + fi.indexOfConstant(varName)
					fi.emitSetTabUp(a, b, vRegs[i])
				} else {
					fi.emitSetTabUp(a, kRegs[i], vRegs[i])
				}
			}
		} else {
			fi.emitSetTable(tRegs[i], kRegs[i], vRegs[i])
		}
	}

	// todo
	fi.usedRegs = oldRegs
}
