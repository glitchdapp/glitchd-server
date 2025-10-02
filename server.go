package main

import (
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/glitchd/glitchd-server/database"
	"github.com/glitchd/glitchd-server/directives"
	"github.com/glitchd/glitchd-server/graph"
	"github.com/glitchd/glitchd-server/middlewares"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

const defaultPort = "8080"

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// connect to database.
	database.DB = database.Connect()

	router := mux.NewRouter()
	router.Use(middlewares.AuthMiddleware)

	c := graph.Config{Resolvers: &graph.Resolver{Rooms: sync.Map{}, Viewers: sync.Map{}}}
	c.Directives.Auth = directives.Auth

	srv := handler.New(graph.NewExecutableSchema(c))
	srv.AddTransport(transport.POST{})
	if os.Getenv("ENVIRONMENT") == "development" {
		srv.Use(extension.Introspection{})
	}
	srv.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// add checking origin logic to decide return true or false
				return true
			},
		},
		KeepAlivePingInterval: 10 * time.Second,
	})

	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
