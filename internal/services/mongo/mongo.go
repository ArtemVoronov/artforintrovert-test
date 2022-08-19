package mongo

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

	GetCollection(db string, collection string) *mongo.Collection

	Insert(collection *mongo.Collection)
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

func (s *Service) GetCollection(db string, collection string) *mongo.Collection {
	return s.client.Database(db).Collection(collection)
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

// TODO: add pagination
func (s *Service) GetAll(dbName string, collectionName string) ([]bson.D, error) {
	collection := s.GetCollection(dbName, collectionName)

	ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		panic(err)
	}

	var results []bson.D
	if err = cursor.All(ctx, &results); err != nil {
		return results, fmt.Errorf("unable to get all documents. Error: %v", err)
	}
	for _, result := range results {
		fmt.Println(result)
	}
	return results, nil
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

func createService() *Service {
	connectTimeout := connectTimeout()
	queryTimeout := queryTimeout()
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

	result, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionURL()))
	if err != nil {
		return result, fmt.Errorf("unable to create mongo client: %v", err)
	}

	return result, nil
}

func connectionURL() string {
	username := utils.EnvVarDefault("DATABASE_USERNAME", "mongo_admin")
	password := utils.EnvVarDefault("DATABASE_PASSWORD", "mongo_admin_password")
	host := utils.EnvVarDefault("DATABASE_HOST", "localhost")
	port := utils.EnvVarDefault("DATABASE_PORT", "27017")
	return "mongodb://" + username + ":" + password + "@" + host + ":" + port
}

func connectTimeout() time.Duration {
	value := utils.EnvVarIntDefault("DATABASE_CONNECT_TIMEOUT_IN_SECONDS", "30")
	return time.Duration(value) * time.Second
}

func queryTimeout() time.Duration {
	value := utils.EnvVarIntDefault("DATABASE_QUERY_TIMEOUT_IN_SECONDS", "30")
	return time.Duration(value) * time.Second
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
