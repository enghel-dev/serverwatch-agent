package network

import (
	"fmt"
	"os"
	"runtime"
	"net"
	"encoding/json"

	"github.com/gorilla/websocket"
)

type RegisterMessage struct {
	Hostname  string `json:"hostname"`
	IPAddress string `json:"ip_address"`
	OS        string `json:"operating_system"`
}

type RegisterResponse struct {
	ServerID int `json:"server_id"`
}

func TestConnection(host string) bool {
	url := "ws://" + host + "/ws/metrics"

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Println("Error al conectar:", err)
		return false
	}
	defer conn.Close()

	return true
}

func Connect(host string) (*websocket.Conn, error) {
	url := "ws://" + host + "/ws/metrics"

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Println("Error al conectar:", err)
		return nil, err
	}
	return conn, nil
}

func GetOperatingSystem() string {
	return runtime.GOOS
}

func GetHostname () (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return hostname, nil
}

func GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}


func Register(conn *websocket.Conn) (int, error) {
	hostname, err := GetHostname()
	if err != nil {
		return 0, err
	}

	ip, err := GetLocalIP()
	if err != nil {
		return 0, err
	}

	msg := RegisterMessage{
		Hostname:  hostname,
		IPAddress: ip,
		OS:        runtime.GOOS,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return 0, err
	}

	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return 0, err
	}

	_, responseData, err := conn.ReadMessage()
	if err != nil {
		return 0, err
	}

	var response RegisterResponse
	err = json.Unmarshal(responseData, &response)
	if err != nil {
		return 0, err
	}

	return response.ServerID, nil
}