package main

import (
	"context"
	"time"

	"github.com/calyptia/enterprise-plugin-dummy/plugin"
)

func init() {
	plugin.RegisterInput("go-dummy-plugin", "Dummy golang plugin for testing", &dummyPlugin{})
}

type dummyPlugin struct {
	foo string
}

func (plug *dummyPlugin) Setup(ctx context.Context, conf plugin.ConfigLoader) error {
	plug.foo = conf.Load("foo")
	return nil
}

func (plug *dummyPlugin) Run(ctx context.Context, w plugin.Writer) error {
	for i := 0; i < 10; i++ {
		data := map[string]string{
			"message": "hello from go-dummy-plugin",
			"foo":     plug.foo,
		}
		if err := w.Write(ctx, time.Now(), data); err != nil {
			return err
		}

		time.Sleep(time.Second)
	}

	return nil
}

func main() {}
