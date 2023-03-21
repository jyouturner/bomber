package example

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	bomber "github.com/jyouturner/bomber/pkg/bombers"
)

type Msg struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func ProcessSqsMessages(c context.Context, task []byte) (bomber.BomberReport, error) {
	//process message this is array of msg
	arrMsg := []Msg{}
	err := json.Unmarshal(task, &arrMsg)
	fmt.Println(arrMsg)
	if err != nil {
		return bomber.BomberReport{
			Name:   "error",
			Result: "error",
		}, err
	}
	return bomber.BomberReport{
		Name:   "error",
		Result: "error",
	}, nil
}

func ListenToSqs(ctx context.Context, queueName string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	var wg sync.WaitGroup
	bombers, fnAddTask, fnClose := bomber.NewBombers(4)

	sqsWorker := bomber.NewSqsWorker(ctx, queueName)

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
	bombers, fnAddTask, fnClose := bomber.NewBombers(20)

	sqsWorker := bomber.NewSqsWorker(ctx, queueName)

	wg.Add(1)
	go func() {
		defer wg.Done()
		bombers.Launch(ctx, func(c context.Context, task []byte) (bomber.BomberReport, error) {

			sqsWorker.SendMessage(ctx, string(task), map[string]types.MessageAttributeValue{})
			return bomber.BomberReport{
				Name:   "",
				Result: "",
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
