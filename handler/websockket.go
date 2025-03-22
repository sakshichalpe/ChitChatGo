package handler

import (
	"log"
	"net/http"
	"realtimechatttask/db"
	"realtimechatttask/model"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// convert http req to websocket conn
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[string]*websocket.Conn)
var clientsMutex sync.Mutex

func HandleWebSocket(c *gin.Context) {
	username := c.GetString("username")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer conn.Close()

	clientsMutex.Lock()
	clients[username] = conn
	clientsMutex.Unlock()

	defer func() {
		clientsMutex.Lock()
		delete(clients, username)
		clientsMutex.Unlock()
	}()

	for {
		var msg model.Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("Read Error:", err)
			break
		}
		msg.Timestamp = time.Now()
		_, err = db.DB.Exec("INSERT INTO messages (sender, receiver, content, timestamp) VALUES ($1, $2, $3, $4)", msg.Sender, msg.Receiver, msg.Content, msg.Timestamp)
		if err != nil {
			log.Println("Database Insert Error:", err)
		}

		clientsMutex.Lock()
		receiverConn, exists := clients[msg.Receiver]
		clientsMutex.Unlock()

		if exists {
			receiverConn.WriteJSON(msg)
		}
	}
}
