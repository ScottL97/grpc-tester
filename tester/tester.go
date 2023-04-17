package tester

import (
	"context"
	"fmt"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type Tester struct {
	CockroachDB *CockroachDB
	C           *Config
	stubs       []*grpcdynamic.Stub
	conns       []*grpc.ClientConn
	workers     []*Worker
}

func NewTester(testDir string) *Tester {
	return &Tester{
		C: NewConfig(testDir),
	}
}

func (t *Tester) WithCockroachDB() *Tester {
	t.C.ParseCockroachDBConfig()
	t.CockroachDB = NewCockroachDB(t.C.CockroachDBConfig)

	return t
}

func (t *Tester) newClientConn() *grpc.ClientConn {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(t.C.creds))

	ctx, _ := context.WithTimeout(context.Background(), t.C.Timeout*time.Second)
	conn, err := grpc.DialContext(ctx, t.C.Host, opts...)
	if err != nil {
		panic(err)
	}
	return conn
}

func (t *Tester) openClientConns() {
	t.C.ConnsNum = t.C.Concurrency/100 + 1
	fmt.Printf("连接数：%d\n", t.C.ConnsNum)
	t.conns = make([]*grpc.ClientConn, t.C.ConnsNum)
	for i := 0; i < t.C.ConnsNum; i++ {
		t.conns[i] = t.newClientConn()
	}
}

func (t *Tester) closeClientConns() {
	for i := 0; i < t.C.ConnsNum; i++ {
		t.conns[i].Close()
		t.conns[i] = nil
	}
}

func (t *Tester) newClientStubs() {
	t.stubs = make([]*grpcdynamic.Stub, t.C.ConnsNum)
	for i := 0; i < t.C.ConnsNum; i++ {
		stub := grpcdynamic.NewStub(t.conns[i])
		t.stubs[i] = &stub
	}
}

func (t *Tester) DoRequestsSequentially(method, proto string, importPaths, requests []string) {
	mtd := getMethodDescFromProto(method, proto, importPaths)

	t.openClientConns()
	defer t.Clean()
	t.newClientStubs()

	// 有多少种请求，就创建多少个worker，每个worker携带一种请求
	t.workers = make([]*Worker, len(requests))
	for i := 0; i < len(requests); i++ {
		stub := t.stubs[i%t.C.ConnsNum]
		t.workers[i] = NewWorker(requests[i], mtd, stub)
	}

	for i := 0; i < t.C.RequestsTimes; i++ {
		for j := 0; j < len(requests); j++ {
			ctx, _ := context.WithTimeout(context.Background(), t.C.Timeout*time.Second)
			t.workers[j].makeUnaryRequestSequentially(&ctx)
		}
	}
}

func (t *Tester) DoRequests(method, proto string, importPaths, requests []string) {
	mtd := getMethodDescFromProto(method, proto, importPaths)

	t.openClientConns()
	defer t.Clean()
	t.newClientStubs()

	// 有多少种请求，就创建多少个worker，每个worker携带一种请求
	t.workers = make([]*Worker, len(requests))
	for i := 0; i < len(requests); i++ {
		stub := t.stubs[i%t.C.ConnsNum]
		t.workers[i] = NewWorker(requests[i], mtd, stub)
	}

	var ch = make(chan struct{}, t.C.Concurrency)
	var wg sync.WaitGroup
	for i := 0; i < t.C.RequestsTimes; i++ {
		for j := 0; j < len(requests); j++ {
			ch <- struct{}{}
			wg.Add(1)
			ctx, _ := context.WithTimeout(context.Background(), t.C.Timeout*time.Second)
			go t.workers[j].makeUnaryRequest(&ctx, &wg, ch)
		}
	}
	wg.Wait()
}

func CalcSecondsWrapper(f func()) func() {
	return func() {
		startTime := time.Now()

		f()

		elapsed := time.Since(startTime)
		// 打印消耗的秒数
		fmt.Printf("消耗的秒数：%f\n", elapsed.Seconds())
	}
}

func (t *Tester) Clean() {
	t.closeClientConns()
	if t.CockroachDB != nil {
		t.CockroachDB.close()
	}
}
