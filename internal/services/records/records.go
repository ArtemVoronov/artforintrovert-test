package records

import (
	"context"
	"fmt"
	"sync"

	"github.com/ArtemVoronov/artforintrovert-test/internal/services/db"
	"github.com/ArtemVoronov/artforintrovert-test/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	RECORDS_COLLECTION_NAME = "records"
)

type Record struct {
	Id   primitive.ObjectID `json:"id" bson:"_id" binding:"required"`
	Data string             `json:"data" bson:"data"  binding:"required"`
}

type RecordsService interface {
	ShutDown()
	Insert(document interface{}) (*primitive.ObjectID, error)
	Upsert(id primitive.ObjectID, document interface{}) (*primitive.ObjectID, error)
	Delete(id primitive.ObjectID) error
	GetAllRecords() ([]Record, error)
}
type Service struct {
	dbName string
}

var once sync.Once
var instance *Service

func Instance() *Service {
	once.Do(func() {
		if instance == nil {
			instance = createService()
		}
	})
	return instance
}

func (s *Service) ShutDown() {
}

func (s *Service) Insert(document interface{}) (*primitive.ObjectID, error) {
	return db.Instance().Insert(s.dbName, RECORDS_COLLECTION_NAME, document)
}

func (s *Service) Upsert(id primitive.ObjectID, document interface{}) (*primitive.ObjectID, error) {
	return db.Instance().Upsert(s.dbName, RECORDS_COLLECTION_NAME, id, document)
}

func (s *Service) Delete(id primitive.ObjectID) error {
	return db.Instance().Delete(s.dbName, RECORDS_COLLECTION_NAME, id)
}

func (s *Service) GetAll() ([]Record, error) {
	var result []Record = make([]Record, 0)

	collection := db.Instance().GetCollection(s.dbName, RECORDS_COLLECTION_NAME)

	ctx, cancel := context.WithTimeout(context.Background(), db.Instance().GetQueryTimeout())
	defer cancel()

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return result, fmt.Errorf("unable to get all documents. Error: %v", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var document Record
		err := cursor.Decode(&document)
		if err != nil {
			return result, fmt.Errorf("unable to get all documents. Error: %v", err)
		}
		result = append(result, document)
	}
	return result, nil
}

func createService() *Service {
	return &Service{
		dbName: utils.EnvVarDefault("DATABASE_NAME", "testdb"),
	}
}
