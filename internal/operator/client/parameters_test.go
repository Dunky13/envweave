package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParameterValidationDetailReachesOperator(t *testing.T) {
	for _, tc := range []struct{ name, body, want string }{
		{"missing", `{"error":{"code":"bad_request","message":"bad request","detail":"required parameter PR_NUMBER is missing"}}`, "PR_NUMBER"},
		{"schema", `{"error":{"code":"bad_request","message":"bad request","detail":"key APP_BASE_URL failed validation"}}`, "APP_BASE_URL"},
		{"unknown", `{"error":{"code":"internal","detail":"do not echo this"}}`, "fetch validation failed"},
		{"malformed", `not JSON containing confidential text`, "fetch validation failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("parameters") != `{"PR_NUMBER":"123"}` {
					t.Errorf("wire omitted parameter object: %s", r.URL.RawQuery)
				}
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			c, err := NewClient(srv.URL, caPEM(t, srv), "parameter-test")
			if err != nil {
				t.Fatal(err)
			}
			out, outcome, err := c.Fetch(t.Context(), FetchRequest{Org: "o", Project: "p", Environment: "e", Bearer: "credential", Parameters: map[string]string{"PR_NUMBER": "123"}})
			if out != nil || outcome != OutcomeFetchFailed || err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("response = %+v, %v, %v", out, outcome, err)
			}
			if strings.Contains(err.Error(), "do not echo") || strings.Contains(err.Error(), "confidential") {
				t.Fatal("untrusted response body echoed")
			}
		})
	}
}
