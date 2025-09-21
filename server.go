package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"sdmht-server/db"
	"sdmht-server/graph"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	_ "github.com/joho/godotenv/autoload"
	"github.com/vektah/gqlparser/v2/ast"
)

func setupGraphqlService() *handler.Server {
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: graph.NewResolver()}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})
	return srv
}

type GinContextKeyType string

const GinContextKey GinContextKeyType = "GinContextKey"

func GinContextToContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), GinContextKey, c)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func setupRouter(endpoint string) *gin.Engine {
	router := gin.Default()
	router.SetTrustedProxies([]string{
		"127.0.0.0/8",
		"10.0.0.0/8",
		"172.16.0.0/12",
	})
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, endpoint)
	})
	srv := setupGraphqlService()
	router.Any(endpoint, func(c *gin.Context) {
		accept := c.GetHeader("Accept")
		if c.Request.Method == "GET" && c.Query("query") == "" && (strings.Contains(accept, "text/html") || strings.Contains(accept, "*/*")) {
			playground.Handler("GraphQL playground", endpoint, playground.WithGraphiqlEnablePluginExplorer(true)).ServeHTTP(c.Writer, c.Request)
		} else {
			srv.ServeHTTP(c.Writer, c.Request)
		}
	})
	router.Use(GinContextToContextMiddleware())
	return router
}

const defaultPort = "8000"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	endpoint := "/api/"

	router := setupRouter(endpoint)

	db.SetupDB()

	log.Printf("http://localhost:%s%s", port, endpoint)
	log.Fatal(router.Run(":" + port))
}
