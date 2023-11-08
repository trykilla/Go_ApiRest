package main

// import "fmt"
import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
)

//import "strings"

type User struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Token    string   `json:"token"`
	DocsID   []string `json:"docs"`
}

var users = make(map[string]User)
var userDocs []map[string]string

//const token = "token 1234"

func openFile(doc_id string) {

	jsonFile := "docs/" + doc_id + ".json"
	fmt.Println(jsonFile)
	file, err := os.Open(jsonFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)

	if err != nil {
		fmt.Println(err)
		return
	}

	err = json.Unmarshal(byteValue, &userDocs)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(userDocs)
}

func getVersion(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"version": "0.0.1"})
}

func signUp(c *gin.Context) {

	tokenString := c.GetHeader("Authorization")

	// if tokenString == token {
	// 	fmt.Println("Token correcto")
	// }

	var user User
	c.BindJSON(&user)

	if _, exists := users[user.Username]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists."})
		return
	}

	user.Token = tokenString
	users[user.Username] = user

	c.IndentedJSON(http.StatusOK, gin.H{"access_token": tokenString})
}

func login(c *gin.Context) {
	var user User
	c.BindJSON(&user)

	if _, exists := users[user.Username]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found."})
		return
	}

	if users[user.Username].Password != user.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password."})
		return
	}

	var token string
	if users[user.Username].Token != "" {
		token = users[user.Username].Token
	} else {
		user = users[user.Username]
		token = "token 1234"
		user.Token = token
		users[user.Username] = user
	}

	c.IndentedJSON(http.StatusOK, gin.H{"access_token": token})

}

func authentification(tokenString string, Username string, c *gin.Context) bool {

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not found or token expired."})
		return false
	}

	if tokenString != users[Username].Token {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token for this user."})
		return false
	}

	if _, exists := users[Username]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found."})
		return false
	}
	return true
}

func getDocs(c *gin.Context) {

	tokenString := c.GetHeader("Authorization")
	Username := c.Param("username")
	DocID := c.Param("doc_id")

	if !authentification(tokenString, Username, c) {
		return
	}

	var doc string

	for i, str := range users[Username].DocsID {
		if str == DocID {
			doc = users[Username].DocsID[i]
			break
		}
	}

	for _, i := range userDocs {

		if value, ok := i[doc]; ok {
			doc = value
			break
		}

	}

	openFile(doc)

	if doc == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wrong id for this user, doc not found."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"doc": userDocs})
	userDocs = nil

}

func postDocs(c *gin.Context) {

}

func main() {

	//fmt.Println("Hello, World!") // prints "Hello, World!"
	router := gin.Default()

	exUser := User{
		Username: "user1",
		Password: "pass1",
		Token:    "token 1234",
		DocsID:   []string{"patatas", "3"},
	}

	users[exUser.Username] = exUser
	// router.GET("/cars/:car", getCars)
	router.GET("/:username/:doc_id", getDocs)
	router.GET("/version", getVersion)
	router.POST("/signup", signUp)
	router.POST("/login", login)
	router.Run("myserver.local:5000")

}
