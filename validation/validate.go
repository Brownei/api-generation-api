package validation

import "github.com/go-playground/validator/v10"

type ValidationErrorResponse struct {
	Error   string                  `json:"error"`
	Details []ValidationErrorDetail `json:"details"`
}

type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Value must be at least " + err.Param()
	case "max":
		return "Value must be at most " + err.Param()
	case "email":
		return "Invalid email format"
	default:
		return "Invalid value"
	}
}

func ValidateStruct(s interface{}) []ValidationErrorDetail {
	validate := validator.New()
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var details []ValidationErrorDetail
	for _, err := range err.(validator.ValidationErrors) {
		details = append(details, ValidationErrorDetail{
			Field:   err.Field(),
			Message: getErrorMessage(err),
		})
	}
	return details
}
