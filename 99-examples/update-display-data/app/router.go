package app

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
)

const plateCtxKey = "plateContext"

func Router() http.Handler {
	r := chi.NewRouter()

	// ログ量を減らしたい場合はアクセスログは無効にしても良いかも
	r.Use(middleware.Logger)

	r.Get("/", healthCheck)
	r.Post("/", loggingEventData)
	return r
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	// 確認用。必要なら、何か参照すべき
	Succeed(w, struct {
		Status string `json:"status"`
	}{string("healthy")})
}

func loggingEventData(w http.ResponseWriter, r *http.Request) {
	var jsonBody map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&jsonBody)
	if err != nil {
		Logger.Warn("Request Body is invalid:", zap.Any("body", r.Body), zap.Error(err))
		Fail(w, http.StatusBadRequest)
		return
	}

	Logger.Info("logging event request data:", zap.Any("header", r.Header), zap.Any("body", jsonBody))
	Succeed(w, struct {
		Status string `json:"status"`
	}{string("healthy")})
}
