package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/alex123012/dependency-bot/pkg/structs"
)

func main() {
	headerName, token := "PRIVATE-TOKEN", os.Getenv("TOKEN")
	req, err := http.NewRequest("GET", "https://gitlab.com/api/v4/projects/33213896/repository/compare?from=lol&to=main", nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set(headerName, token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	compare := new(structs.Compare)
	err = getJsonStruct(resp, compare)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(compare)
}

func getJsonStruct(resp *http.Response, jsonVar interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(jsonVar)
}
