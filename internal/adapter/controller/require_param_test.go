package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// reqParamTestOutput is a minimal *WebOutput stand-in for exercising
// requireParam: it just needs a JSON-decodable message field so the test can
// confirm the supplied param-error message is propagated (issue #2102).
type reqParamTestOutput struct {
	Message string `json:"message"`
}

func newReqParamTestOutput(msg string) *reqParamTestOutput {
	return &reqParamTestOutput{Message: msg}
}

// TestRequireParam_MissingWrites400 verifies the helper writes a 400 carrying
// the supplied message (built via newDefault) and returns false when the
// required parameter is missing.
func TestRequireParam_MissingWrites400(t *testing.T) {
	bc := &baseController{}
	w := httptest.NewRecorder()

	ok := requireParam(bc, w, newReqParamTestOutput, true, "param error: col is required.")

	if ok {
		t.Fatal("requireParam returned true for a missing param; want false")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	var got reqParamTestOutput
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.Message != "param error: col is required." {
		t.Errorf("message = %q, want the supplied param-error message", got.Message)
	}
}

// TestRequireParam_PresentDoesNothing verifies the helper returns true and
// writes no response when the required parameter is present.
func TestRequireParam_PresentDoesNothing(t *testing.T) {
	bc := &baseController{}
	w := httptest.NewRecorder()

	ok := requireParam(bc, w, newReqParamTestOutput, false, "param error: col is required.")

	if !ok {
		t.Fatal("requireParam returned false for a present param; want true")
	}
	if w.Body.Len() != 0 {
		t.Errorf("body = %q, want empty (helper must not write when param present)", w.Body.String())
	}
}
