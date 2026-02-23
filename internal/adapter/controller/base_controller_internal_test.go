package controller

import (
	"errors"
	"net/http"
	"testing"

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

func TestWritePresenterResponse_WriteJsonError_Success(t *testing.T) {
	bc := &baseController{}
	fw := newFailWriter()

	// Valid JSON response string — triggers the success WriteJson path (line 24).
	bc.writePresenterResponse(fw, `{"ok":true}`, "fallback")
	assert.Equal(t, http.StatusOK, fw.headerCode)
}

func TestWritePresenterResponse_WriteJsonError_Error(t *testing.T) {
	bc := &baseController{}
	fw := newFailWriter()

	// Empty response string — triggers the error WriteJson path (line 18).
	bc.writePresenterResponse(fw, "", "fallback")
	assert.Equal(t, http.StatusBadRequest, fw.headerCode)
}
