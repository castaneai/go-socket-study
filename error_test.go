package go_socket_study

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestClosingByClient(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	nch := make(chan int)
	errch := make(chan error)
	go func() {
		c, err := lis.Accept()
		if err != nil {
			log.Fatal(err)
		}
		buf := make([]byte, 100)
		n, err := c.Read(buf);
		nch <-n
		errch <-err
	}()

	c, err := net.Dial("tcp", lis.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	c.Close()

	assert.Equal(t, 0, <-nch)
	assert.Equal(t, io.EOF, <-errch)
}

func TestClosingByServer(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		c, err := lis.Accept()
		if err != nil {
			log.Fatal(err)
		}
		if err := c.Close(); err != nil {
			log.Fatal(err)
		}
		log.Printf("closed by server")
		wg.Done()
	}()

	c, err := net.Dial("tcp", lis.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	{
		n, err := c.Write([]byte("hello"))
		assert.Equal(t, 5, n)
		assert.Nil(t, err)
	}

	{
		n, err := c.Write([]byte("hello"))
		assert.Equal(t, 0, n)
		assert.True(t, errors.Is(err, syscall.EPIPE))
	}
}

func TestClosingOnReading(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	nch := make(chan int)
	errch := make(chan error)
	go func() {
		c, err := lis.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			time.Sleep(100 * time.Millisecond)
			c.Close()
			log.Printf("closing client on reading...")
		}()
		buf := make([]byte, 100)
		n, err := c.Read(buf);
		nch <-n
		errch <-err
	}()

	c, err := net.Dial("tcp", lis.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 100)
	n, err := c.Read(buf)
	assert.Equal(t, 0, <-nch)
	assert.Contains(t, (<-errch).Error(), "use of closed network connection")
	assert.Equal(t, 0, n)
	assert.True(t, errors.Is(err, io.EOF))
}
