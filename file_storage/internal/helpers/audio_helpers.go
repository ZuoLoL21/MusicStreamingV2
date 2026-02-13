package helpers

import (
	"bytes"
	"file-storage/internal/general"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gorilla/mux"
)

type AudioResult struct {
	ID   string
	Data io.Reader
}

const MaxAudioSize = 10 << 20 // 10MB

func ParseAudioFromRequest(r *http.Request) (*AudioResult, *general.ErrorResult) {
	vars := mux.Vars(r)
	validated := general.ValidateUUID(vars["id"])
	if !validated {
		return nil, &general.ErrorResult{Message: "Invalid id provided", Status: http.StatusBadRequest}
	}

	id := vars["id"]

	// Get audio file
	reader, err := r.MultipartReader()
	if err != nil {
		return nil, &general.ErrorResult{Message: "failed to read multipart data", Status: http.StatusBadRequest}
	}
	part, err := reader.NextPart()
	if err != nil {
		return nil, &general.ErrorResult{Message: "failed to get file part", Status: http.StatusBadRequest}
	}

	if part.FormName() != "audio" {
		_ = part.Close()
		return nil, &general.ErrorResult{Message: "expected 'audio' field", Status: http.StatusBadRequest}
	}

	// Validate filename exists
	filename := part.FileName()
	if filename == "" {
		return nil, &general.ErrorResult{Message: "missing filename", Status: http.StatusBadRequest}
	}

	// Validate MP3 format by peeking at the header
	combined, err2 := testIfMP3(part)
	if err2 != nil {
		_ = part.Close()
		return nil, err2
	}

	// Limit to MaxAudioSize+1 so the handler can detect overflow after writing
	limitedReader := io.LimitReader(combined, int64(MaxAudioSize)+1)

	return &AudioResult{ID: id, Data: limitedReader}, nil
}

func testIfMP3(filePart *multipart.Part) (io.Reader, *general.ErrorResult) {
	header := make([]byte, 3)
	n, err := io.ReadFull(filePart, header)
	if err != nil || n != 3 {
		return nil, &general.ErrorResult{Message: "Failed to read file header", Status: http.StatusBadRequest}
	}
	isID3 := header[0] == 0x49 && header[1] == 0x44 && header[2] == 0x33
	isMPEG := header[0] == 0xFF && (header[1]&0xE0) == 0xE0
	if !isID3 && !isMPEG {
		return nil, &general.ErrorResult{Message: "Invalid MP3 file format", Status: http.StatusBadRequest}
	}
	return io.MultiReader(bytes.NewReader(header), filePart), nil
}
