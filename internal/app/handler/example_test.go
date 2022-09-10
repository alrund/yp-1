package handler_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func ExampleAdd() {
	serverAddress := "http://localhost:8080"
	endpoint := "/"
	stringData := "https://ya.ru"

	r, err := http.Post(
		serverAddress+endpoint,
		"text/plain",
		bytes.NewBufferString(stringData),
	)

	if err != nil {
		fmt.Println("get error", err)
		return
	}

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("read error", err)
		return
	}

	fmt.Println(string(buf))
	// serverAddress + "/oTHlXx"
}

func ExampleAddJSON() {
	serverAddress := "http://localhost:8080"
	endpoint := "/api/shorten"
	data := `{"url": "https://ya.ru"}`

	r, err := http.Post(
		serverAddress+endpoint,
		"application/json; charset=utf-8",
		strings.NewReader(data),
	)

	if err != nil {
		fmt.Println("get error", err)
		return
	}

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("read error", err)
		return
	}

	fmt.Println(string(buf))
	// {"result":"http://localhost:8080/oTHlXx"}
}

func ExampleAddBatchJSON() {
	serverAddress := "http://localhost:8080"
	endpoint := "/api/shorten/batch"
	data := `[{"correlation_id":"xxx","original_url":"https://ya.ru"},{"correlation_id":"yyy","original_url":"https://google.com"}]`

	r, err := http.Post(
		serverAddress+endpoint,
		"application/json; charset=utf-8",
		strings.NewReader(data),
	)

	if err != nil {
		fmt.Println("get error", err)
		return
	}

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("read error", err)
		return
	}

	fmt.Println(string(buf))
	//	[
	//		{"correlation_id":"xxx","short_url":"http://localhost:8080/oTHlXx"},
	//		{"correlation_id":"yyy","short_url":"http://localhost:8080/FaMvXd"}
	//	]
}

func ExampleDeleteURLs() {
	serverAddress := "http://localhost:8080"
	endpoint := "/api/user/urls"
	data := `["oTHlXx", "bjHoyQ"]`

	client := &http.Client{}
	req, err := http.NewRequest("DELETE", serverAddress+endpoint, strings.NewReader(data))
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

func ExampleGet() {
	serverAddress := "http://localhost:8080"
	endpoint := "/oTHlXx"

	r, err := http.Get(serverAddress + endpoint)

	if err != nil {
		fmt.Println("get error", err)
		return
	}

	fmt.Println(r.StatusCode)
	// 200
}

func ExampleGetUserURLs() {
	serverAddress := "http://localhost:8080"
	endpoint := "/api/user/urls"

	r, err := http.Get(serverAddress + endpoint)

	if err != nil {
		fmt.Println("get error", err)
		return
	}

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("read error", err)
		return
	}

	fmt.Println(string(buf))
	//	[
	//		{"short_url": "http://localhost:8080/koRTZS", "original_url": "https://google.ru"}
	//	]
}
