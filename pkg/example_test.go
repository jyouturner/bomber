package bomber

import (
	"context"
	"fmt"
	"syscall"
	"testing"
	"time"
)

func TestBombUrl(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(time.Millisecond*1000))
	defer cancel()
	type args struct {
		ctx context.Context
		c   HttpConfig
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
				c: HttpConfig{
					Url:    "http://localhost:3000/mandrill",
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
		c   HttpConfig
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
				c: HttpConfig{
					Url:    "http://localhost:3000/mandrill",
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

func TestListenToSqs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(time.Millisecond*1000))
	defer cancel()
	type args struct {
		ctx       context.Context
		queueName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// listen to sqs queue and process messages
		{
			name: "listen to sqs",
			args: args{
				ctx:       ctx,
				queueName: "loadtest-webhook-mandrill-app2",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		go func() {

			time.Sleep(1000 * time.Millisecond)
			fmt.Println("sending signal...")
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}()
		t.Run(tt.name, func(t *testing.T) {
			if err := ListenToSqs(tt.args.ctx, tt.args.queueName); (err != nil) != tt.wantErr {
				t.Errorf("ListenToSqs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBombSqs(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(time.Millisecond*1000))
	defer cancel()
	type args struct {
		ctx       context.Context
		queueName string
		n         int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "bomb sqs",
			args: args{
				ctx:       ctx,
				queueName: "loadtest-webhook-mandrill-app2",
				n:         100000,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := BombSqs(tt.args.ctx, tt.args.queueName, tt.args.n); (err != nil) != tt.wantErr {
				t.Errorf("BombSqs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
