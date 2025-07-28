package validation

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

type testIsOlderThan struct {
    birthdate time.Time
    age       int
    isOlder   bool
}

type testhHasOnlyDigits struct {
    s     string
    valid bool
}

type testIsPasswordValid struct {
    pass  string
    valid bool
}

func TestIsPasswordValid(t *testing.T) {
    testCases := []testIsPasswordValid{
        {pass: "adgA4$qq", valid: true},
        {pass: "11gg@@TQTQ", valid: true},
        {pass: "@$%#ssTT436", valid: true},
        {pass: "6611aB#1111", valid: true},
        {pass: "#)(2241aaFg", valid: true},
        {pass: "1aa##5(&R", valid: true},
        {pass: "", valid: false},
        {pass: " ", valid: false},
        {pass: "ASD$%123TY@@", valid: false},      // no lower
        {pass: "asd495$##%dksadf", valid: false},  // no upper
        {pass: "asdASHOAIH8952935", valid: false}, // no special
        {pass: "aR#5", valid: false},              // short
    }

    for _, tc := range testCases {
        t.Run(tc.pass, func(t *testing.T) {
            assert.Equal(t, tc.valid, IsPasswordValid(tc.pass))
        })
    }
}

func TestIsOlderThan(t *testing.T) {
    testCases := []testIsOlderThan{
        {birthdate: tparse("02 May 03 15:04 MST"), age: 18, isOlder: true},
        {birthdate: tparse("02 Jan 02 15:04 EET"), age: 18, isOlder: true},
        {birthdate: tparse("02 Jan 04 15:04 MST"), age: 18, isOlder: false},
        {birthdate: tparse("02 Sep 03 15:04 MST"), age: 18, isOlder: false}, // TODO: will fail in September
        {birthdate: time.Now().Add(time.Second), age: 0, isOlder: true},
        {birthdate: time.Now().UTC().Add(time.Second), age: 0, isOlder: true},
    }

    for _, tc := range testCases {
        t.Run(tc.birthdate.String(), func(t *testing.T) {
            assert.Equal(t, tc.isOlder, IsOlderThan(tc.birthdate, tc.age))
        })
    }
}

func tparse(val string) time.Time {
    out, err := time.Parse(time.RFC822, val)
    if err != nil {
        panic(err)
    }
    return out
}

func TestHasOnlyDigits(t *testing.T) {
    testCases := []testhHasOnlyDigits{
        {s: "", valid: false},
        {s: " ", valid: false},
        {s: "123123", valid: true},
        {s: "5123958315735815", valid: true},
        {s: "(123)1255", valid: false},
        {s: "123-5515", valid: false},
        {s: "3 53158", valid: false},
    }

    for _, tc := range testCases {
        t.Run(tc.s, func(t *testing.T) {
            assert.Equal(t, tc.valid, HasOnlyDigits(&tc.s))
        })
    }
}