package controller

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"
	"message-pusher/channel"
	"message-pusher/common"
	"message-pusher/model"
	"net/http"
)

func GetPushMessage(c *gin.Context) {
	message := channel.Message{
		Title:       c.Query("title"),
		Description: c.Query("description"),
		Content:     c.Query("content"),
		URL:         c.Query("url"),
		Channel:     c.Query("channel"),
		Token:       c.Query("token"),
	}
	if message.Description == "" {
		// Keep compatible with ServerChan
		message.Description = c.Query("desp")
	}
	pushMessageHelper(c, &message)
}

func PostPushMessage(c *gin.Context) {
	message := channel.Message{}
	err := json.NewDecoder(c.Request.Body).Decode(&message)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无法解析请求体，请检查其是否为合法 JSON",
		})
		return
	}
	pushMessageHelper(c, &message)
}

func pushMessageHelper(c *gin.Context, message *channel.Message) {
	user := model.User{Username: c.Param("username")}
	user.FillUserByUsername()
	if user.Status == common.UserStatusNonExisted {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "用户不存在",
		})
		return
	}
	if user.Status == common.UserStatusDisabled {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "用户已被封禁",
		})
		return
	}
	if user.Token != "" {
		if message.Token == "" {
			message.Token = c.Request.Header.Get("Authorization")
		}
		if user.Token != message.Token {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "无效的 token",
			})
			return
		}
	}
	if message.Title == "" {
		message.Title = common.SystemName
	}
	if message.Content != "" {
		var buf bytes.Buffer
		err := goldmark.Convert([]byte(message.Content), &buf)
		if err != nil {
			common.SysLog(err.Error())
		} else {
			message.Content = buf.String()
		}
	} else {
		if message.Description != "" {
			message.Content = message.Description
		} else {
			message.Content = "无内容"
		}
	}
	if message.Channel == "" {
		message.Channel = user.Channel
		if message.Channel == "" {
			message.Channel = channel.TypeEmail
		}
	}
	err := message.Send(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
	})
	return
}
