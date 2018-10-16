package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

const (
	cert = "/docker/mainStack/letsencrypt_certs/live/chat.kernelpanics.it/fullchain.pem"
	key  = "/docker/mainStack/letsencrypt_certs/live/chat.kernelpanics.it/privkey.pem"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var clients = make(map[string]*Client)
var broadcast = make(chan string)


func routeMessageToUser(client *Client) {
	for {
		var msg Message
		err := client.ReadJSON(&msg)
		if err != nil {
			delete(clients, client.Username)
			log.Printf("client deleted, info: %v\n", err)
			return
		}

		handleMessage(&msg)
	}
}

func handleWebSockets(w http.ResponseWriter, r *http.Request) {
	log.Println("New WS connection")
	username := r.URL.Query()["username"][0]

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
        return
	}

	if _, ok := clients[username]; ok {
		conn.WriteJSON(map[string]map[int]string{"error": {420: "username already taken"}})
		conn.Close()
		return
	}
    
    client := &Client{conn, username}
	clients[username] = client
	conn.WriteJSON(map[string]string{"success": "connection succeeded"})
	conn.WriteJSON(map[string][]string{"clients": getListOfClients()})
	broadcast <- username
	go routeMessageToUser(client)
}

func getListOfClients() []string {
	keys := make([]string, len(clients))
	counter := 0
	for username := range clients {
		keys[counter] = username
		counter++
	}
	return keys
}

func handleMessage(msg *Message) {
	log.Println("NEW MESSAGE", msg)
	recipient := clients[msg.Recipient]
	if recipient == nil {
		log.Printf("Client %s does not exist\n", msg.Recipient)
		recipient.WriteJSON(map[string]string{"Error": "client does not exist"})
		return
	}
	sendMessage(recipient.conn, msg, msg.Recipient)
}

func handleBroadcast() {
	for newClient := range broadcast {
        m := map[string]string {"client":newClient} 
        for username, conn := range clients {
			sendMessage(conn, m, username)
		}
	}
}

func sendMessage(conn *websocket.Conn, msg interface{}, recipient string) {
	err := conn.WriteJSON(msg)
	if err != nil {
		log.Printf("error writing to client: %v\n", err)
		conn.Close()
		delete(clients, recipient)
	}
}

func main() {
	http.HandleFunc("/chat", handleWebSockets)
	http.HandleFunc("/updateInfo", handleUpdateInfo)
	http.HandleFunc("/download", handleDownload)
	go handleBroadcast()

    err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
