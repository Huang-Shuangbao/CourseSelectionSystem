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

var currentMaxUserID int64 = 10

func CreateMember(c *gin.Context) {
	response := new(types.CreateMemberResponse)

	userID, err := c.Cookie("camp-session")
	if err != nil {
		response.Code = types.LoginRequired
		c.JSON(http.StatusOK, response)
		return
	}

	currentUser := new(types.TMember)
	if err = initial.Db.First(currentUser, userID).Error; err != nil {
		response.Code = types.UnknownError
		c.JSON(http.StatusOK, response)
		return
	}

	if currentUser.UserType.String() != "Admin" {
		response.Code = types.PermDenied
		c.JSON(http.StatusOK, response)
		return
	}

	request := new(types.CreateMemberRequest)
	if err = c.BindJSON(request); err != nil {
		response.Code = types.ParamInvalid
		c.JSON(http.StatusOK, response)
		return
	}

	theUser := new(types.TMember)
	err = initial.Db.First(theUser, "username = ?", request.Username).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		response.Code = types.UserHasExisted
		c.JSON(http.StatusOK, response)
		return
	}

	currentMaxUserID++
	var newUser = types.TMember{
		UserID:    strconv.FormatInt(currentMaxUserID, 10),
		Nickname:  request.Nickname,
		Username:  request.Username,
		Password:  request.Password,
		UserType:  request.UserType,
		IsDeleted: false,
	}
	if err = initial.Db.Omit("LearnCourses").Create(&newUser).Error; err != nil {
		response.Code = types.UnknownError
		c.JSON(http.StatusOK, response)
		return
	}
	response.Code = types.OK
	response.Data.UserID = strconv.FormatInt(currentMaxUserID, 10)
	c.JSON(http.StatusOK, response)
}

func GetMember(c *gin.Context) {
	request := new(types.GetMemberRequest)
	response := new(types.GetMemberResponse)

	request.UserID = c.Query("UserID")
	if request.UserID == "" {
		response.Code = types.ParamInvalid
		c.JSON(http.StatusOK, response)
		return
	}

	theUser := new(types.TMember)
	err := initial.Db.First(theUser, request.UserID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Code = types.UserNotExisted
		c.JSON(http.StatusOK, response)
		return
	}

	if theUser.IsDeleted == true {
		response.Code = types.UserHasDeleted
		c.JSON(http.StatusOK, response)
		return
	}

	response.Code = types.OK
	response.Data = *theUser
	c.JSON(http.StatusOK, response)
}

func GetMemberList(c *gin.Context) {
	request := new(types.GetMemberListRequest)
	response := new(types.GetMemberListResponse)

	var err1, err2 error
	request.Limit, err1 = strconv.Atoi(c.Query("Limit"))
	request.Offset, err2 = strconv.Atoi(c.Query("Offset"))
	if err1 != nil || err2 != nil {
		response.Code = types.ParamInvalid
		c.JSON(http.StatusOK, response)
		return
	}

	members := make([]types.TMember, 0)
	if err := initial.Db.Limit(request.Limit).Offset(request.Offset).Find(&members).Error; err != nil {
		response.Code = types.UnknownError
		response.Data.MemberList = make([]types.TMember, 0)
		c.JSON(http.StatusOK, response)
		return
	}
	response.Code = types.OK
	response.Data.MemberList = members
	c.JSON(http.StatusOK, response)
}

func UpdateMember(c *gin.Context) {
	request := new(types.UpdateMemberRequest)
	response := new(types.UpdateMemberResponse)

	if err := c.BindJSON(request); err != nil {
		response.Code = types.ParamInvalid
		c.JSON(http.StatusOK, response)
		return
	}

	theUser := new(types.TMember)
	err := initial.Db.First(theUser, request.UserID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Code = types.UserNotExisted
		c.JSON(http.StatusOK, response)
		return
	}

	toUpdate := &types.TMember{
		Nickname: request.Nickname,
	}
	if err := initial.Db.Where(request.UserID).Updates(toUpdate).Error; err != nil {
		response.Code = types.UnknownError
		c.JSON(http.StatusOK, response)
		return
	}
	response.Code = types.OK
	c.JSON(http.StatusOK, response)
}

func DeleteMember(c *gin.Context) {
	request := new(types.DeleteMemberRequest)
	response := new(types.DeleteMemberResponse)

	if err := c.BindJSON(request); err != nil {
		response.Code = types.ParamInvalid
		c.JSON(http.StatusOK, response)
		return
	}

	theUser := new(types.TMember)
	err := initial.Db.First(theUser, request.UserID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Code = types.UserNotExisted
		c.JSON(http.StatusOK, response)
		return
	}

	toUpdate := &types.TMember{
		IsDeleted: true,
	}
	if err := initial.Db.Where(request.UserID).Updates(toUpdate).Error; err != nil {
		response.Code = types.UnknownError
		c.JSON(http.StatusOK, response)
		return
	}
	response.Code = types.OK
	c.JSON(http.StatusOK, response)
}
