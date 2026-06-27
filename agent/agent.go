package agent

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/enghel-dev/serverwatch-agent/collector"
	"github.com/enghel-dev/serverwatch-agent/config"
	"github.com/enghel-dev/serverwatch-agent/network"
	"github.com/gorilla/websocket"
)

func Run(cfg *config.AgentConfig) {

	conectado, err := network.Connect(cfg.BackendHost)
	if err != nil {
		fmt.Println("No se pudo conectar al backend:", cfg.BackendHost)
		return
	}

	serverID, err := network.Register(conectado)
	if err != nil {
		fmt.Println("Error al registrar el agente:", err)
		return
	}
	fmt.Println("Agente registrado con server_id:", serverID)	
																												
	var buffer []*collector.Metrics
	reconectando := false
	canalReconexion := make(chan *websocket.Conn)

	for {
		// Si hay una reconexión en curso, revisamos sin bloquear si ya terminó
		if reconectando {
			select {
			case nuevaConn := <-canalReconexion:
				conectado = nuevaConn
				reconectando = false
				fmt.Println("Reconectado. Enviando", len(buffer), "métricas pendientes...")
				for _, pendiente := range buffer {
					dataPendiente, errMarshal := json.Marshal(pendiente)
					if errMarshal != nil {
						continue
					}
					conectado.WriteMessage(websocket.TextMessage, dataPendiente)
				}
				buffer = nil
			default:
				// todavía no reconecta, seguimos sin bloquear
			}
		}
		

		metricas, err := collector.GetAllMetrics()
		if err != nil {
			fmt.Println("Error leyendo métricas:", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Mientras se reconecta, solo acumulamos en buffer, no intentamos enviar
		if reconectando {
			buffer = append(buffer, metricas)
			fmt.Println("Sin conexión, métrica guardada en buffer. Total:", len(buffer))
			time.Sleep(5 * time.Second)
			continue
		}

		jsonData, err := json.Marshal(metricas)
		if err != nil {
			fmt.Println("Error serializando métricas:", err)
			time.Sleep(5 * time.Second)
			continue
		}

		err = conectado.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			fmt.Println("Conexión perdida, guardando métrica en buffer:", err)
			buffer = append(buffer, metricas)
			reconectando = true
			go reconectar(cfg.BackendHost, canalReconexion)
		} else {
			fmt.Println("Métricas enviadas al backend:", cfg.BackendHost)
		}

		time.Sleep(5 * time.Second)
	}
}

func reconectar(host string, canal chan *websocket.Conn) {
	espera := 1 * time.Second
	const esperaMaxima = 60 * time.Second

	for {
		fmt.Println("Intentando reconectar en", espera, "...")
		time.Sleep(espera)

		if network.TestConnection(host) {
			conn, err := network.Connect(host)
			if err == nil {
				canal <- conn
				return
			}
		}

		espera *= 2
		if espera > esperaMaxima {
			espera = esperaMaxima
		}
	}
}