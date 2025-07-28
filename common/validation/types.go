package validation

import (
    "fmt"
    "regexp"

    "bitbucket.org/ouadreams/common/event"
)

// TODO: move to any errors, not only validation?

const (
    notImplementedCode = "not_implemented"
    unknownField       = "unknown"
)

// Result is a data structure expected by FE for each 4xx responses
// structure of backend errors
// https://oua.atlassian.net/wiki/spaces/ORIENTCODE/pages/1535770629/Unified+server+error+reporting
type Result struct {
    Details string                 `json:"details"`
    Code    string                 `json:"code"`
    Errors  []*Error               `json:"errors"`
    Meta    map[string]interface{} `json:"meta,omitempty"`

    // used by limit ms to receive events
    Events event.Payload `json:"events,omitempty"`
}

type Error struct {
    Name  string                 `json:"name"`
    Index int                    `json:"-"` // used for request of array types to determine the correct field
    Codes []ErrorDetails         `json:"codes"`
    Data  map[string]interface{} `json:"data,omitempty"`
}

type ErrorDetails struct {
    Message string `json:"message"`
    Code    string `json:"code"`
}

func NewResult() *Result {
    return &Result{
        Details: "",
        Errors:  make([]*Error, 0),
    }
}

func UnmarshalError(err error) *Result {
    return &Result{
        Details: err.Error(),
        Code:    notImplementedCode,
    }
}

var (
    rxFieldName    = regexp.MustCompile("field=(.+?),")
    rxErrorDetails = regexp.MustCompile(`internal=(.+)`)
)

func UnmarshalDetailedError(err error) *Result {
    return unmarshalError(notImplementedCode, err)
}

func unmarshalError(code string, err error) *Result {
    out := NewResult()

    fieldNameGroup := rxFieldName.FindStringSubmatch(err.Error())
    errorDetailsGroup := rxErrorDetails.FindStringSubmatch(err.Error())

    fieldName := unknownField // TODO: ticket was created to not allow this, need some fancy parsing to always have the
    if len(fieldNameGroup) > 1 {
        fieldName = fieldNameGroup[1]
    }

    errorDetails := err.Error()
    if len(errorDetailsGroup) > 1 {
        errorDetails = errorDetailsGroup[1]
    }

    out.AddFieldError(fieldName, ErrorDetails{
        Message: errorDetails,
        Code:    code,
    })

    return out
}

func DBOperationError(err error) *Result {
    return &Result{
        Details: err.Error(),
        Code:    notImplementedCode,
    }
}

func BothEmailAndPhoneProvided() *Result {
    return &Result{
        Details: "can't provide both email and phone during registration",
        Code:    "both_email_phone_provided",
    }
}

func CaptchaError(err error) *Result {
    return &Result{
        Details: err.Error(),
        Code:    "captcha_error",
    }
}

// NoCodeError is a generic error, on request from FE will refactor responses
// that uses this if they need a code
func NoCodeError(err error) *Result {
    return &Result{
        Details: err.Error(),
    }
}

// CodeError is a generic error with code value
func CodeError(code string, err error) *Result {
    return &Result{
        Code:    code,
        Details: err.Error(),
    }
}

// CodeErrorWithEvents is a generic error with code value events
func CodeErrorWithEvents(code string, err error, events event.Payload) *Result {
    return &Result{
        Code:    code,
        Details: err.Error(),
        Events:  events,
    }
}

func (r *Result) IsValid() bool {
    return len(r.Errors) == 0 && r.Details == ""
}

func (r *Result) AddResult(result *Result) {
    r.Errors = append(r.Errors, result.Errors...)
}

// AddMetaInfo will add key:value to validation result Meta field
func (r *Result) AddMetaInfo(key string, value interface{}) *Result {
    if r.Meta == nil {
        r.Meta = make(map[string]interface{})
    }
    r.Meta[key] = value
    return r
}

func (r *Result) AddFieldError(field string, ed ErrorDetails) *Result {
    for _, e := range r.Errors {
        if e.Name == field {
            e.Codes = append(e.Codes, ed)
            return r
        }
    }
    r.Errors = append(r.Errors, &Error{
        Name:  field,
        Codes: []ErrorDetails{ed},
    })
    return r
}

// AddFieldErrorWithData used for handling errors regarding array request
// index used to determine which request object is having error , eg, request
//  {
//      "addresses":
//      [
//          {
//              "asset":"BTC",
//              "network:"Tron"
//          },
//          {
//              "asset":"ETH",
//              "network":"Erc20"
//          }
//      ]
//  }
// usage :
//  out := new(validation.Result)
//  for i := range newAddresses {
//      if newAddresses[i].Asset==""{
//          out.AddFieldErrorWithData(messages.FieldAddress, errorx.ErrorDetails(errorx.ErrEmptyAsset), map[string]interface{}{
//              messages.FieldAsset: newAddresses[i].Asset,
//              messages.FieldNetwork: newAddresses[i].Network,
//          }, i)
//      }
//  }
func (r *Result) AddFieldErrorWithData(field string, ed ErrorDetails, data map[string]interface{}, index int) *Result {
    for _, e := range r.Errors {
        if e.Name == field && e.Index == index {
            e.Codes = append(e.Codes, ed)
            return r
        }
    }

    r.Errors = append(r.Errors, &Error{
        Name:  field,
        Codes: []ErrorDetails{ed},
        Data:  data,
        Index: index,
    })
    return r
}

