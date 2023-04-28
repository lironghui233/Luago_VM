## 安装说明

1. 安装Lua5.3.4
2. 安装go1.10.2



## 编译说明

```
cd Luago_VM

export GOPATH=$PWD

go install luago
```

然后就在/bin目录下出现可执行程序luago



## 目录说明

- bin: 生成的可执行文件luago

- pkg:

  - linux_amd64
    - luago
      - compiler: 生成的Lua编译器相关静态库

- src

  - luago

    - api：lua虚拟机的api接口

    - binchunk：Lua字节码二进制chunk相关（Lua解析器）
    - compiler：Lua编译器相关
    - number：Lua虚拟机数值运算相关
    - state：Lua虚拟机相关
    - stdlib：Lua标准库
    - vm：Lua指令集操作虚拟机相关

