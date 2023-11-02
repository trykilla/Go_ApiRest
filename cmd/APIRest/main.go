package main

// import "fmt"
import "github.com/gin-gonic/gin"
import "net/http"
import "fmt"

//import "strings"

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var users = make(map[string]User)
const token = "1234"

func getVersion(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"version": "0.0.1"})
}

func signUp(c *gin.Context) {

	tokenString := c.GetHeader("Authorization")

	if tokenString == token {
		fmt.Println("Token correcto")
	}

	var user User
	c.BindJSON(&user)

	if _, exists := users[user.Username]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists."})
		return
	}

	users[user.Username] = user

	c.IndentedJSON(http.StatusOK, gin.H{"message": "User created"})
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

	c.IndentedJSON(http.StatusOK, gin.H{"access_token": token})

}

// func getCars(c *gin.Context) {

// 	Carro := c.Param("car")
// 	CarroMinus := strings.ToLower(Carro)

// 	if CarroMinus == "focus" {
// 		cars := Car{Brand: "Ford", Model: "Focus", Year: 2010, Color: "Red"}
// 		c.IndentedJSON(http.StatusOK, cars)
// 		return
// 	}
// cars := []Car{
// 	{Brand: "Ford", Model: "Focus", Year: 2010, Color: "Red"},
// 	{Brand: "Ford", Model: "Fiesta", Year: 2012, Color: "Blue"},
// 	{Brand: "Ford", Model: "Mustang", Year: 2015, Color: "Yellow"},
// 	{Brand: "Ford", Model: "F150", Year: 2017, Color: "Black"},
// }
// 	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Car not found"})
// }

func main() {
	//fmt.Println("Hello, World!") // prints "Hello, World!"
	router := gin.Default()
	// router.GET("/cars/:car", getCars)
	router.GET("/version", getVersion)
	router.POST("/signup", signUp)
	router.POST("/login", login)
	router.Run("myserver.local:5000")

}
