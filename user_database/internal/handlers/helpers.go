package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	sql_handler "backend/sql/sqlc"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
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
