package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Пропускаем любой запрос
	},
}

type Server struct {
	route         *gin.Engine
	clients       map[*websocket.Conn]bool
	handleMessage func(message []byte) // хандлер новых сообщений
	port          string
	sync.Mutex
}

func HttpServerNew(port string) *Server {
	if os.Getenv("HTTP_SERVER_DEBUG") == "0" {
		gin.SetMode(gin.ReleaseMode)
	}
	server := Server{
		route:   gin.Default(),
		clients: make(map[*websocket.Conn]bool),
		port:    port,
	}

	return &server
}

func (server *Server) GetRoute() *gin.Engine {
	return server.route
}

func (server *Server) SetWSHandle(wshandle func(message []byte)) {
	server.handleMessage = wshandle
}

func (server *Server) Start() *Server {

	server.route.GET("/ws", server.echo)
	server.route.StaticFile("/", "./web/build/index.html")
	server.route.StaticFS("/static", http.Dir("./web/build/static/"))

	go server.route.Run(":" + server.port)

	return server
}

func (server *Server) echo(c *gin.Context) {
	connection, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
	}
	defer connection.Close()

	server.Lock()
	server.clients[connection] = true // Сохраняем соединение, используя его как ключ
	server.Unlock()
	defer func() {
		server.Lock()
		defer server.Unlock()
		delete(server.clients, connection)
	}()

	for {
		mt, message, err := connection.ReadMessage()
		fmt.Println(mt, message, err)

		if err != nil || mt == websocket.CloseMessage {
			break // Выходим из цикла, если клиент пытается закрыть соединение или связь прервана
		}

		go server.handleMessage(message)
	}

}

func (server *Server) WriteMessage(message []byte) {
	for conn := range server.clients {
		conn.WriteMessage(websocket.TextMessage, message)
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
