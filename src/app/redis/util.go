package redis

import (
	"os"
	"os/signal"
    "syscall"
    "time"
	redisgo "github.com/garyburd/redigo/redis"
)

var pool *redisgo.Pool

func init() {
	pool = newPool("10.10.6.8:8501", "kjt@123")
	cleanupHook()
}

func newPool(server, password string) *redisgo.Pool {
	return &redisgo.Pool{
        MaxIdle:     3, //最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态
        MaxActive:   6, // 最大的激活连接数，表示同时最多有N个连接 ，为0事表示没有限制
        IdleTimeout: 240 * time.Second, //最大的空闲连接等待时间，超过此时间后，空闲连接将被关闭
        //Dial 是创建链接的方法
		Dial: func() (redisgo.Conn, error) {
            c, err := redisgo.Dial("tcp", server)
            if err != nil {
                return nil, err
            }
            if _, err := c.Do("AUTH", password); err != nil {
                c.Close()
                return nil, err
            }
			if _, err := c.Do("SELECT", 15); err != nil {
				c.Close()
				return nil, err
			}
            return c, err
        },
        TestOnBorrow: func(c redisgo.Conn, t time.Time) error {
            if time.Since(t) < time.Minute {
                return nil
            }
            _, err := c.Do("PING")
            return err
        },
    }
}

func cleanupHook() {

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    signal.Notify(c, syscall.SIGTERM)
    signal.Notify(c, syscall.SIGKILL)
    go func() {
        <-c
        pool.Close()
    }()
}

/*
使用：
conn := pool.Get()
defer conn.Close()

conn.Do("", "","")
*/