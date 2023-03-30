package tester

import (
	"context"
	"errors"
	"fmt"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"google.golang.org/grpc"
	"os"
	"path"
	"sync"
	"time"
)

type Tester struct {
	cockroachDB *CockroachDB
	c           *Config
	stubs       []*grpcdynamic.Stub
	conns       []*grpc.ClientConn
	workers     []*Worker
}

func NewTester(testDir string) *Tester {
	return &Tester{
		c: NewConfig(testDir),
	}
}

func (t *Tester) WithCockroachDB() *Tester {
	t.c.ParseCockroachDBConfig()
	t.cockroachDB = NewCockroachDB(t.c.CockroachDBConfig)

	return t
}

func (t *Tester) newClientConn() *grpc.ClientConn {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(t.c.creds))

	ctx, _ := context.WithTimeout(context.Background(), t.c.Timeout*time.Second)
	conn, err := grpc.DialContext(ctx, t.c.Host, opts...)
	if err != nil {
		panic(err)
	}
	return conn
}

func (t *Tester) openClientConns() {
	t.c.ConnsNum = t.c.Concurrency / 100
	fmt.Printf("连接数：%d\n", t.c.ConnsNum)
	t.conns = make([]*grpc.ClientConn, t.c.ConnsNum)
	for i := 0; i < t.c.ConnsNum; i++ {
		t.conns[i] = t.newClientConn()
	}
}

func (t *Tester) closeClientConns() {
	for i := 0; i < t.c.ConnsNum; i++ {
		t.conns[i].Close()
		t.conns[i] = nil
	}
}

func (t *Tester) newClientStubs() {
	t.stubs = make([]*grpcdynamic.Stub, t.c.ConnsNum)
	for i := 0; i < t.c.ConnsNum; i++ {
		stub := grpcdynamic.NewStub(t.conns[i])
		t.stubs[i] = &stub
	}
}

func transToAbsPath(p string) string {
	if p == "" {
		panic(errors.New("path is empty"))
	}
	currentPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return path.Join(currentPath, p)
}

func (t *Tester) DoRequests(method, proto string, importPaths, requests []string) {
	proto = transToAbsPath(proto)
	for i := 0; i < len(importPaths); i++ {
		importPaths[i] = transToAbsPath(importPaths[i])
	}
	mtd := getMethodDescFromProto(method, proto, importPaths)

	t.openClientConns()
	defer t.Clean()
	t.newClientStubs()

	// 有多少种请求，就创建多少个worker，每个worker携带一种请求
	t.workers = make([]*Worker, len(requests))
	for i := 0; i < len(requests); i++ {
		stub := t.stubs[i%t.c.ConnsNum]
		t.workers[i] = NewWorker(requests[i], mtd, stub)
	}
	ctx, _ := context.WithTimeout(context.Background(), t.c.Timeout*time.Second)

	startTime := time.Now()

	var ch = make(chan struct{}, t.c.Concurrency)
	var wg sync.WaitGroup
	for i := 0; i < t.c.RequestsTimes; i++ {
		for j := 0; j < len(requests); j++ {
			ch <- struct{}{}
			wg.Add(1)
			go t.workers[j].makeUnaryRequest(&ctx, &wg, ch)
		}
	}
	wg.Wait()

	elapsed := time.Since(startTime)
	// 打印消耗的秒数
	fmt.Printf("消耗的秒数：%f\n", elapsed.Seconds())
}

func (t *Tester) Clean() {
	t.closeClientConns()
	if t.cockroachDB != nil {
		t.cockroachDB.close()
	}
}
