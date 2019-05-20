package redis

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mediocregopher/radix"
	cli "gopkg.in/urfave/cli.v1"
)

var (
	Pool *radix.Pool
)

// InitFromCli is the same as Init but receives it's parameters
// from a Context struct of the cli package (aka from program arguments)
func InitFromCli(c *cli.Context) error {
	host := c.String("redishost")
	port := c.Int("redisport")
	poolSize := c.Int("redispoolsize")

	return Init(host, port, poolSize)
}

// Init creates a Redis Pool.
func Init(host string, port, poolSize int) (err error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	if Pool, err = radix.NewPool("tcp", addr, poolSize); err != nil {
		return err
	}
	cleanupHook()
	return
}

func cleanupHook() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	go func() {
		<-c
		Pool.Close()
		os.Exit(0)
	}()
}
