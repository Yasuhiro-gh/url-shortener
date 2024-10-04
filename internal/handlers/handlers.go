package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/Yasuhiro-gh/url-shortener/internal/auth"
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
	storage.URLStorages
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
	r.Handle("/api/user/urls", gzipMiddleware(logger.Logging(uh.UserURLS())))
	r.Handle("/ping", logger.Logging(CheckDBConnection(ctx, pdb)))
	return r
}

func GetUserIDFromCookie(r *http.Request) (int, error) {
	uidCookie, err := r.Cookie("userIDToken")
	if err != nil {
		return 0, err
	}
	userID, err := auth.GetUserID(uidCookie.Value)
	if err != nil {
		return 0, err
	}
	return userID, nil
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

func (h *URLHandler) Auth(w http.ResponseWriter, r *http.Request) error {
	cookie, cookieErr := r.Cookie("userIDToken")

	if cookieErr == nil {
		_, err := auth.GetUserID(cookie.Value)
		if err == nil {
			return nil
		}
	}

	newUserID := h.GetUserID()
	token, err := auth.BuildJWTString(newUserID + 1)
	if err != nil {
		return err
	}
	newCookie := http.Cookie{Name: "userIDToken", Value: token, Expires: time.Now().Add(time.Minute * 10), Path: "/"}
	http.SetCookie(w, &newCookie)
	return errors.New("unauthorized")
}

func (h *URLHandler) ShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is supported.", http.StatusBadRequest)
			return
		}

		_ = h.Auth(w, r)

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

		var userID int
		if uid, err := GetUserIDFromCookie(r); err == nil {
			userID = uid
		} else {
			userID = h.GetUserID() + 1
		}

		urlHash := utils.HashURL(urlString)
		urlStore := &storage.Store{OriginalURL: urlString, ShortURL: config.Options.BaseURL + "/" + urlHash, UserID: userID}

		repeatErr := h.Set(urlHash, urlStore)
		if repeatErr != nil && repeatErr.Error() == pgerrcode.UniqueViolation {
			httpStatus = http.StatusConflict
		} else {
			repeatErr = filestore.MakeRecord(urlStore)
			if repeatErr != nil {
				http.Error(w, repeatErr.Error(), http.StatusBadRequest)
			}
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(httpStatus)
		_, _ = w.Write([]byte(urlStore.ShortURL))
	}
}

func (h *URLHandler) GetShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET method is supported.", http.StatusBadRequest)
			return
		}

		_ = h.Auth(w, r)

		shortURL := r.PathValue("id")

		if shortURL == "" {
			http.Error(w, "Please provide a URL.", http.StatusBadRequest)
			return
		}

		urlStore, exist := h.Get(shortURL)
		if !exist {
			http.Error(w, "Invalid URL.", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Location", urlStore.OriginalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (h *URLHandler) UserURLS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET method is supported.", http.StatusBadRequest)
		}

		err := h.Auth(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		uid, _ := GetUserIDFromCookie(r)

		urlStores, err := h.GetUserURLS(r.Context(), uid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if len(urlStores) == 0 {
			http.Error(w, "There is no your urls", http.StatusNoContent)
			return
		}

		resp, err := json.Marshal(urlStores)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(resp)
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

		_ = h.Auth(w, r)

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

		var userID int
		if uid, err := GetUserIDFromCookie(r); err == nil {
			userID = uid
		} else {
			userID = h.GetUserID() + 1
		}

		urlHash := utils.HashURL(shortenRequest.URL)
		urlStore := &storage.Store{OriginalURL: shortenRequest.URL, ShortURL: config.Options.BaseURL + "/" + urlHash, UserID: userID}
		repeatErr := h.Set(urlHash, urlStore)
		var httpStatus = http.StatusCreated

		shortenResponse.Result = config.Options.BaseURL + "/" + urlHash

		resp, err := json.Marshal(shortenResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		if repeatErr != nil && repeatErr.Error() == pgerrcode.UniqueViolation {
			httpStatus = http.StatusConflict
		} else {
			err = filestore.MakeRecord(urlStore)
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

		_ = h.Auth(w, r)

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

			var userID int
			if uid, err := GetUserIDFromCookie(r); err == nil {
				userID = uid
			} else {
				userID = h.GetUserID() + 1
			}

			urlHash := utils.HashURL(val.OriginalURL)
			urlStore := &storage.Store{OriginalURL: val.OriginalURL, ShortURL: config.Options.BaseURL + "/" + urlHash, UserID: userID}
			repeatErr := h.Set(urlHash, urlStore)

			if repeatErr != nil && repeatErr.Error() == pgerrcode.UniqueViolation {
				httpStatus = http.StatusConflict
			} else {
				err = filestore.MakeRecord(urlStore)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}
			}

			response := ShortenResponse{val.CorrelationID, urlStore.ShortURL}

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
