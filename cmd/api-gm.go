package main

import "github.com/gin-gonic/gin"

func handleIndex(c *gin.Context) {
	c.String(200, "Hello, World!")
}

func main() {
	r := gin.Default()
	r.GET("/", handleIndex)
	r.Run(":8080")
}
