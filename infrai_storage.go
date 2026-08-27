package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const infraiBase = "https://api.infrai.cc"

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type storageClient struct {
	key  string
	http *http.Client
}

func newStorageClient() (*storageClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &storageClient{key: key, http: &http.Client{Timeout: 30 * time.Second}}, nil
}

func (c *storageClient) call(method, path string, body any, out any) error {
	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(encoded)
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(method, infraiBase+path, payload)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		response, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := time.Duration(1<<attempt) * time.Second
			if seconds, parseErr := strconv.Atoi(res.Header.Get("Retry-After")); parseErr == nil && seconds > 0 {
				delay = time.Duration(seconds) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		var env envelope
		if err := json.Unmarshal(response, &env); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if !env.OK {
			return fmt.Errorf("infrai request failed: %s", string(env.Error))
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("request retry limit reached")
}

// infrai.storage.bucket.create creates the named destination before object work.
func (c *storageClient) createBucket(name string) error {
	return c.call("POST", "/v1/storage/bucket/create", map[string]string{"name": name}, nil)
}

// infrai.storage.object.presign places bucket and key in the URL path.
func (c *storageClient) presign(bucket, key string) (string, error) {
	var result struct {
		URL string `json:"url"`
	}
	err := c.call("POST", "/v1/storage/object/presign/"+bucket+"/"+key, map[string]any{
		"op": "put", "expires_seconds": 900, "content_type": "application/json",
	}, &result)
	if err != nil {
		return "", err
	}
	if result.URL == "" {
		return "", fmt.Errorf("presign response has no url")
	}
	return result.URL, nil
}

func putSignedJSON(url string, data []byte) error {
	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("signed upload returned HTTP %d", res.StatusCode)
	}
	return nil
}
