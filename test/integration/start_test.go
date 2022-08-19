//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"

	recordsApi "github.com/ArtemVoronov/artforintrovert-test/internal/api/rest/v1/records"
	"github.com/ArtemVoronov/artforintrovert-test/internal/services/cache"
	"github.com/ArtemVoronov/artforintrovert-test/internal/services/db"
	"github.com/ArtemVoronov/artforintrovert-test/internal/services/records"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var TestRouter *gin.Engine

func TestMain(m *testing.M) {
	Setup()
	TestRouter = SetupRouter()
	code := m.Run()
	Shutdown()
	os.Exit(code)
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/records/", recordsApi.GetRecords)
	r.PUT("/records/", recordsApi.UpdateRecord)
	r.DELETE("/records/", recordsApi.DeleteRecord)

	return r
}

func GetRootPath() string {
	_, b, _, _ := runtime.Caller(0)
	d1 := path.Join(path.Dir(b))
	return d1[:len(d1)-len("/test/integration")]
}

func InitTestEnv() {
	if err := godotenv.Load(GetRootPath() + "/.env.test"); err != nil {
		fmt.Println("No .env.test file found")
	}
}

type TestFunc func(t *testing.T)

func RunWithRecreateDB(f TestFunc) func(t *testing.T) {
	collection := db.Instance().GetCollection("testdb", "records")

	return func(t *testing.T) {
		err := collection.Drop(context.TODO())
		assert.Nil(t, err)
		f(t)
	}
}

func Setup() {
	InitTestEnv()
	db.Instance()
	cache.Instance()
	records.Instance()
}

func Shutdown() {
	cache.Instance().ShutDown()
	records.Instance().ShutDown()

	collection := db.Instance().GetCollection("testdb", "records")
	collection.Drop(context.TODO())

	db.Instance().ShutDown()
}
