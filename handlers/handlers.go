package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

var wsChannel = make(chan WsPayload)

var clients = make(map[WebSocketConnection]string)

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WebSocketConnection struct {
	*websocket.Conn
}

// defines the response sent back from websocket
type WsResponse struct {
	Action      string `json:"action"`
	Message     string `json:"message"`
	MessageType string `json:"messageType"`
	ConnectedUsers []string `json:"connectedUsers"`
}

type WsPayload struct {
	Action  string              `json:"action"`
	Username string `json:"username"`
	Message string              `json:"message"`
	Conn    WebSocketConnection `json:"-"`
}

// Upgrades connection to websocket
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client connected to endpoint")

	var response WsResponse
	response.Message = `<em><small>Connected to server</small></em>`

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = ""

	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}

	go ListenForWs(&conn)
}

func ListenForWs(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			// do nothing (there is no payload)
		} else {
			payload.Conn = *conn
			wsChannel <- payload
		}
	}
}

func ListenToWsChannel() {
	var response WsResponse

	for {
		evt := <-wsChannel

		switch evt.Action {
		case "username":
			// get a list of all users and send it back via broadcast
			clients[evt.Conn] = evt.Username
			users := getUserList()
			response.Action = "listUsers"
			response.ConnectedUsers = users
			broadcastToAll(response)
		case "left":
			response.Action = "listUsers"
			delete(clients, evt.Conn)
			users := getUserList()
			response.ConnectedUsers = users
			broadcastToAll(response)
		}
	}
}

func broadcastToAll(response WsResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Println("websocket err")
			_ = client.Close()
			delete(clients, client)
		}
	}
}

func getUserList() []string {
	var userList []string
	for _, u := range clients {
		if u != "" {
			userList = append(userList, u)
		}
	}
	// sort.Strings(userList)
	return userList
}

// Renders the Home page
func Home(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("Home address")
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}
}

func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	err = view.Execute(w, data, nil) // the third argument is for context, which we are ignoring
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
