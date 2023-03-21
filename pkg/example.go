package bomber

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// ReadFileAndBombHttp example to load test HTTP with bombers, configuration including request url, method, body and headers come from CSV file
func ReadFileAndBombHttp(ctx context.Context, f string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	var wg sync.WaitGroup
	bombers, fnAddTask, fnClose := NewBombers(4)

	httpBomber, err := NewHttpBomber(2)
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		bombers.Launch(ctx, httpBomber.bomb)
	}()

	wg.Add(1)
	go func() {
		defer fnClose()
		defer wg.Done()
		//read lines from file and send to bombers
		file, err := os.Open(f)
		if err != nil {
			return
		}
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			fnAddTask(ctx, scanner.Bytes())
		}

		fmt.Println("sender is done")
		file.Close()
	}()
	wg.Wait()
	fmt.Println("done")

	return nil
}

func BombUrl(ctx context.Context, c HttpConfig, n int) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	
	bombers, fnAddTask, fnClose := NewBombers(4)

	httpBomber, err := NewHttpBomber(2)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		bombers.Launch(ctx, httpBomber.bomb)
		fmt.Println("bombers is done")
	}()

	// send the N tasks
	wg.Add(1)
	go func() {
		defer fnClose()
		defer wg.Done()

		task, err := json.Marshal(c)
		if err != nil {
			return
		}
		i := 0
	Loop:
		for i = 0; i < n; i++ {
			select {
			case <-ctx.Done():
				break Loop
			default:
				fnAddTask(ctx, task)
			}

		}

		fmt.Println("sender is done", i)
	}()

	wg.Wait()
	fmt.Println("done")
	return nil
}

type Msg struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func ProcessSqsMessages(c context.Context, task []byte) (BomberReport, error) {
	//process message this is array of msg
	arrMsg := []Msg{}
	err := json.Unmarshal(task, &arrMsg)
	fmt.Println(arrMsg)
	if err != nil {
		return BomberReport{
			name:   "error",
			result: "error",
		}, err
	}
	return BomberReport{
		name:   "error",
		result: "error",
	}, nil
}

func ListenToSqs(ctx context.Context, queueName string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	var wg sync.WaitGroup
	bombers, fnAddTask, fnClose := NewBombers(4)

	sqsWorker := NewSqsWorker(ctx, queueName)

	wg.Add(1)
	go func() {
		defer wg.Done()
		bombers.Launch(ctx, ProcessSqsMessages)
		fmt.Println("bombers is done")
	}()

	// send the N tasks
	wg.Add(1)
	go func() {
		defer fnClose()
		defer wg.Done()
		//keep receiving message
	Loop:
		for {
			msgs, err := sqsWorker.Receive(ctx)
			if err != nil {
				return
			}
			for _, msg := range msgs {
				//send message to workers
				select {
				case <-ctx.Done():
					break Loop
				default:
					fnAddTask(ctx, []byte(*msg.Body))
					//TODO this is problem
					sqsWorker.Delete(ctx, msg)
				}
			}
		}

		fmt.Println("sender is done")
	}()

	wg.Wait()
	fmt.Println("done")
	return nil
}

func BombSqs(ctx context.Context, queueName string, n int) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	var wg sync.WaitGroup
	bombers, fnAddTask, fnClose := NewBombers(20)

	sqsWorker := NewSqsWorker(ctx, queueName)

	wg.Add(1)
	go func() {
		defer wg.Done()
		bombers.Launch(ctx, func(c context.Context, task []byte) (BomberReport, error) {

			sqsWorker.SendMessage(ctx, string(task), map[string]types.MessageAttributeValue{})
			return BomberReport{
				name:   "",
				result: "",
			}, nil
		})
		fmt.Println("bombers is done")
	}()

	// send the N tasks
	wg.Add(1)
	go func() {
		defer fnClose()
		defer wg.Done()

		task, err := json.Marshal(Msg{
			Name:  "hello",
			Value: "world",
		})
		if err != nil {
			return
		}
		i := 0
	Loop:
		for i = 0; i < n; i++ {
			select {
			case <-ctx.Done():
				break Loop
			default:
				fnAddTask(ctx, task)
			}

		}

		fmt.Println("sender is done", i)
	}()

	wg.Wait()
	fmt.Println("done")
	return nil
}
