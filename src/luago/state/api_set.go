package state

import . "luago/api"

//把键值对写入表，其中键和值从栈里弹出，表则位于指定索引处
func (self *luaState) SetTable(idx int) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	k := self.stack.pop()
	self.setTable(t, k, v, false)
}

//把键值对写入表，其中键由参数传入的字符串，值从栈里弹出，表则位于指定索引处。该方法是专门用来从记录表（map）里写入字段
func (self *luaState) SetField(idx int, k string) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	self.setTable(t, k, v, false)
}

//把键值对写入表，其中键由参数传入的数字，值从栈里弹出，表则位于指定索引处。该方法是专门用来从数组（arr）里写入字段
func (self *luaState) SetI(idx int, i int64) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	self.setTable(t, i, v, false)
}

func (self *luaState) RawSet(idx int) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	k := self.stack.pop()
	self.setTable(t, k, v, true)
}

func (self *luaState) RawSetI(idx int, i int64) {
	t := self.stack.get(idx)
	v := self.stack.pop()
	self.setTable(t, i, v, true)
}

func (self *luaState) setTable(t, k, v luaValue, raw bool) {	//raw表示是否忽略元方法
	if tbl, ok := t.(*luaTable); ok {
		tbl.put(k, v)
		return
	}

	if tbl, ok := t.(*luaTable); ok {
		if raw || tbl.get(k) != nil || !tbl.hasMetafield("__newindex") {
			tbl.put(k, v)
			return
		}
	}

	if !raw {
		if mf := getMetafield(t, "__newindex", self); mf != nil {
			switch x := mf.(type) {
			case *luaTable:
				self.setTable(x, k, v, false)
				return
			case *closure:
				self.stack.push(mf)
				self.stack.push(t)
				self.stack.push(k)
				self.stack.push(v)
				self.Call(3, 0)
				return
			}
		}
	}

	panic("not a table!")
}

//往全局环境中写入一个值，其中字段名由参数指定，值从栈顶弹出
func (self *luaState) SetGlobal(name string) {
	t := self.registry.get(LUA_RIDX_GLOBALS)
	v := self.stack.pop()
	self.setTable(t, name, v, false)
}

//专门用于给全局环境注册Go函数，该方法仅操作全局环境，字段名和Go函数从参数传入，不改变Lua栈的状态
func (self *luaState) Register(name string, f GoFunction) {
	self.PushGoFunction(f)
	self.SetGlobal(name)
}

//从栈顶弹出一个表，然后把指定索引处值的元表设置成该表
func (self *luaState) SetMetatable(idx int) {
	val := self.stack.get(idx)
	mtVal := self.stack.pop()

	if mtVal == nil {
		setMetatable(val, nil, self)
	} else if mt, ok := mtVal.(*luaTable); ok {
		setMetatable(val, mt, self)
	} else {
		panic("table expected!") // todo
	}
}