//go:build integrationtesting
// +build integrationtesting

package example

import (
	"context"
	"fmt"

	"syscall"
	"testing"
	"time"

	bomber "github.com/jyouturner/bomber/pkg/bombers"
	bombers "github.com/jyouturner/bomber/pkg/bombers"
)

func TestBombUrl(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(time.Millisecond*1000))
	defer cancel()
	type args struct {
		ctx context.Context
		c   bomber.HttpConfig
		n   int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// test localhost
		{
			name: "localhost http load testing",
			args: args{
				ctx: ctx,
				c: bombers.HttpConfig{
					Url:    "http://localhost:3000/hello",
					Method: "POST",
					Body:   "",
					Headers: map[string]string{
						"Content-Type": "application/json",
					}},
				n: 100,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := BombUrl(tt.args.ctx, tt.args.c, tt.args.n); (err != nil) != tt.wantErr {
				t.Errorf("BombUrl() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBombUrl_Interrupt(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(time.Millisecond*3000))
	defer cancel()
	type args struct {
		ctx context.Context
		c   bomber.HttpConfig
		n   int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// run long process and send a signal to kill
		{
			name: "localhost http load testing",
			args: args{
				ctx: ctx,
				c: bombers.HttpConfig{
					Url:    "http://localhost:3000/hello",
					Method: "POST",
					Body:   "",
					Headers: map[string]string{
						"Content-Type": "application/json",
					}},
				n: 10000000,
			},
			wantErr: false,
		},
	}
	go func() {

		time.Sleep(100 * time.Millisecond)
		fmt.Println("sending signal...")
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := BombUrl(tt.args.ctx, tt.args.c, tt.args.n); (err != nil) != tt.wantErr {
				t.Errorf("BombUrl() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
