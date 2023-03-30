package example

import "grpctester/tester"

type HelloWorldTester struct {
	*tester.Tester // 使用框架能力
}

func NewHelloWorldTester(testDir string) *HelloWorldTester {
	return &HelloWorldTester{
		tester.NewTester(testDir),
	}
}

// 编写测试函数
func (t *HelloWorldTester) TestSayHello() {
	t.DoRequests("helloworld.Greeter.SayHello", // 要测试的 gRPC 方法，格式为 package.service.method
		".\\example\\protos\\helloworld.proto",                  // 要测试的 gRPC 方法所在的 proto 文件
		[]string{".\\example\\protos"},                          // 要测试的 gRPC 方法所在的 proto 文件所依赖的 import 路径
		[]string{"{\"name\":\"world\"}", "{\"name\":\"grpc\"}"}, // 要测试的 gRPC 方法的请求参数，可以写多个，调用接口次数为 requestsTimes * len(requests)
	)
}
