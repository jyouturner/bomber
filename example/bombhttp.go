package example

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"

	bomber "github.com/jyouturner/bomber/pkg/bombers"
)

// ReadFileAndBombHttp example to load test HTTP with bombers, configuration including request url, method, body and headers come from CSV file
func ReadFileAndBombHttp(ctx context.Context, f string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	var wg sync.WaitGroup
	bombers, fnAddTask, fnClose := bomber.NewBombers(4)

	httpBomber, err := bomber.NewHttpBomber(2)
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		bombers.Launch(ctx, httpBomber.Bomb)
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

func BombUrl(ctx context.Context, c bomber.HttpConfig, n int) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	bombers, fnAddTask, fnClose := bomber.NewBombers(4)

	httpBomber, err := bomber.NewHttpBomber(2)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		bombers.Launch(ctx, httpBomber.Bomb)
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
