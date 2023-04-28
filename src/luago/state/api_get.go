package state

import . "luago/api"

//创建一个空的Lua表，将其推入栈顶。两个参数用于指定数组部分和哈希表部分的初始容量，预分配足够的空间以避免后续对表进行频繁扩容。
func (self *luaState) CreateTable(nArr, nRec int) {
	t := newLuaTable(nArr, nRec)
	self.stack.push(t)
}

//如果无法估计表的用法和容量。
func (self *luaState) NewTable() {
	self.CreateTable(0, 0)
}

//根据键（从栈顶弹出）从表（索引由参数指定）里取值，然后把值推入栈顶并返回值的类型
func (self *luaState) GetTable(idx int) LuaType {
	t := self.stack.get(idx)
	k := self.stack.pop()
	return self.getTable(t, k, false)
}

//根据键（值由参数指定）从表（索引由参数指定）里取值，然后把值推入栈顶并返回值的类型。该方法是专门用来从记录表（map）里获取字段
func (self *luaState) GetField(idx int, k string) LuaType {
	t := self.stack.get(idx)
	return self.getTable(t, k, false)
}

//根据键（值由参数指定）从表（索引由参数指定）里取值，然后把值推入栈顶并返回值的类型。该方法是专门用来从数组（arr）里获取数组元素
func (self *luaState) GetI(idx int, i int64) LuaType {
	t := self.stack.get(idx)
	return self.getTable(t, i, false)
}

func (self *luaState) RawGet(idx int) LuaType {
	t := self.stack.get(idx)
	k := self.stack.pop()
	return self.getTable(t, k, true)
}

func (self *luaState) RawGetI(idx int, i int64) LuaType {
	t := self.stack.get(idx)
	return self.getTable(t, i, true)
}

func (self *luaState) getTable(t, k luaValue, raw bool) LuaType {	//raw表示是否忽略元方法
	if tbl, ok := t.(*luaTable); ok {
		v := tbl.get(k)
		if raw || v != nil || !tbl.hasMetafield("__index") {
			self.stack.push(v)
			return typeOf(v)
		}
	}

	if !raw {
		if mf := getMetafield(t, "__index", self); mf != nil {
			switch x := mf.(type) {
			case *luaTable:
				return self.getTable(x, k, false)
			case *closure:
				self.stack.push(mf)
				self.stack.push(t)
				self.stack.push(k)
				self.Call(2, 1)
				v := self.stack.get(-1)
				return typeOf(v)
			}
		}
	}

	panic("not a table!") 
}

//从全局环境（_G，其实就是注册表中的一个普通的lua_table，注册表也是普通的lua_table）中根据key获取value并推入栈顶，返回value的类型
func (self *luaState) GetGlobal(name string) LuaType {
	t := self.registry.get(LUA_RIDX_GLOBALS)
	return self.getTable(t, name, false)
}

//看指定索引处的值是否有元表，如果有，则把元表推入栈顶并返回true，否则栈的状态不改变，返回false
func (self *luaState) GetMetatable(idx int) bool {
	val := self.stack.get(idx)

	if mt := getMetatable(val, self); mt != nil {
		self.stack.push(mt)
		return true
	} else {
		return false
	}
}