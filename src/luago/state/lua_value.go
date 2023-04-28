package state

import "fmt"
import . "luago/api"
import "luago/number"

//用于表示lua值
type luaValue interface{}

func typeOf(val luaValue) LuaType {
	switch val.(type) {
	case nil:				return LUA_TNIL
	case bool:				return LUA_TBOOLEAN 
	case int64, float64:	return LUA_TNUMBER
	case string:			return LUA_TSTRING
	case *luaTable:			return LUA_TTABLE
	case *closure:			return LUA_TFUNCTION
	case *luaState: 		return LUA_TTHREAD
	default :				panic("todo!")
	}
}

func convertToBoolean(val luaValue) bool {
	switch x := val.(type) {
	case nil:		return false
	case bool:		return x
	default:		return true	
	}
}

func convertToFloat(val luaValue) (float64, bool) {
	switch x := val.(type) {
	case float64:		return x, true
	case int64:			return float64(x), true
	case string:		return number.ParseFloat(x)
	default:			return 0, false
	}
}

func convertToInteger(val luaValue) (int64, bool) {
	switch x := val.(type) {
	case int64:			return x, true
	case float64:		return number.FloatToInteger(x)
	case string:		return _stringToInteger(x)
	default:			return 0, false
	}
}

func _stringToInteger(s string) (int64, bool) {
	if i, ok := number.ParseInteger(s); ok {
		return i, true
	}
	if f, ok := number.ParseFloat(s); ok {
		return number.FloatToInteger(f)
	}
	return 0, false
}

/* metatable */

func getMetatable(val luaValue, ls *luaState) *luaTable {
	if t, ok := val.(*luaTable); ok {
		return t.metatable
	}
	key := fmt.Sprintf("_MT%d", typeOf(val))
	if mt := ls.registry.get(key); mt != nil {
		return mt.(*luaTable)
	}
	return nil
}

func setMetatable(val luaValue, mt *luaTable, ls *luaState) {
	if t, ok := val.(*luaTable); ok {
		t.metatable = mt
		return
	}
	//根据变量类型把元表存储在注册表里，这样就达到了其他类型共享元表的目的
	key := fmt.Sprintf("_MT%d", typeOf(val))
	ls.registry.put(key, mt)
}

func callMetamethod(a, b luaValue, mmName string, ls *luaState) (luaValue, bool) {
	var mm luaValue
	if mm = getMetafield(a, mmName, ls); mm == nil {
		if mm = getMetafield(b, mmName, ls); mm == nil {
			return nil, false
		}
	}

	ls.stack.check(4)
	ls.stack.push(mm)
	ls.stack.push(a)
	ls.stack.push(b)
	ls.Call(2, 1)
	return ls.stack.pop(), true
}

func getMetafield(val luaValue, fieldName string, ls *luaState) luaValue {
	if mt := getMetatable(val, ls); mt != nil {
		return mt.get(fieldName)
	}
	return nil
}