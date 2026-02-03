package main

import (
	"fmt"
	// "github.com/joho/godotenv"
	"io"
	"net/http"
	"os"
	// "strings"
)

func e() {
	url := "https://margincalculator.angelbroking.com/OpenAPI_File/files/OpenAPIScripMaster.json"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Request creation error:", err)
		return
	}
	res, err := client.Do(req)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}
	if err := os.WriteFile("dictionary.json", body, 0644); err != nil {
		fmt.Println("File write error:", err)
		return
	}
	fmt.Println("Saved response to dictionary.json")
}
