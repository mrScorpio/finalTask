package handlers

import (
	"net/http"
	"os"
)

func HandleMain(w http.ResponseWriter, req *http.Request) {
	indexContent, err := os.ReadFile("./web/index.html")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(indexContent)
}
