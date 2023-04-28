package state

import . "luago/api"

type luaState struct {
	stack 	*luaStack					//Lua栈		
	registry *luaTable					//注册表
	
	/* coroutine */
	coStatus int						//协程的状态（运行running、挂起suspended、正常normal、死亡dead）
	coCaller *luaState
	coChan   chan int					//两个线程通过彼此的coChan字段来相互合作
}

//创建luaState实例
func New() *luaState {
	ls := &luaState{}

	registry := newLuaTable(8, 0)						//注册表
	registry.put(LUA_RIDX_MAINTHREAD, ls)				//注册表里的 主线程索引
	registry.put(LUA_RIDX_GLOBALS, newLuaTable(0, 20))	//注册表里的 全局环境索引

	ls.registry = registry
	ls.pushLuaStack(newLuaStack(LUA_MINSTACK, ls))
	return ls
}

//判断线程是否为主线程（通过上面New()函数创建的线程就是主线程，主线程在注册表的索引是1）
func (self *luaState) isMainThread() bool {
	return self.registry.get(LUA_RIDX_MAINTHREAD) == self
}

func (self *luaState) pushLuaStack(stack *luaStack) {
	stack.prev = self.stack
	self.stack = stack
}

func (self *luaState) popLuaStack() {
	stack := self.stack
	self.stack = stack.prev
	stack.prev = nil
}
