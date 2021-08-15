package main 

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func doREST(req *http.Request) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("failed to get http response: %q\n", err)
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("failed to read http response: %q\n", err)
		return "", err
	}

	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, body, "", "  ")

	return prettyJSON.String(), nil
}
