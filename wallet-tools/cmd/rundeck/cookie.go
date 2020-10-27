package main

import (
	"net/http"
	"net/url"
)

type cookieJar struct {
	cookies []*http.Cookie
}

func newCookieJar() *cookieJar {
	return &cookieJar{}
}

func (j *cookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	if u.Path == authPath {
		j.cookies = cookies
	}
}

func (j *cookieJar) Cookies(u *url.URL) []*http.Cookie {
	return j.cookies
}
