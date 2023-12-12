package main

import (
	"github.com/gin-gonic/gin"

	"io"
	"fmt"
	"strings"
	"net/http"
	"encoding/json"

)

const (
	authServiceURL  = "http://myserver.local:8084/auth"
	filesServiceURL = "http://myserver.local:8082/files"
)

type User struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Token    string   `json:"token"`
	DocsID   []string `json:"docsID"`
}

var users = make(map[string]User)

func redirectToService(c *gin.Context, targetURL string, variables map[string]string) {
	client := &http.Client{}
	originalRequest := c.Request
	newRequest, err := http.NewRequest(originalRequest.Method, targetURL+originalRequest.URL.Path, originalRequest.Body)
	if err != nil {
		fmt.Println("Error creating new request:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating new request"})
		return
	}
	newRequest.Header = originalRequest.Header

	if variables != nil {
		queryParams := newRequest.URL.Query()
		for k, v := range variables {
			queryParams.Add(k, v)
		}
		newRequest.URL.RawQuery = queryParams.Encode()
	}

	response, err := client.Do(newRequest)
	if err != nil {
		fmt.Println("Error sending request:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending request"})
		return
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading response body"})
		return
	}

	responseBodyJson := make(map[string]interface{})
	if err := json.Unmarshal(responseBody, &responseBodyJson); err != nil {
		fmt.Println("Error unmarshalling response body:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unmarshalling response body"})
		return
	}

	c.JSON(response.StatusCode, responseBodyJson)
}

func handleBrokerRoute(c *gin.Context) {
	serviceName := determineService(c)

	switch serviceName {
	case "auth":

		redirectToService(c, authServiceURL, nil)
	case "files":
		username := c.Param("username")

		user, ok := users[username]
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found."})
			return

		}

		docsIDString := strings.Join(user.DocsID, ",")

		variables := map[string]string{
			"username": user.Username,
			"doc_id":   c.Param("doc_id"),
			"token":    user.Token,
			"docsID":   docsIDString,
		}

		redirectToService(c, filesServiceURL, variables)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Service not found"})
	}
}

func determineService(c *gin.Context) string {
	if c.Request.URL.Path == "/login" || c.Request.URL.Path == "/signup" {
		return "auth"
	} else {
		return "files"
	}
}

func main() {
	router := gin.Default()

	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": "1.0.0"})
	})
	router.POST("/:username/:doc_id", handleBrokerRoute)
	router.POST("/login", handleBrokerRoute)
	// router.POST("/signup", reverseProxy(authServiceURL))
	// router.POST("/:username/:doc_id", reverseProxy(filesServiceURL))
	router.POST("/auth_rec", manageAuthRec)

	// esperar 1 minut

	router.Run("myserver.local:5000")

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

	fmt.Println("Received information from auth service:", users)

}

// func reverseProxy(targetURL string) gin.HandlerFunc {

// 	target, err := url.Parse(targetURL)
// 	if err != nil {
// 		panic(err)
// 	}

// 	proxy := httputil.NewSingleHostReverseProxy(target)

// 	return func(c *gin.Context) {

// 		usersCopy := make(map[string]User)
// 		for k, v := range users {
// 			usersCopy[k] = v
// 		}

// 		ctx := context.WithValue(c.Request.Context(), "users", usersCopy)
// 		req := c.Request.WithContext(ctx)
// 		c.Request = req

// 		proxy.ServeHTTP(c.Writer, c.Request)

// 	}

// }
