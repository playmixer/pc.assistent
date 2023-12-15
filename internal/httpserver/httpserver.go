package httpserver

import (
	"fmt"
	"io"
	"log"
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

func New(port string) *Server {
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

// func (server *Server) index(w http.ResponseWriter, r *http.Request) {
func (server *Server) index(c *gin.Context) {
	file, err := os.Open("./web/build/index.html")
	if err != nil {
		log.Println(err)
	}
	b, err := io.ReadAll(file)
	if err != nil {
		log.Println(err)
	}
	_, err = c.Writer.Write(b)
	if err != nil {
		log.Println(err)
	}
}

// func (server *Server) echo(w http.ResponseWriter, r *http.Request) {
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
