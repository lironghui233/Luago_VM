//其他类型指令（2条）

package vm

import . "luago/api"

func move(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1; b += 1	//寄存器索引加1才是相应的栈索引
	vm.Copy(b, a) 
}

func jmp(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()

	vm.AddPC(sBx)
	if a != 0 {
		vm.CloseUpvalues(a)
	}
}
