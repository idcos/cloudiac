package ws

import (
	"github.com/gorilla/websocket"
	"net/http"
)

var defaultUpgrader  = websocket.Upgrader{}

func Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error) {
	return defaultUpgrader.Upgrade(w, r, responseHeader)
}
