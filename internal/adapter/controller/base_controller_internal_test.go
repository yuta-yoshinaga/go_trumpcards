package controller

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- deref helper tests ---

func TestDeref(t *testing.T) {
	t.Run("bool nil returns zero", func(t *testing.T) {
		assert.False(t, deref[bool](nil))
	})
	t.Run("bool non-nil returns value", func(t *testing.T) {
		v := true
		assert.True(t, deref(&v))
	})
	t.Run("int nil returns zero", func(t *testing.T) {
		assert.Equal(t, 0, deref[int](nil))
	})
	t.Run("int non-nil returns value", func(t *testing.T) {
		v := 42
		assert.Equal(t, 42, deref(&v))
	})
}

func TestDerefDefault(t *testing.T) {
	t.Run("bool nil returns default true", func(t *testing.T) {
		assert.True(t, derefDefault[bool](nil, true))
	})
	t.Run("bool nil returns default false", func(t *testing.T) {
		assert.False(t, derefDefault[bool](nil, false))
	})
	t.Run("bool non-nil returns value", func(t *testing.T) {
		v := false
		assert.False(t, derefDefault(&v, true))
	})
	t.Run("int nil returns default", func(t *testing.T) {
		assert.Equal(t, 99, derefDefault[int](nil, 99))
	})
	t.Run("int non-nil returns value", func(t *testing.T) {
		v := 7
		assert.Equal(t, 7, derefDefault(&v, 99))
	})
}

// failWriter is a minimal http.ResponseWriter implementation where Write always fails.
type failWriter struct {
	header     http.Header
	headerCode int
}

func newFailWriter() *failWriter {
	return &failWriter{header: http.Header{}}
}

func (fw *failWriter) Header() http.Header         { return fw.header }
func (fw *failWriter) WriteHeader(code int)        { fw.headerCode = code }
func (fw *failWriter) Write(_ []byte) (int, error) { return 0, errors.New("write failed") }

func TestWritePresenterResponse_WriteError(t *testing.T) {
	bc := &baseController{}
	fw := newFailWriter()

	// Valid JSON response string — triggers the Write error path.
	bc.writePresenterResponse(fw, `{"ok":true}`)
	assert.Equal(t, http.StatusOK, fw.headerCode)
}

// --- writeJsonResponse tests ---

func TestWriteJsonResponse_Success(t *testing.T) {
	bc := &baseController{}
	rec := httptest.NewRecorder()

	bc.writeJsonResponse(rec, http.StatusOK, map[string]string{"msg": "ok"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"msg":"ok"`)
}

func TestWriteJsonResponse_WriteError(t *testing.T) {
	bc := &baseController{}
	fw := newFailWriter()

	// Write fails — should log but not panic.
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
	provider := NewMemorySessionProvider[*testInteractor]()
	defer provider.Stop()
	rec := httptest.NewRecorder()
	req := makeHTTPRequest(`{invalid json}`)

	execWithSession(bc, rec, req, provider,
		func() *testInteractor { return &testInteractor{} },
		func(msg string) any { return map[string]string{"message": msg} },
		func(_ http.ResponseWriter, _ *testInteractor, _ testInput) bool { return true },
	)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "param error.")
}

func TestExecWithSession_EmptyCommand(t *testing.T) {
	bc := &baseController{}
	provider := NewMemorySessionProvider[*testInteractor]()
	defer provider.Stop()
	rec := httptest.NewRecorder()
	req := makeHTTPRequest(`{"command":"","sessionId":"s1"}`)

	execWithSession(bc, rec, req, provider,
		func() *testInteractor { return &testInteractor{} },
		func(msg string) any { return map[string]string{"message": msg} },
		func(_ http.ResponseWriter, _ *testInteractor, _ testInput) bool { return true },
	)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "param error.")
}

func TestExecWithSession_EmptySessionId(t *testing.T) {
	bc := &baseController{}
	provider := NewMemorySessionProvider[*testInteractor]()
	defer provider.Stop()
	rec := httptest.NewRecorder()
	req := makeHTTPRequest(`{"command":"r","sessionId":""}`)

	execWithSession(bc, rec, req, provider,
		func() *testInteractor { return &testInteractor{} },
		func(msg string) any { return map[string]string{"message": msg} },
		func(_ http.ResponseWriter, _ *testInteractor, _ testInput) bool { return true },
	)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "param error.")
}

func TestExecWithSession_QuitCommand(t *testing.T) {
	for _, cmd := range []string{"q", "quit"} {
		t.Run(cmd, func(t *testing.T) {
			bc := &baseController{}
			provider := NewMemorySessionProvider[*testInteractor]()
			defer provider.Stop()
			rec := httptest.NewRecorder()
			req := makeHTTPRequest(`{"command":"` + cmd + `","sessionId":"s1"}`)

			execWithSession(bc, rec, req, provider,
				func() *testInteractor { return &testInteractor{} },
				func(msg string) any { return map[string]string{"message": msg} },
				func(_ http.ResponseWriter, _ *testInteractor, _ testInput) bool { return true },
			)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), "bye.")
		})
	}
}

func TestExecWithSession_SessionRetrievalFailure(t *testing.T) {
	bc := &baseController{}
	provider := NewMemorySessionProvider[*testInteractor]()
	defer provider.Stop()
	rec := httptest.NewRecorder()
	// Session ID exceeding max length triggers Acquire failure.
	longID := strings.Repeat("x", SessionMaxIDLen+1)
	req := makeHTTPRequest(`{"command":"r","sessionId":"` + longID + `"}`)

	execWithSession(bc, rec, req, provider,
		func() *testInteractor { return &testInteractor{} },
		func(msg string) any { return map[string]string{"message": msg} },
		func(_ http.ResponseWriter, _ *testInteractor, _ testInput) bool { return true },
	)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "param error.")
}

func TestExecWithSession_HandlerReturnsTrue(t *testing.T) {
	bc := &baseController{}
	provider := NewMemorySessionProvider[*testInteractor]()
	defer provider.Stop()
	rec := httptest.NewRecorder()
	req := makeHTTPRequest(`{"command":"r","sessionId":"s1"}`)

	execWithSession(bc, rec, req, provider,
		func() *testInteractor { return &testInteractor{} },
		func(msg string) any { return map[string]string{"message": msg} },
		func(w http.ResponseWriter, _ *testInteractor, _ testInput) bool {
			bc.writeJsonResponse(w, http.StatusOK, map[string]string{"message": "handled"})
			return true
		},
	)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "handled")
}

func TestExecWithSession_HandlerReturnsFalse(t *testing.T) {
	bc := &baseController{}
	provider := NewMemorySessionProvider[*testInteractor]()
	defer provider.Stop()
	rec := httptest.NewRecorder()
	req := makeHTTPRequest(`{"command":"xyz","sessionId":"s1"}`)

	execWithSession(bc, rec, req, provider,
		func() *testInteractor { return &testInteractor{} },
		func(msg string) any { return map[string]string{"message": msg} },
		func(_ http.ResponseWriter, _ *testInteractor, _ testInput) bool {
			return false
		},
	)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Unsupported command.")
}
