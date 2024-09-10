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

		urlHash := utils.HashURL(urlString)
		if _, exist := h.URLS.Get(urlHash); !exist {
			h.URLS.Set(urlHash, urlString)
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)

		err := filestore.MakeRecord(urlHash, urlString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

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
		if _, exist := h.URLS.Get(urlHash); !exist {
			h.URLS.Set(urlHash, shortenRequest.URL)
		}
		shortenResponse.Result = config.Options.BaseURL + "/" + urlHash

		resp, err := json.Marshal(shortenResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		err = filestore.MakeRecord(urlHash, shortenRequest.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(resp)
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
