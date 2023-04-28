package binchunk

//LUA 二进制chunk
type binaryChunk struct {
	header							//头部
	sizeUpvalues 	byte			//主函数upvalue数量
	mainFunc 		*Prototype		//主函数原型
}


//header信息用于检测该二进制chunk是否与本机lua虚拟机匹配，如果不匹配，则拒绝加载该二进制chunk
type header struct {
	signature 		[4]byte		//签名
	version 		byte		//版本号
	format 			byte		//格式号
	luacData 		[6]byte		//校验
	cintSize		byte		//cint 数据类型在二进制chunk里占用的字节数
	sizetSize 		byte		//size_t 数据类型在二进制chunk里占用的字节数
	instructionSize byte		//lua虚拟机指令 数据类型在二进制chunk里占用的字节数
	luaIntegerSize	byte		//lua整数 数据类型在二进制chunk里占用的字节数
	luaNumberSize	byte		//lua浮点数 数据类型在二进制chunk里占用的字节数
	luacInt			byte		//检查该二进制chunk的大小端方式与本机是否匹配
	luacNum			byte		//检查该二进制chunk浮点数格式和本机是否匹配
}

//函数原型
type Prototype struct {
	Source				string			//源文件名，只有主函数原型才有值，其他嵌套的函数原型该字段存放空字符串
	LineDefined			uint32			//起始行号
	LastLineDefined		uint32			//结束行号
	NumParams			byte			//固定参数个数
	IsVararg			byte			//是否Vararg函数，即是否有变长参数
	MaxStackSize		byte			//寄存器数量
	Code				[]uint32		//指令表，每条指令占4个字节
	Constants			[]interface{}	//常量表，用于存放lua代码里出现的字面量，包括nil、布尔值、整数、浮点数和字符串五种
	Upvalues			[]Upvalue		//Upvalue表
	Protos				[]*Prototype	//子函数原型表
	//调试信息
	//行号表、局部变量表和Upvalue名列表，这三个表里存储的都是调试信息，对于程序的执行不必要。如果在编译lua脚本时指定了“-s”选项，lua编译器会在二进制chunk中巴这三个表清空。
	LineInfo			[]uint32		//行号表，行号表中的行号和指令表中的指令一一对应，分别记录每条指令在源代码中对应的行号。
	LocVars				[]LocVar		//局部变量表
	UpvalueNames		[]string		//Upvalue名列表，该列表中的元素和前面Upvalue表中的元素一一对应，分别记录每个Upvalue在源代码中的名字。
} 

//Upvalue
type Upvalue struct {
	Instack		byte
	Idx			byte
}

//局部变量
type LocVar struct {
	VarName		string			//变量名
	StartPC		uint32			//开始指令索引
	EndPC		uint32			//结束指令索引
}

const (
	LUA_SIGNATURE		= "\x1bLua"
	LUAC_VERSION		= 0x53
	LUAC_FORMAT			= 0
	LUAC_DATA			= "\x19\x93\r\n\x1a\n"
	CINT_SIZE			= 4
	CSIZET_SIZE			= 8
	INSTRUCTION_SIZE	= 4
	LUA_INTEGER_SIZE	= 8
	LUA_NUMBER_SIZE		= 8
	LUAC_INT			= 0x5678
	LUAC_NUM			= 370.5
)

const (
	TAG_NIL			= 0x00
	TAG_BOOLEAN		= 0x01
	TAG_NUMBER		= 0x03
	TAG_INTEGER		= 0x13
	TAG_SHORT_STR	= 0x04
	TAG_LONG_STR	= 0x14
)

func IsBinaryChunk(data []byte) bool {
	return len(data) > 4 &&
		string(data[:4]) == LUA_SIGNATURE
}

//简化版lua反编译器
func Undump(data []byte) *Prototype {
	reader := &reader{data}
	reader.checkHeader()			//校验头部
	reader.readByte()				//跳过Upvalue数量
	return reader.readProto("")		//读取函数原型
}