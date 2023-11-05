package main
import "github.com/gin-contrib/cors"
import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	
)

var messages = []map[string]string{}
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan map[string]string)
var mutex = &sync.Mutex{}

func main() {
    r := gin.Default()
    r.Use(cors.Default())  // Add this line for CORS

    r.GET("/get_messages", GetMessageHistory)
    r.GET("/ws", HandleWebSocket)

    go HandleMessages()

    r.Run(":5000")
}

func GetMessageHistory(c *gin.Context) {
	c.JSON(http.StatusOK, messages)
}

func HandleWebSocket(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Failed to set WebSocket upgrade")
		return
	}
	defer ws.Close()

	clients[ws] = true

	for {
		var msg map[string]string
		err := ws.ReadJSON(&msg)
		if err != nil {
			delete(clients, ws)
			break
		}

		broadcast <- msg
	}
}

func HandleMessages() {
	for {
		msg := <-broadcast
		mutex.Lock()
		messages = append(messages, msg)
		mutex.Unlock()
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
	}
}

