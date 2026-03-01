package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/stretchr/testify/assert"
)

// failWriter is a minimal rest.ResponseWriter implementation where WriteJson always fails.
type failWriter struct {
	header     http.Header
	headerCode int
}

func newFailWriter() *failWriter {
	return &failWriter{header: http.Header{}}
}

func (fw *failWriter) Header() http.Header   { return fw.header }
func (fw *failWriter) WriteHeader(code int)  { fw.headerCode = code }
func (fw *failWriter) WriteJson(_ any) error { return errors.New("write failed") }
func (fw *failWriter) EncodeJson(v any) ([]byte, error) {
	return nil, errors.New("encode failed")
}

// successWriter is a minimal rest.ResponseWriter that records the status code and JSON body.
type successWriter struct {
	header     http.Header
	headerCode int
	body       any
}

func newSuccessWriter() *successWriter {
	return &successWriter{header: http.Header{}}
}

func (sw *successWriter) Header() http.Header  { return sw.header }
func (sw *successWriter) WriteHeader(code int) { sw.headerCode = code }
func (sw *successWriter) WriteJson(v any) error {
	sw.body = v
	return nil
}
func (sw *successWriter) EncodeJson(v any) ([]byte, error) { return json.Marshal(v) }

func TestWritePresenterResponse_WriteJsonError(t *testing.T) {
	bc := &baseController{}
	fw := newFailWriter()

	// Valid JSON response string — triggers the WriteJson error path.
	bc.writePresenterResponse(fw, `{"ok":true}`)
	assert.Equal(t, http.StatusOK, fw.headerCode)
}

// --- writeJsonResponse tests ---

func TestWriteJsonResponse_Success(t *testing.T) {
	bc := &baseController{}
	sw := newSuccessWriter()

	bc.writeJsonResponse(sw, http.StatusOK, map[string]string{"msg": "ok"})
	assert.Equal(t, http.StatusOK, sw.headerCode)
	assert.Equal(t, map[string]string{"msg": "ok"}, sw.body)
}

func TestWriteJsonResponse_WriteJsonError(t *testing.T) {
	bc := &baseController{}
	fw := newFailWriter()

	// WriteJson fails — should log but not panic.
	bc.writeJsonResponse(fw, http.StatusBadRequest, "error body")
	assert.Equal(t, http.StatusBadRequest, fw.headerCode)
}

// --- execWithSession tests ---

// testInput is a minimal WebInput implementation for testing.
type testInput struct {
	Command   string `json:"command"`
	SessionId string `json:"sessionId"`
}

func (i testInput) GetCommand() string   { return i.Command }
func (i testInput) GetSessionID() string { return i.SessionId }

// testInteractor is a trivial interactor for testing.
type testInteractor struct{}

func TestExecWithSession_DecodeError(t *testing.T) {
	bc := &baseController{}
	store := NewSessionStore[*testInteractor]()
	defer store.Stop()
	sw := newSuccessWriter()
	req := makeRestRequest(`{invalid json}`)

	execWithSession(bc, sw, req, store,
		func() *testInteractor { return &testInteractor{} },
		func(msg string) any { return map[string]string{"message": msg} },
		nil,
		func(_ rest.ResponseWriter, _ *testInteractor, _ testInput) bool { return true },
	)
	assert.Equal(t, http.StatusBadRequest, sw.headerCode)
	assert.Equal(t, map[string]string{"message": "param error."}, sw.body)
}

func TestExecWithSession_EmptyCommand(t *testing.T) {
	bc := &baseController{}
	store := NewSessionStore[*testInteractor]()
	defer store.Stop()
	sw := newSuccessWriter()
	req := makeRestRequest(`{"command":"","sessionId":"s1"}`)

	execWithSession(bc, sw, req, store,
		func() *testInteractor { return &testInteractor{} },
		func(msg string) any { return map[string]string{"message": msg} },
		nil,
		func(_ rest.ResponseWriter, _ *testInteractor, _ testInput) bool { return true },
	)
	assert.Equal(t, http.StatusBadRequest, sw.headerCode)
	assert.Equal(t, map[string]string{"message": "param error."}, sw.body)
}

func TestExecWithSession_EmptySessionId(t *testing.T) {
	bc := &baseController{}
	store := NewSessionStore[*testInteractor]()
	defer store.Stop()
	sw := newSuccessWriter()
	req := makeRestRequest(`{"command":"r","sessionId":""}`)

	execWithSession(bc, sw, req, store,
		func() *testInteractor { return &testInteractor{} },
		func(msg string) any { return map[string]string{"message": msg} },
		nil,
		func(_ rest.ResponseWriter, _ *testInteractor, _ testInput) bool { return true },
	)
	assert.Equal(t, http.StatusBadRequest, sw.headerCode)
	assert.Equal(t, map[string]string{"message": "param error."}, sw.body)
}

