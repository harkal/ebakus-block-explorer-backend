package redis

import (
	"fmt"

	"github.com/mediocregopher/radix"
)

func Set(key string, value []byte) error {
	if err := Pool.Do(radix.FlatCmd(nil, "SET", key, value)); err != nil {
		v := string(value)
		if len(v) > 15 {
			v = v[0:12] + "..."
		}
		return fmt.Errorf("error setting key %s to %s: %v", key, v, err)
	}
	return nil
}

func Exists(key string) (ok bool, err error) {
	err = Pool.Do(radix.FlatCmd(&ok, "EXISTS", key))
	return
}

func Get(key string) ([]byte, error) {
	var data []byte
	if err := Pool.Do(radix.Cmd(&data, "GET", key)); err != nil {
		return nil, fmt.Errorf("error getting key %s: %v", key, err)
	}
	return data, nil
}

func Delete(key string) error {
	return Pool.Do(radix.Cmd(nil, "DEL", key))
}

func Expire(key string, seconds uint64) error {
	var res uint64
	if err := Pool.Do(radix.FlatCmd(&res, "EXPIRE", key, seconds)); err != nil {
		return fmt.Errorf("error setting expire for key %s: %v", key, err)
	}
	if res != 1 {
		return fmt.Errorf("couldn't find the key %s for setting expire", key)
	}
	return nil
}
