package main

import (
	"flag"
	"grpctester/example"
)

var testDir string

func init() {
	flag.StringVar(&testDir, "t", ".\\example", "测试目录")
}

func main() {
	flag.Parse()

	// 创建测试对象，调用测试函数
	example.NewHelloWorldTester(testDir).TestSayHello()
}
