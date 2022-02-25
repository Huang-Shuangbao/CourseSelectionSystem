package initial

import (
	"camp-backend/types"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
	"sync"
)

const (
	redisAddr = "180.184.68.166:6379"
	password  = ""
)

var RedisContext = context.Background()
var RedisClient *redis.Client

func SetupRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: password,
		DB:       0,
	})

	pong, err := RedisClient.Ping(RedisContext).Result()
	fmt.Println("Redis ping result: ", pong)
	if err != nil {
		panic(fmt.Sprintf("redis ping failed, err is %s", err))
	}
}

var CourseMutexes []sync.Mutex

func InsertDataToRedis() {
	students := make([]types.TMember, 0)
	Db.Select("user_id").Find(&students, "user_type = ? AND is_deleted = ?", 2, 0)
	for _, v := range students {
		RedisClient.SAdd(RedisContext, "students", v.UserID)
	}

	courses := make([]types.TCourse, 0)
	Db.Select("course_id", "capacity").Find(&courses)
	CourseMutexes = make([]sync.Mutex, len(courses))
	for _, v := range courses {
		intCourseID, _ := strconv.Atoi(v.CourseID)
		c := &redis.Z{
			Score:  float64(intCourseID),
			Member: v.CourseID,
		}
		RedisClient.ZAdd(RedisContext, "courses", c)
		RedisClient.HSet(RedisContext, "course:"+v.CourseID, "capacity", v.Capacity, "count", 0)
	}
}
