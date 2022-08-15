package palette

import (
	"bufio"
	"io"
)

func putBack(line string, inCh <-chan string) <-chan string {
	outCh := make(chan string, 1)
	outCh <- line

	go func() {
		defer close(outCh)
		for v := range inCh {
			outCh <- v
		}
	}()

	return outCh
}

func readLines(done <-chan struct{}, r io.Reader) (<-chan string, <-chan error) {
	ch, errCh := make(chan string), make(chan error, 1)

	go func() {
		defer close(ch)
		defer close(errCh)

		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			select {
			case <-done:
				return
			case ch <- scanner.Text():
			}
		}

		if scanner.Err() != nil {
			errCh <- scanner.Err()
		}
	}()

	return ch, errCh
}
