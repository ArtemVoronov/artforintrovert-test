package db

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ArtemVoronov/artforintrovert-test/internal/api"
	"github.com/ArtemVoronov/artforintrovert-test/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoService interface {
	ShutDown()

	GetCollection(dbName string, collectionName string) *mongo.Collection
	Insert(dbName string, collectionName string, document interface{}) (*primitive.ObjectID, error)
	Upsert(dbName string, collectionName string, id primitive.ObjectID, document interface{}) (*primitive.ObjectID, error)
	Delete(dbName string, collectionName string, id primitive.ObjectID) error
}

type Service struct {
	connectTimeout time.Duration
	queryTimeout   time.Duration
	client         *mongo.Client
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
	ctx, cancel := context.WithTimeout(context.Background(), s.connectTimeout)
	defer cancel()
	defer func() {
		err := s.client.Disconnect(ctx)
		if err != nil {
			log.Printf("mongo client unable to disconnect: %v", err)
		}
	}()
}

func (s *Service) GetCollection(dbName string, collectionName string) *mongo.Collection {
	return s.client.Database(dbName).Collection(collectionName)
}

func (s *Service) Insert(dbName string, collectionName string, document interface{}) (*primitive.ObjectID, error) {
	collection := s.GetCollection(dbName, collectionName)

	ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
	defer cancel()

	insertResult, err := collection.InsertOne(ctx, document)
	if err != nil {
		return nil, fmt.Errorf("unable to insert document '%v'. Error: %v", document, err)
	}

	result, ok := insertResult.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("unable to insert document: %s", api.ERROR_ASSERT_RESULT_TYPE)
	}

	return &result, nil
}

func (s *Service) Upsert(dbName string, collectionName string, id primitive.ObjectID, document interface{}) (*primitive.ObjectID, error) {
	collection := s.GetCollection(dbName, collectionName)

	ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
	defer cancel()

	opts := options.Update().SetUpsert(true)
	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", document}}
	result, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, fmt.Errorf("unable to update document. ID: '%v'. Document: '%v'. Error: %v", id, document, err)
	}

	if result.MatchedCount != 0 {
		return nil, nil
	}

	if result.UpsertedCount != 0 {
		id, ok := result.UpsertedID.(primitive.ObjectID)
		if !ok {
			return nil, fmt.Errorf("unable to update document: %s", api.ERROR_ASSERT_RESULT_TYPE)
		}
		return &id, nil
	}

	return nil, nil
}

func (s *Service) Delete(dbName string, collectionName string, id primitive.ObjectID) error {
	collection := s.GetCollection(dbName, collectionName)

	ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
	defer cancel()

	_, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("unable to delete document. ID: '%v'. Error: %v", id, err)
	}
	return err
}

func (s *Service) GetQueryTimeout() time.Duration {
	return s.queryTimeout
}

func createService() *Service {
	connectTimeout := ConnectTimeout()
	queryTimeout := QueryTimeout()
	client, err := createClient(connectTimeout)
	if err != nil {
		log.Fatalf("unable to setup mongo service: %v", err)
	}
	return &Service{
		connectTimeout: connectTimeout,
		queryTimeout:   queryTimeout,
		client:         client,
	}
}

func createClient(connectTimeout time.Duration) (*mongo.Client, error) {
	var result *mongo.Client

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	result, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoConnectionURL()))
	if err != nil {
		return result, fmt.Errorf("unable to create mongo client: %v", err)
	}

	return result, nil
}

func mongoConnectionURL() string {
	username := utils.EnvVarDefault("DATABASE_USERNAME", "mongo_admin")
	password := utils.EnvVarDefault("DATABASE_PASSWORD", "mongo_admin_password")
	host := utils.EnvVarDefault("DATABASE_HOST", "localhost")
	port := utils.EnvVarDefault("DATABASE_PORT", "27017")
	return "mongodb://" + username + ":" + password + "@" + host + ":" + port
}

// TODO: add processing of case when we have replica set of mongos and need to use sessions + tx
type QueryFuncVoid func(sc mongo.SessionContext) error

func Tx(f QueryFuncVoid) func() error {
	service := Instance()

	ctx, cancel := context.WithTimeout(context.Background(), service.queryTimeout)
	defer cancel()

	return func() error {
		session, err := service.client.StartSession()
		if err != nil {
			return fmt.Errorf("unable to start session: %v", err)
		}
		defer session.EndSession(ctx)

		err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
			err = session.StartTransaction()
			if err != nil {
				return fmt.Errorf("unable to start tx: %v", err)
			}
			defer session.AbortTransaction(sc)

			err := f(sc)
			if err != nil {
				return err
			}

			err = session.CommitTransaction(sc)
			if err != nil {
				return fmt.Errorf("unable to commit tx: %v", err)
			}
			return nil
		})

		return err
	}
}
