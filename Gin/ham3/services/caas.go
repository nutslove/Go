package services

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateCaas(c *gin.Context) {
	caas_id := c.Param("caas_id")
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Create caas %s", caas_id),
	})
}

func GetCaas(c *gin.Context) {
	caas_id := c.Param("caas_id")
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Get caas %s", caas_id),
	})
}

func DeleteCaas(c *gin.Context) {
	caas_id := c.Param("caas_id")
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Delete caas %s", caas_id),
	})
}

func GetCaases(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get caases",
	})
}
