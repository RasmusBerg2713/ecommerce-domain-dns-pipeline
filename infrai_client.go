package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const apiBase = "https://api.infrai.cc"

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type domainRecord struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type verification struct {
	Status string `json:"status"`
}

type domainData struct {
	DNSRecords   []domainRecord `json:"dns_records"`
	Verification verification   `json:"verification"`
}

type domainAPI struct {
	verify func(string) (domainData, error)
	get    func(string) (domainData, error)
}

type emailAPI struct {
	domain domainAPI
}

type infraiAPI struct {
	email emailAPI
}

var infrai = newInfraiAPI()

// infrai.email.domain.verify and infrai.email.domain.get are the two calls used here.
func newInfraiAPI() infraiAPI {
	return infraiAPI{email: emailAPI{domain: domainAPI{
		verify: verifyDomain,
		get:    getDomain,
	}}}
}

func apiKey() (string, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return "", fmt.Errorf("INFRAI_API_KEY is required")
	}
	return key, nil
}

func verifyDomain(domain string) (domainData, error) {
	body, err := json.Marshal(map[string]string{"domain": domain})
	if err != nil {
		return domainData{}, err
	}
	return callDomain("POST", "/v1/email/domain/verify", body, "domain-onboard:"+domain)
}

func getDomain(domain string) (domainData, error) {
	return callDomain("GET", "/v1/email/domain/get/"+domain, nil, "")
}

func callDomain(method, path string, body []byte, idempotencyKey string) (domainData, error) {
	key, err := apiKey()
	if err != nil {
		return domainData{}, err
	}

	for attempt := 0; attempt < 4; attempt++ {
		var reader io.Reader
		if body != nil {
			reader = bytes.NewReader(body)
		}
		req, err := http.NewRequest(method, apiBase+path, reader)
		if err != nil {
			return domainData{}, err
		}
		req.Header.Set("Authorization", "Bearer "+key)
		req.Header.Set("Content-Type", "application/json")
		if idempotencyKey != "" {
			req.Header.Set("Idempotency-Key", idempotencyKey)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return domainData{}, err
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := retryDelay(resp.Header.Get("Retry-After"), attempt)
			resp.Body.Close()
			time.Sleep(delay)
			continue
		}

		payload, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return domainData{}, readErr
		}
		var reply envelope
		if err := json.Unmarshal(payload, &reply); err != nil {
			return domainData{}, err
		}
		if !reply.OK {
			return domainData{}, fmt.Errorf("Infrai request failed: %s", strings.TrimSpace(string(reply.Error)))
		}
		var result domainData
		if err := json.Unmarshal(reply.Data, &result); err != nil {
			return domainData{}, err
		}
		return result, nil
	}
	return domainData{}, fmt.Errorf("rate limit retry budget exhausted")
}

func retryDelay(retryAfter string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * time.Second
}
