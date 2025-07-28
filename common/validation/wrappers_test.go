package validation

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

type testEmail struct {
    email    string
    expected bool
}

type testIPAddress struct {
    ip       string
    expected bool
}

func TestValidateEmail(t *testing.T) {
    testCases := []testEmail{
        {email: "", expected: false},
        {email: "abc@xyz.com", expected: true},
        {email: "$!#@google.com", expected: true},
        {email: "a_bc@google.com", expected: true},
        {email: "a.bc@go_gle.com", expected: false},
        {email: "a.bc@go-ogle.com", expected: true},
        {email: "abc@googl", expected: false},
    }

    for i := range testCases {
        t.Run(testCases[i].email, func(t *testing.T) {
            assert.Equal(t, testCases[i].expected, IsEmailValid(testCases[i].email))
        })
    }
}

func TestValidateIpv4(t *testing.T) {
    testCases := []testIPAddress{
        {ip: "", expected: false},
        {ip: " ", expected: false},
        {ip: "1.0.0.0", expected: true},
        {ip: "114.114.141.29", expected: true},
        {ip: "127.0.0.1", expected: true},
        {ip: "82.28.28.28", expected: true},
        {ip: "5.140.105.291", expected: false},
        {ip: "5.140.1253.291", expected: false},
        {ip: "256.0.0.0", expected: false},
    }

    for i := range testCases {
        t.Run(testCases[i].ip, func(t *testing.T) {
            assert.Equal(t, testCases[i].expected, IsIpv4Valid(testCases[i].ip))
        })
    }
}

func TestValidateIpv6(t *testing.T) {
    testCases := []testIPAddress{
        {ip: "", expected: false},
        {ip: " ", expected: false},
        {ip: "::00:192.168.10.184", expected: true},
        {ip: "::1", expected: true},
        {ip: "ae34:ae:fe:12:51:5af:bcde:123", expected: true},
        {ip: "fe80::219:7eff:fe46:6c42", expected: true},
        {ip: "::", expected: true},
    }

    for i := range testCases {
        t.Run(testCases[i].ip, func(t *testing.T) {
            assert.Equal(t, testCases[i].expected, IsIpv6Valid(testCases[i].ip))
        })
    }
}