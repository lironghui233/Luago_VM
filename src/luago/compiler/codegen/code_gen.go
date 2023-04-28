//定义GenProto()函数，把我们的中间结构和处理过程隐藏起来

package codegen

import . "luago/binchunk"
import . "luago/compiler/ast"

//代码生成器
func GenProto(chunk *Block) *Prototype {
	fd := &FuncDefExp{
		IsVararg: true,
		Block:    chunk,
	}

	fi := newFuncInfo(nil, fd)
	fi.addLocVar("_ENV")
	cgFuncDefExp(fi, fd, 0)
	return toProto(fi.subFuncs[0])
}
