package redis

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mediocregopher/radix/v3"
	"github.com/urfave/cli"
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
	dbSelect := c.Int("redisdbselect")
	return Init(host, port, poolSize, dbSelect)
}

// Init creates a Redis Pool.
func Init(host string, port, poolSize int, dbSelect int) (err error) {
	connFunc := func(network, addr string) (radix.Conn, error) {
		options := []radix.DialOpt{
			radix.DialSelectDB(dbSelect),
		}
		return radix.Dial(network, addr,
			options...,
		)
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	Pool, err = radix.NewPool("tcp", addr, poolSize, radix.PoolConnFunc(connFunc))
	return err
}

func CleanupHook() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-c
		Pool.Close()
		os.Exit(0)
	}()
}
