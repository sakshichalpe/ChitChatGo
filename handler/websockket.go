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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Track active WebSocket connections
var (
	mu              sync.Mutex
	userConnections sync.Map // Stores active WebSocket connections per user
	userMessages    sync.Map // Stores user message count
	userResetTime   sync.Map // Stores reset time per user
	messageLimit    = 5
	resetTime       = 10 * time.Second
)

func RateLimitCheck(username string) bool {
	mu.Lock()
	defer mu.Unlock()

	msgCount, _ := userMessages.LoadOrStore(username, 0)
	lastReset, _ := userResetTime.LoadOrStore(username, time.Now())

	// Reset if time exceeded
	if time.Since(lastReset.(time.Time)) > resetTime {
		log.Printf("Resetting rate limit for user %s", username)
		userMessages.Store(username, 0)
		userResetTime.Store(username, time.Now())
	}

	if msgCount.(int) >= messageLimit {
		log.Printf("User %s exceeded rate limit", username)
		return false
	}

	userMessages.Store(username, msgCount.(int)+1)
	log.Printf("User %s sent %d messages", username, msgCount.(int)+1)
	return true
}

func HandleWebSocket(c *gin.Context) {
	username := c.GetString("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer conn.Close()

	log.Printf("User %s connected", username)
	userConnections.Store(username, conn)

	for {
		var msg model.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Read Error:", err)
			break
		}

		// Apply rate limiting for sender
		if !RateLimitCheck(msg.Sender) {
			conn.WriteJSON(gin.H{"error": "Rate limit exceeded: Max 5 messages per 10 sec"})
			continue
		}

		// Store message in DB
		_, err = db.DB.Exec("INSERT INTO messages (sender, receiver, content, timestamp) VALUES ($1, $2, $3, $4)", msg.Sender, msg.Receiver, msg.Content, time.Now())
		if err != nil {
			log.Println("DB Insert Error:", err)
			continue
		}

		log.Printf("Message from %s to %s: %s", msg.Sender, msg.Receiver, msg.Content)

		// Send message to receiver if online with rate limit check
		if receiverConn, ok := userConnections.Load(msg.Receiver); ok {
			if RateLimitCheck(msg.Receiver) {
				receiverConn.(*websocket.Conn).WriteJSON(msg)
			} else {
				log.Printf("User %s exceeded rate limit and cannot receive messages", msg.Receiver)
			}
		} else {
			log.Printf("User %s is offline, message stored for later retrieval", msg.Receiver)
		}
	}

	// Remove user from active connections when they disconnect
	userConnections.Delete(username)
}
