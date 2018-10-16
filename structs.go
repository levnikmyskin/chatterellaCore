package main
import(
	"github.com/gorilla/websocket"
)

type Message struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
}

type Client struct {
    conn *websocket.Conn
    Username string `json:"username"`
}

func (c *Client) WriteJSON (msg interface{}) error {
    return c.conn.WriteJSON(msg)
}

func (c *Client) ReadJSON (msg interface{}) error {
    return c.conn.ReadJSON(msg)
}

func (c *Client) Close () error {
    return c.conn.Close()
}
