package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/Yasuhiro-gh/url-shortener/internal/config"
	"github.com/Yasuhiro-gh/url-shortener/internal/db"
	"github.com/Yasuhiro-gh/url-shortener/internal/logger"
	"github.com/Yasuhiro-gh/url-shortener/internal/usecase/compress"
	"github.com/Yasuhiro-gh/url-shortener/internal/usecase/storage"
	"github.com/Yasuhiro-gh/url-shortener/internal/usecase/storage/filestore"
	"github.com/Yasuhiro-gh/url-shortener/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"
	"io"
	"net/http"
	"strings"
	"time"
)

type URLHandler struct {
	*storage.URLS
}

func NewURLHandler(us *storage.URLS) *URLHandler {
	return &URLHandler{us}
}

func URLRouter(ctx context.Context, us *storage.URLS, pdb *db.PostgresDB) chi.Router {
	r := chi.NewRouter()

	uh := NewURLHandler(us)

	r.Handle("/", gzipMiddleware(logger.Logging(uh.ShortURL())))
	r.Handle("/{id}", gzipMiddleware(logger.Logging(uh.GetShortURL())))
	r.Handle("/api/shorten", gzipMiddleware(logger.Logging(uh.ShortURLJSON())))
	r.Handle("/api/shorten/batch", gzipMiddleware(logger.Logging(uh.ShortURLBatch())))
	r.Handle("/ping", logger.Logging(CheckDBConnection(ctx, pdb)))
	return r
}

func gzipMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			cw := compress.NewGzipWriter(w)
			ow = cw
			defer cw.Close()
		}

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := compress.NewGzipReader(r.Body)
			if err != nil {
				ow.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		next.ServeHTTP(ow, r)
	}
}

func (h *URLHandler) ShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is supported.", http.StatusBadRequest)
			return
		}

		body, _ := io.ReadAll(r.Body)

		urlString := string(body)
		if urlString == "" {
			http.Error(w, "Please provide a URL.", http.StatusBadRequest)
			return
		}

		if !utils.IsValidURL(urlString) {
			http.Error(w, "Invalid URL.", http.StatusBadRequest)
			return
		}

		var httpStatus = http.StatusCreated

		urlHash := utils.HashURL(urlString)
		err := h.URLS.Set(urlHash, urlString)
		if err != nil && err.Error() == pgerrcode.UniqueViolation {
			httpStatus = http.StatusConflict
		} else {
			err = filestore.MakeRecord(urlHash, urlString)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(httpStatus)
		_, _ = w.Write([]byte(config.Options.BaseURL + "/" + urlHash))
	}
}

func (h *URLHandler) GetShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET method is supported.", http.StatusBadRequest)
			return
		}

		shortURL := r.PathValue("id")

		if shortURL == "" {
			http.Error(w, "Please provide a URL.", http.StatusBadRequest)
			return
		}

		url, exist := h.URLS.Get(shortURL)
		if !exist {
			http.Error(w, "Invalid URL.", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (h *URLHandler) ShortURLJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is supported.", http.StatusBadRequest)
			return
		}

		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "Only JSON content type is supported.", http.StatusBadRequest)
			return
		}

		var buf bytes.Buffer

		type ShortenJSON struct {
			URL string `json:"url"`
		}

		var shortenRequest ShortenJSON

		type ResponseJSON struct {
			Result string `json:"result"`
		}

		var shortenResponse ResponseJSON

		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(buf.Bytes(), &shortenRequest); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if shortenRequest.URL == "" {
			http.Error(w, "Please provide a URL.", http.StatusBadRequest)
			return
		}

		if !utils.IsValidURL(shortenRequest.URL) {
			http.Error(w, "Invalid URL.", http.StatusBadRequest)
			return
		}

		urlHash := utils.HashURL(shortenRequest.URL)
		repeatErr := h.URLS.Set(urlHash, shortenRequest.URL)
		var httpStatus = http.StatusCreated

		shortenResponse.Result = config.Options.BaseURL + "/" + urlHash

		resp, err := json.Marshal(shortenResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		if err != nil && repeatErr.Error() == pgerrcode.UniqueViolation {
			httpStatus = http.StatusConflict
		} else {
			err = filestore.MakeRecord(urlHash, shortenRequest.URL)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		_, _ = w.Write(resp)
	}
}

func (h *URLHandler) ShortURLBatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is supported.", http.StatusBadRequest)
		}

		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "Only JSON content type is supported.", http.StatusBadRequest)
			return
		}

		var buf bytes.Buffer

		type ShortenJSON struct {
			CorrelationID string `json:"correlation_id"`
			OriginalURL   string `json:"original_url"`
		}

		var shortenRequest []ShortenJSON

		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(buf.Bytes(), &shortenRequest); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		type ShortenResponse struct {
			CorrelationID string `json:"correlation_id"`
			ShortURL      string `json:"short_url"`
		}

		var shortenResponse []ShortenResponse

		var httpStatus = http.StatusCreated

		for _, val := range shortenRequest {
			if val.OriginalURL == "" {
				http.Error(w, "Please provide a URL.", http.StatusBadRequest)
				return
			}

			if !utils.IsValidURL(val.OriginalURL) {
				http.Error(w, "Invalid URL.", http.StatusBadRequest)
				return
			}

			urlHash := utils.HashURL(val.OriginalURL)
			repeatErr := h.URLS.Set(urlHash, val.OriginalURL)

			if err != nil && repeatErr.Error() == pgerrcode.UniqueViolation {
				httpStatus = http.StatusConflict
			} else {
				err = filestore.MakeRecord(urlHash, val.OriginalURL)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}
			}

			response := ShortenResponse{val.CorrelationID, config.Options.BaseURL + "/" + urlHash}

			shortenResponse = append(shortenResponse, response)
		}

		marshaledResponse, err := json.Marshal(shortenResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		_, _ = w.Write(marshaledResponse)
	}
}

func CheckDBConnection(ctx context.Context, pdb *db.PostgresDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET method is supported.", http.StatusBadRequest)
		}

		ctxTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		if err := pdb.DB.PingContext(ctxTimeout); err != nil {
			http.Error(w, "Cannot connect to database", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Database connected"))
	}
}
