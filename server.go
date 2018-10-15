package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

const(
	cert = "/docker/mainStack/letsencrypt_certs/live/chat.kernelpanics.it/fullchain.pem"
	key = "/docker/mainStack/letsencrypt_certs/live/chat.kernelpanics.it/privkey.pem"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var clients = make(map[string]*websocket.Conn)
var broadcast = make(chan string)

type Message struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
}

type NewClient struct {
	Client string `json:"client"`
}

func routeMessageToUser(client *websocket.Conn, username string) {
	for {
		var msg Message
		err := client.ReadJSON(&msg)
		if err != nil {
			delete(clients, username)
			log.Printf("client ws error: %v", err)
			return
		}

		handleMessage(&msg)
	}
}

func handleWebSockets(w http.ResponseWriter, r *http.Request) {
	log.Println("New WS connection")
	username := r.URL.Query()["username"][0]

	client, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	if _, ok := clients[username]; ok {
		client.WriteJSON(map[string]map[int]string{"error": {420: "username already taken"}})
		client.Close()
		return
	}


	clients[username] = client
	client.WriteJSON(map[string]string{"success": "connection succeeded"})
	client.WriteJSON(map[string][]string{"clients": getListOfClients()})
	broadcast <- username
	go routeMessageToUser(client, username)
}

func getListOfClients() []string {
	keys := make([]string, len(clients), len(clients))
	counter := 0
	for k := range clients {
		keys[counter] = k
		counter++
	}
	return keys
}

func handleMessage(msg *Message) {
	log.Println("NEW MESSAGE", msg)
	client := clients[msg.Recipient]
	if client == nil {
		log.Printf("Client %s does not exist", msg.Recipient)
		client.WriteJSON("Error: client does not exist")
		return
	}
	sendMessage(client, msg, msg.Recipient)
}

func handleBroadcast() {
	for newClient := range broadcast {
		for username, conn := range clients {
			sendMessage(conn, &NewClient{Client: newClient}, username)
		}
	}
}

func sendMessage(client *websocket.Conn, msg interface{}, recipient string) {
	err := client.WriteJSON(msg)
	if err != nil {
		log.Fatalf("error writing to client: %v", err)
		client.Close()
		delete(clients, recipient)
	}
}

func main() {
	http.HandleFunc("/chat", handleWebSockets)
	go handleBroadcast()

	err := http.ListenAndServeTLS("chat.kernelpanics.it:4043", cert, key, nil)
	if err != nil {
		log.Fatal(err)
	}
}
