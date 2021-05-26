package ws

import (
	"github.com/gorilla/websocket"
	"net/http"
)

var defaultUpgrader  = websocket.Upgrader{}

func Upgrade(w http.ResponseWriter, r *http.Request, h http.Header) (*websocket.Conn, error) {
	return defaultUpgrader.Upgrade(w, r, h)
}

func UpgradeWithNotifyClosed(w http.ResponseWriter, r *http.Request, h http.Header) (
	*websocket.Conn, <- chan struct{}, error) {
	wsConn, err := Upgrade(w, r, h)
	if err != nil {
		return nil, nil, err
	}

	// 通知处理进程，对端主动断开了连接
	closedCh := make(chan struct{})
	go func() {
		for {
			// 通过调用 ReadMessage() 来检测对端是否断开连接，
			// 如果对端关闭连接，该调用会返回 error，其他消息我们忽略
			_, _, err := wsConn.ReadMessage()
			if err != nil {
				close(closedCh)
				return
			}
		}
	}()

	return  wsConn, closedCh, nil
}
