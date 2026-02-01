package helpers

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gorilla/mux"
)

func ParseAudioFromRequest(r *http.Request) (string, *multipart.Part, error, int) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get audio file
	reader, err := r.MultipartReader()
	if err != nil {
		return "", nil, errors.New("failed to read multipart data"), http.StatusBadRequest
	}
	part, err := reader.NextPart()
	if err != nil {
		return "", nil, errors.New("failed to get file part"), http.StatusBadRequest
	}

	if part.FormName() != "audio" {
		_ = part.Close()
		return "", nil, errors.New("expected 'audio' field"), http.StatusBadRequest
	}
	err = testIfMP3(part)
	if err != nil {
		_ = part.Close()
		return "", nil, err, http.StatusBadRequest
	}

	return id, part, nil, http.StatusOK
}

func testIfMP3(filePart *multipart.Part) error {
	header := make([]byte, 3)
	n, err := io.ReadFull(filePart, header)
	if err != nil || n != 3 {
		return errors.New("Failed to read file header")
	}
	isID3 := header[0] == 0x49 && header[1] == 0x44 && header[2] == 0x33
	isMPEG := header[0] == 0xFF && (header[1]&0xE0) == 0xE0
	if !isID3 && !isMPEG {
		return errors.New("Invalid MP3 file format")
	}
	return nil
}
