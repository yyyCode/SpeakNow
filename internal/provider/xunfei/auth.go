package xunfei

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const defaultHostURL = "wss://iat-api.xfyun.cn/v2/iat"

func assembleAuthURL(hostURL, apiKey, apiSecret string) (string, error) {
	ul, err := url.Parse(hostURL)
	if err != nil {
		return "", fmt.Errorf("parse host url: %w", err)
	}

	date := time.Now().UTC().Format(time.RFC1123)
	signString := strings.Join([]string{
		"host: " + ul.Host,
		"date: " + date,
		"GET " + ul.Path + " HTTP/1.1",
	}, "\n")

	signature := hmacSHA256Base64(signString, apiSecret)
	authURL := fmt.Sprintf(
		`hmac username="%s", algorithm="hmac-sha256", headers="host date request-line", signature="%s"`,
		apiKey, signature,
	)
	authorization := base64.StdEncoding.EncodeToString([]byte(authURL))

	v := url.Values{}
	v.Add("host", ul.Host)
	v.Add("date", date)
	v.Add("authorization", authorization)

	return hostURL + "?" + v.Encode(), nil
}

func hmacSHA256Base64(data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
