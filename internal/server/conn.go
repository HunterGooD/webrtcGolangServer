package server

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/HunterGooD/webrtcGolangServer/internal/util"
	"github.com/chuckpreslar/emission"
	"github.com/gorilla/websocket"
)

//Интервал между отправкой пакетов
const pingPeriod = 5 * time.Second

//Структура соединения сокета
type WebSocketConn struct {
	//диспетчер событий
	emission.Emitter
	//socket соединение
	socket *websocket.Conn
	mutex  *sync.Mutex
	closed bool
}

//Создание структуры сокет соединения
func NewWebSocketConn(socket *websocket.Conn) *WebSocketConn {
	//переменная соединения
	var conn WebSocketConn
	// создание тригера событий
	conn.Emitter = *emission.NewEmitter()
	//подключение к сокет
	conn.socket = socket

	conn.mutex = new(sync.Mutex)
	//Открыт
	conn.closed = false
	// функция обратного вызова
	conn.socket.SetCloseHandler(func(code int, text string) error {
		util.Warnf("%s [%d]", text, code)
		//Событие закрытия
		conn.Emit("close", code, text)
		//отключаем
		conn.closed = true
		return nil
	})
	return &conn
}

//Читать сообщения
func (conn *WebSocketConn) ReadMessage() {
	in := make(chan []byte)
	stop := make(chan struct{})
	pingTicker := time.NewTicker(pingPeriod)

	var c = conn.socket
	go func() {
		for {
			//Чтение сокета
			_, message, err := c.ReadMessage()
			if err != nil {
				util.Warnf("Ошибка на сокете: %v", err)
				//Закрыть сокет при ошибке
				if c, k := err.(*websocket.CloseError); k {
					//Событие закрытия
					conn.Emit("close", c.Code, c.Text)
				} else {
					if c, k := err.(*net.OpError); k {
						conn.Emit("close", 1008, c.Error())
					}
				}
				//Закрыть канал
				close(stop)
				break
			}
			//Запись сообщения в канал
			in <- message
		}
	}()

	//Прием данных
	for {
		select {
		case _ = <-pingTicker.C:
			util.Infof("Отправление пакета...")
			heartPackage := map[string]interface{}{
				//Тип сообщения
				"type": "heartPackage",
				// Пустой пакет
				"data": "",
			}
			//Отправить контрольный пакет партнеру, который в данный момент отправляет сообщение
			if err := conn.Send(util.Marshal(heartPackage)); err != nil {
				util.Errorf("Ошибка отправки контрольного пакета")
				pingTicker.Stop()
				return
			}
		case message := <-in:
			{
				util.Infof("Данные получены: %s", message)
				//Отправьте полученные данные, тип сообщения - сообщение
				conn.Emit("message", []byte(message))
			}
		case <-stop:
			return
		}
	}
}

//Послать сообщение
func (conn *WebSocketConn) Send(message string) error {
	util.Infof("отправка сообщения: %s", message)

	conn.mutex.Lock()

	defer conn.mutex.Unlock()
	//Закрыто ли соединение
	if conn.closed {
		return errors.New("websocket: write closed")
	}
	// послать сообщение
	return conn.socket.WriteMessage(websocket.TextMessage, []byte(message))
}

func (conn *WebSocketConn) Close() {
	//блокировка соединения
	conn.mutex.Lock()
	//Отложенное выполнение разблокировки соединения
	defer conn.mutex.Unlock()
	if conn.closed == false {
		util.Infof("Закрытие соединения сокета: ", conn)
		conn.socket.Close()
		conn.closed = true
	} else {
		util.Warnf("Соединение закрыто :", conn)
	}
}
