package plugin

/*
#include <stdlib.h>
*/
import "C"
import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
	"unsafe"

	"github.com/fluent/fluent-bit-go/input"
)

var theName string
var theDesc string
var theInput Input
var once sync.Once
var msgCh chan Message
var runCtx context.Context
var runCancel context.CancelFunc
var theWriter *queueWriter

type Message struct {
	Time time.Time
	Data map[string]string
}

//export FLBPluginRegister
func FLBPluginRegister(def unsafe.Pointer) int {
	return input.FLBPluginRegister(def, theName, theDesc)
}

type ConfigLoader interface {
	Load(key string) string
}

type configLoader struct {
	ptr unsafe.Pointer
}

func (f *configLoader) Load(key string) string {
	return input.FLBPluginConfigKey(f.ptr, key)
}

type Input interface {
	Setup(ctx context.Context, conf ConfigLoader) error
	Run(ctx context.Context, w Writer) error
}

func RegisterInput(name, desc string, in Input) {
	theName = name
	theDesc = desc
	theInput = in
}

//export FLBPluginInit
func FLBPluginInit(ptr unsafe.Pointer) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf := &configLoader{ptr: ptr}
	if err := theInput.Setup(ctx, conf); err != nil {
		fmt.Fprintf(os.Stderr, "init: %s\n", err)
		return input.FLB_ERROR
	}

	return input.FLB_OK
}

//export FLBPluginInputCallback
func FLBPluginInputCallback(data *unsafe.Pointer, size *C.size_t) int {
	var err error
	once.Do(func() {
		runCtx, runCancel = context.WithCancel(context.Background())
		theWriter = &queueWriter{ch: make(chan Message, 1)}
		err = theInput.Run(runCtx, theWriter)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "run: %s\n", err)
		return input.FLB_ERROR
	}

	select {
	case msg := <-theWriter.ch:
		t := input.FLBTime{Time: msg.Time}
		b, err := input.NewEncoder().Encode([]any{t, msg.Data})
		if err != nil {
			fmt.Fprintf(os.Stderr, "encode: %s\n", err)
			return input.FLB_ERROR
		}

		*data = C.CBytes(b)
		*size = C.size_t(len(b))
	case <-runCtx.Done():
		if err := runCtx.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "run: %s\n", err)
			return input.FLB_ERROR
		}

		return input.FLB_OK
	}

	return input.FLB_OK
}

type Writer interface {
	Write(ctx context.Context, t time.Time, data map[string]string) error
}

type queueWriter struct {
	ch chan Message
}

func (w *queueWriter) Write(ctx context.Context, t time.Time, data map[string]string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	w.ch <- Message{Time: t, Data: data}
	return nil
}

//export FLBPluginExit
func FLBPluginExit() int {
	runCancel()
	close(msgCh)
	return input.FLB_OK
}
