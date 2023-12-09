package socketserver

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Пропускаем любой запрос
	},
}

type Server struct {
	clients       map[*websocket.Conn]bool
	handleMessage func(message []byte) // хандлер новых сообщений
	sync.Mutex
}

func StartServer(handleMessage func(message []byte)) *Server {
	server := Server{
		clients:       make(map[*websocket.Conn]bool),
		handleMessage: handleMessage,
	}

	http.HandleFunc("/", server.echo)
	go http.ListenAndServe(":8080", nil) // Уводим http сервер в горутину

	return &server
}

func (server *Server) echo(w http.ResponseWriter, r *http.Request) {
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}
	defer connection.Close()

	server.Lock()
	server.clients[connection] = true // Сохраняем соединение, используя его как ключ
	defer server.Unlock()
	defer delete(server.clients, connection) // Удаляем соединение

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
