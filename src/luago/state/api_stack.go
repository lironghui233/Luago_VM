//下面方法对栈本身进行操作

package state

import . "luago/api"

//返回栈顶索引
func (self *luaState) GetTop() int {
	return self.stack.top
}

//把索引转换为绝对索引
func (self *luaState) AbsIndex(idx int) int {
	return self.stack.absIndex(idx)
}

//检查栈的空间是否还可以容纳（推入）至少n个值，如果不满足条件，则扩容
func (self *luaState) CheckStack(n int) bool {
	self.stack.check(n)
	return true	//never fails
}

//从栈顶弹出n个值
func (self *luaState) Pop(n int) {
	// for i := 0; i < n; i++ {
	// 	self.stack.pop()
	// }
	self.SetTop(-n-1)
}

//把值从一个位置复制到另一个位置
func (self *luaState) Copy(fromIdx, toIdx int) {
	val := self.stack.get(fromIdx)
	self.stack.set(toIdx, val)
}

//把指定索引处的值推入栈顶
func (self *luaState) PushValue(idx int) {
	val := self.stack.get(idx)
	self.stack.push(val)
}

//将栈顶值弹出，然后写入指定位置
func (self *luaState) Replace(idx int) {
	val := self.stack.pop()
	self.stack.set(idx, val)
}

//将栈顶值弹出，然后插入指定位置
func (self *luaState) Insert(idx int) {
	self.Rotate(idx, 1)
}

//删除指定索引处的值
func (self *luaState) Remove(idx int) {
	self.Rotate(idx, -1)
	self.Pop(1)
}

//将[idx, top]索引区间内的值朝栈顶方向旋转n个位置。如果n是负数，那么实际效果是朝栈底方向旋转
func (self *luaState) Rotate(idx, n int) {
	t := self.stack.top - 1
	p := self.stack.absIndex(idx) - 1
	var m int
	if n >= 0 {
		m = t - n
	} else {
		m = p - n - 1
	}
	self.stack.reverse(p, m)
	self.stack.reverse(m+1, t)
	self.stack.reverse(p, t)
}

//将栈顶索引设置为指定值
//如果指定值小于当前栈顶索引，效果相当于弹出操作（指定值为0相当于清空栈）
//如果指定值大于当前栈顶索引，效果相当于推入多个nil值
func (self *luaState) SetTop(idx int) {
	newTop := self.stack.absIndex(idx)
	if newTop < 0 {
		panic("stack underflow!")
	}

	n := self.stack.top - newTop
	if n > 0 {
		for i := 0; i < n; i++ {
			self.stack.pop()
		}
	} else if n < 0 {
		for i := 0; i > n; i-- {
			self.stack.push(nil)
		}
	}
}

// [-?, +?, –]
// http://www.lua.org/manual/5.3/manual.html#lua_xmove
// 栈操作方法，用于在两个线程的栈之间移动元素
func (self *luaState) XMove(to LuaState, n int) {
	vals := self.stack.popN(n)
	to.(*luaState).stack.pushN(vals, n)
}