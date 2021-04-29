package services

import (
	"fmt"
	"net/http"
)

func DoGiteaRequest(request *http.Request, token string) (*http.Response, error) {
	client := &http.Client{}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	response, err := client.Do(request)
	defer response.Body.Close()
	if err != nil {
		return nil, err
	}
	return response, nil
}
