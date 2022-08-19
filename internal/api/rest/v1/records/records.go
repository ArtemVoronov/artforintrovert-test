package records

import (
	"log"
	"net/http"

	"github.com/ArtemVoronov/artforintrovert-test/internal/api"
	"github.com/ArtemVoronov/artforintrovert-test/internal/api/validation"
	"github.com/ArtemVoronov/artforintrovert-test/internal/services/mongo"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UpdateRecordDTO struct {
	Id   primitive.ObjectID `json:"id,omitempty"`
	Data string             `json:"data" binding:"required"`
}

type DeleteRecordDTO struct {
	Id primitive.ObjectID `json:"id" binding:"required"`
}

type RawData struct {
	Data string `json:"data" binding:"required"`
}

const (
	dbName         = "testdb" // TODO: read from config, put somewhere else
	collectionName = "records"
)

func GetRecords(c *gin.Context) {
	// TODO: add pagination
	// TODO: add using cache
	result, err := mongo.Instance().GetAll(dbName, collectionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ERROR_INTERNAL_SERVER_ERROR)
		log.Printf("unable to create record: %v", err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func UpdateRecord(c *gin.Context) {
	var record UpdateRecordDTO

	if err := c.BindJSON(&record); err != nil {
		validation.SendError(c, err)
		return
	}

	if record.Id == primitive.NilObjectID {
		id, err := mongo.Instance().Insert(dbName, collectionName, RawData{record.Data})
		if err != nil {
			c.JSON(http.StatusInternalServerError, api.ERROR_INTERNAL_SERVER_ERROR)
			log.Printf("unable to create record: %v", err)
			return
		}
		c.JSON(http.StatusCreated, id)
		return
	}

	id, err := mongo.Instance().Upsert(dbName, collectionName, record.Id, RawData{record.Data})
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ERROR_INTERNAL_SERVER_ERROR)
		log.Printf("unable to update record: %v", err)
		return
	}

	if id != nil {
		c.JSON(http.StatusCreated, id)
		return
	}

	c.JSON(http.StatusOK, api.DONE)

}

func DeleteRecord(c *gin.Context) {
	var record DeleteRecordDTO

	if err := c.BindJSON(&record); err != nil {
		validation.SendError(c, err)
		return
	}

	if record.Id == primitive.NilObjectID {
		c.JSON(http.StatusBadRequest, api.ERROR_MISSED_ID)
		return
	}

	err := mongo.Instance().Delete(dbName, collectionName, record.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ERROR_INTERNAL_SERVER_ERROR)
		log.Printf("unable to update record: %v", err)
		return
	}

	c.JSON(http.StatusOK, api.DONE)
}
