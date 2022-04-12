package main

import (
	"context"
	"time"
)

var Plugin = dummy{}

type dummy struct {
	foo string
}

func (dummy) Name() string { return "go-dummy-plugin" }

func (dummy) Desc() string { return "dummy golang plugin" }

func (dummy *dummy) Setup(ctx context.Context, load func(string) string) error {
	dummy.foo = load("foo")
	return nil
}

func (dummy *dummy) Process(ctx context.Context, send func(time.Time, map[string]string) error) error {
	for i := 0; i < 10; i++ {
		if err := send(time.Now(), map[string]string{
			"now": time.Now().Format(time.RFC3339),
			"foo": dummy.foo,
		}); err != nil {
			return err
		}

		time.Sleep(time.Second)
	}

	return nil
}
