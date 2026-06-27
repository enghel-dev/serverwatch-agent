package installer

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/enghel-dev/serverwatch-agent/config"
	"github.com/enghel-dev/serverwatch-agent/network"
)

var hostPattern = regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d{1,5}$`)

func IsValidHost(host string) bool {
	return hostPattern.MatchString(host)
}

func RunCLI() {
	reader := bufio.NewReader(os.Stdin)
	var host string
	var name string
	for {
		//Obtener el host
		fmt.Print("Host del backend (ej. 192.168.1.100:8001): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if IsValidHost(input) {
			if network.TestConnection(input) == true {
				host = input
				break	
			} 
		} else {
			fmt.Println("Formato inválido. Debe ser IP:puerto, ej. 192.168.1.100:8001")
		}
	}
	for{
		//Obtener el nombre del servidor
		fmt.Print("Nombre del servidor: ")
		nameInput, _ := reader.ReadString('\n')
		nameInput = strings.TrimSpace(nameInput)
		if nameInput == "" {
			fmt.Println("El nombre del servidor no puede estar vacío.")
			continue
		}else {
			name = nameInput
			break
		}
	}

	//Guardar la configuración
	var cfg config.AgentConfig
	cfg.BackendHost = host
	cfg.DisplayName = name
	err := config.SaveConfig(&cfg)
	if err != nil {
		fmt.Println("Error guardando config:", err)
		return
	}

	fmt.Println("Host válido recibido:", host)
	fmt.Println("Nombre del servidor:", name)
}