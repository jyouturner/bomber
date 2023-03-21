## Implementation of Fanout and Workers Process

### Code

The implementation is in the [bombers.go](/pkg/bombers.go)

### To Use the Package


````Go
ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
defer stop()
````

then crate the workers, note this function returns back two functions for you can call - one to send "task" to the workers, the other to call of the workers (by closing the task channel)

````Go
bombers, fnAddTask, fnClose := NewBombers(4)
````

then create a waitgroup and couple of go routines. First to launch all the workers, and provide the function to process the task.

````Go
var wg sync.WaitGroup
wg.Add(1)
go func() {
	defer wg.Done()
	bombers.Launch(ctx, httpBomber.bomb)
}()
````

another go routine to send tasks to the workers. Note we call the close function to close the task channel (only sender cloes the channel). 

````Go
wg.Add(1)
	go func() {
		defer fnClose()
		defer wg.Done()
	Loop:
		for i = 0; i < 10000000; i++ {
			select {
			case <-ctx.Done():
				break Loop
			default:
				fnAddTask(ctx, task)
			}

		}
	}()
````

and wait for the them

````Go
wg.Wait()
fmt.Println("done")
````

### Test

run some test cases

````sh
go test -v -run TestBombSqs ./pkg/bomber
````

run all the test cases

````sh
go test -v -run Test ./pkg/bomber
````

### Examples

below examples can be found in [example.go](/pkg/example.go)

1. Simply Hit One Single URL

2. Read from CSV file and Stress Test HTTP

3. Simply Send Large Mound of Messages to SQS

4. Listen to SQS and Process Incoming Messages