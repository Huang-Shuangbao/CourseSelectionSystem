package handler

import (
	"camp-backend/initial"
	"camp-backend/types"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

var insertedToRedis = false

func BookCourse(c *gin.Context) {
	if !insertedToRedis {
		initial.InsertDataToRedis()
		insertedToRedis = true
	}

	request := new(types.BookCourseRequest)
	response := new(types.BookCourseResponse)
	if err := c.BindJSON(request); err != nil {
		response.Code = types.ParamInvalid
		c.JSON(http.StatusOK, response)
		return
	}

	hasStudent, err := initial.RedisClient.SIsMember(initial.RedisContext, "students", request.StudentID).Result()
	if err != nil {
		response.Code = types.UnknownError
		c.JSON(http.StatusOK, response)
		return
	}
	if !hasStudent {
		response.Code = types.StudentNotExisted
		c.JSON(http.StatusOK, response)
		return
	}

	courseRank, err := initial.RedisClient.ZRank(initial.RedisContext, "courses", request.CourseID).Result()
	if err != nil {
		response.Code = types.CourseNotExisted
		c.JSON(http.StatusOK, response)
		return
	}

	initial.CourseMutexes[courseRank].Lock() //加锁防止超卖

	hasBookedThisCourse, err := initial.RedisClient.SIsMember(initial.RedisContext, "student:"+request.StudentID+":courses", request.CourseID).Result()
	if err != nil {
		initial.CourseMutexes[courseRank].Unlock()
		response.Code = types.UnknownError
		c.JSON(http.StatusOK, response)
		return
	}
	if hasBookedThisCourse {
		initial.CourseMutexes[courseRank].Unlock()
		response.Code = types.StudentHasCourse
		c.JSON(http.StatusOK, response)
		return
	}

	//studentHasSomeCourses, err := initial.RedisClient.Exists(initial.RedisContext, "student:"+request.StudentID+":courses").Result()
	//if err != nil {
	//	initial.CourseMutexes[courseRank].Unlock()
	//	response.Code = types.UnknownError
	//	c.JSON(http.StatusOK, response)
	//	return
	//}
	//if studentHasSomeCourses > 0 {
	//	hasBookedThisCourse, err := initial.RedisClient.SIsMember(initial.RedisContext, "student:"+request.StudentID+":courses", request.CourseID).Result()
	//	if err != nil {
	//		initial.CourseMutexes[courseRank].Unlock()
	//		response.Code = types.UnknownError
	//		c.JSON(http.StatusOK, response)
	//		return
	//	}
	//	if hasBookedThisCourse {
	//		initial.CourseMutexes[courseRank].Unlock()
	//		response.Code = types.StudentHasCourse
	//		c.JSON(http.StatusOK, response)
	//		return
	//	}
	//}

	course, err := initial.RedisClient.HGetAll(initial.RedisContext, "course:"+request.CourseID).Result()
	if err != nil {
		initial.CourseMutexes[courseRank].Unlock()
		response.Code = types.UnknownError
		c.JSON(http.StatusOK, response)
		return
	}
	count, _ := strconv.Atoi(course["count"])
	capacity, _ := strconv.Atoi(course["capacity"])
	if count >= capacity {
		initial.CourseMutexes[courseRank].Unlock()
		response.Code = types.CourseNotAvailable
		c.JSON(http.StatusOK, response)
		return
	} else {
		initial.RedisClient.HIncrBy(initial.RedisContext, "course:"+request.CourseID, "count", 1)

		initial.CourseMutexes[courseRank].Unlock()

		initial.RedisClient.SAdd(initial.RedisContext, "student:"+request.StudentID+":courses", request.CourseID)
		response.Code = types.OK
		c.JSON(http.StatusOK, response)
	}
}

func GetStudentCourse(c *gin.Context) {
	if !insertedToRedis {
		initial.InsertDataToRedis()
		insertedToRedis = true
	}

	request := new(types.GetStudentCourseRequest)
	response := new(types.GetStudentCourseResponse)

	request.StudentID = c.Query("StudentID")
	if request.StudentID == "" {
		response.Code = types.ParamInvalid
		c.JSON(http.StatusOK, response)
		return
	}

	theUser := new(types.TMember)
	err := initial.Db.First(theUser, request.StudentID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || theUser.UserType.String() != "Student" {
		response.Code = types.StudentNotExisted
		c.JSON(http.StatusOK, response)
		return
	}

	courseIDs, err := initial.RedisClient.SMembers(initial.RedisContext, "student:"+request.StudentID+":courses").Result()
	if err != nil {
		response.Code = types.UnknownError
		c.JSON(http.StatusOK, response)
		return
	}
	courses := make([]types.TCourse, 0)
	for _, v := range courseIDs {
		theCourse := new(types.TCourse)
		if err := initial.Db.First(theCourse, v).Error; err != nil {
			response.Code = types.UnknownError
			c.JSON(http.StatusOK, response)
			return
		}
		courses = append(courses, *theCourse)
	}
	//if err := initial.Db.Model(theUser).Association("LearnCourses").Find(&courses); err != nil {
	//	response.Code = types.UnknownError
	//	c.JSON(http.StatusOK, response)
	//	return
	//}

	if len(courses) == 0 {
		response.Code = types.StudentHasNoCourse
	} else {
		response.Code = types.OK
	}
	response.Data.CourseList = courses
	c.JSON(http.StatusOK, response)
}
