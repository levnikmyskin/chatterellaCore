package main

import (
	"log"
	"net/http"
)

func replyBadRequest(w http.ResponseWriter, msg []byte) {
	w.WriteHeader(http.StatusBadRequest)
	sendReply(w, msg)
}

func replyInternalServerError(w http.ResponseWriter, msg []byte) {
	w.WriteHeader(http.StatusInternalServerError)
	sendReply(w, msg)
}

func sendReply(w http.ResponseWriter, msg []byte) {
	_, err := w.Write(msg)
	if err != nil {
		log.Printf("Error while replying to client: %s", err)
	}
}
