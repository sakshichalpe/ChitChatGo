package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"realtimechatttask/db"
	"realtimechatttask/handler"
	"realtimechatttask/middleware"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	db.InitDB()
	r := gin.Default()

	r.POST("/createJWTToken", handler.CreateJWTToken)
	r.GET("/ws", middleware.AuthenticateJWT, handler.HandleWebSocket)
	r.GET("/history/:sender/:receiver", middleware.AuthenticateJWT, handler.GetChatHistory)

	//r.Run()
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
