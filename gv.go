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

type GV struct {
	login_email string
	password    string
	page_map    map[string]string
	http_client http.Client
	rnr_se      string //special code for connections?
}

func (g *GV) Login(email string, passwd string) error {
	g.login_email = email
	g.password = passwd

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

	fmt.Println(resp.Status)
	body, err := httputil.DumpResponse(resp, true)
	resp.Body.Close()
	if err != nil {
		return ErrLoginFailed
	}
	//fmt.Println(string(body[:]))
	galx := string(regexp.MustCompile(
		`name="GALX" type="hidden"\s*value="(.+)"`).FindSubmatch(body)[1])
	service := string(regexp.MustCompile(
		`name="service" type="hidden"\s*value="(.+)"`).FindSubmatch(body)[1])
	utf8 := string(regexp.MustCompile(
		`type="hidden" .* name="_utf8" *value="(.+)"`).FindSubmatch(body)[1])
	bgresponse := string(regexp.MustCompile(
		`type="hidden" name="bgresponse" .* *value="(.+)"`).FindSubmatch(body)[1])

	pstMsg := "1"

	fmt.Println("Galx " + string(galx))
	fmt.Println("Service " + string(service))
	fmt.Println("Utf8 " + string(utf8))
	fmt.Println("bgresponse " + string(bgresponse))

	resp, err = g.http_client.PostForm(
		"https://accounts.google.com/ServiceLogin?service=grandcentral",
		url.Values{"Email": {g.login_email}, "Passwd": {g.password},
			"GALX": {galx}, "_utf8": {utf8},
			"bgresponse": {bgresponse}, "pstMsg": {pstMsg},
			"service":  {service},
			"continue": {"https://www.google.com/voice/"},
			"followup": {"https://www.google.com/voice/"}})
	if err != nil {
		return ErrLoginFailed
	}

	fmt.Println(resp.Status)
	body, err = httputil.DumpResponse(resp, true)
	resp.Body.Close()
	if err != nil {
		return ErrLoginFailed
	}

	g.rnr_se = string(regexp.MustCompile(
		"'_rnr_se': '(.+)'").FindSubmatch(body)[1])
	//fmt.Println(string(body[:]))
	fmt.Println("rnr_se " + g.rnr_se)

	return ErrLoginFailed
}

func (g *GV) SendSms(phoneNumber string, text string) error {

	resp, err := g.http_client.PostForm(
		"https://www.google.com/voice/sms/send",
		url.Values{"phoneNumber": {phoneNumber}, "text": {text},
			"_rnr_se": {g.rnr_se}})
	defer resp.Body.Close()
	fmt.Println("Send SMS response " + resp.Status)
	body, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return ErrSendSmsFailed
	}

	fmt.Println(string(body[:]))

	return nil
}

func (g *GV) Logout() error {
	resp, err := g.http_client.Get(
		"https://www.google.com/voice/account/signout",
	)
	if err != nil {
		return ErrLogoutFailed
	}

	defer resp.Body.Close()
	fmt.Println("Send SMS response " + resp.Status)
	body, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return ErrSendSmsFailed
	}
	fmt.Println(string(body[:]))
	return nil
}
