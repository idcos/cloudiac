package utils

import (
	"github.com/gorilla/websocket"
	"net/url"
	"path"
	"time"
)

func WebsocketDail(server string, urlPath string, params url.Values) (*websocket.Conn, error) {
	u, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, urlPath)
	if u.Scheme == "http" {
		u.Scheme = "ws"
	} else if u.Scheme == "https" {
		u.Scheme = "wss"
	}
	u.RawQuery = params.Encode()

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	return c, err
}

func WebsocketClose(conn *websocket.Conn) error {
	return WebsocketCloseWithCode(conn, websocket.CloseNormalClosure, "")
}

func WebsocketCloseWithCode(conn *websocket.Conn, code int, text string) error {
	_ = conn.WriteControl(websocket.CloseMessage,
		websocket.FormatCloseMessage(code, text),
		time.Now().Add(time.Second))
	return conn.Close()
}
