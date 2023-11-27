package main

import (
	"os"
	"fmt"
	"time"
	"strings"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"github.com/dgrijalva/jwt-go"
)

type User struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Token    string   `json:"token"`
	DocsID   []string `json:"docsID"`
}

var users = make(map[string]User)
var userDocs []map[string]string

// crear una variable userDocs que sea una lista para ir añadiendo strings

func createToken(username string) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(5 * time.Minute).Unix(),
	})

	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {

		return "", err
	}

	return tokenString, nil

}

func checkExp(c *gin.Context, userToken string, expired *bool) {

	tokenString := c.GetHeader("Authorization")
	if tokenString == "" || tokenString == "token" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not given."})
		*expired = true
		return
	}

	fmt.Println("Entering checkExp")

	parts := strings.Split(tokenString, " ")
	if len(parts) != 2 || parts[0] != "token" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wrong format for token."})
		*expired = true
		c.Abort()
		return
	}

	tokenString = parts[1]
	validatedToken, err := validateToken(tokenString)

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {

				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is malformed."})
				c.Abort()
				*expired = true
				return
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {

				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired or not valid yet, please go to /login to get a new access token."})
				c.Abort()
				*expired = true
				return
			} else {

				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not valid."})
				c.Abort()
				*expired = true
				return
			}
		}

		c.Abort()
		*expired = true
		return
	}

	claims, ok := validatedToken.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error getting claims."})
		c.Abort()
		*expired = true
		return
	}

	expTime := time.Unix(int64(claims["exp"].(float64)), 0)
	if time.Now().After(expTime) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired, you need to acces /login again."})
		c.Abort()
		*expired = true
		return
	}

	if strings.Split(userToken, " ")[1] != tokenString {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token for this user."})
		c.Abort()
		*expired = true
		return
	}

}

func validateToken(tokenString string) (*jwt.Token, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {

			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("secret"), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {

		return nil, fmt.Errorf("token is not valid")
	}

	return token, err

}

func openFile(username string, doc_id string) {

	jsonFile := "./cmd/APIRest/docs/" + username + "/" + doc_id + ".json"

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

}

func writeFile(username string, doc_id string, bodyContent []byte, bytesWriten *int) {

	jsonFile := "./cmd/APIRest/docs/" + username + "/" + doc_id + ".json"
	fmt.Println(jsonFile)
	file, err := os.OpenFile(jsonFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY|os.O_APPEND, 0644)
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

	var user User
	c.BindJSON(&user)

	if user.Username == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request."})
		return
	}

	if _, exists := users[user.Username]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists."})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password."})
		return
	}

	tokenString, err := createToken(user.Username)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating token."})
		return
	}

	// fmt.Print(tokenString)

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

	
	if user.Username == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request or wrong format (must be {})."})
		return
	}

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

	tokenString, err := createToken(user.Username)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating token."})
		return
	}
	user.Username = users[user.Username].Username
	user.Password = users[user.Username].Password
	user.Token = tokenString
	user.DocsID = users[user.Username].DocsID

	users[user.Username] = user

	c.IndentedJSON(http.StatusOK, gin.H{"access_token": tokenString})

}

func authentification(tokenString string, user User, c *gin.Context) bool {

	if user.Token == "token" || user.Token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not logged in."})
		return false
	}

	if tokenString == "" || tokenString == "token" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not given."})
		return false
	}

	if user.Token != tokenString[6:] {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token for this user."})
		return false
	}
	expired := false
	checkExp(c, tokenString, &expired)

	if expired {
		return false
	}

	if _, exists := users[user.Username]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found."})
		return false
	}
	return true
}

func getDocs(c *gin.Context) {

	tokenString := c.GetHeader("Authorization")
	Username := c.Param("username")
	DocID := c.Param("doc_id")

	if !authentification(tokenString, users[Username], c) {
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

	c.JSON(http.StatusOK, userDocs)
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

func putDocs(c *gin.Context) {

	tokenString := c.GetHeader("Authorization")
	Username := c.Param("username")
	DocID := c.Param("doc_id")
	flag := false

	for _, str := range users[Username].DocsID {
		if str == DocID {
			flag = true
			break
		}
	}

	if !flag {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wrong id for this user, doc not found."})
		return
	}

	if !authentification(tokenString, users[Username], c) {
		return
	}

	bodyContent, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request."})
		return
	}

	var bytesWriten int

	writeFile(Username, DocID, bodyContent, &bytesWriten)

	c.JSON(http.StatusOK, gin.H{"size": bytesWriten})
}

func postDocs(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	Username := c.Param("username")
	DocID := c.Param("doc_id")
	fmt.Print(DocID)

	for _, str := range users[Username].DocsID {
		if str == DocID {
			c.JSON(http.StatusConflict, gin.H{"error": "Doc already exists."})
			return
		}
	}

	if !authentification(tokenString, users[Username], c) {
		return
	}

	bodyContent, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request."})
		return
	}

	if bodyContent[0] != '[' || bodyContent[len(bodyContent)-1] != ']' {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Wrong format, must be [{}]."})
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

func deleteDocs(c *gin.Context) {

	tokenString := c.GetHeader("Authorization")
	Username := c.Param("username")
	DocID := c.Param("doc_id")
	flag := false

	if !authentification(tokenString, users[Username], c) {
		return
	}

	for _, str := range users[Username].DocsID {
		if str == DocID {
			flag = true
			break
		}
	}

	if !flag {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wrong id for this user, doc not found."})
		return
	}

	jsonFile := "./cmd/APIRest/docs/" + Username + "/" + DocID + ".json"
	err := os.Remove(jsonFile)
	if err != nil {
		fmt.Println(err)

	}

	c.JSON(http.StatusOK, gin.H{})
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

func getAllDocsFromUser(c *gin.Context) {

	allDocs := make(map[string]map[string]interface{})

	tokenString := c.GetHeader("Authorization")
	Username := c.Param("username")

	if !authentification(tokenString, users[Username], c) {
		return
	}

	for _, str := range users[Username].DocsID {

		openFile(Username, str)

		for _, i := range userDocs {
			docInterface := make(map[string]interface{})

			for key, value := range i {

				docInterface[key] = value
			}
			allDocs[str] = docInterface

		}
		userDocs = nil

	}

	c.JSON(http.StatusOK, allDocs)

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

	router := gin.Default()

	createDirectories()
	importUsers()
	fmt.Println("Users in the system:")
	for _, user := range users {
		fmt.Println(user.Username)
	}

	router.GET("/:username/:doc_id", getDocs)
	router.GET("/version", getVersion)
	router.POST("/signup", signUp)
	router.POST("/login", login)
	router.POST("/:username/:doc_id", postDocs)
	router.PUT("/:username/:doc_id", putDocs)
	router.DELETE("/:username/:doc_id", deleteDocs)
	router.GET("/:username/_all_docs", getAllDocsFromUser)

	err := http.ListenAndServeTLS("myserver.local:5000", "certificates/myserver.local.pem", "certificates/myserver.local-key.pem", router)
	if err != nil {
		fmt.Println(err)
		return
	}

}
