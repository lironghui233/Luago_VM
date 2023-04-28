package state

func (self *luaState) PC() int {
	return self.stack.pc
}

func (self *luaState) AddPC(n int) {
	self.stack.pc += n
}

//根据PC索引从函数原型的指令表里取出当前指令，然后把PC加1，这样下次再调用该方法取出的值就是下一条指令。
func (self *luaState) Fetch() uint32 {
	i := self.stack.closure.proto.Code[self.stack.pc]
	self.stack.pc++
	return i
}

//根据索引从函数原型的常量表中取出一个常量值，然后把它推入栈顶
func (self *luaState) GetConst(idx int) {
	c := self.stack.closure.proto.Constants[idx]
	self.stack.push(c)
}

func (self *luaState) GetRK(rk int) {
	if rk > 0xFF { // constant
		self.GetConst(rk & 0xFF)
	} else { // register
		self.PushValue(rk + 1)
	}
}

//返回当前Lua函数所操作的寄存器数量
func (self *luaState) RegisterCount() int {
	return int(self.stack.closure.proto.MaxStackSize)
}

//把传递给当前Lua函数的变长参数推入栈顶（多退少补）
func (self *luaState) LoadVararg(n int) {
	if n < 0 {
		n = len(self.stack.varargs)
	}

	self.stack.check(n)
	self.stack.pushN(self.stack.varargs, n)
}

//把当前Lua函数的子函数的原型实例化为闭包推入栈顶
func (self *luaState) LoadProto(idx int) {
	stack := self.stack
	subProto := stack.closure.proto.Protos[idx]
	closure := newLuaClosure(subProto)
	stack.push(closure)

	for i, uvInfo := range subProto.Upvalues {
		uvIdx := int(uvInfo.Idx)
		if uvInfo.Instack == 1 {	//一个Upvalue捕获的是当前函数的局部变量，那么我们只要访问当前函数的局部变量即可
			if stack.openuvs == nil {	//记录暂时还处于开放状态（Open）的Upvalue的map
				stack.openuvs = map[int]*upvalue{}
			}

			if openuv, found := stack.openuvs[uvIdx]; found {	//把处于开放状态（Open）的Upvalue闭合
				closure.upvals[i] = openuv
			} else {	//暂存处于开放状态（Open）的Upvalue，当局部变量推出作用域时，需要把处于开放状态（Open）的Upvalue闭合，即上面if那个分支的情况
				closure.upvals[i] = &upvalue{&stack.slots[uvIdx]}
				stack.openuvs[uvIdx] = closure.upvals[i]
			}
		} else {	//一个Upvalue捕获的是更外围的函数中的局部变量，该Upvalue已被当前函数捕获，我们只要把该Upvalue传递给闭包即可
			closure.upvals[i] = stack.closure.upvals[uvIdx]
		}
	}
}

func (self *luaState) CloseUpvalues(a int) {
	for i, openuv := range self.stack.openuvs {
		if i >= a-1 {
			val := *openuv.val
			openuv.val = &val
			delete(self.stack.openuvs, i)
		}
	}
}
