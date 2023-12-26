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

func enviarInformacionAlBroker(user User) {
	// Aquí implementa la lógica para enviar información al servicio broker
	// Puedes utilizar bibliotecas como "net/http" para hacer una solicitud POST al servicio broker.

	// Ejemplo:
	userJSON, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Error marshalling user:", err)
		return
	}


    // Realiza la solicitud POST al servicio broker
    url := "https://myserver.local:5000/auth_rec"
    _, err1 := http.Post(url, "application/json", strings.NewReader(string(userJSON)))
    if err1 != nil {
        fmt.Println("Error sending information to broker:", err)
        return
    }
}

func signUp(c *gin.Context) {
	importUsers()
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

	user.Password = string(hashedPassword)
	user.Token = tokenString
	user.DocsID = make([]string, 0)
	users[user.Username] = user

	insertUser(user)

	enviarInformacionAlBroker(user)

	c.IndentedJSON(http.StatusOK, gin.H{"access_token": tokenString})
}

func login(c *gin.Context) {
	importUsers()
	var user User
	c.BindJSON(&user)

	if user.Username == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request or wrong format (must be {})."})
		return
	}

	if _, exists := users[user.Username]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found."})
		fmt.Println("User not found.")
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

	enviarInformacionAlBroker(user)
	fmt.Println(users)

	c.IndentedJSON(http.StatusOK, gin.H{"access_token": tokenString})

}

func importUsers() {

	tempUsers := readUsersFromFile()
	for _, user := range tempUsers {

		users[user.Username] = user

	}

}

func main() {

	importUsers()
	router := gin.Default()

	router.POST("/auth/signup", signUp)
	router.POST("/auth/login", login)

	err := http.ListenAndServeTLS("10.0.1.3:8084", "certificates/auth.pem", "certificates/auth-key.pem", router)
	if err != nil {
		fmt.Println(err)
		return
	}
}
