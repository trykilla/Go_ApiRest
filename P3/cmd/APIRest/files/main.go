package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"net/http"
)

type User struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Token    string   `json:"token"`
	DocsID   []string `json:"docsID"`
}

// ...

func postDocs(c *gin.Context) {

	tokenString := c.GetHeader("Authorization")
	Username := c.Param("username")
	DocID := c.Param("doc_id")
	fmt.Println("tokenString:", tokenString)
	fmt.Println("Username:", Username)
	fmt.Println("DocID:", DocID)
	
	// username := c.Query("username")
	// fmt.Println("username:", username)
	// doc_id := c.Query("doc_id")
	// fmt.Println("doc_id:", doc_id)
	// token := c.Query("token")
	// fmt.Println("token:", token)
	// docs_id := c.Query("docsID")
	// fmt.Println("docs_id:", docs_id)

	// haz un json con los datos del usuario

	

	c.JSON(http.StatusOK, gin.H{"size": ""})

}

// ...

func main() {

	router := gin.Default()
	router.POST("/files/:username/:doc_id", postDocs)
	router.Run("myserver.local:8082")
}
