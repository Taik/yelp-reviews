package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/rs/xhandler"
	"github.com/rs/xmux"
	"golang.org/x/net/context"
)

type yelpReviewFilter struct {
	Type  string
	Value string
}

type yelpReviewRequest struct {
	URL     string
	filters []yelpReviewFilter
}

func yelpReviewHandle(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	request := &yelpReviewRequest{}

	err := decoder.Decode(request)
	if err != nil {
		fmt.Fprintf(w, `{"status": "ERROR", "msg": "%s"}`, err)
		return
	}

	fmt.Printf("%v\n", request)

	fmt.Fprintf(w, `{"status": "OK", "msg": "It works!"}`)
}

func main() {
	c := xhandler.Chain{}

	c.UseC(xhandler.CloseHandler)
	c.Use(handlers.CompressHandler)
	c.Use(func(next http.Handler) http.Handler {
		return handlers.ContentTypeHandler(next, "application/json")
	})
	c.Use(func(next http.Handler) http.Handler {
		return handlers.LoggingHandler(os.Stdout, next)
	})

	mux := xmux.New()
	mux.POST("/", xhandler.HandlerFuncC(yelpReviewHandle))

	if err := http.ListenAndServe(":1234", c.Handler(mux)); err != nil {
		log.Fatal(err)
	}
}
