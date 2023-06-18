package handlers

import (
	"log"
	"net/http"
)

func WriteError(w http.ResponseWriter, err error) {
	Write(w, []byte(`{"error": "`+err.Error()+`"}`))
}

func Write(w http.ResponseWriter, bz []byte) {
	if _, err := w.Write(bz); err != nil {
		log.Default().Println(err)
	}
}
