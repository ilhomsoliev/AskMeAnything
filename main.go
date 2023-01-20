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
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	r := mux.NewRouter()
	r.HandleFunc("/ask", AskQuestion).Methods("POST")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		res, _ := json.Marshal("Hello world")
		w.Write(res)
	})
	fmt.Print("IT is running actually")
	log.Fatal(http.ListenAndServe(":8080", r))
}

type Question struct {
	Question string `json:"question"`
}

type Answer struct {
	Answer string `json:"answer"`
}

func AskQuestion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatalln("Api Key is missing")
	}
	ctx := context.Background()
	client := gpt3.NewClient(apiKey)
	question := &Question{}
	ParseBody(r, question)
	answer := &Answer{Answer: ""}
	err := client.CompletionStreamWithEngine(ctx, gpt3.TextDavinci003Engine, gpt3.CompletionRequest{
		Prompt: []string{
			question.Question,
		},
		MaxTokens:   gpt3.IntPtr(3000),
		Temperature: gpt3.Float32Ptr(0),
	}, func(resp *gpt3.CompletionResponse) {
		answer.Answer += resp.Choices[0].Text
		//fmt.Print(resp.Choices[0].Text)
	})
	if err != nil {
		log.Fatal(err)
	}
	res, _ := json.Marshal(answer)
	w.WriteHeader(http.StatusOK)
	w.Write(res)
	fmt.Printf("\n")
	//GetResponse(client, ctx, question)

}

func GetResponse(client gpt3.Client, ctx context.Context, question string) {

}

func ParseBody(r *http.Request, x interface{}) {
	if body, err := ioutil.ReadAll(r.Body); err == nil {
		if err := json.Unmarshal([]byte(body), x); err != nil {
			return
		}
	}
}
