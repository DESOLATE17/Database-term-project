package utils

import (
	"encoding/json"
	"fmt"
	"github.com/DESOLATE17/Database-term-project/internal/models"
	"net/http"
)

func Response(w http.ResponseWriter, status int, body interface{}) {
	if body != nil {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(status)
	if status == http.StatusNotFound && body != nil {
		jsn, err := json.Marshal(models.ErrorResponse{Message: fmt.Sprintf("Can't find user with id #%s\\n", body)})
		if err != nil {
			return
		}
		_, _ = w.Write(jsn)
		return
	}
	if body != nil {
		jsn, err := json.Marshal(body)
		if err != nil {
			return
		}
		_, _ = w.Write(jsn)
	}
}
