package traefik_middleware_sigv4_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	plugin "github.com/LiveOakLabs/traefik_middleware_sigv4"

	"testing"
)

func TestHandler(t *testing.T) {
	c := plugin.CreateConfig()
	c.Service = "lambda"
	c.Endpoint = "s6jt2rgliirxvwbhfq7bpbakoe0aigbr.lambda-url.us-east-1.on.aws"
	c.Region = "us-east-1"

	ctx := context.Background()

	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	handler, err := plugin.New(ctx, next, c, "foo")

	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	reqURL := fmt.Sprintf("https://%s/health", c.Endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	body, err := io.ReadAll(recorder.Result().Body)

	if err != nil {
		t.Fatal(err)
	}

	if recorder.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v: %v", recorder.Result().StatusCode, string(body))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
	}
	defer resp.Body.Close()

	log.Println(req.RemoteAddr, " ", resp.Status)
}
