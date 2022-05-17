package app

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

const plateCtxKey = "plateContext"

func Router() http.Handler {
	r := chi.NewRouter()

	// ログ量を減らしたい場合はアクセスログは無効にしても良いかも
	r.Use(middleware.Logger)

	r.Get("/", healthCheck)
	r.With(validateCollectPlatesRequest()).Post("/v1/plates", collectPlateStates)
	return r
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	// 確認用。必要なら、何か参照すべき
	Succeed(w, struct {
		Status string `json:"status"`
	}{string("healthy")})
}

func collectPlateStates(w http.ResponseWriter, r *http.Request) {
	b, _ := r.Context().Value(plateCtxKey).(PlateRequestBody)
	data, err := AddPlateStates(r.Context(), Plate{
		QrID:       b.QrID,
		ShopNumber: *b.ShopNumber,
		Hostname:   b.Hostname,
		PopNumber:  *b.PopNumber,
		State:      *b.State,
	})
	if err != nil {
		Logger.Error("Failed to add plate states:", zap.Error(err))
		Fail(w, http.StatusInternalServerError)
		return
	}
	Created(w, data)
}

func validateCollectPlatesRequest() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var body = PlateRequestBody{}
			err := json.NewDecoder(r.Body).Decode(&body)
			if err != nil {
				Logger.Warn("Request Body is invalid:", zap.Any("body", r.Body), zap.Error(err))
				Fail(w, http.StatusBadRequest)
				return
			}

			errMessages := make([]string, 0, 5)
			if body.QrID == "" {
				errMessages = append(errMessages, "Missing qrId")
			}
			if body.ShopNumber == nil {
				errMessages = append(errMessages, "Missing shopNumber")
			}
			if body.Hostname == "" {
				errMessages = append(errMessages, "Missing hostname")
			}
			if body.PopNumber == nil {
				errMessages = append(errMessages, "Missing popNumber")
			}
			if body.State == nil {
				errMessages = append(errMessages, "Missing state")
			}
			if len(errMessages) != 0 {
				Logger.Warn("Request Body value is invalid:",
					zap.Any("body", r.Body),
					zap.String("errs", strings.Join(errMessages, ",")),
				)
				Fail(w, http.StatusBadRequest)
				return
			}
			Logger.Debug("body:", zap.Any("body", body))

			c := r.Context()
			c = context.WithValue(c, plateCtxKey, body)
			next.ServeHTTP(w, r.WithContext(c))
		})
	}
}

type PlateRequestBody struct {
	QrID       string `json:"qrId"`
	ShopNumber *int64 `json:"shopNumber"`
	Hostname   string `json:"hostname"`
	PopNumber  *int16 `json:"popNumber"`
	State      *int8  `json:"state"`
}
