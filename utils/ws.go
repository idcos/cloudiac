package utils

import (
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"path"
	"time"
)

func WebsocketDail(server string, urlPath string, params url.Values) (*websocket.Conn, *http.Response, error) {
	u, err := url.Parse(server)
	if err != nil {
		return nil, nil, err
	}

	u.Path = path.Join(u.Path, urlPath)
	if u.Scheme == "http" {
		u.Scheme = "ws"
	} else if u.Scheme == "https" {
		u.Scheme = "wss"
	}
	u.RawQuery = params.Encode()

	c, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	return c, resp, err
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
