/*
 31       22       13       5    0
  +-------+^------+-^-----+-^-----
  |b=9bits |c=9bits |a=8bits|op=6|
  +-------+^------+-^-----+-^-----
  |    bx=18bits    |a=8bits|op=6|
  +-------+^------+-^-----+-^-----
  |   sbx=18bits    |a=8bits|op=6|
  +-------+^------+-^-----+-^-----
  |    ax=26bits            |op=6|
  +-------+^------+-^-----+-^-----
 31      23      15       7      0
*/

package vm

import "luago/api"

const MAXARG_Bx = 1<<18 - 1			// 2^18 - 1 = 262143
const MAXARG_sBx = MAXARG_Bx >> 1		// 262143 / 2 = 131071

//32bit表示一个指令
type Instruction uint32

//返回指令的操作码名字
func (self Instruction) OpName() string {
	return opcodes[self.Opcode()].name
}

//返回指令的编码模式
func (self Instruction) OpMode() byte {
	return opcodes[self.Opcode()].opMode
}

//返回指令的操作数B的使用模式
func (self Instruction) BMode() byte {
	return opcodes[self.Opcode()].argBMode
}

//返回指令的操作数C的使用模式
func (self Instruction) CMode() byte {
	return opcodes[self.Opcode()].argCMode
}

//从指令中提取操作码
func (self Instruction) Opcode() int {
	return int(self & 0x3F)
}

//从iABC模式指令中提取参数
func (self Instruction) ABC() (a, b, c int) {
	a = int(self >> 6 & 0xFF)
	c = int(self >> 14 & 0x1FF)
	b = int(self >> 23 & 0x1FF)
	return 
}

//从iABx模式指令中提取参数
func (self Instruction) ABx() (a, bx int) {
	a = int(self >> 6 & 0xFF)
	bx = int(self >> 14)
	return 
}

//从iAsBx模式指令中提取参数
func (self Instruction) AsBx() (a, sbx int) {
	a, bx := self.ABx()
	return a, bx - MAXARG_sBx
}

//从iAx模式指令中提取参数
func (self Instruction) Ax() int {
	return int(self >> 6)
}

func (self Instruction) Execute(vm api.LuaVM) {
	action := opcodes[self.Opcode()].action
	if action != nil {
		action(self, vm)
	} else {
		panic(self.OpName())
	}
}