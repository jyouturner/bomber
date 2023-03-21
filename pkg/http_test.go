package bomber

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func TestHttpBomber_bomb(t *testing.T) {
	cfg := HttpConfig{
		Url:    "http://google.com",
		Method: "GET",
		Body:   "",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
	task, err := json.Marshal(cfg)
	if err != nil {
		t.Fail()
	}
	type fields struct {
		client http.Client
	}
	type args struct {
		task []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    BomberReport
		wantErr bool
	}{
		// test making http request
		{
			name: "making http request GET",
			fields: fields{
				client: *http.DefaultClient,
			},
			args: args{
				task: task,
			},
			want: BomberReport{
				name:   "StatusCode",
				result: "200",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &HttpBomber{
				client: tt.fields.client,
			}
			got, err := s.bomb(context.TODO(), tt.args.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("HttpBomber.bomb() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HttpBomber.bomb() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpBomber_translateTask(t *testing.T) {
	cfg := HttpConfig{
		Url:    "http://google.com",
		Method: "GET",
		Body:   "",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
	task, err := json.Marshal(cfg)
	if err != nil {
		t.Fail()
	}
	type fields struct {
		client http.Client
	}
	type args struct {
		task []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *HttpConfig
		wantErr bool
	}{
		// test transform task from bytes to struct
		{
			name: "json unmarshalling",
			fields: fields{
				client: *http.DefaultClient,
			},
			args: args{
				task: task,
			},
			want:    &cfg,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &HttpBomber{
				client: tt.fields.client,
			}
			got, err := s.translateTask(tt.args.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("HttpBomber.translateTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HttpBomber.translateTask() = %v, want %v", got, tt.want)
			}
		})
	}
}
