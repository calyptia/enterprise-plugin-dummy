package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"context"
	"fmt"
	"os"
	"plugin"
	"sync"
	"time"
	"unsafe"

	"github.com/fluent/fluent-bit-go/input"
)

var thePlugin Plugin
var once sync.Once
var ch chan message
var processCtx context.Context
var processCancel context.CancelFunc

type Plugin interface {
	Name() string
	Desc() string
	Setup(ctx context.Context, load func(key string) string) error
	Process(ctx context.Context, send func(t time.Time, data map[string]string) error) error
}

type message struct {
	Time time.Time
	Data map[string]string
}

//export FLBPluginRegister
func FLBPluginRegister(def unsafe.Pointer) int {
	f, err := plugin.Open("/data/go-dummy-plugin.so") // TODO: take plugin file from somewhere.
	if err != nil {
		fmt.Fprintf(os.Stderr, "plugin open: %v\n", err)
		return input.FLB_ERROR
	}

	sym, err := f.Lookup("Plugin")
	if err != nil {
		fmt.Fprintf(os.Stderr, "plugin lookup: %v\n", err)
		return input.FLB_ERROR
	}

	var ok bool
	thePlugin, ok = sym.(Plugin)
	if !ok {
		fmt.Fprintf(os.Stderr, "plugin type assertion: %T\n", sym)
		return input.FLB_ERROR
	}

	return input.FLBPluginRegister(def, thePlugin.Name(), thePlugin.Desc())
}

//export FLBPluginInit
func FLBPluginInit(ptr unsafe.Pointer) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := thePlugin.Setup(ctx, func(key string) string {
		return input.FLBPluginConfigKey(ptr, key)
	}); err != nil {
		fmt.Fprintf(os.Stderr, "setup: %v\n", err)
		return input.FLB_ERROR
	}

	return input.FLB_OK
}

//export FLBPluginInputCallback
func FLBPluginInputCallback(data *unsafe.Pointer, size *C.size_t) int {
	var err error

	once.Do(func() {
		processCtx, processCancel = context.WithCancel(context.Background())

		ch = make(chan message, 1)

		go func() {
			err = thePlugin.Process(processCtx, func(t time.Time, data map[string]string) error {
				ch <- message{t, data}
				return processCtx.Err()
			})
		}()
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "process: %v\n", err)
		return input.FLB_ERROR
	}

	msg := <-ch

	enc := input.NewEncoder()
	b, err := enc.Encode([]any{input.FLBTime{Time: msg.Time}, msg.Data})
	if err != nil {
		fmt.Fprintf(os.Stderr, "msgpack encode: %v\n", err)
		return input.FLB_ERROR
	}

	length := len(b)
	*data = C.CBytes(b)
	*size = C.size_t(length)

	return input.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
	processCancel()
	close(ch)
	return input.FLB_OK
}

func main() {}
