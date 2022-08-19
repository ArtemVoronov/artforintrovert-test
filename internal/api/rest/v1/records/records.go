package records

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	NOT_IMPLEMENTED = "Not Implemented"
)

func GetRecords(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, NOT_IMPLEMENTED)
}

func GetRecord(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, NOT_IMPLEMENTED)
}

func CreateRecord(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, NOT_IMPLEMENTED)
}

func UpdateRecord(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, NOT_IMPLEMENTED)
}

func DeleteRecord(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, NOT_IMPLEMENTED)
}
