// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package utils

import (
	"bytes"
	"cloudiac/utils/logs"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func httpClient(conntimeout, deadline int) *http.Client {
	c := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(time.Duration(deadline) * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Duration(conntimeout)*time.Second)
				if err != nil {
					return nil, err
				}
				return c, c.SetDeadline(deadline)
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	return c
}

func HttpService(reqUrl, method string, header *http.Header, data interface{}, conntimeout, deadline int) ([]byte, error) {
	c := httpClient(conntimeout, deadline)

	var req *http.Request
	var err error
	if header == nil {
		header = &http.Header{}
	}
	if header.Get("Content-Type") == "" {
		header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if http.MethodPost == method {
		if data != nil {
			if header.Get("Content-Type") == "application/json" {
				b, err := json.Marshal(data)
				if err != nil {
					return nil, err
				}
				req, err = http.NewRequest(method, reqUrl, bytes.NewReader(b))
			} else if value, ok := data.(string); ok {
				req, err = http.NewRequest(method, reqUrl, bytes.NewReader([]byte(value)))
			} else {
				return nil, fmt.Errorf("params err")
			}
		} else {
			req, err = http.NewRequest(method, reqUrl, nil)
		}
	} else if http.MethodDelete == method {
		if data != nil {
			if header.Get("Content-Type") == "application/json" {
				b, err := json.Marshal(data)
				if err != nil {
					return nil, err
				}
				req, err = http.NewRequest(method, reqUrl, bytes.NewReader(b))
			} else if value, ok := data.(string); ok {
				req, err = http.NewRequest(method, reqUrl, bytes.NewReader([]byte(value)))
			} else {
				return nil, fmt.Errorf("params err")
			}
		} else {
			req, err = http.NewRequest(method, reqUrl, nil)
		}
	} else if http.MethodPut == method {
		if data != nil {
			if header.Get("Content-Type") == "application/json" {
				b, err := json.Marshal(data)
				if err != nil {
					return nil, err
				}
				req, err = http.NewRequest(method, reqUrl, bytes.NewReader(b))
			} else if value, ok := data.(string); ok {
				req, err = http.NewRequest(method, reqUrl, bytes.NewReader([]byte(value)))
			} else {
				return nil, fmt.Errorf("params err")
			}
		} else {
			req, err = http.NewRequest(method, reqUrl, nil)
		}
	} else if http.MethodGet == method {
		req, err = http.NewRequest(method, reqUrl, nil)
	}
	if header != nil {
		req.Header = *header
	}

	if err != nil {
		return nil, err
	}

	logs.Get().Debugf("%s %s", req.Method, req.URL.String())
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resp != nil {
			_ = resp.Body.Close()
		}
	}()

	return ioutil.ReadAll(resp.Body)
}

type FormPart struct {
	Key  string
	Name string
	Part io.Reader
}

func HttpPostFiles(reqUrl string, header *http.Header, formParts []FormPart, connTimeout, deadline int) (resp *http.Response, err error) {
	c := httpClient(connTimeout, deadline)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for idx := range formParts {
		part := formParts[idx]
		var fw io.Writer
		if x, ok := part.Part.(*os.File); ok {
			if fw, err = w.CreateFormFile(part.Key, filepath.Base(x.Name())); err != nil {
				return
			}
		} else {
			if fw, err = w.CreateFormField(part.Key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, part.Part); err != nil {
			return
		}
	}
	if err = w.Close(); err != nil {
		return
	}

	var req *http.Request
	req, err = http.NewRequest(http.MethodPost, reqUrl, &b)
	if err != nil {
		return
	}
	if header != nil {
		req.Header = *header
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err = c.Do(req)
	return
}
