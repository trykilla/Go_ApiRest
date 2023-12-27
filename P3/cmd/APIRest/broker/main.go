package main

import (
	"crypto/tls"
	"io/ioutil"
	"os"

	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

const (
	authServiceURL  = "https://10.0.2.3:8084/auth"
	filesServiceURL = "https://10.0.2.4:8082/files"
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

	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
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
			"password": user.Password,
			"doc_id":   c.Param("doc_id"),
			"token":    user.Token,
			"docsID":   docsIDString,
		}
		// importUsers()
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
	importUsers()

	router := gin.Default()

	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": "1.0.0"})
	})
	router.POST("/:username/:doc_id", handleBrokerRoute)
	router.POST("/login", handleBrokerRoute)
	router.POST("/signup", handleBrokerRoute)
	router.GET("/:username/:doc_id", handleBrokerRoute)
	router.GET("/:username/_all_docs", handleBrokerRoute)
	router.PUT("/:username/:doc_id", handleBrokerRoute)
	router.DELETE("/:username/:doc_id", handleBrokerRoute)
	router.POST("/auth_rec", manageAuthRec)

	printColouredRoutes(router)
	// esperar 1 minut

	err := http.ListenAndServeTLS("myserver.local:5000", "certificates/myserver.local.pem", "certificates/myserver.local-key.pem", router)
	if err != nil {
		fmt.Println(err)
		return
	}
	//esperar un minuto

}

func insertDocs(Username string, DocID string) {

	tempUsers := readUsersFromFile()

	for i, user := range tempUsers {
		if user.Username == Username {
			tempUsers[i].DocsID = append(tempUsers[i].DocsID, DocID)
			break
		}
	}

	updateJsonFile, err := json.MarshalIndent(tempUsers, "", " ")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ioutil.WriteFile("./cmd/APIRest/users.json", updateJsonFile, 0644)
	if err != nil {

		fmt.Println(err)
		return
	}

}

func manageAuthRec(c *gin.Context) {
	var user User
	flag := false

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		fmt.Println("Error receiving information from auth service:", err)
		return
	}

	for _, us := range users {
		if user.Username == users[us.Username].Username {
			flag = true
			fmt.Println("user docs ids:", user.DocsID)
			if len(user.DocsID) == 2 {
				insertDocs(user.Username, user.DocsID[1])
			} else if len(user.DocsID) != 0 {

				for i, doc := range users[us.Username].DocsID {
					if doc == user.DocsID[i] {
						fmt.Println("Doc already exists")
						continue
					} else {
						insertDocs(user.Username, user.DocsID[i])
						break
					}
				}
			}
		}
	}

	users[user.Username] = user

	if !flag {
		fmt.Println("User does not exist, inserting...")
		insertUser(user)
	}

	fmt.Println("Received information from auth service:", users)

}

func importUsers() {

	tempUsers := readUsersFromFile()
	for _, user := range tempUsers {

		users[user.Username] = user

	}

}

func insertUser(user User) {

	tempUsers := readUsersFromFile()

	tempUsers = append(tempUsers, user)

	updateJsonFile, err := json.MarshalIndent(tempUsers, "", " ")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ioutil.WriteFile("cmd/APIRest/users.json", updateJsonFile, 0644)
	if err != nil {

		fmt.Println(err)
		return
	}

	parentRoute := "./cmd/APIRest/docs/" + user.Username + "/"
	route_err := os.MkdirAll(parentRoute, 0777)

	if route_err != nil {
		fmt.Println("Error creating directory", route_err)
		return
	}

}

func readUsersFromFile() (users []User) {
	jsonFile := "./cmd/APIRest/users.json"

	file, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Println("Error reading the users file", err)
		return
	}

	var tempUsers []User
	err = json.Unmarshal(file, &tempUsers)
	if err != nil {

		fmt.Println("Error unmarshalling the users file", err)
		return
	}

	return tempUsers

}

func printColouredRoutes(r *gin.Engine) {
	fmt.Println("Routes:")

	for _, route := range r.Routes() {
		method := route.Method
		path := route.Path

		switch method {
		case "GET":
			color.Green("%s %s", method, path)
		case "POST":
			color.Blue("%s %s", method, path)
		case "PUT":
			color.Yellow("%s %s", method, path)
		case "DELETE":
			color.Red("%s %s", method, path)
		default:
			color.White("%s %s", method, path)
		}
	}
	fmt.Println()
}
