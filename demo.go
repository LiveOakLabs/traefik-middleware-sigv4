package traefik_middleware_sigv4

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Config struct {
	AccessKey    string  `json:"accessKey"`
	SecretKey    string  `json:"secretKey"`
	SessionToken *string `json:"sessionToken,omitempty"`
	Service      string  `json:"service"`
	Endpoint     string  `json:"endpoint"`
	Region       string  `json:"region"`
}

type Plugin struct {
	name         string
	next         http.Handler
	AccessKey    string  `json:"accessKey"`
	SecretKey    string  `json:"secretKey"`
	SessionToken *string `json:"sessionToken,omitempty"`
	Service      string  `json:"service"`
	Endpoint     string  `json:"endpoint"`
	Region       string  `json:"region"`
}

func (p *Plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	host := p.Endpoint
	uri := r.URL.Path
	query := r.URL.RawQuery

	now := time.Now().UTC()
	dateStamp := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")

	signedHeaders := "host;x-amz-date"
	algorithm := "AWS4-HMAC-SHA256"

	var payload []byte
	if r.Body != nil {
		payload, _ = io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(payload))
	}
	payloadHash := sha256.Sum256(payload)

	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-date:%s\n",
		host, amzDate)

	if p.SessionToken != nil {
		r.Header.Set("X-Amz-Security-Token", *p.SessionToken)
		canonicalHeaders += fmt.Sprintf("x-amz-security-token:%s\n", *p.SessionToken)
		signedHeaders += ";x-amz-security-token"
	}

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		method, uri, query, canonicalHeaders, signedHeaders, hex.EncodeToString(payloadHash[:]))
	canonicalRequestHash := sha256.Sum256([]byte(canonicalRequest))
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, p.Region, p.Service)
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s",
		algorithm, amzDate, credentialScope, hex.EncodeToString(canonicalRequestHash[:]))

	signingKey := getSigningKey(p.SecretKey, dateStamp, p.Region, p.Service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))

	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm, p.AccessKey, credentialScope, signedHeaders, signature)

	r.Header.Set("X-Amz-Date", amzDate)
	r.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(payloadHash[:]))
	r.Header.Set("Authorization", authorization)

	p.next.ServeHTTP(w, r)
}

func CreateConfig() *Config {
	return &Config{
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}
}

func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &Plugin{
		name:         name,
		next:         next,
		AccessKey:    config.AccessKey,
		SecretKey:    config.SecretKey,
		SessionToken: config.SessionToken,
		Service:      config.Service,
		Endpoint:     config.Endpoint,
		Region:       config.Region,
	}, nil
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

func getSigningKey(secretKey, date, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "aws4_request")
	return kSigning
}
