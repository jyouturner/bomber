package bomber

import (
	"context"
	"fmt"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

// bombers represents the workers to execute tasks in typical fanout worker pattern.
type bombers struct {
	numberOfWorkers int
	tasksChan       chan []byte
}

// BomberReport has the result of the task
type BomberReport struct {
	Name   string
	Result string
}

const BufferSize = 100

// NewBombers create number of workers or bombers, and return two functions, one for the caller to add task for workers to process, another one to close or dismiss the mission.
// the caller need to make sure to defer the "close" to avoid any leaking go routines.
func NewBombers(n int) (s *bombers, fnToAddTask func(ctx context.Context, task []byte), fnToClose func()) {
	b := &bombers{
		numberOfWorkers: n,
		tasksChan:       make(chan []byte, BufferSize),
	}
	return b, b.addTask, b.close
}

// AddTask add a task to the task channel for workers to work on.
func (b *bombers) addTask(ctx context.Context, task []byte) {
	select {
	case <-ctx.Done():
	case b.tasksChan <- task:

	}
}

// Close function to dismiss the mission. Must call it after finishing adding tasks, otherwise the go routines will be blocked.
func (b *bombers) close() {
	close(b.tasksChan)
	fmt.Println("bombers task channel is closed")
}

// Launch start the workers and execute tasks, and report results
func (b *bombers) Launch(ctx context.Context, fnBomb func(c context.Context, task []byte) (BomberReport, error)) (map[string]string, error) {
	g, egctx := errgroup.WithContext(ctx)

	reports := make(chan BomberReport)
	nWorkers := int32(b.numberOfWorkers)
	// map tasks to bombers
	for i := 0; i < b.numberOfWorkers; i++ {
		//i2 := i
		g.Go(func() error {
			defer func() {
				// Last one out closes shop
				if atomic.AddInt32(&nWorkers, -1) == 0 {
					close(reports)
				}
			}()
			for task := range b.tasksChan {
				err := b.onTask(egctx, task, fnBomb, reports)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}

	//reduce
	ret := make(map[string]string, b.numberOfWorkers)
	g.Go(func() error {
		//read from the reports
		for report := range reports {
			ret[report.Name] = report.Result

		}
		fmt.Println("finished reporting, all done")
		return nil
	})
	return ret, g.Wait()

}

// onTask execute the given task with the given function
func (b *bombers) onTask(ctx context.Context, task []byte, fnBomb func(c context.Context, task []byte) (BomberReport, error), reports chan BomberReport) error {

	result, err := fnBomb(ctx, task)
	if err != nil {
		return fmt.Errorf("worker failed with %v", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case reports <- result:
	}
	return nil
}
