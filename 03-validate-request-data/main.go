package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

const plateCtxKey = "plateContext"

var logger *zap.Logger

func init() {
	var err error
	if os.Getenv("ENV") == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	r := chi.NewRouter()

	// Logging the start and end of each request
	r.Use(middleware.Logger)

	// Path, Method 単位で定義
	r.With(validateCollectPlatesRequest()).Post("/plates", collectPlateStates)

	// Listen port
	http.ListenAndServe(":8080", r)
}

func collectPlateStates(w http.ResponseWriter, r *http.Request) {
	b, _ := r.Context().Value(plateCtxKey).(PlateRequestBody)
	j, err := json.Marshal(b)
	if err != nil {
		log.Fatal(err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8;")
	w.WriteHeader(http.StatusCreated)
	w.Write(j)
}

func validateCollectPlatesRequest() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var body = PlateRequestBody{}
			err := json.NewDecoder(r.Body).Decode(&body)
			if err != nil {
				logger.Warn("Request Body is invalid:", zap.Any("body", r.Body), zap.Error(err))
				w.Header().Set("Content-Type", "application/json; charset=utf-8;")
				w.WriteHeader(http.StatusBadRequest)
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
				logger.Warn("Request Body value is invalid:",
					zap.Any("body", r.Body),
					zap.String("errs", strings.Join(errMessages, ",")),
				)
				w.Header().Set("Content-Type", "application/json; charset=utf-8;")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			logger.Debug("body:", zap.Any("body", body))

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
