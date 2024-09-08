package filestore

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/Yasuhiro-gh/url-shortener/internal/config"
	"github.com/Yasuhiro-gh/url-shortener/internal/usecase/storage"
	"os"
)

var IDCounter int

func CreateFileStorage() error {
	_, err := os.OpenFile(config.Options.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	return err
}

type Record struct {
	ID          int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func MakeRecord(shortURL, originalURL string) error {
	r := Record{ID: IDCounter + 1, ShortURL: shortURL, OriginalURL: originalURL}

	rm, err := json.Marshal(r)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(config.Options.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(rm)
	if err != nil {
		return err
	}

	_, err = file.WriteString("\n")
	if err != nil {
		return err
	}

	IDCounter++
	return nil
}

func Restore(us *storage.URLS) error {
	err := CreateFileStorage()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(config.Options.FileStoragePath)
	if err != nil {
		return err
	}

	scn := bufio.NewScanner(bytes.NewReader(data))

	for scn.Scan() {
		line := scn.Text()
		record := Record{}
		err := json.Unmarshal([]byte(line), &record)
		record.ID = IDCounter + 1
		if err != nil {
			return err
		}
		IDCounter++
		us.Set(record.ShortURL, record.OriginalURL)
	}
	return nil
}
