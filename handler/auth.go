package handler

import (
	"fmt"
	"net/http"
	"realtimechatttask/model"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var JwtKey = []byte("secret_key")

func GenerateToken(username string) (string, error) {
	//expirationTime := time.Now().Add(24 * time.Hour)
	claims := &model.Claims{
		Username:       username,
		StandardClaims: jwt.StandardClaims{
			//ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	fmt.Println("token::20:", token) //Prints the token object before it's converted into a string
	return token.SignedString(JwtKey)
}

func CreateJWTToken(c *gin.Context) {
	username := c.PostForm("username")

	token, err := GenerateToken(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}
	fmt.Println("token::", token)
	c.JSON(http.StatusOK, gin.H{"token": token})
}
