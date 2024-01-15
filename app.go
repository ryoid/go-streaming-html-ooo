package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

//go:embed templates/*
var resources embed.FS

var t = template.Must(template.ParseFS(resources, "templates/*"))

type Food struct {
	Id   int
	Name string
}

func sendFoodInDelayedOrder(data []Food, order []int, ch chan Food) {
	go func() {
		for _, index := range order {
			time.Sleep(500 * time.Millisecond)
			ch <- data[index]
		}
		close(ch)
	}()
}

var Foods = []Food{
	{Id: 1, Name: "Ice Cream"},
	{Id: 2, Name: "Pizza"},
	{Id: 3, Name: "Chocolate"},
	{Id: 4, Name: "Cheseburger"},
	{Id: 5, Name: "Oreo"},
}

type AwaitedSlot struct {
	Slot string
	Html string
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"

	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		ch := make(chan Food)
		// Send the food to channel in a delayed fixed order
		go sendFoodInDelayedOrder(Foods, []int{4, 0, 3, 2, 1}, ch)

		t.ExecuteTemplate(w, "head", nil)
		w.(http.Flusher).Flush()

		time.Sleep(1 * time.Second)
		t.ExecuteTemplate(w, "content", nil)
		w.(http.Flusher).Flush()

		// Stream the food to the browser, item by item
		for item := range ch {
			t.ExecuteTemplate(w, "slot", AwaitedSlot{
				Slot: fmt.Sprintf("slot-%d", item.Id),
				Html: item.Name,
			})
			w.(http.Flusher).Flush()
		}

		t.ExecuteTemplate(w, "tail", nil)
		w.(http.Flusher).Flush()
	})

	log.Println("listening on", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
