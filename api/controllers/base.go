package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"siot/api/models"

	_ "github.com/jinzhu/gorm/dialects/postgres" //postgres database driver
	"github.com/rs/cors"
)

type Server struct {
	DB     *gorm.DB
	Router *mux.Router
	MDB    *mongo.Client
}

func (server *Server) Initialize(DbUser, DbPassword, DbPort, DbHost, DbName, mongoHost string) {

	// connect to postgres
	var err error
	DBURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", DbHost, DbPort, DbUser, DbName, DbPassword)
	server.DB, err = gorm.Open("postgres", DBURL)
	if err != nil {
		fmt.Println("Cannot connect to postgres database")
		log.Fatal("This is the error:", err)
	} else {
		fmt.Println("We are connected to the postgres database")
	}

	// connect to mongodb
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	server.MDB, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+mongoHost+":27017"))
	if err != nil {
		fmt.Println("Cannot connect to mongodb database")
		log.Fatal("This is the error:", err)
	} else {
		fmt.Println("We are connected to the mongodb database")
	}

	// database migration
	server.DB.AutoMigrate(&models.User{})

	server.Router = mux.NewRouter()
	server.initializeRoutes()
}

func (server *Server) Run(addr string) {
	fmt.Println("Listening to port 8080")
	handler := cors.Default().Handler(server.Router)
	log.Fatal(http.ListenAndServe(addr, handler))
}
