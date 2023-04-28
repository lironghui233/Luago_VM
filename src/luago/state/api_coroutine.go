package state

import . "luago/api"

//协程共有4种状态：运行（running）、挂起（suspende）、正常（normal）、死亡（dead）
//任何时刻，只有一个协程处于运行状态，通过running()函数可以获取这个协程 

// [-0, +1, m]
// http://www.lua.org/manual/5.3/manual.html#lua_newthread
// lua-5.3.4/src/lstate.c#lua_newthread()
// 线程其实就是luaState结构。这个函数创建一个新的线程，把它推入栈顶，同时也作为返回值返回。新创建的线程和创建它的线程共享相同的全局变量，但是有各自的调用栈。新创建的线程处于挂起状态
func (self *luaState) NewThread() LuaState {
	t := &luaState{registry: self.registry}
	t.pushLuaStack(newLuaStack(LUA_MINSTACK, t))
	self.stack.push(t)
	return t
}

// [-?, +?, –]
// http://www.lua.org/manual/5.3/manual.html#lua_resume
// 让处于挂起状态的协程开始或回复运行状态
func (self *luaState) Resume(from LuaState, nArgs int) int {
	lsFrom := from.(*luaState)
	if lsFrom.coChan == nil {
		lsFrom.coChan = make(chan int)
	}

	if self.coChan == nil {
		// start coroutine
		self.coChan = make(chan int)
		self.coCaller = lsFrom
		//启动一个Go语言协程（Goroutine）来执行其主函数
		go func() {
			self.coStatus = self.PCall(nArgs, -1, 0)
			lsFrom.coChan <- 1
		}()
	} else {
		// resume coroutine
		if self.coStatus != LUA_YIELD { // todo
			self.stack.push("cannot resume non-suspended coroutine")
			return LUA_ERRRUN
		}
		self.coStatus = LUA_OK
		self.coChan <- 1
	}

	<-lsFrom.coChan // wait coroutine to finish or yield
	return self.coStatus
}

// [-?, +?, e]
// http://www.lua.org/manual/5.3/manual.html#lua_yield
// 挂起自己
func (self *luaState) Yield(nResults int) int {
	if self.coCaller == nil { // todo
		panic("attempt to yield from outside a coroutine")
	}
	self.coStatus = LUA_YIELD
	self.coCaller.coChan <- 1
	<-self.coChan
	return self.GetTop()
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isyieldable
func (self *luaState) IsYieldable() bool {
	if self.isMainThread() {
		return false
	}
	return self.coStatus != LUA_YIELD // todo
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_status
// lua-5.3.4/src/lapi.c#lua_status()
//获取任意协程的状态
func (self *luaState) Status() int {
	return self.coStatus
}

// debug
func (self *luaState) GetStack() bool {
	return self.stack.prev != nil
}
