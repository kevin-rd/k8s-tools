package socks5

import (
	"context"
	"io"
	"sync"
)

var bufPool512 = sync.Pool{
	New: func() interface{} {
		return make([]byte, 512)
	},
}

// from https://ixday.github.io/post/golang-cancel-copy/

type readerFunc func(p []byte) (n int, err error)

func (rf readerFunc) Read(p []byte) (n int, err error) { return rf(p) }

func copyWithCtx(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	// Copy will call the Reader and Writer interface multiple time,
	// in order to copy by chunk (avoiding loading the whole file in memory).
	// I insert the ability to cancel before read time as it is the earliest possible in the call process.
	return io.Copy(dst, readerFunc(func(p []byte) (n int, err error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			return src.Read(p)
		}
	}))
}
