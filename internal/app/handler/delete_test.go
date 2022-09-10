package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// nolint
func ExampleDeleteURLs() {
	serverAddress := "http://localhost:8080"
	endpoint := "/api/user/urls"
	data := `["oTHlXx", "bjHoyQ"]`

	client := &http.Client{}
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodDelete,
		serverAddress+endpoint,
		strings.NewReader(data),
	)
	if err != nil {
		fmt.Println("get error", err)
		return
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	r, err := client.Do(req)
	if err != nil {
		fmt.Println("get error", err)
		return
	}
	defer r.Body.Close()

	fmt.Println(r.StatusCode)
	// 202
}
