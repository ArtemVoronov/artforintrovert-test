package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	recordsApi "github.com/ArtemVoronov/artforintrovert-test/internal/api/rest/v1/records"
	"github.com/ArtemVoronov/artforintrovert-test/internal/services/cache"
	"github.com/ArtemVoronov/artforintrovert-test/internal/services/db"
	"github.com/ArtemVoronov/artforintrovert-test/internal/services/records"
	"github.com/ArtemVoronov/artforintrovert-test/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func Start() {
	setup()
	defer Shutdown()
	srv := &http.Server{
		Addr:    host(),
		Handler: router(),
	}

	go func() {
		log.Printf("App starting at localhost%s ...\n", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Println("Server was closed")
		} else if err != nil {
			log.Fatalf("Unable to start app: %v\n", err)
		}
	}()

	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := srv.Shutdown(ctx)
	if err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server has been shutdown")
}

func setup() {
	loadEnv()
	db.Instance()
	cache.Instance()
	records.Instance()

}
func Shutdown() {
	db.Instance().ShutDown()
	cache.Instance().ShutDown()
	records.Instance().ShutDown()
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Print("No .env file found")
	}
}

func host() string {
	port := utils.EnvVarDefault("APP_PORT", "3000")
	host := ":" + port
	return host
}

func mode() string {
	return utils.EnvVarDefault("APP_MODE", "debug")
}

func router() *gin.Engine {
	router := gin.Default()
	gin.SetMode(mode())
	router.Use(cors())
	router.Use(gin.Logger())

	v1 := router.Group("/api/v1")
	v1.GET("/records/", recordsApi.GetRecords)
	v1.PUT("/records/", recordsApi.UpdateRecord)
	v1.DELETE("/records/", recordsApi.DeleteRecord)

	return router
}

func cors() gin.HandlerFunc {
	cors := utils.EnvVarDefault("CORS", "*")
	return func(c *gin.Context) {
		c.Writer.Header().Add("Access-Control-Allow-Origin", cors)
		c.Next()
	}
}
