package handler

import (
	"camp-backend/initial"
	"camp-backend/types"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

var currentMaxCourseID int64 = 10

func CreateCourse(c *gin.Context) {
	request := new(types.CreateCourseRequest)
	response := new(types.CreateCourseResponse)

	if err := c.BindJSON(request); err != nil {
		response.Code = types.ParamInvalid
		c.JSON(http.StatusOK, response)
		return
	}

	theCourse := new(types.TCourse)
	err := initial.Db.First(theCourse, "name = ?", request.Name).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		response.Code = types.UnknownError
		c.JSON(http.StatusOK, response)
		return
	}

	currentMaxCourseID++
	var newCourse = types.TCourse{
		CourseID: strconv.FormatInt(currentMaxCourseID, 10),
		Name:     request.Name,
		Capacity: request.Cap,
	}
	if err = initial.Db.Omit("TeacherID").Create(&newCourse).Error; err != nil {
		response.Code = types.UnknownError
		c.JSON(http.StatusOK, response)
		return
	}
	response.Code = types.OK
	response.Data.CourseID = strconv.FormatInt(currentMaxCourseID, 10)
	c.JSON(http.StatusOK, response)
}

func GetCourse(c *gin.Context) {
	request := new(types.GetCourseRequest)
	response := new(types.GetCourseResponse)

	request.CourseID = c.Query("CourseID")
	if request.CourseID == "" {
		response.Code = types.ParamInvalid
		c.JSON(http.StatusOK, response)
		return
	}

	theCourse := new(types.TCourse)
	err := initial.Db.First(theCourse, request.CourseID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Code = types.UnknownError
		c.JSON(http.StatusOK, response)
		return
	}

	response.Code = types.OK
	response.Data = *theCourse
	c.JSON(http.StatusOK, response)
}

func BindCourse(c *gin.Context) {
	request := new(types.BindCourseRequest)
	response := new(types.BindCourseResponse)

	if err := c.BindJSON(request); err != nil {
		response.Code = types.ParamInvalid
		c.JSON(http.StatusOK, response)
		return
	}

	theTeacher := new(types.TMember)
	err := initial.Db.First(theTeacher, request.TeacherID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Code = types.UserNotExisted
		c.JSON(http.StatusOK, response)
		return
	}
	if theTeacher.UserType.String() != "Teacher" {
		response.Code = types.ParamInvalid
		c.JSON(http.StatusOK, response)
		return
	}

	theCourse := new(types.TCourse)
	err = initial.Db.First(theCourse, request.CourseID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Code = types.CourseNotExisted
		c.JSON(http.StatusOK, response)
		return
	}
	if theCourse.TeacherID != "" {
		response.Code = types.CourseHasBound
		c.JSON(http.StatusOK, response)
		return
	}

	if err := initial.Db.Model(theCourse).Association("Teacher").Append(theTeacher); err != nil {
		response.Code = types.UnknownError
		c.JSON(http.StatusOK, response)
		return
	}
	response.Code = types.OK
	c.JSON(http.StatusOK, response)
}

func UnbindCourse(c *gin.Context) {
	request := new(types.UnbindCourseRequest)
	response := new(types.UnbindCourseResponse)

	if err := c.BindJSON(request); err != nil {
		response.Code = types.ParamInvalid
		c.JSON(http.StatusOK, response)
		return
	}

	theCourse := new(types.TCourse)
	err := initial.Db.First(theCourse, request.CourseID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Code = types.CourseNotExisted
		c.JSON(http.StatusOK, response)
		return
	}
	if theCourse.TeacherID != request.TeacherID {
		response.Code = types.CourseNotBind
		c.JSON(http.StatusOK, response)
		return
	}

	if err := initial.Db.Model(theCourse).Association("Teacher").Clear(); err != nil {
		response.Code = types.UnknownError
		c.JSON(http.StatusOK, response)
		return
	}
	response.Code = types.OK
	c.JSON(http.StatusOK, response)
}

func GetTeacherCourse(c *gin.Context) {
	request := new(types.GetTeacherCourseRequest)
	response := new(types.GetTeacherCourseResponse)

	request.TeacherID = c.Query("TeacherID")
	if request.TeacherID == "" {
		response.Code = types.ParamInvalid
		c.JSON(http.StatusOK, response)
		return
	}

	courses := make([]types.TCourse, 0)
	initial.Db.Where("teacher_id = ?", request.TeacherID).Find(&courses)
	fmt.Println(courses)
	response.Data.CourseList = make([]*types.TCourse, len(courses))
	response.Code = types.OK
	for i, _ := range courses {
		response.Data.CourseList[i] = &courses[i]
	}
	c.JSON(http.StatusOK, response)
}

func dfs(teacher string, courseList map[string][]string, used map[string]bool, pre map[string]string) bool {
	for _, course := range courseList[teacher] {
		if !used[course] {
			used[course] = true
			if pre[course] == "" || dfs(pre[course], courseList, used, pre) {
				pre[course] = teacher
				return true
			}
		}
	}
	return false
}

func ScheduleCourse(c *gin.Context) {
	request := new(types.ScheduleCourseRequest)
	response := new(types.ScheduleCourseResponse)

	pre := make(map[string]string)

	if err := c.BindJSON(request); err != nil {
		response.Code = types.ParamInvalid
		c.JSON(http.StatusBadRequest, response)
		return
	}
	g := request.TeacherCourseRelationShip

	for teacher, _ := range g {
		used := make(map[string]bool)
		if dfs(teacher, g, used, pre) {
			continue
		}
	}

	response.Data = make(map[string]string)
	for course, teacher := range pre {
		response.Data[teacher] = course
	}
	response.Code = types.OK
	c.JSON(http.StatusOK, response)
}
