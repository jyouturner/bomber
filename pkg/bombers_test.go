package bomber

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func simplebomb(ctx context.Context, task []byte) (BomberReport, error) {
	return BomberReport{
		name:   fmt.Sprintf("bomber %s", task),
		result: fmt.Sprintf("bombed %s", task),
	}, nil

}

func sleepbomb(ctx context.Context, task []byte) (BomberReport, error) {
	return BomberReport{
		name:   fmt.Sprintf("bomber %s", task),
		result: fmt.Sprintf("bombed %s", task),
	}, SleepWithContext(ctx, time.Duration(100)*time.Second)

}

func SleepWithContext(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		fmt.Println("sleep end due to context done")
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

func TestBombers_Bomb(t *testing.T) {

	type fields struct {
		numberOfWorkers int
	}
	type args struct {
		ctx  context.Context
		bomb func(c context.Context, task []byte) (BomberReport, error)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]string
		wantErr bool
	}{
		//simple
		{
			name: "simple case",
			fields: fields{
				numberOfWorkers: 4,
			},
			args: args{
				ctx:  context.TODO(),
				bomb: simplebomb,
			},
			want: map[string]string{
				"bomber A": "bombed A",
				"bomber B": "bombed B",
				"bomber C": "bombed C",
				"bomber D": "bombed D",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, fnAddTask, fnClose := NewBombers(tt.fields.numberOfWorkers)
			go func() {
				defer fnClose()
				//add tasks
				fnAddTask(tt.args.ctx, []byte("A"))
				fnAddTask(tt.args.ctx, []byte("B"))
				fnAddTask(tt.args.ctx, []byte("C"))
				fnAddTask(tt.args.ctx, []byte("D"))

			}()
			got, err := b.Launch(tt.args.ctx, tt.args.bomb)
			fmt.Println("got", got)
			if (err != nil) != tt.wantErr {
				t.Errorf("bombers.Bomb() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("bombers.Bomb() = %v, want %v", got, tt.want)
			}

		})
	}
}

func TestBombers_Http(t *testing.T) {
	httpbomber, _ := NewHttpBomber(10)
	type fields struct {
		numberOfWorkers int
	}
	type args struct {
		ctx  context.Context
		bomb func(c context.Context, task []byte) (BomberReport, error)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "test http bomber",
			fields: fields{
				numberOfWorkers: 4,
			},
			args: args{
				ctx:  context.TODO(),
				bomb: httpbomber.bomb,
			},
			want: map[string]string{
				"StatusCode": "200",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, fnAddTask, fnClose := NewBombers(tt.fields.numberOfWorkers)
			go func() {
				defer fnClose()
				//add tasks
				fnAddTask(tt.args.ctx, createHttpGetTask("http://www.google.com"))
				fnAddTask(tt.args.ctx, createHttpGetTask("http://www.microsoft.com"))

			}()
			got, err := b.Launch(tt.args.ctx, tt.args.bomb)
			fmt.Println("got", got)
			if (err != nil) != tt.wantErr {
				t.Errorf("bombers.Bomb() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("bombers.Bomb() = %v, want %v", got, tt.want)
			}

		})
	}
}

func createHttpGetTask(url string) []byte {
	cfg := HttpConfig{
		Url:    url,
		Method: "GET",
		Body:   "",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
	task, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	return task
}
