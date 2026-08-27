package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const bucket = "creator-nightly-snapshots"

func main() {
	var input CreatorSnapshot
	if len(os.Args) > 1 {
		data, err := os.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(data, &input); err != nil {
			panic(err)
		}
	}
	if !shouldSnapshot(input) {
		fmt.Println("snapshot skipped: processing is incomplete or there are no subscribers")
		return
	}
	client, err := newStorageClient()
	if err != nil {
		panic(err)
	}
	if err := client.createBucket(bucket); err != nil {
		panic(err)
	}
	key := input.CreatorID + "/" + time.Now().UTC().Format("2006-01-02") + ".json"
	payload, err := json.Marshal(struct {
		CreatedAt string          `json:"created_at"`
		Data      CreatorSnapshot `json:"data"`
	}{time.Now().UTC().Format(time.RFC3339), input})
	if err != nil {
		panic(err)
	}
	url, err := client.presign(bucket, key)
	if err != nil {
		panic(err)
	}
	if err := putSignedJSON(url, payload); err != nil {
		panic(err)
	}
	fmt.Printf("snapshot uploaded: bucket=%s key=%s\n", bucket, key)
}
