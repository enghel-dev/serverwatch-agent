package network

import (
	"fmt"

	"github.com/gorilla/websocket"
)

func TestConnection(host string) bool {
	url := "ws://" + host + "/ws/metrics/test"

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Println("Error al conectar:", err)
		return false
	}
	defer conn.Close()

	return true
}

func Connect(host string) (*websocket.Conn, error) {
	url := "ws://" + host + "/ws/metrics/test"

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Println("Error al conectar:", err)
		return nil, err
	}
	return conn, nil
}