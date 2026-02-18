package handlers

import (
	"backend/internal/di"
	sqlhandler "backend/sql/sqlc"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

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

func decodeBody[T any](w http.ResponseWriter, r *http.Request, returns *di.ReturnManager) (T, bool) {
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

func userUUIDFromCtx(w http.ResponseWriter, r *http.Request, config *di.Config, returns *di.ReturnManager) (pgtype.UUID, bool) {
	uuidStr, _ := r.Context().Value(config.UserUUIDKey).(string)
	userUUID, err := uuidToPgtype(uuidStr)
	if err != nil {
		returns.ReturnError(w, "invalid user uuid", http.StatusInternalServerError)
		return pgtype.UUID{}, false
	}
	return userUUID, true
}

var roleWeight = map[sqlhandler.ArtistMemberRole]int{
	sqlhandler.ArtistMemberRoleMember:  1,
	sqlhandler.ArtistMemberRoleManager: 2,
	sqlhandler.ArtistMemberRoleOwner:   3,
}

func checkArtistRole(ctx context.Context, q *sqlhandler.Queries, artistUUID pgtype.UUID, userUUID pgtype.UUID, minRole sqlhandler.ArtistMemberRole) bool {
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
