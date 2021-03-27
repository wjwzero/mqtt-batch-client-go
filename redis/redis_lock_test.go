package redis

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	id := "redisTest"
	key := "a_redis_test"
	l := NewRedisPool("", "123456")

	if l.AddLock(key, id, 30000) {
		fmt.Println("添加lock成功")
	} else {
		fmt.Println("添加lock失败")
	}

	if l.GetLock(key) == id {
		fmt.Println("相等")
	} else {
		fmt.Println("不相等")
	}

	/*	if DelLock(key, id) {
			fmt.Println("删除成功")
		} else {
			fmt.Println("删除失败")
		}*/
}