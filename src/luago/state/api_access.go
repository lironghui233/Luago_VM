//access系列方法用于从栈里获取信息。
//access系列方法基本上仅使用索引访问栈里存储的信息，不会改变栈的状态

package state

import "fmt"
import . "luago/api"

func (self *luaState) RawLen(idx int) uint {
	val := self.stack.get(idx)
	switch x := val.(type) {
	case string:
		return uint(len(x))
	case *luaTable:
		return uint(x.len())
	default:
		return 0
	}
}

//把给定Lua类型转换成对应的字符串表示
func (self *luaState) TypeName(tp LuaType) string {
	switch tp {
	case LUA_TNONE:				return "no value"
	case LUA_TNIL:				return "nil"
	case LUA_TBOOLEAN:			return "boolean"
	case LUA_TNUMBER:			return "number"
	case LUA_TSTRING:			return "string"
	case LUA_TTABLE:			return "table"
	case LUA_TFUNCTION:			return "function"
	case LUA_TTHREAD:			return "thread"
	default:					return "userdata"
	}
}

//根据索引返回值的类型，如果索引无效，然会LUA_TNONE
func (self *luaState) Type(idx int) LuaType {
	if self.stack.isValid(idx) {
		val := self.stack.get(idx)
		return typeOf(val)
	}
	return LUA_TNONE
}

func (self *luaState) IsNone(idx int) bool {
	return self.Type(idx) == LUA_TNONE
}

func (self *luaState) IsNil(idx int) bool {
	return self.Type(idx) == LUA_TNIL
}

func (self *luaState) IsNoneOrNil(idx int) bool {
	return self.Type(idx) <= LUA_TNIL
}

func (self *luaState) IsBoolean(idx int) bool {
	return self.Type(idx) == LUA_TBOOLEAN
}

func (self *luaState) IsString(idx int) bool {
	t := self.Type(idx)
	return t == LUA_TSTRING || t == LUA_TNUMBER
}

func (self *luaState) IsNumber(idx int) bool {
	_, ok := self.ToNumberX(idx)
	return ok
}

func (self *luaState) IsInteger(idx int) bool {
	val := self.stack.get(idx)
	_, ok := val.(int64)
	return ok
}

func (self *luaState) ToBoolean(idx int) bool {
	val := self.stack.get(idx)
	return convertToBoolean(val)
}

func (self *luaState) IsTable(idx int) bool {
	return self.Type(idx) == LUA_TTABLE
}

func (self *luaState) IsFunction(idx int) bool {
	return self.Type(idx) == LUA_TFUNCTION
}

func (self *luaState) IsThread(idx int) bool {
	return self.Type(idx) == LUA_TTHREAD
}

func (self *luaState) IsGoFunction(idx int) bool {
	val := self.stack.get(idx)
	if c, ok := val.(*closure); ok {
		return c.goFunc != nil
	}
	return false
}

func (self *luaState) ToNumber(idx int) float64 {
	n, _ := self.ToNumberX(idx)
	return n
}

func (self *luaState) ToNumberX(idx int) (float64, bool) {
	val := self.stack.get(idx)
	return convertToFloat(val)
}

func (self *luaState) ToInteger(idx int) int64 {
	i, _ := self.ToIntegerX(idx)
	return i
}

func (self *luaState) ToIntegerX(idx int) (int64, bool) {
	val := self.stack.get(idx)
	return convertToInteger(val)
}

func (self *luaState) ToString(idx int) string {
	s, _ := self.ToStringX(idx)
	return s
}

func (self *luaState) ToStringX(idx int) (string, bool) {
	val := self.stack.get(idx)
	switch x := val.(type) {
	case string:
		return x, true
	case int64, float64:
		s := fmt.Sprintf("%v", x)
		self.stack.set(idx, s)	//注意这里会修改栈！
		return s, true
	default:	
		return "", false	
	}
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tocfunction
func (self *luaState) ToGoFunction(idx int) GoFunction {
	val := self.stack.get(idx)
	if c, ok := val.(*closure); ok {
		return c.goFunc
	}
	return nil
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_topointer
func (self *luaState) ToPointer(idx int) interface{} {
	// todo
	return self.stack.get(idx)
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tothread
// 把指定索引处的值转换为线程并返回，如果值不是线程，就返回nil
func (self *luaState) ToThread(idx int) LuaState {
	val := self.stack.get(idx)
	if val != nil {
		if ls, ok := val.(*luaState); ok {
			return ls
		}
	}
	return nil
}