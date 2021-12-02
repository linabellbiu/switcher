# switcher


#### switcher诞生是为了解决proto转dto繁琐的编码工作,解放工作量.

<img src="https://big-c.oss-cn-hangzhou.aliyuncs.com/cms/img/2021/12/03/axG8D8Zi5X1638466008241-1638466008241.gif" alt="图片" width="510" height="333" align="bottom" />

### 安装
`go get github.com/wangxudong123/switcher`

### 使用

#### .proto文件必须要在项目下才能执行
#### switcher会搜索当前的项目下的.proto文件找到有注解的文件并执行

#### 在proto文件中头部写入包的注解

```protobuf
//@switcher protoGoSrc [github.com/wangxudong123/switcher/example/proto]
//@switcher out ./example/belong.go
syntax = "proto3";

package user;

```

#### 在`message`添加转换注解
```protobuf
//@switcher struct HelloWorldDto
message HelloWorld {
    string  field1 = 1;
    string  field2 = 2;
}
```
#### 生成代码
在该项目的根路径执行命令  
> switcher make  

生成文件在 `./example/belong.go`

### 参数
对以上注解解释
- struct 转换后的结构体名称 
- out  输出到文件路径(当前项目根路径的相对路径)
- protoGoSrc 当前`.proto`文件执行protoc生成的go代码包路径(需要拿到这个路径在生成的代码中导入这个包)

### 不支持
 - 不支持 `enum`
 - 不支持 `reserved`
 - 不支持 类型嵌套
 - 不支持 类型引用(下个迭代支持)