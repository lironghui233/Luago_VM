package vm

import . "luago/api"

// R(A) := UpValue[B]
func getUpval(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	vm.Copy(LuaUpvalueIndex(b), a)
}

// UpValue[B] := R(A)
func setUpval(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	vm.Copy(a, LuaUpvalueIndex(b))
}

// R(A) := UpValue[B][RK(C)]
// 如果当前闭包的某个Upvalue是表，则GETTABUP指令（iABC模式）可以根据键从该表里取值，然后把值放入目标寄存器中
func getTabUp(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	vm.GetRK(c)
	vm.GetTable(LuaUpvalueIndex(b))
	vm.Replace(a)
}

// UpValue[A][RK(B)] := RK(C)
// 如果当前闭包的某个Upvalue是表，则SETTABUP指令（iABC）可以根据键往该表里写入值
func setTabUp(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1

	vm.GetRK(b)
	vm.GetRK(c)
	vm.SetTable(LuaUpvalueIndex(a))
}