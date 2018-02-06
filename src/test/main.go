package main

import "app/redis"
//import "fmt"

//import "github.com/garyburd/redigo/redis"

//测试
func main() {
	
	/*
	c, err := redis.Dial("tcp", "10.10.6.8:8501")
	if err != nil {
		fmt.Println(err)
		return
	}
	//密码授权
	c.Do("AUTH", "kjt@123")
	c.Do("SELECT", 15)
	c.Do("SET", "a", "1223456789")
	a, err := redis.String(c.Do("GET", "a"))

	fmt.Println(a)

	defer c.Close()
	*/
	
	redis.Set("123456789", "sdfasfdasdfasdfadsfasdf")
	
	for {
		
	}
}