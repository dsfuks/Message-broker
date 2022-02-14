package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Время отправки GET запроса.
var getTime = make(map[string][]time.Time)

type Queue struct {
	Data map[string][]string
}

func NewQueue() *Queue {
	return &Queue{Data: make(map[string][]string)}
}
func (q *Queue) handler(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPut {
		q.putMessage(w, req)
	} else if req.Method == http.MethodGet {
		q.getMessage(w, req)
	}
}
func (q *Queue) putMessage(w http.ResponseWriter, req *http.Request) {
	// Имя очереди.
	qName := req.URL.Path[1:]
	// Сообщение, которое надо положить в очередь.
	qMessage := req.URL.Query().Get("v")
	if qMessage == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	q.Data[qName] = append(q.Data[qName], qMessage)
	w.Write(nil)
}
func (q *Queue) getMessage(w http.ResponseWriter, req *http.Request) {
	// Имя очереди.
	qName := req.URL.Path[1:]
	gettingTime := time.Now()
	// Добавляем время отправки запроса на случай возникновения очереди из получателей.
	getTime[qName] = append(getTime[qName], gettingTime)
	// True, если сообщение пришло во время ожидания timeout.
	var isSent = false
	if len(q.Data[qName]) == 0 {
		if tm, err := strconv.Atoi(req.URL.Query().Get("timeout")); err == nil {
			for i := 0; i < tm; i++ {
				if len(q.Data[qName]) != 0 {
					// Проверяем первый ли в очереди на получение.
					if gettingTime == getTime[qName][0] {
						isSent = true
						break
					}
				}
				time.Sleep(time.Second)
			}
		}
		if !isSent {
			http.Error(w, "", http.StatusNotFound)
			return
		}
	}
	js, err := json.Marshal(q.Data[qName][0])
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
	}
	q.Data[qName] = q.Data[qName][1:]
	w.Header().Set("Content-Type", "application/json")
	getTime[qName] = getTime[qName][1:]
	w.Write(js)
}
func main() {
	server := NewQueue()
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handler)
	log.Fatal(http.ListenAndServe("localhost:"+os.Args[1], mux))
}
