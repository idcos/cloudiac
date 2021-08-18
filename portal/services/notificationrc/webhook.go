package notificationrc

import "github.com/unliar/utils/go/http"

type Webhook struct {
	Url string
}

func (w Webhook) Send(massage string) error {
	baseURL := w.Url
	_, err := http.Post(baseURL, massage, nil)
	if err != nil {
		return err
	}
	return nil
}
