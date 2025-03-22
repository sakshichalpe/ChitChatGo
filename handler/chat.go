package handler

import (
	"net/http"
	"realtimechatttask/db"
	"realtimechatttask/model"

	"github.com/gin-gonic/gin"
)

func GetChatHistory(c *gin.Context) {
	sender := c.Param("sender")
	receiver := c.Param("receiver")

	rows, err := db.DB.Query("SELECT sender, receiver, content, timestamp FROM messages WHERE (sender=$1 AND receiver=$2) OR (sender=$2 AND receiver=$1) ORDER BY timestamp DESC", sender, receiver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var msg model.Message
		rows.Scan(&msg.Sender, &msg.Receiver, &msg.Content, &msg.Timestamp)
		messages = append(messages, msg)
	}

	c.JSON(http.StatusOK, messages)
}
