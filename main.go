package main

import (
	"os"

	"github.com/gin-gonic/gin"

	"github.com/jeffandersen/listbot/actions"
)

func main() {
	r := gin.Default()
	r.POST("/webhook", actions.HandleWebhook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	r.Run(":" + port)
}