func TestExecWithSession_ValidateReturnsError(t *testing.T) {
	bc := &baseController{}
	store := NewSessionStore[*testInteractor]()
	defer store.Stop()
	sw := newSuccessWriter()
	req := makeRestRequest(`{"command":"r","sessionId":"s1"}`)

	execWithSession(bc, sw, req, store,
		func() *testInteractor { return &testInteractor{} },
		func(msg string) any { return map[string]string{"message": msg} },
		func(_ testInput) error { return errors.New("invalid") },
		func(_ rest.ResponseWriter, _ *testInteractor, _ testInput) bool { return true },
	)
	assert.Equal(t, http.StatusBadRequest, sw.headerCode)
	assert.Equal(t, map[string]string{"message": "param error."}, sw.body)
}

func TestExecWithSession_ValidateNilSkipped(t *testing.T) {
	bc := &baseController{}
	store := NewSessionStore[*testInteractor]()
	defer store.Stop()
	sw := newSuccessWriter()
	req := makeRestRequest(`{"command":"r","sessionId":"s1"}`)
	handlerCalled := false

	execWithSession(bc, sw, req, store,
		func() *testInteractor { return &testInteractor{} },
		func(msg string) any { return map[string]string{"message": msg} },
		nil,
		func(_ rest.ResponseWriter, _ *testInteractor, _ testInput) bool {
			handlerCalled = true
			return true
		},
	)
	assert.True(t, handlerCalled)
}

func TestExecWithSession_QuitCommand(t *testing.T) {
	for _, cmd := range []string{"q", "quit"} {
		t.Run(cmd, func(t *testing.T) {
			bc := &baseController{}
			store := NewSessionStore[*testInteractor]()
			defer store.Stop()
			sw := newSuccessWriter()
			req := makeRestRequest(`{"command":"` + cmd + `","sessionId":"s1"}`)

			execWithSession(bc, sw, req, store,
				func() *testInteractor { return &testInteractor{} },
				func(msg string) any { return map[string]string{"message": msg} },
				nil,
				func(_ rest.ResponseWriter, _ *testInteractor, _ testInput) bool { return true },
			)
			assert.Equal(t, http.StatusOK, sw.headerCode)
			assert.Equal(t, map[string]string{"message": "bye."}, sw.body)
		})
	}
}

func TestExecWithSession_SessionRetrievalFailure(t *testing.T) {
	bc := &baseController{}
	store := NewSessionStore[*testInteractor]()
	defer store.Stop()
	sw := newSuccessWriter()
	// Session ID exceeding max length triggers GetWithLock failure.
	longID := strings.Repeat("x", SessionMaxIDLen+1)
	req := makeRestRequest(`{"command":"r","sessionId":"` + longID + `"}`)

	execWithSession(bc, sw, req, store,
		func() *testInteractor { return &testInteractor{} },
		func(msg string) any { return map[string]string{"message": msg} },
		nil,
		func(_ rest.ResponseWriter, _ *testInteractor, _ testInput) bool { return true },
	)
	assert.Equal(t, http.StatusBadRequest, sw.headerCode)
	assert.Equal(t, map[string]string{"message": "param error."}, sw.body)
}

func TestExecWithSession_HandlerReturnsTrue(t *testing.T) {
	bc := &baseController{}
	store := NewSessionStore[*testInteractor]()
	defer store.Stop()
	sw := newSuccessWriter()
	req := makeRestRequest(`{"command":"r","sessionId":"s1"}`)

	execWithSession(bc, sw, req, store,
		func() *testInteractor { return &testInteractor{} },
		func(msg string) any { return map[string]string{"message": msg} },
		nil,
		func(w rest.ResponseWriter, _ *testInteractor, _ testInput) bool {
			bc.writeJsonResponse(w, http.StatusOK, map[string]string{"message": "handled"})
			return true
		},
	)
	assert.Equal(t, http.StatusOK, sw.headerCode)
	assert.Equal(t, map[string]string{"message": "handled"}, sw.body)
}

func TestExecWithSession_HandlerReturnsFalse(t *testing.T) {
	bc := &baseController{}
	store := NewSessionStore[*testInteractor]()
	defer store.Stop()
	sw := newSuccessWriter()
	req := makeRestRequest(`{"command":"xyz","sessionId":"s1"}`)

	execWithSession(bc, sw, req, store,
		func() *testInteractor { return &testInteractor{} },
		func(msg string) any { return map[string]string{"message": msg} },
		nil,
		func(_ rest.ResponseWriter, _ *testInteractor, _ testInput) bool {
			return false
		},
	)
	assert.Equal(t, http.StatusBadRequest, sw.headerCode)
	assert.Equal(t, map[string]string{"message": "Unsupported command."}, sw.body)
}
