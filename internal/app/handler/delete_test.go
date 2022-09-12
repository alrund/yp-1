package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func ExampleDeleteURLs() {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodDelete,
		"http://localhost:8080/api/user/urls",
		strings.NewReader(`["oTHlXx", "bjHoyQ"]`),
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
