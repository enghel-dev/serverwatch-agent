# ServerWatch Agent

Agente de monitoreo escrito en Go para el proyecto **ServerWatch**. Se instala como un único ejecutable en cada servidor monitoreado (Linux o Windows), recolecta métricas del sistema cada 5 segundos y las transmite en tiempo real al backend (`ingestion-svc`) a través de WebSocket.

---

## Características

- **Instalador CLI integrado** — al ejecutarse por primera vez (sin `config.yaml`), pide el host del backend y el nombre del servidor, valida el formato, prueba la conexión real antes de guardar, y pasa automáticamente al modo agente sin reiniciar el proceso.
- **Recolección de métricas reales** vía [`gopsutil`](https://github.com/shirou/gopsutil):
  - CPU (%)
  - RAM usada / total (MB)
  - Disco — uso por partición real del sistema
  - Red — bytes/s de subida y bajada
  - Top 5 procesos por consumo de CPU
- **Registro automático** — al conectar, el agente envía `hostname`, `ip_address` y `operating_system`, y recibe un `server_id` asignado por el backend.
- **Conexión WebSocket persistente** — envío continuo cada 5 segundos por un único socket abierto.
- **Resiliencia ante caídas de red** — si se pierde la conexión, las métricas se acumulan en un buffer en memoria mientras una goroutine en segundo plano reintenta reconectar con backoff exponencial (1s → 2s → 4s... hasta un máximo de 60s, indefinidamente). Al reconectar, reenvía todo lo acumulado.

---

## Requisitos

- Go 1.22 o superior
- Acceso de red al host donde corre `serverwatch-ingestion`

---

## Estructura del proyecto

```
serverwatch-agent/
├── main.go              # Punto de entrada: decide modo instalador o modo agente
├── config/               # Lectura/escritura de config.yaml según el OS
├── installer/            # Flujo CLI de instalación (host, nombre, validación)
├── network/              # Conexión WebSocket, registro inicial, prueba de conexión
├── collector/            # Recolección de métricas del sistema (gopsutil)
├── agent/                # Loop principal: conectar, registrar, medir, enviar, reconectar
└── go.mod
```

---

## Instalación y uso

### Compilar

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o serverwatch-agent

# Windows
GOOS=windows GOARCH=amd64 go build -o serverwatch-agent.exe
```

### Primera ejecución (instalación)

```bash
./serverwatch-agent
```

El programa va a solicitar:

1. **Host del backend** — formato `IP:puerto` (ej. `192.168.1.100:8001`). Se valida el formato y se prueba la conexión real antes de continuar.
2. **Nombre del servidor** — nombre descriptivo que se mostrará en el dashboard.

Al finalizar, guarda la configuración y entra automáticamente en modo agente — no es necesario volver a ejecutar el programa.

### Ejecuciones posteriores

Si ya existe el archivo de configuración, el agente arranca directo en modo monitoreo:

```bash
./serverwatch-agent
```

### Ubicación del archivo de configuración

| Sistema | Ruta |
|---|---|
| Windows | `C:\ProgramData\ServerWatch\config.yaml` |
| Linux | `/etc/serverwatch/config.yaml` |

```yaml
backend_host: "192.168.1.100:8001"
display_name: "Servidor Web Producción"
agent_uuid: ""   # reservado para uso futuro con tokens persistentes
```

> El campo `backend_host` debe ser **solo** `host:puerto`, sin `ws://` ni ninguna ruta — el agente construye la URL completa internamente.

---

## Contrato de comunicación con el backend

### 1. Registro (primer mensaje, una vez por conexión)

El agente envía:

```json
{
  "hostname": "ENGHEL",
  "ip_address": "192.168.1.130",
  "operating_system": "windows"
}
```

El backend responde:

```json
{"server_id": 1}
```

### 2. Métricas (cada 5 segundos, después del registro)

```json
{
  "cpu_percent": 18.45,
  "ram_used_mb": 10732.1,
  "ram_total_mb": 16055.36,
  "disk_usage": [
    {"partition": "C:", "used_gb": 350.91, "total_gb": 475.95}
  ],
  "network_out_kbps": 214,
  "network_in_kbps": 722,
  "top_processes": [
    {"name": "Discord.exe", "pid": 17684, "cpu_percent": 47.58}
  ]
}
```

Ruta WebSocket actual: `/ws/metrics` (sin `server_id` en la ruta — el backend identifica al agente por la conexión activa, ya que el flujo de tokens persistentes aún no existe).

---

## Limitaciones conocidas

### ⚠️ Re-registro pendiente tras reconexión

El backend actual **no tiene memoria entre conexiones** — cada conexión WebSocket nueva es tratada como un servidor desconocido hasta que recibe el mensaje de registro, y le asigna un `server_id` nuevo (no recupera el anterior). Esto es consecuencia del modo de auto-registro temporal, sin tokens persistentes (eso está planeado para cuando `api-svc` implemente el flujo real de tokens).

**Estado actual del agente:** la goroutine de reconexión (`agent.go`) reabre la conexión tras una caída de red, pero **no vuelve a enviar el mensaje de registro** en la conexión nueva — solo reenvía el buffer de métricas acumuladas directamente.

**Consecuencia práctica:** en el backend actual, una reconexión real probablemente sea rechazada (o tratada de forma inconsistente) porque el primer mensaje que recibe en la conexión nueva es una métrica, no un registro.

**Pendiente:** llamar a `network.Register(conn)` también dentro de la rama de reconexión exitosa en `agent.Run()`, antes de vaciar el buffer — no solo en la conexión inicial. Se decidió posponer esta corrección para una iteración posterior del proyecto.

---

## Componentes internos

| Paquete | Responsabilidad |
|---|---|
| `config` | Ruta de configuración según OS, carga y guardado del YAML |
| `installer` | Flujo de instalación por CLI, validación de host por regex |
| `network` | `Connect`, `TestConnection`, `Register`, obtención de hostname/IP local |
| `collector` | Lectura de CPU, RAM, disco, red y procesos vía `gopsutil` |
| `agent` | Orquestación: conectar → registrar → loop de métricas → reconexión |

---

## Dependencias principales

- [`gorilla/websocket`](https://github.com/gorilla/websocket) — cliente WebSocket
- [`shirou/gopsutil/v3`](https://github.com/shirou/gopsutil) — métricas de sistema multiplataforma
- [`gopkg.in/yaml.v3`](https://github.com/go-yaml/yaml) — lectura/escritura de configuración