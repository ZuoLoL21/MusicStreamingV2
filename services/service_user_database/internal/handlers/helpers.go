package handlers

import (
	"backend/internal/consts"
	"backend/internal/di"
	"libs/helpers"

	sqlhandler "backend/sql/sqlc"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	libsdi "libs/di"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgtype"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

func decodeBody[T any](w http.ResponseWriter, r *http.Request, returns *libsdi.ReturnManager) (T, bool) {
	var body T
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		returns.ReturnError(w, "invalid request body", http.StatusBadRequest)
		return body, false
	}
	if err := validateBody(&body); err != nil {
		returns.ReturnError(w, err.Error(), http.StatusBadRequest)
		return body, false
	}
	return body, true
}

func validateBody(body any) error {
	if err := validate.Struct(body); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return fmt.Errorf("%s: %s", ve[0].Field(), ve[0].Tag())
		}
		return err
	}
	return nil
}

func parseUUID(r *http.Request, key string) (pgtype.UUID, bool) {
	vars := mux.Vars(r)
	uuidStr, ok := vars[key]
	if !ok {
		return pgtype.UUID{}, false
	}
	var id pgtype.UUID
	if err := id.Scan(uuidStr); err != nil {
		return pgtype.UUID{}, false
	}
	return id, true
}

func uuidToPgtype(uuidStr string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(uuidStr); err != nil {
		return pgtype.UUID{}, err
	}
	return id, nil
}

// TODO remove config from this function everywhere
func userUUIDFromCtx(w http.ResponseWriter, r *http.Request, config *di.Config, returns *libsdi.ReturnManager) (pgtype.UUID, bool) {
	uuidStr := helpers.GetUserUUIDFromContext(r.Context())
	if uuidStr == "" {
		returns.ReturnError(w, "unauthorized", http.StatusUnauthorized)
		return pgtype.UUID{}, false
	}
	userUUID, err := uuidToPgtype(uuidStr)
	if err != nil {
		returns.ReturnError(w, "invalid user uuid", http.StatusUnauthorized)
		return pgtype.UUID{}, false
	}
	return userUUID, true
}

// parsePagination reads limit (default 20), cursor_ts, and cursor_id from the
// request's URL query parameters.
func parsePagination(r *http.Request) (limit int32, cursorTS pgtype.Timestamptz, cursorID pgtype.UUID) {
	limit = 20
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			limit = int32(n)
		}
	}
	if s := r.URL.Query().Get("cursor_ts"); s != "" {
		if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
			_ = cursorTS.Scan(t)
		}
	}
	if s := r.URL.Query().Get("cursor_id"); s != "" {
		_ = cursorID.Scan(s)
	}
	return
}

// parsePaginationName reads limit (default 20) and cursor_name from the
// request's URL query parameters.
func parsePaginationName(r *http.Request) (limit int32, cursorName string) {
	limit = 20
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			limit = int32(n)
		}
	}
	cursorName = r.URL.Query().Get("cursor_name")
	return
}

// parsePaginationAlpha reads limit (default 20), cursor_name, and cursor_ts
// from the request's URL query parameters (used for alphabetically sorted
// queries where the cursor is (name, created_at)).
func parsePaginationAlpha(r *http.Request) (limit int32, cursorName string, cursorTS pgtype.Timestamptz) {
	limit, cursorName = parsePaginationName(r)
	if s := r.URL.Query().Get("cursor_ts"); s != "" {
		if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
			_ = cursorTS.Scan(t)
		}
	}
	return
}

// parsePaginationPos reads limit (default 20) and cursor_pos from the
// request's URL query parameters (used for position-ordered queries).
func parsePaginationPos(r *http.Request) (limit int32, cursorPos int32) {
	limit = 20
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			limit = int32(n)
		}
	}
	if s := r.URL.Query().Get("cursor_pos"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			cursorPos = int32(n)
		}
	}
	return
}

// parsePaginationSearch parses search pagination query params (similarity score + timestamp cursor)
func parsePaginationSearch(r *http.Request) (limit int32, cursorScore pgtype.Float8, cursorTS pgtype.Timestamptz) {
	limit = 20
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 100 {
			limit = int32(n)
		}
	}

	cursorScoreStr := r.URL.Query().Get("cursor_score")
	if cursorScoreStr != "" {
		if score, err := strconv.ParseFloat(cursorScoreStr, 64); err == nil && score >= 0 && score <= 1 {
			cursorScore = pgtype.Float8{Float64: score, Valid: true}
		}
	}

	cursorTSStr := r.URL.Query().Get("cursor_ts")
	if cursorTSStr != "" {
		if t, err := time.Parse(time.RFC3339, cursorTSStr); err == nil {
			cursorTS = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}

	return
}

var roleWeight = map[sqlhandler.ArtistMemberRole]int{
	sqlhandler.ArtistMemberRoleMember:  1,
	sqlhandler.ArtistMemberRoleManager: 2,
	sqlhandler.ArtistMemberRoleOwner:   3,
}

func checkArtistRole(ctx context.Context, q consts.DB, artistUUID pgtype.UUID, userUUID pgtype.UUID, minRole sqlhandler.ArtistMemberRole) bool {
	members, err := q.GetUsersRepresentingArtist(ctx, artistUUID)
	if err != nil {
		return false
	}
	for _, m := range members {
		if m.Uuid.Bytes == userUUID.Bytes {
			return roleWeight[m.Role] >= roleWeight[minRole]
		}
	}
	return false
}

// validateStringField validates that a string is not empty or whitespace-only
func validateStringField(field string, fieldName string, minLen, maxLen int) (string, error) {
	trimmed := strings.TrimSpace(field)
	if trimmed == "" {
		return "", fmt.Errorf("%s cannot be empty or whitespace-only", fieldName)
	}
	if strings.Contains(trimmed, "\x00") {
		return "", fmt.Errorf("%s contains invalid null byte characters", fieldName)
	}
	if len(trimmed) < minLen {
		return "", fmt.Errorf("%s must be at least %d characters", fieldName, minLen)
	}
	if maxLen > 0 && len(trimmed) > maxLen {
		return "", fmt.Errorf("%s must not exceed %d characters", fieldName, maxLen)
	}
	return trimmed, nil
}

// isValidEmail checks if an email address is valid (basic validation)
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	atIndex := strings.Index(email, "@")
	if atIndex <= 0 || atIndex >= len(email)-1 {
		return false
	}

	domainPart := email[atIndex+1:]
	if !strings.Contains(domainPart, ".") || strings.HasPrefix(domainPart, ".") || strings.HasSuffix(domainPart, ".") {
		return false
	}
	return true
}
