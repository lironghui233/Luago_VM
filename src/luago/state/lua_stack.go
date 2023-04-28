package state

import . "luago/api"

type luaStack struct {
	/* virtual stack */
	slots 		[]luaValue				//存放lua值
	top 		int						//记录栈顶索引
	/* call info */
	state   	*luaState				//用于间接访问注册表	
	closure		*closure				//闭包
	varargs		[]luaValue				//变长参数
	openuvs 	map[int]*upvalue		//记录暂时还处于开放状态（Open）的Upvalue
	pc 			int						//程序计数器
	/* linked list */
	prev		*luaStack				//让调用帧变成链表节点
}

//创建指定容量的栈
func newLuaStack(size int, state *luaState) *luaStack {
	return &luaStack{
		slots: make([]luaValue, size),
		top:   0,
		state: state,
	}
}

//检查栈的空间是否还可以容纳（推入）至少n个值，如果不满足条件，则调用Go语言内置的append()函数进行扩容
func (self *luaStack) check(n int) {
	free := len(self.slots) - self.top
	for i := free; i < n; i++ {
		self.slots = append(self.slots, nil)
	}
}

//将值推入栈顶
func (self *luaStack) push(val luaValue) {
	if self.top == len(self.slots) {
		panic("stack overflow!")
	}
	self.slots[self.top] = val
	self.top++
}

//从栈顶弹出一个值
func (self *luaStack) pop() luaValue {
	if self.top < 1 {
		panic("stack underflow!")
	}
	self.top--
	val := self.slots[self.top]
	self.slots[self.top] = nil
	return val
}

//将n个值推入栈顶
func (self *luaStack) pushN(vals []luaValue, n int) {
	nVals := len(vals)
	if n < 0 {
		n = nVals
	}

	for i := 0; i < n; i++ {
		if i < nVals {
			self.push(vals[i])
		} else {
			self.push(nil)
		}
	}
}

//将n个值从栈顶弹出
func (self *luaStack) popN(n int) []luaValue {
	vals := make([]luaValue, n)
	for i := n - 1; i >= 0; i-- {
		vals[i] = self.pop()
	}
	return vals
}

//把索引转换成绝对索引
func (self *luaStack) absIndex(idx int) int {
	if idx >= 0 || idx <= LUA_REGISTRYINDEX {
		return idx
	}
	return idx + self.top + 1
}

//判断索引是否有效
func (self *luaStack) isValid(idx int) bool {
	if idx < LUA_REGISTRYINDEX { /* upvalues 伪索引*/
		uvIdx := LUA_REGISTRYINDEX - idx - 1	//转成真实索引（从0开始）
		c := self.closure
		return c != nil && uvIdx < len(c.upvals)
	}
	if idx == LUA_REGISTRYINDEX {
		return true
	}
	absIdx := self.absIndex(idx)
	return absIdx > 0 && absIdx <= self.top
}

//根据索引从栈里取值
func (self *luaStack) get(idx int) luaValue {
	if idx < LUA_REGISTRYINDEX { /* upvalues 伪索引*/
		uvIdx := LUA_REGISTRYINDEX - idx - 1	//转成真实索引（从0开始）
		c := self.closure
		if c == nil || uvIdx >= len(c.upvals) {
			return nil
		}
		return *(c.upvals[uvIdx].val)
	}

	if idx == LUA_REGISTRYINDEX {
		return self.state.registry
	}

	absIdx := self.absIndex(idx)
	if absIdx > 0 && absIdx <= self.top {
		return self.slots[absIdx-1]
	}
	return nil
}

//根据索引往栈里写入值
func (self *luaStack) set(idx int, val luaValue) {
	if idx < LUA_REGISTRYINDEX { /* upvalues 伪索引*/
		uvIdx := LUA_REGISTRYINDEX - idx - 1	//转成真实索引（从0开始）
		c := self.closure
		if c != nil && uvIdx < len(c.upvals) {
			*(c.upvals[uvIdx].val) = val
		}
		return
	}

	if idx == LUA_REGISTRYINDEX {
		self.state.registry = val.(*luaTable)
		return
	}
	
	absIdx := self.absIndex(idx)
	if absIdx > 0 && absIdx <= self.top {
		self.slots[absIdx-1] = val
		return
	}
	panic("invalid index")
}

func (self *luaStack) reverse(from, to int) {
	slots := self.slots
	for from < to {
		slots[from], slots[to] = slots[to], slots[from]
		from++
		to--
	}
}

