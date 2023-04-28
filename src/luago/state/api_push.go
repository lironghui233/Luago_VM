//push系列方法用于将lua值从外部推入栈顶

package state

import "fmt"
import . "luago/api"

func (self *luaState) PushNil()					{ self.stack.push(nil) }
func (self *luaState) PushBoolean(b bool)		{ self.stack.push(b) }
func (self *luaState) PushInteger(n int64)		{ self.stack.push(n) }
func (self *luaState) PushNumber(n float64)		{ self.stack.push(n) }
func (self *luaState) PushString(s string)		{ self.stack.push(s) }

// [-0, +1, e]
// http://www.lua.org/manual/5.3/manual.html#lua_pushfstring
func (self *luaState) PushFString(fmtStr string, a ...interface{}) {
	str := fmt.Sprintf(fmtStr, a...)
	self.stack.push(str)
}

func (self *luaState) PushGoFunction(f GoFunction) { 
	self.stack.push(newGoClosure(f, 0)) 
}

//从注册表中获取全局环境（_G），然后把全局环境（_G）放入Lua栈
func (self *luaState) PushGlobalTable() {
	global := self.registry.get(LUA_RIDX_GLOBALS)
	self.stack.push(global)
}

func (self *luaState) PushGoClosure(f GoFunction, n int) {
	closure := newGoClosure(f, n)
	for i := n; i > 0; i-- {
		val := self.stack.pop()
		closure.upvals[i-1] = &upvalue{&val}
	}
	self.stack.push(closure)
}

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_pushthread
// 把线程（线程就是luaState类型）推入栈顶，返回的布尔值表示线程是否为主线程
func (self *luaState) PushThread() bool {
	self.stack.push(self)
	return self.isMainThread()
}
