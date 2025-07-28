package validation

import (
    "strings"
)

const (
    EmailField       = "email"
    PhoneField       = "phone"
    deviceField      = "device"
    ipAddressField   = "ip_address"
    antiPhishingCode = "code"
)

/* These functions are not used on structs so no TrimSpaces occur on fields,
bee careful to prepare the fields and change if needed
*/

func ValidateIPAddressAndDevice(ipAddress, device string) *Result {
    out := new(Result)

    // Check ip address is valid i.e. IP4 or IP6
    if !IsIpv4Valid(ipAddress) && !IsIpv6Valid(ipAddress) {
        out.AddFieldError(ipAddressField, InvalidIPAddress(ipAddress))
    }

    // Check device is not empty
    if strings.TrimSpace(device) == "" {
        out.AddFieldError(deviceField, EmptyDevice())
    }

    return out
}

func ValidateIPAddress(in string) *Result {
    out := new(Result)

    // Check ip address is valid i.e. IP4 or IP6
    if !IsIpv4Valid(in) && !IsIpv6Valid(in) {
        out.AddFieldError(ipAddressField, InvalidIPAddress(in))
    }

    return out
}

func ValidateAntiPhishingCode(in string) *Result {
    out := new(Result)

    if !IsAntiPhishingCodeValid(in) {
        out.AddFieldError(antiPhishingCode, InvalidAntiPhishingCode())
    }

    return out
}