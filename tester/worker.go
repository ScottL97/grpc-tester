package tester

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"sync"
)

func createPayloadFromJSON(json string, mtd *desc.MethodDescriptor) *dynamic.Message {
	msg := dynamic.NewMessage(mtd.GetInputType())
	err := jsonpb.UnmarshalString(json, msg)
	if err != nil {
		panic(err)
	}
	return msg
}

type Worker struct {
	msg  *dynamic.Message
	stub *grpcdynamic.Stub
	mtd  *desc.MethodDescriptor
}

func NewWorker(request string, mtd *desc.MethodDescriptor, stub *grpcdynamic.Stub) *Worker {
	fmt.Println("new worker")
	return &Worker{
		msg:  createPayloadFromJSON(request, mtd),
		stub: stub,
		mtd:  mtd,
	}
}

func (w *Worker) makeUnaryRequest(ctx *context.Context, wg *sync.WaitGroup, ch <-chan struct{}) {
	defer func() {
		<-ch
		wg.Done()
	}()
	_, err := w.stub.InvokeRpc(*ctx, w.mtd, w.msg)
	if err != nil {
		panic(err)
	}
}
