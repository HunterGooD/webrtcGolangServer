package server

import (
	"net/http"
	"strconv"

	"github.com/HunterGooD/webrtcGolangServer/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// SFUServerConfig для конфигурации сервера
type SFUServerConfig struct {
	Host          string //IP
	Port          int    //Порт
	CertFile      string //Cert
	KeyFile       string //Key cert
	WebSocketPath string // url path
}

//默认WebSocket服务配置
func DefaultConfig() SFUServerConfig {
	return SFUServerConfig{
		Host:          "127.0.0.1",
		Port:          8080,
		WebSocketPath: "/ws",
	}
}

//SFU server struct
type SFUServer struct {
	handleWebSocket func(ws *WebSocketConn, request *http.Request) //функция обработчик сокета
	upgrader        websocket.Upgrader                             // обновление сокета
}

//Создать структуру сервера по функции обработчик
func NewSFUServer(wsHandler func(ws *WebSocketConn, request *http.Request)) *SFUServer {
	var server = &SFUServer{
		handleWebSocket: wsHandler,
	}

	server.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return server
}

//WebSocket обработчик
func (server *SFUServer) handleWebSocketRequest(c *gin.Context) {
	//вернуть заголовок
	responseHeader := http.Header{}
	//responseHeader.Add("Sec-WebSocket-Protocol", "protoo")
	//получить соединение сокет
	socket, err := server.upgrader.Upgrade(c.Writer, c.Request, responseHeader)
	if err != nil {
		util.Panicf("%v", err)
	}
	//Создание экземпляра WebSocketConn
	wsTransport := NewWebSocketConn(socket)
	//Обработка соединений сокета
	server.handleWebSocket(wsTransport, c.Request)
	wsTransport.ReadMessage()
}

func (server *SFUServer) Bind(cfg SFUServerConfig) {
	router := gin.Default()
	router.Any(cfg.WebSocketPath, server.handleWebSocketRequest)

	util.Infof("SFU Server listening on: %s:%d", cfg.Host, cfg.Port)

	router.RunTLS(cfg.Host+":"+strconv.Itoa(cfg.Port), cfg.CertFile, cfg.KeyFile)
}
