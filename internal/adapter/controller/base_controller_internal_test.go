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

// --- deref helper tests ---

func TestDerefBool(t *testing.T) {
	t.Run("nil returns false", func(t *testing.T) {
		assert.False(t, derefBool(nil))
	})
	t.Run("non-nil returns value", func(t *testing.T) {
		v := true
		assert.True(t, derefBool(&v))
	})
}

func TestDerefBoolDefault(t *testing.T) {
	t.Run("nil returns default true", func(t *testing.T) {
		assert.True(t, derefBoolDefault(nil, true))
	})
	t.Run("nil returns default false", func(t *testing.T) {
		assert.False(t, derefBoolDefault(nil, false))
	})
	t.Run("non-nil returns value", func(t *testing.T) {
		v := false
		assert.False(t, derefBoolDefault(&v, true))
	})
}

func TestDerefInt(t *testing.T) {
	t.Run("nil returns 0", func(t *testing.T) {
		assert.Equal(t, 0, derefInt(nil))
	})
	t.Run("non-nil returns value", func(t *testing.T) {
		v := 42
		assert.Equal(t, 42, derefInt(&v))
	})
}

func TestDerefIntDefault(t *testing.T) {
	t.Run("nil returns default", func(t *testing.T) {
		assert.Equal(t, 99, derefIntDefault(nil, 99))
	})
	t.Run("non-nil returns value", func(t *testing.T) {
		v := 7
		assert.Equal(t, 7, derefIntDefault(&v, 99))
	})
}

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
	BaseWebInput
}

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
		func(_ rest.ResponseWriter, _ *testInteractor, _ testInput) bool { return true },
	)
	assert.Equal(t, http.StatusBadRequest, sw.headerCode)
	assert.Equal(t, map[string]string{"message": "param error."}, sw.body)
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
		func(_ rest.ResponseWriter, _ *testInteractor, _ testInput) bool {
			return false
		},
	)
	assert.Equal(t, http.StatusBadRequest, sw.headerCode)
	assert.Equal(t, map[string]string{"message": "Unsupported command."}, sw.body)
}
