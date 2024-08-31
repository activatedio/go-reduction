package testing

import (
	"github.com/go-resty/resty/v2"
	"net/http"
)

type restyCookieStore struct {
	cookies map[string]*http.Cookie
}

func (s *restyCookieStore) onAfterResponse(client *resty.Client, response *resty.Response) error {

	for _, c := range response.Cookies() {
		if c.MaxAge < 0 {
			delete(s.cookies, c.Name)
		} else {
			s.cookies[c.Name] = c
		}
	}

	return nil
}

func (s *restyCookieStore) onBeforeRequest(client *resty.Client, request *resty.Request) error {
	var cookies []*http.Cookie
	for _, v := range s.cookies {
		cookies = append(cookies, v)
	}
	request.Cookies = cookies
	return nil
}

func RestyWithCookieSupport(client *resty.Client) *resty.Client {
	cs := &restyCookieStore{
		cookies: make(map[string]*http.Cookie),
	}

	client.OnAfterResponse(cs.onAfterResponse)
	client.OnBeforeRequest(cs.onBeforeRequest)
	return client
}
