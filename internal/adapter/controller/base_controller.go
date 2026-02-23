package controller

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
)

// baseController 各WebController共通のレスポンス書き込みロジック
type baseController struct{}

// writePresenterResponse プレゼンターの出力を再エンコードせず直接書き込む
func (bc *baseController) writePresenterResponse(w rest.ResponseWriter, responseStr string, errorOutput any) {
	if responseStr == "" || !json.Valid([]byte(responseStr)) {
		w.WriteHeader(http.StatusBadRequest)
		if err := w.WriteJson(errorOutput); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := w.WriteJson(json.RawMessage(responseStr)); err != nil {
		log.Printf("WriteJson error: %v", err)
	}
}
