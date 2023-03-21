package bomber

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type HttpBomber struct {
	client http.Client
}

type HttpConfig struct {
	Url     string            `json:"url"`
	Method  string            `json:"method"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
}

const Error = "Error"
const StatusCode = "StatusCode"

func NewHttpBomber(timeoutSecond int) (*HttpBomber, error) {
	return &HttpBomber{
		client: http.Client{
			Timeout: time.Duration(timeoutSecond) * time.Second,
		},
	}, nil

}
func (s *HttpBomber) translateTask(task []byte) (*HttpConfig, error) {
	//expect task is JSON string
	config := HttpConfig{}
	err := json.Unmarshal(task, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall JSON to HttpConfig %v", task)
	}
	return &config, nil
}

func (s *HttpBomber) Bomb(ctx context.Context, task []byte) (BomberReport, error) {
	config, err := s.translateTask(task)
	if err != nil {
		return BomberReport{
			Name:   Error,
			Result: "1",
		}, fmt.Errorf("failed to translate the JSON task, error %v", err)
	}

	req, err := http.NewRequest(config.Method, config.Url, strings.NewReader(config.Body))
	if err != nil {
		return BomberReport{
			Name:   Error,
			Result: fmt.Sprintf("client: error creating http request, error %v", err),
		}, nil
	}
	for name, value := range config.Headers {
		req.Header.Set(name, value)
	}

	res, err := s.client.Do(req)
	if err != nil {
		return BomberReport{
			Name:   Error,
			Result: fmt.Sprintf("client: error making http request, error %v", err),
		}, nil
	}
	return BomberReport{
		Name:   StatusCode,
		Result: fmt.Sprintf("%v", res.StatusCode),
	}, nil
}
