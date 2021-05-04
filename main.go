package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/masgustavos/alertmanager-discord-webhook/alertmanager"
	"github.com/masgustavos/alertmanager-discord-webhook/config"
	"github.com/masgustavos/alertmanager-discord-webhook/discord"
)

func main() {
	configs := config.LoadUserConfig()

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "Application is healthy!",
		})
	})

	router.POST("/:channel", func(c *gin.Context) {
		channelName := c.Param("channel")

		var alertmanagerBody alertmanager.MessageBody
		if err := c.ShouldBindJSON(&alertmanagerBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := discord.SendAlerts(channelName, alertmanagerBody, *configs); err != nil {
			log.Println("[ERROR] ", err)
		}

		c.String(http.StatusOK, "Channel: %s", channelName)
	})

	router.Run()

}
