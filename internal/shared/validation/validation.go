package validation

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var Validator = validator.New()

// FieldErrors turns validator errors into the JSON field names consumed by the
// client, so users can correct the exact input rather than decoding a generic 422.
func FieldErrors(err error, input any) map[string][]string {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return map[string][]string{
			"request": {"Format data tidak dapat diproses. Periksa kembali isian yang dikirim."},
		}
	}

	typeOfInput := reflect.TypeOf(input)
	for typeOfInput.Kind() == reflect.Pointer {
		typeOfInput = typeOfInput.Elem()
	}

	errors := make(map[string][]string, len(validationErrors))
	for _, fieldError := range validationErrors {
		field := jsonFieldName(typeOfInput, fieldError.StructField())
		errors[field] = append(errors[field], validationMessage(fieldError))
	}
	return errors
}

func jsonFieldName(typeOfInput reflect.Type, structField string) string {
	field, found := typeOfInput.FieldByName(structField)
	if !found {
		return strings.ToLower(structField)
	}
	name := strings.Split(field.Tag.Get("json"), ",")[0]
	if name == "" || name == "-" {
		return strings.ToLower(structField)
	}
	return name
}

func validationMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "Wajib diisi agar data dapat disimpan."
	case "uuid":
		return "Pilih data yang valid dari daftar yang tersedia."
	case "email":
		return "Masukkan alamat email dengan format yang benar."
	case "oneof":
		return "Pilih salah satu opsi yang tersedia."
	case "min":
		return "Nilai yang diisi masih terlalu pendek atau terlalu kecil."
	case "max":
		return "Nilai yang diisi melebihi batas yang diperbolehkan."
	default:
		return "Nilai pada bidang ini belum sesuai ketentuan."
	}
}
