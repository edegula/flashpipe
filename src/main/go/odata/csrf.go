package odata

import (
	"fmt"
	"github.com/engswee/flashpipe/httpclnt"
	"github.com/engswee/flashpipe/logger"
	"net/http"
)

type Csrf struct {
	exe         *httpclnt.HTTPExecuter
	token       string
	csrfCookies []*http.Cookie
}

// NewCsrf returns an initialised Csrf instance.
func NewCsrf(exe *httpclnt.HTTPExecuter) *Csrf {
	c := new(Csrf)
	c.exe = exe
	return c
}

func (c *Csrf) GetToken() (string, []*http.Cookie, error) {
	if c.token == "" {
		logger.Debug("Get CSRF Token")
		headers := map[string]string{
			"x-csrf-token": "fetch",
		}
		resp, err := c.exe.ExecGetRequest("/api/v1/", headers)

		if err != nil {
			return "", nil, err
		}
		if resp.StatusCode == 200 {
			c.token = resp.Header.Get("x-csrf-token")
			c.csrfCookies = resp.Cookies()
			logger.Debug(fmt.Sprintf("Received CSRF Token - %v", c.token))
		} else {
			return "", nil, c.exe.LogError(resp, "Get CSRF Token")
		}
	}
	return c.token, c.csrfCookies, nil
}
