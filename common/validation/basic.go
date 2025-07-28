package validation

import (
    "strconv"
    "strings"
    "time"
    "unicode"

    "github.com/biter777/countries"
)

// IsAlpha returns true if s contains only letters
func IsAlpha(s string) bool {
    return alphaRegex.MatchString(s)
}

func IsPasswordValid(s string) bool {
    var number, upper, lower, special bool
    var letters int 

    for _, c := range s { 
        letters++
        switch {
        case unicode.IsLower(c):
            lower = true
        case unicode.IsNumber(c):
            number = true
        case unicode.IsUpper(c):
            upper = true
        case unicode.IsPunct(c) || unicode.IsSymbol(c):
            special = true
        default:
            return false
        }   
    }   

    return number && upper && lower && special && letters >= 8 && letters <= 255 
}

func IsEmailValid(email string) bool {
    // A maximum 64 characters in the "local part" (before the "@") and
    // a maximum of 255 characters (octets) in the domain part (after
    // the "@") for a total length of 320 characters.
    if len(email) < 3 || len(email) > 319 {
        return false
    }

    // Check email format
    if !emailRegex.MatchString(email) {
        return false
    }

    // TODO should we check lookupMX?

    return true
}

func IsIpv4Valid(ip string) bool {
    ip = strings.TrimSpace(ip)

    return ipv4Regex.MatchString(ip)
}

func IsIpv6Valid(ip string) bool {
    ip = strings.TrimSpace(ip)

    return ipv6Regex.MatchString(ip)
}

func IsAntiPhishingCodeValid(in string) bool {
    // Anti-phishing code must have minimum length of 4
    // and maximum length of 20.
    if len(in) < 4 || len(in) > 20 {
        return false
    }

    // Anti-phishing code must exclude special characters.
    // special character is a character that is not an
    // alphabetic or numeric character.
    for _, char := range in {
        if char >= 48 && char <= 57 {
            continue
        }
        if char >= 65 && char <= 90 {
            continue
        }
        if char >= 97 && char <= 122 {
            continue
        }

        // Oops! There's a special character in code
        return false
    }

    return true
}

// IsNameValid allows names with no less than 2 characters
func IsNameValid(name string) bool {
    return len(name) >= 2
}

func IsOlderThan(birthdate time.Time, years int) bool {
    yearsDiff, _, _, _, _, _ := diff(time.Now(), birthdate) //nolint:dogsled
    return yearsDiff >= years
}

// diff calculates the absolute difference between 2 time instances in
// years, months, days, hours, minutes and seconds.
//
// For details, see https://stackoverflow.com/a/36531443/1705598
func diff(a, b time.Time) (year, month, day, hour, min, sec int) {
    if a.Location() != b.Location() {
        b = b.In(a.Location())
    }
    if a.After(b) {
        a, b = b, a
    }
    y1, M1, d1 := a.Date()
    y2, M2, d2 := b.Date()

    h1, m1, s1 := a.Clock()
    h2, m2, s2 := b.Clock()

    year = y2 - y1
    month = int(M2 - M1)
    day = d2 - d1
    hour = h2 - h1
    min = m2 - m1
    sec = s2 - s1

    // Normalize negative values
    if sec < 0 {
        sec += 60
        min--
    }
    if min < 0 {
        min += 60
        hour--
    }
    if hour < 0 {
        hour += 24
        day--
    }
    if day < 0 {
        // days in month:
        t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
        day += 32 - t.Day()
        month--
    }
    if month < 0 {
        month += 12
        year--
    }

    return //nolint:nakedret
}

func HasOnlyDigits(s *string) bool {
    if s == nil {
        return false
    }
    _, err := strconv.ParseUint(*s, 10, 64)
    return err == nil
}

func IsCountryCallingCodeValid(country, callingCode string) bool {
    callingCode = "+" + callingCode
    for _, code := range countries.ByName(country).CallCodes() {
        if code.String() == callingCode {
            return true
        }
    }

    return false
}