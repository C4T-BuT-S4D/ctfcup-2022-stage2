package main

import (
	"log"

	"proxyz/pkg/pb/zavod"

	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	b, err := protojson.Marshal(&zavod.LoginResponse{
		Result: &zavod.LoginResponse_Error{Error: "invalid credentials"},
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(string(b))

	b, err = protojson.Marshal(&zavod.LoginResponse{
		Result: &zavod.LoginResponse_Token{Token: "aboba"},
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(string(b))

	var response zavod.LoginResponse
	if err := protojson.Unmarshal([]byte(`{"error":"invalid credentials"}`), &response); err != nil {
		log.Fatalln(err)
	}
	log.Println(response.GetError())
}
