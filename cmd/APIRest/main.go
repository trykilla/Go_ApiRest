package main

// import "fmt"
import "github.com/gin-gonic/gin"
import "net/http"
import "strings"

type Car struct {
	Brand string `json:"brand"`
	Model string `json:"model"`
	Year  int    `json:"year"`
	Color string `json:"color"`
}

func getVersion(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"version": "0.0.1"})
}

func getCars(c *gin.Context) {

	Carro := c.Param("car")
	CarroMinus := strings.ToLower(Carro)

	if CarroMinus == "focus" {
		cars := Car{Brand: "Ford", Model: "Focus", Year: 2010, Color: "Red"}
		c.IndentedJSON(http.StatusOK, cars)
		return
	}
	// cars := []Car{
	// 	{Brand: "Ford", Model: "Focus", Year: 2010, Color: "Red"},
	// 	{Brand: "Ford", Model: "Fiesta", Year: 2012, Color: "Blue"},
	// 	{Brand: "Ford", Model: "Mustang", Year: 2015, Color: "Yellow"},
	// 	{Brand: "Ford", Model: "F150", Year: 2017, Color: "Black"},
	// }
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Car not found"})
}

func main() {
	//fmt.Println("Hello, World!") // prints "Hello, World!"
	router := gin.Default()
	router.GET("/cars/:car", getCars)
	router.GET("/version", getVersion)
	router.Run("myserver.local:5000")

}
