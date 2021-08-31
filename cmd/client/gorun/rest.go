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
	body, err := readAll(resp.Body)
	if err != io.EOF {
		fmt.Printf("failed to read http response: %q\n", err)
		return "", err
	}

	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, body, "", "  ")

	return prettyJSON.String(), nil
}

func readAll(in io.ReadCloser) ([]byte, error) {
	var out bytes.Buffer
	var buf = make([]byte, 4*1024)
	for {
		n, err := in.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
		}

		if err != nil {
			return out.Bytes(), err
		}
	}
}
