package helpers

import (
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gorilla/mux"
)

type AudioResult struct {
	ID   string
	Data *multipart.Part
}

func ParseAudioFromRequest(r *http.Request) (*AudioResult, *ErrorResult) {
	vars := mux.Vars(r)
	validated := ValidateUUID(vars["id"])
	if !validated {
		return nil, &ErrorResult{Message: "Invalid id provided", Status: http.StatusBadRequest}
	}

	id := vars["id"]

	// Get audio file
	reader, err := r.MultipartReader()
	if err != nil {
		return nil, &ErrorResult{Message: "failed to read multipart data", Status: http.StatusBadRequest}
	}
	part, err := reader.NextPart()
	if err != nil {
		return nil, &ErrorResult{Message: "failed to get file part", Status: http.StatusBadRequest}
	}

	if part.FormName() != "audio" {
		_ = part.Close()
		return nil, &ErrorResult{Message: "expected 'audio' field", Status: http.StatusBadRequest}
	}
	err2 := testIfMP3(part)
	if err2 != nil {
		_ = part.Close()
		return nil, err2
	}

	return &AudioResult{ID: id, Data: part}, nil
}

func testIfMP3(filePart *multipart.Part) *ErrorResult {
	header := make([]byte, 3)
	n, err := io.ReadFull(filePart, header)
	if err != nil || n != 3 {
		return &ErrorResult{Message: "Failed to read file header", Status: http.StatusBadRequest}
	}
	isID3 := header[0] == 0x49 && header[1] == 0x44 && header[2] == 0x33
	isMPEG := header[0] == 0xFF && (header[1]&0xE0) == 0xE0
	if !isID3 && !isMPEG {
		return &ErrorResult{Message: "Invalid MP3 file format", Status: http.StatusBadRequest}
	}
	return nil
}
