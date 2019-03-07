package util

import (
	"fmt"
	"io"
	"runtime"
)

func MustScan(a ...interface{}) {
	if _, err := fmt.Scan(a...); err != nil {
		panic(err)
	}
}

func MustScanln(a ...interface{}) {
	if _, err := fmt.Scanln(a...); err != nil {
		panic(err)
	}
}

func MustScanf(f string, a ...interface{}) {
	if _, err := fmt.Scanf(f, a...); err != nil {
		panic(err)
	}
}

func MustFscan(r io.Reader, a ...interface{}) {
	if _, err := fmt.Fscan(r, a...); err != nil {
		panic(err)
	}
}

func MustFscanln(r io.Reader, a ...interface{}) {
	if _, err := fmt.Fscanln(r, a...); err != nil {
		panic(err)
	}
}

func MustFscanf(r io.Reader, f string, a ...interface{}) {
	if _, err := fmt.Fscanf(r, f, a...); err != nil {
		panic(err)
	}
}

func WorkerPool(num int, work func(interface{}) interface{}) (chan<- interface{}, <-chan interface{}) {
	inputCh := make(chan interface{}, 1)
	outputCh := make(chan interface{}, 1)

	if num <= 1 {
		go func() {
			defer close(outputCh)
			for input := range inputCh {
				outputCh <- work(input)
			}
		}()
		return inputCh, outputCh
	}

	type task struct {
		input    interface{}
		outputCh chan interface{}
	}

	buffIn := make(chan chan interface{}, 1)
	buffOut := make(chan chan interface{}, 1)
	tasks := make(chan task, 1)

	go func() {
		defer close(buffOut)
		var buff []chan interface{}
		for {
			if len(buff) == 0 {
				in, ok := <-buffIn
				if !ok {
					return
				}
				buff = append(buff, in)
			} else {
				select {
				case in, ok := <-buffIn:
					if !ok {
						for len(buff) != 0 {
							buffOut <- buff[0]
							buff = buff[1:]
						}
						return
					}
					buff = append(buff, in)
				case buffOut <- buff[0]:
					buff = buff[1:]
				}
			}
		}
	}()

	go func() {
		defer func() {
			close(buffIn)
			close(tasks)
		}()
		for input := range inputCh {
			taskOutputCh := make(chan interface{}, 1)
			buffIn <- taskOutputCh
			tasks <- task{
				input:    input,
				outputCh: taskOutputCh,
			}
		}
	}()

	go func() {
		defer close(outputCh)
		for s := range buffOut {
			outputCh <- <-s
		}
	}()

	for i := 0; i < num; i++ {
		go func() {
			for task := range tasks {
				func() {
					defer close(task.outputCh)
					task.outputCh <- work(task.input)
				}()
			}
		}()
	}

	return inputCh, outputCh
}

func AllCpusWorkerPool(work func(interface{}) interface{}) (chan<- interface{}, <-chan interface{}) {
	return WorkerPool(runtime.NumCPU(), work)
}

func Main(readInput func(chan<- interface{}), work func(interface{}) interface{}) {
	inputCh, outputCh := AllCpusWorkerPool(work)

	go func() {
		defer close(inputCh)
		readInput(inputCh)
	}()

	for out := range outputCh {
		if out != nil {
			fmt.Println(out)
		}
	}
}
