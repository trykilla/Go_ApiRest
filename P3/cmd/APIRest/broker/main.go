package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

const (
	authServiceURL  = "http://localhost:8084/auth"
	filesServiceURL = "http://localhost:8082/files"
)

type User struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Token    string   `json:"token"`
	DocsID   []string `json:"docsID"`
}

var users = make(map[string]User)

func main() {
	router := gin.Default()

	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": "1.0.0"})
	})
	router.POST("/login", reverseProxy(authServiceURL))
	router.POST("/signup", reverseProxy(authServiceURL))
	router.POST("/auth_rec", manageAuthRec)

	// esperar 1 minut

	router.Run(":8080")

	//esperar un minuto

}

func manageAuthRec(c *gin.Context) {
	var user User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		fmt.Println("Error receiving information from auth service:", err)
		return
	}

	users[user.Username] = user

	fmt.Println("Received information from auth service:", user)
}

func reverseProxy(targetURL string) gin.HandlerFunc {

	target, err := url.Parse(targetURL)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)

	}

}
