package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type User struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Token    string   `json:"token"`
	DocsID   []string `json:"docsID"`
}

var userDocs map[string]string

func getDocs(c *gin.Context) {

	tokenString := c.GetHeader("Authorization")
	Username := c.Param("username")
	DocID := c.Param("doc_id")
	docsId := c.Query("docsID")
	userToken := c.Query("token")
	userPass := c.Query("password")

	docsIdSlice := strings.Split(docsId, ",")

	user := User{
		Username: Username,
		Password: userPass,
		Token:    userToken,
		DocsID:   docsIdSlice,
	}

	fmt.Println("user:", user)

	if !authentification(tokenString, user, c) {
		return
	}

	var doc string

	for i, str := range user.DocsID {
		if str == DocID {
			doc = user.DocsID[i]
			break
		}
	}

	if doc == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wrong id for this user, doc not found or removed."})
		return
	}

	openFile(Username, doc)

	if doc == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wrong id for this user, doc not found."})
		return
	}

	c.JSON(http.StatusOK, userDocs)
	userDocs = nil

}

func postDocs(c *gin.Context) {

	tokenString := c.GetHeader("Authorization")
	Username := c.Param("username")
	DocID := c.Param("doc_id")
	docsId := c.Query("docsID")
	userToken := c.Query("token")
	userPass := c.Query("password")
	// fmt.Println("tokenString:", tokenString)
	// fmt.Println("Username:", Username)
	// fmt.Println("DocID:", DocID)

	// fmt.Println("docsId:", docsId)
	docsIdSlice := strings.Split(docsId, ",")

	for _, docId := range docsIdSlice {
		if docId == DocID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Doc already exists."})
			return
		}
	}

	user := User{
		Username: Username,
		Password: userPass,
		Token:    userToken,
		DocsID:   docsIdSlice,
	}

	fmt.Println("user:", user)

	if !authentification(tokenString, user, c) {
		return
	}

	bodyContent, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request."})
		return
	}

	user.DocsID = append(user.DocsID, DocID)

	var bytesWriten int

	writeFile(Username, DocID, bodyContent, &bytesWriten)


	fmt.Println("user after postin doc:", user)

	//insertDocs(Username, DocID)


	enviarInformacionAlBroker(user)

	c.JSON(http.StatusOK, gin.H{"size": bytesWriten})

}

