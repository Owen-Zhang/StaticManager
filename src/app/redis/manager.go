package redis

import (
	"fmt"
	redisgo "github.com/garyburd/redigo/redis"
)

func Ping() error {
    conn := pool.Get()
    defer conn.Close()

    _, err := redisgo.String(conn.Do("PING"))
    if err != nil {
        return fmt.Errorf("cannot 'PING' db: %v", err)
    }
    return nil
}

func Get(key string) (string, error) {

    conn := pool.Get()
    defer conn.Close()

    var data string
    data, err := redisgo.String(conn.Do("GET", key))
    if err != nil {
        return data, fmt.Errorf("error getting key %s: %v", key, err)
    }
    return data, err
}

func Set(key string, value string) error {

    conn := pool.Get()
    defer conn.Close()
	
	//前面有错，conn会反映出来
	
    _, err := conn.Do("SET", key, value)
    if err != nil {
        v := string(value)
        if len(v) > 15 {
            v = v[0:12] + "..."
        }
        return fmt.Errorf("error setting key %s to %s: %v", key, v, err)
    }
    return err
}

func Exists(key string) (bool, error) {

    conn := pool.Get()
    defer conn.Close()

    ok, err := redisgo.Bool(conn.Do("EXISTS", key))
    if err != nil {
        return ok, fmt.Errorf("error checking if key %s exists: %v", key, err)
    }
    return ok, err
}

func Delete(key string) error {

    conn := pool.Get()
    defer conn.Close()

    _, err := conn.Do("DEL", key)
    return err
}

func GetKeys(pattern string) ([]string, error) {

    conn := pool.Get()
    defer conn.Close()

    iter := 0
    keys := []string{}
    for {
        arr, err := redisgo.Values(conn.Do("SCAN", iter, "MATCH", pattern))
        if err != nil {
            return keys, fmt.Errorf("error retrieving '%s' keys", pattern)
        }

        iter, _ = redisgo.Int(arr[0], nil)
        k, _ := redisgo.Strings(arr[1], nil)
        keys = append(keys, k...)

        if iter == 0 {
            break
        }
    }

    return keys, nil
}

func Incr(counterKey string) (int, error) {

    conn := pool.Get()
    defer conn.Close()

    return redisgo.Int(conn.Do("INCR", counterKey))
}