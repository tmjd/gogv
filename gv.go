package gv

import (
	"errors"
	"fmt"
	"golang.org/x/net/publicsuffix"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"regexp"
)

var ErrLoginFailed = errors.New("Failed to login")
var ErrLogoutFailed = errors.New("Failed to logout")
var ErrSendSmsFailed = errors.New("Failed to send sms")
var ErrNotLoggedIn = errors.New("Not logged in")

type GV struct {
	http_client http.Client
	rnr_se      string //special code for connections?
	ShowStatus  bool
	logged_in   bool
}

func (g *GV) Login(email string, password string) error {
	g.logged_in = false
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	var err error
	g.http_client.Jar, err = cookiejar.New(&options)
	if err != nil {
		return ErrLoginFailed
	}

	var resp *http.Response
	resp, err = g.http_client.Get(
		"https://accounts.google.com/ServiceLogin?service=grandcentral")
	if err != nil {
		return ErrLoginFailed
	}

	if g.ShowStatus {
		fmt.Println("Primer request " + resp.Status)
	}
	body, err := httputil.DumpResponse(resp, true)
	resp.Body.Close()
	if err != nil {
		return ErrLoginFailed
	}
	re := regexp.MustCompile(`name="GALX" .*\s*value="(.+)"`)
	galx := string(re.FindSubmatch(body)[1])
	re = regexp.MustCompile(`name="service" .*\s*value="(.+)"`)
	service := string(re.FindSubmatch(body)[1])
	re = regexp.MustCompile(`name="_utf8" *value="(.+)"`)
	utf8 := string(re.FindSubmatch(body)[1])
	re = regexp.MustCompile(`name="bgresponse" .* *value="(.+)"`)
	bgresponse := string(re.FindSubmatch(body)[1])

	pstMsg := "1"

	resp, err = g.http_client.PostForm(
		"https://accounts.google.com/ServiceLogin?service=grandcentral",
		url.Values{"Email": {email}, "Passwd": {password},
			"GALX": {galx}, "_utf8": {utf8},
			"bgresponse": {bgresponse}, "pstMsg": {pstMsg},
			"service":  {service},
			"continue": {"https://www.google.com/voice/"},
			"followup": {"https://www.google.com/voice/"}})
	if err != nil {
		return ErrLoginFailed
	}

	if g.ShowStatus {
		fmt.Println("login respone: " + resp.Status)
	}
	body, err = httputil.DumpResponse(resp, true)
	resp.Body.Close()
	if err != nil {
		return ErrLoginFailed
	}

	re = regexp.MustCompile("'_rnr_se': '(.+)'")
	g.rnr_se = string(re.FindSubmatch(body)[1])

	if len(g.rnr_se) == 0 {
		return ErrLoginFailed
	}
	g.logged_in = true

	return nil
}

func (g *GV) SendSms(phoneNumber string, text string) error {
	if g.logged_in == false {
		return ErrNotLoggedIn
	}

	resp, err := g.http_client.PostForm(
		"https://www.google.com/voice/sms/send",
		url.Values{"phoneNumber": {phoneNumber}, "text": {text},
			"_rnr_se": {g.rnr_se}})
	defer resp.Body.Close()
	if g.ShowStatus {
		fmt.Println("Send SMS response " + resp.Status)
	}
	_, err = httputil.DumpResponse(resp, true)
	if err != nil {
		return ErrSendSmsFailed
	}

	return nil
}

func (g *GV) Logout() error {
	if g.logged_in == false {
		return nil
	}
	g.logged_in = false
	resp, err := g.http_client.Get(
		"https://www.google.com/voice/account/signout",
	)
	if err != nil {
		return ErrLogoutFailed
	}

	defer resp.Body.Close()
	if g.ShowStatus {
		fmt.Println("Logout response status " + resp.Status)
	}
	_, err = httputil.DumpResponse(resp, true)
	if err != nil {
		return ErrSendSmsFailed
	}
	return nil
}