func putDocs(c *gin.Context) {

	tokenString := c.GetHeader("Authorization")
	Username := c.Param("username")
	DocID := c.Param("doc_id")
	docsId := c.Query("docsID")
	userToken := c.Query("token")
	userPass := c.Query("password")
	fmt.Println("tokenString:", tokenString)
	fmt.Println("Username:", Username)
	fmt.Println("DocID:", DocID)

	fmt.Println("docsId:", docsId)
	docsIdSlice := strings.Split(docsId, ",")
	flag := false

	for _, str := range docsIdSlice {
		if str == DocID {
			flag = true
			break
		}
	}

	if !flag {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wrong id for this user, doc not found."})
		return
	}

	user := User{
		Username: Username,
		Password: userPass,
		Token:    userToken,
		DocsID:   docsIdSlice,
	}
	if !authentification(tokenString, user, c) {
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

func deleteDocs(c *gin.Context) {

	tokenString := c.GetHeader("Authorization")
	Username := c.Param("username")
	DocID := c.Param("doc_id")
	docsId := c.Query("docsID")
	userToken := c.Query("token")
	userPass := c.Query("password")
	fmt.Println("tokenString:", tokenString)
	fmt.Println("Username:", Username)
	fmt.Println("DocID:", DocID)
	flag := false

	user := User{
		Username: Username,
		Password: userPass,
		Token:    userToken,
		DocsID:   strings.Split(docsId, ","),
	}

	if !authentification(tokenString, user, c) {
		return
	}

	for _, str := range user.DocsID {
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

	for i, str := range user.DocsID {
		if str == DocID {
			user.DocsID[i] = user.DocsID[len(user.DocsID)-1]
			user.DocsID[len(user.DocsID)-1] = ""
			user.DocsID = user.DocsID[:len(user.DocsID)-1]
			break
		}
	}

	fmt.Println("user_after_delete", user)

	overWriteUser(user)

	enviarInformacionAlBroker(user)

	c.JSON(http.StatusOK, gin.H{})
}

func getAllDocsFromUser(c *gin.Context) {

	allDocs := make(map[string]map[string]string)

	tokenString := c.GetHeader("Authorization")
	Username := c.Param("username")

	docsId := c.Query("docsID")
	userToken := c.Query("token")
	userPass := c.Query("password")
	fmt.Println("tokenString:", tokenString)
	fmt.Println("Username:", Username)

	var docsIdSlice []string

	if len(docsId) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No docs for this user"})
		return
	} else if len(docsId) != 1 {
		fmt.Println("docsId:", docsId)
		docsIds := strings.Split(docsId, ",")
		for _, docId := range docsIds {
			docsIdSlice = append(docsIdSlice, strings.TrimSpace(docId))
		}

	} else {
		docsIdSlice = []string{strings.TrimSpace(docsId)}

	}

	fmt.Println("docsIdSlice:", docsIdSlice)

	user := User{
		Username: Username,
		Password: userPass,
		Token:    userToken,
		DocsID:   docsIdSlice,
	}

	if !authentification(tokenString, user, c) {
		return
	}

	print("docs_ids", user.DocsID)

	for _, doc := range user.DocsID {
		openFile(Username, doc)
		allDocs[doc] = userDocs
		userDocs = nil
	}

	print("allDocs:", allDocs)

	if len(allDocs) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No docs found for this user."})
		return
	}

	c.JSON(http.StatusOK, allDocs)

}

func overWriteUser(user User) {

	existingUsers := readUsersFromFile()

	for i, existingUser := range existingUsers {
		if strings.EqualFold(existingUser.Username, user.Username) {
			existingUsers[i] = user
		}
	}

	updateJsonFile, err := json.MarshalIndent(existingUsers, "", " ")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ioutil.WriteFile("cmd/APIRest/users.json", updateJsonFile, 0644)
	if err != nil {

		fmt.Println(err)
		return
	}

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

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	// Realiza la solicitud POST al servicio broker
	url := "https://myserver.local:5000/auth_rec"
	_, err1 := client.Post(url, "application/json", strings.NewReader(string(userJSON)))
	if err1 != nil {
		fmt.Println("Error sending information to broker:", err)
		return
	}
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

// func insertDocs(Username string, DocID string) {

// 	tempUsers := readUsersFromFile()

// 	for i, user := range tempUsers {
// 		if user.Username == Username {
// 			tempUsers[i].DocsID = append(tempUsers[i].DocsID, DocID)
// 			break
// 		}
// 	}

// 	updateJsonFile, err := json.MarshalIndent(tempUsers, "", " ")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

	

// 	err = ioutil.WriteFile("./cmd/APIRest/users.json", updateJsonFile, 0644)
// 	if err != nil {

// 		fmt.Println(err)
// 		return
// 	}

// }



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

func writeFile(username string, doc_id string, bodyContent []byte, bytesWriten *int) {

	parentRoute := "./cmd/APIRest/docs/" + username + "/"
	route_err := os.MkdirAll(parentRoute, 0777)

	if route_err != nil {
		fmt.Println("Error creating directory", route_err)
		return
	}

	jsonFile := "./cmd/APIRest/docs/" + username + "/" + doc_id + ".json"

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

	return !expired

}

func checkExp(c *gin.Context, userToken string, expired *bool) {

	tokenString := c.GetHeader("Authorization")
	if tokenString == "" || tokenString == "token" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not given."})
		*expired = true
		return
	}

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

// ...

func main() {

	router := gin.Default()
	router.DELETE("/files/:username/:doc_id", deleteDocs)
	router.POST("/files/:username/:doc_id", postDocs)
	router.PUT("/files/:username/:doc_id", putDocs)
	router.GET("/files/:username/:doc_id", getDocs)
	router.GET("/files/:username/_all_docs", getAllDocsFromUser)
	err := http.ListenAndServeTLS("10.0.2.4:8082", "certificates/files.pem", "certificates/files-key.pem", router)
	if err != nil {
		fmt.Println(err)
		return
	}
}
