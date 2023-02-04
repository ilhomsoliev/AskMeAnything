package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/gorilla/websocket"

	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{}

func main() {
	godotenv.Load()
	http.HandleFunc("/todo", func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade failed: ", err)
			return
		}
		defer conn.Close()

		// Continuosly read and write message
		for {
			mt, output, err := conn.ReadMessage()
			questionActual := string(output)
			// ->
			apiKey := os.Getenv("API_KEY")
			if apiKey == "" {
				log.Fatalln("Api Key is missing")
			}
			ctx := context.Background()
			client := gpt3.NewClient(apiKey)
			question := &Question{}
			question.Question = questionActual
			fmt.Println(questionActual)
			client.CompletionStreamWithEngine(ctx, gpt3.TextDavinci003Engine, gpt3.CompletionRequest{
				Prompt: []string{
					question.Question,
				},
				MaxTokens:   gpt3.IntPtr(3000),
				Temperature: gpt3.Float32Ptr(0),
			}, func(resp *gpt3.CompletionResponse) {
				message := []byte(string(resp.Choices[0].Text))
				fmt.Println(string(resp.Choices[0].Text))
				err = conn.WriteMessage(mt, message)
				if err != nil {
					log.Println("write failed:", err)
					return
				}
			})
			err = conn.WriteMessage(mt, []byte(string("END")))
			if err != nil {
				log.Println("write failed:", err)
				return
			}
		}
	})

	fmt.Println("Server has started")
	http.ListenAndServe(":8080", nil)
}

type Question struct {
	Question string `json:"question"`
}

type Answer struct {
	Answer string `json:"answer"`
}

func ParseBody(r *http.Request, x interface{}) {
	if body, err := ioutil.ReadAll(r.Body); err == nil {
		if err := json.Unmarshal([]byte(body), x); err != nil {
			return
		}
	}
}
