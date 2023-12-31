package rlutils

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// testHTTPRequest is a utility function that creates a new http.Request with a RemoteAddr set
func testHTTPRequest(remoteAddr string) *http.Request {
	request, _ := http.NewRequest("GET", "http://example.com", nil)
	request.RemoteAddr = remoteAddr
	return request
}

func TestCountryLimiter(t *testing.T) {
	abspath, _ := filepath.Abs("./testdata/GeoIP2-Country-Test.mmdb")
	reqLimit := 10

	// Define your test cases
	testCases := []struct {
		name            string
		request         *http.Request
		countries       []string
		skipCountries   []string
		expectedCountry string
		expectedError   bool
		shouldLimit     bool
	}{
		{
			name:            "Valid IP from United States With Port",
			request:         testHTTPRequest("50.114.0.1:1234"),
			expectedCountry: "US",
			countries:       []string{"US"},
			shouldLimit:     true,
			expectedError:   false,
		},
		{
			name:            "Valid IP from United States",
			request:         testHTTPRequest("50.114.0.1"),
			expectedCountry: "US",
			shouldLimit:     true,
			expectedError:   false,
		},
		{
			name:            "Invalid IP format",
			request:         testHTTPRequest("invalid-ip"),
			expectedCountry: "",
			shouldLimit:     false,
			expectedError:   true,
		},
		{
			name:            "Valid IP from United States With WildCard",
			request:         testHTTPRequest("50.114.0.1"),
			expectedCountry: "US",
			countries:       []string{"*"},
			shouldLimit:     true,
			expectedError:   false,
		},
		{
			name:            "Valid IP from United States With Skip country",
			request:         testHTTPRequest("50.114.0.1"),
			expectedCountry: "US",
			countries:       []string{"*"},
			skipCountries:   []string{"US"},
			shouldLimit:     false,
			expectedError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run the country function to get the ISO country code
			cl, err := NewCountryLimiter(
				abspath,
				[]string{"US"},
				tc.skipCountries,
				reqLimit,
				1*time.Hour,
				nil,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}
			remoteAddr := strings.Split(tc.request.RemoteAddr, ":")[0]
			country, err := cl.country(remoteAddr)
			if tc.expectedError {
				if err == nil {
					t.Errorf("Expected an error but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got: %v", err)
				}
				if country != tc.expectedCountry {
					t.Errorf("Expected country %s, but got %s", tc.expectedCountry, country)
				}
			}

			rule, ruleErr := cl.Rule(tc.request)
			if ruleErr != nil {
				if !tc.expectedError {
					t.Errorf("Rule method returned an unexpected error: %v", ruleErr)
				}
			} else {
				if tc.shouldLimit && (rule == nil || rule.ReqLimit != reqLimit) {
					t.Errorf("Expected shouldLimit rule with limit %d, but got %+v", reqLimit, rule)
				}
				if !tc.shouldLimit && (rule == nil || rule.ReqLimit != -1) {
					t.Errorf("Expected disshouldLimit rule with no limiting, but got %+v", rule)
				}
			}

		})
	}
}

func TestCountryLimiterWithContext(t *testing.T) {
	abspath, _ := filepath.Abs("./testdata/GeoIP2-Country-Test.mmdb")
	reqLimit := 10

	// Define your test cases
	testCases := []struct {
		name            string
		request         *http.Request
		countries       []string
		expectedCountry string
		shouldLimit     bool
	}{
		{
			name:        "Valid IP from United States With Port",
			request:     testHTTPRequest("50.114.0.1:1234"),
			countries:   []string{"US"},
			shouldLimit: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run the country function to get the ISO country code
			cl, err := NewCountryLimiter(
				abspath,
				[]string{"US"},
				[]string{},
				reqLimit,
				1*time.Hour,
				nil,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}
			ctx := tc.request.Context()
			ctx = context.WithValue(ctx, ContextCountryKey, "US")
			tc.request = tc.request.WithContext(ctx)
			rule, _ := cl.Rule(tc.request)
			if tc.shouldLimit && (rule == nil || rule.ReqLimit != reqLimit) {
				t.Errorf("Expected shouldLimit rule with limit %d, but got %+v", reqLimit, rule)
			}
			if !tc.shouldLimit && (rule == nil || rule.ReqLimit != -1) {
				t.Errorf("Expected disshouldLimit rule with no limiting, but got %+v", rule)
			}

		})
	}
}
