package utils

import (
	"github.com/gorilla/websocket"
	"net/url"
	"path"
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
