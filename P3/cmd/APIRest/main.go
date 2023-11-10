package main

// import "fmt"
import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"net/http"
	"os"
)

type User struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Token    string   `json:"token"`
	DocsID   []string `json:"docsID"`
}

var users = make(map[string]User)
var userDocs []map[string]string

func openFile(username string, doc_id string) {

	jsonFile := "./cmd/APIRest/docs/" + username + "/" + doc_id + ".json"
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

func writeFile(username string, doc_id string, bodyContent []byte, bytesWriten *int) {

	jsonFile := "./cmd/APIRest/docs/" + username + "/" + doc_id + ".json"
	fmt.Println(jsonFile)
	file, err := os.Create(jsonFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()

	*bytesWriten, err = file.Write(bodyContent)
	if err != nil {
		fmt.Println(err)
		return
	}

}

func getVersion(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"version": "0.0.1"})
}

func signUp(c *gin.Context) {

	// tokenString := c.GetHeader("Authorization")
	tokenString := "token"
	// if tokenString == token {
	// 	fmt.Println("Token correcto")
	// }

	var user User
	c.BindJSON(&user)

	if _, exists := users[user.Username]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists."})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password."})
		return
	}

	user.Password = string(hashedPassword)
	user.Token = tokenString
	user.DocsID = make([]string, 0)
	users[user.Username] = user

	insertUser(user)

	c.IndentedJSON(http.StatusOK, gin.H{"access_token": tokenString})
}

func login(c *gin.Context) {
	var user User
	c.BindJSON(&user)

	if _, exists := users[user.Username]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found."})
		return
	}

	hashedPassword := users[user.Username].Password
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))

	if err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password."})
		return
	}

	var token string
	if users[user.Username].Token != "token " {
		token = users[user.Username].Token
	} else {
		user = users[user.Username]
		user.Token += "1234"
		token = user.Token
		users[user.Username] = user
	}

	c.IndentedJSON(http.StatusOK, gin.H{"access_token": token})

}

func authentification(tokenString string, Username string, c *gin.Context) bool {

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not found or token expired."})
		return false
	}

	fmt.Println("Token string", tokenString)
	fmt.Println("User token", users[Username].Token)

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
	fmt.Print(users[Username].DocsID)
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

	openFile(Username, doc)

	if doc == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wrong id for this user, doc not found."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"doc": userDocs})
	userDocs = nil

}

func insertUser(user User) {

	tempUsers := readUsersFromFile()

	tempUsers = append(tempUsers, user)

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

	parentRoute := "./cmd/APIRest/docs/" + user.Username + "/"
	route_err := os.MkdirAll(parentRoute, 0777)

	if route_err != nil {
		fmt.Println("Error creating directory", err)
		return
	}

}

func postDocs(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	Username := c.Param("username")
	DocID := c.Param("doc_id")
	fmt.Print(DocID)

	if !authentification(tokenString, Username, c) {
		return
	}

	bodyContent, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request."})
		return
	}

	user := users[Username]
	user.DocsID = append(user.DocsID, DocID)
	users[Username] = user

	var bytesWriten int

	fmt.Print(users[Username].DocsID)

	writeFile(Username, DocID, bodyContent, &bytesWriten)

	insertDocs(Username, DocID)

	c.JSON(http.StatusOK, gin.H{"size": bytesWriten})

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

func createDirectories() {
	parentRoute := "./cmd/APIRest/docs"
	err := os.MkdirAll(parentRoute, 0777)

	if err != nil {
		fmt.Println("Error creating directory", err)
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

func importUsers() {

	tempUsers := readUsersFromFile()
	for _, user := range tempUsers {

		users[user.Username] = user

	}

}

func main() {

	//fmt.Println("Hello, World!") // prints "Hello, World!"
	router := gin.Default()

	createDirectories()
	importUsers()
	fmt.Println("Users in the system:")
	for _, user := range users {
		fmt.Println(user.Username)
	}

	// router.GET("/cars/:car", getCars)
	router.GET("/:username/:doc_id", getDocs)
	router.GET("/version", getVersion)
	router.POST("/signup", signUp)
	router.POST("/login", login)
	router.POST("/:username/:doc_id", postDocs)

	err := http.ListenAndServeTLS("myserver.local:5000", "myserver.local.pem", "myserver.local-key.pem", router)
	if err != nil {
		fmt.Println(err)
		return
	}

}
