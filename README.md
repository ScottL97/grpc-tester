# grpc-tester

基于 protoreflect 的 gRPC 接口通用压测框架。

## 参数

### 测试函数参数

参考`main.go`和`example`目录下的代码注释。

### 配置文件

参考`config.go`的配置注释。

### 命令行参数

```powershell
-t string
    测试目录 (default ".\\example")
```

## 示例

启动示例的 gRPC server：

```powershell
> cd .\\grpc-tester\\example\\grpcserver
> go run main.go
```

编译并执行测试：

```powershell
> cd grpc-tester
> go mod tidy
> go build -o grpc-tester.exe main.go
> grpc-tester -h
> grpc-tester -t .\\example
```
