package main_test

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	oLoadEnv := main.loadEnv
	main.loadEnv = func(filename ...string) (err error) {
		os.Setenv("PORT", "8899")
		return
	}
	defer func() {
		main.loadEnv = oLoadEnv
		r := recover()
		if r != nil {
			t.Fail()
		}
	}()
	srv := main.createServer()
	time.Sleep(1 * time.Second)
	srv.Shutdown(context.TODO())
}
