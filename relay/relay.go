package relay

import (
	"go.uber.org/multierr"
	"io"
)

func CloseAfterRelay(dst, src io.ReadWriteCloser) error {
	ch := make(chan error)
	go relay(dst, src, ch)
	go relay(src, dst, ch)

	err := <-ch
	return multierr.Combine(err, dst.Close(), src.Close())
}

func relay(dst io.Writer, src io.Reader, ch chan<- error) {
	_, err := io.Copy(dst, src)
	ch <- err
}