func (r *Result) AddDetails(details string, formatArgs ...interface{}) *Result {
    r.Details = fmt.Sprintf(details, formatArgs...)
    return r
}

func (r *Result) AddCode(code string) *Result {
    r.Code = code
    return r
}

func EitherPhoneOrEmail() ErrorDetails {
    return ErrorDetails{
        Message: "either phone or e-mail must be provided",
        Code:    "email_or_phone_must_be_provided",
    }
}

func EmptyBirthDate() ErrorDetails {
    return ErrorDetails{
        Message: "empty birthday",
        Code:    "empty_birthday",
    }
}

func NotOnlyLetters() ErrorDetails {
    return ErrorDetails{
        Message: "field must contain only letters",
        Code:    "only_letters_allowed",
    }
}

func UserAlreadyExists() ErrorDetails {
    return ErrorDetails{
        Message: "user already exists",
        Code:    "user_already_exists",
    }
}

func Unauthorized() ErrorDetails {
    return ErrorDetails{
        Message: "unauthorized",
        Code:    "unauthorized",
    }
}

func EmptyPassword() ErrorDetails {
    return ErrorDetails{
        Message: "empty password provided",
        Code:    "empty_password",
    }
}

func InvalidPassword() ErrorDetails {
    return ErrorDetails{
        Message: "password must contain at least 1 lowercased letter, 1 capital letter, 1 digit, 1 special char and be minimum 8 chars long",
        Code:    "wrong_password_format",
    }
}

func RulesNotAccepted() ErrorDetails {
    return ErrorDetails{
        Message: "rules were not accepted",
        Code:    "rules_not_accepted",
    }
}

func NameIsTooShort() ErrorDetails {
    return ErrorDetails{
        Message: "name cannot be less than 2 characters",
        Code:    "name_too_short",
    }
}

func TooYoungAge() ErrorDetails {
    return ErrorDetails{
        Message: "age must be more than 18 years",
        Code:    "age_must_be_over_18",
    }
}

func WrongPhoneFormat() ErrorDetails {
    return ErrorDetails{
        Message: "wrong phone format: it should contain only digits",
        Code:    "phone_wrong_format",
    }
}

func InvalidEmail() ErrorDetails {
    return ErrorDetails{
        Message: "email is empty or has invalid format",
        Code:    "invalid_email",
    }
}

func UnknownCountry() ErrorDetails {
    return ErrorDetails{
        Message: "such country does not exist",
        Code:    "unknown_country",
    }
}

func InvalidPhone() ErrorDetails {
    return ErrorDetails{
        Message: "phone is empty",
        Code:    "empty_phone",
    }
}

func InvalidIPAddress(ip string) ErrorDetails {
    return ErrorDetails{
        Message: fmt.Sprintf("ip_address is empty or has invalid format: %s", ip),
        Code:    "invalid_ip_address",
    }
}

func EmptyDevice() ErrorDetails {
    return ErrorDetails{
        Message: "empty device provided",
        Code:    "empty_device",
    }
}

func EmptyRefreshToken() ErrorDetails {
    return ErrorDetails{
        Message: "empty refresh token provided",
        Code:    "empty_refresh_token",
    }
}

func InvalidAntiPhishingCode() ErrorDetails {
    return ErrorDetails{
        Message: "anti-phishing code is invalid",
        Code:    "invalid_anti_phishing_code",
    }
}

func InvalidCountryCallingCodeFormat() ErrorDetails {
    return ErrorDetails{
        Message: "country calling code format is invalid",
        Code:    "wrong_country_code_format",
    }
}

func WrongCountryCallingCode() ErrorDetails {
    return ErrorDetails{
        Message: "country calling code does not match selected country",
        Code:    "mismatch_county_code",
    }
}

func InvalidPageLimit() ErrorDetails {
    return ErrorDetails{
        Message: "page limit is invalid",
        Code:    "limit_is_invalid",
    }
}

func InvalidOrderColumn() ErrorDetails {
    return ErrorDetails{
        Message: "order column is invalid",
        Code:    "ordering_is_invalid",
    }
}

func InvalidOtpCode() ErrorDetails {
    return ErrorDetails{
        Message: "invalid otp code provided",
        Code:    "invalid_otp_code",
    }
}

func InvalidAction() ErrorDetails {
    return ErrorDetails{
        Message: "action is invalid",
        Code:    "invalid_action",
    }
}

func Invalid2FAMethod() ErrorDetails {
    return ErrorDetails{
        Message: "method is invalid",
        Code:    "invalid_method",
    }
}

func InvalidCode() ErrorDetails {
    return ErrorDetails{
        Message: "code is invalid",
        Code:    "invalid_code",
    }
}

func InvalidKey() ErrorDetails {
    return ErrorDetails{
        Message: "key is empty",
        Code:    "empty_key",
    }
}

func InvalidResetToken() ErrorDetails {
    return ErrorDetails{
        Message: "reset-token is invalid",
        Code:    "invalid_reset_token",
    }
}

func InvalidConfirmPassword() ErrorDetails {
    return ErrorDetails{
        Message: "confirm password is invalid",
        Code:    "invalid_confirm_password",
    }
}

func InvalidKeys() ErrorDetails {
    return ErrorDetails{
        Message: "phone or email method is missing",
        Code:    "phone_email_method_missing",
    }
}