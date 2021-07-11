package main

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	oLoadEnv := loadEnv
	loadEnv = func(filename ...string) (err error) {
		os.Setenv("PORT", "8899")
		return
	}
	defer func() {
		loadEnv = oLoadEnv
		r := recover()
		if r != nil {
			t.Fail()
		}
	}()
	srv := createServer()
	time.Sleep(1 * time.Second)
	srv.Shutdown(context.TODO())
}
