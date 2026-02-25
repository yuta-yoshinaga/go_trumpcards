package controller

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// makeRestRequest creates a rest.Request with the given JSON body.
func makeRestRequest(body string) *rest.Request {
	httpReq, _ := http.NewRequest("POST", "http://localhost/test", io.NopCloser(bytes.NewBufferString(body)))
	httpReq.Header.Set("Content-Type", "application/json")
	return &rest.Request{Request: httpReq}
}

// --- BlackJack WriteJson error tests ---

type mockBlackJackIF struct{ mock.Mock }

func (m *mockBlackJackIF) Reset() string            { return m.Called().String(0) }
func (m *mockBlackJackIF) Hit() string              { return m.Called().String(0) }
func (m *mockBlackJackIF) Stand() string            { return m.Called().String(0) }
func (m *mockBlackJackIF) Bet(a int) string         { return m.Called(a).String(0) }
func (m *mockBlackJackIF) DoubleDown() string       { return m.Called().String(0) }
func (m *mockBlackJackIF) Split() string            { return m.Called().String(0) }
func (m *mockBlackJackIF) Insurance() string        { return m.Called().String(0) }
func (m *mockBlackJackIF) DeclineInsurance() string { return m.Called().String(0) }

func TestBlackJackWebController_WriteJsonErrors(t *testing.T) {
	bjMock := &mockBlackJackIF{}
	factory := func() usecase.BlackJackInteractorIF { return bjMock }
	ctrl := NewBlackJackWebController(factory)
	defer ctrl.store.Stop()

	t.Run("param error WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "", "sessionId": "s1"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusBadRequest, fw.headerCode)
	})

	t.Run("quit WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "q", "sessionId": "s1"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusOK, fw.headerCode)
	})

	t.Run("session error WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "r", "sessionId": "` + strings.Repeat("a", SessionMaxIDLen+1) + `"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusBadRequest, fw.headerCode)
	})

	t.Run("unsupported command WriteJson fails", func(t *testing.T) {
		bjMock2 := &mockBlackJackIF{}
		factory2 := func() usecase.BlackJackInteractorIF { return bjMock2 }
		ctrl2 := NewBlackJackWebController(factory2)
		defer ctrl2.store.Stop()
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "xyz", "sessionId": "s-bj-unsupported"}`)
		ctrl2.Exec(fw, req)
		assert.Equal(t, http.StatusOK, fw.headerCode)
	})
}

// --- Poker WriteJson error tests ---

type mockPokerIF struct{ mock.Mock }

func (m *mockPokerIF) Reset() string           { return m.Called().String(0) }
func (m *mockPokerIF) Exchange(i []int) string { return m.Called(i).String(0) }
func (m *mockPokerIF) Stand() string           { return m.Called().String(0) }
func (m *mockPokerIF) Bet(a int) string        { return m.Called(a).String(0) }
func (m *mockPokerIF) Call() string            { return m.Called().String(0) }
func (m *mockPokerIF) Raise(a int) string      { return m.Called(a).String(0) }
func (m *mockPokerIF) Fold() string            { return m.Called().String(0) }
func (m *mockPokerIF) Check() string           { return m.Called().String(0) }

func TestPokerWebController_WriteJsonErrors(t *testing.T) {
	pkMock := &mockPokerIF{}
	factory := func() usecase.PokerInteractorIF { return pkMock }
	ctrl := NewPokerWebController(factory)
	defer ctrl.store.Stop()

	t.Run("param error WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "", "sessionId": "s1"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusBadRequest, fw.headerCode)
	})

	t.Run("quit WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "q", "sessionId": "s1"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusOK, fw.headerCode)
	})

	t.Run("session error WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "r", "sessionId": "` + strings.Repeat("a", SessionMaxIDLen+1) + `"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusBadRequest, fw.headerCode)
	})

	t.Run("unsupported command WriteJson fails", func(t *testing.T) {
		pkMock2 := &mockPokerIF{}
		factory2 := func() usecase.PokerInteractorIF { return pkMock2 }
		ctrl2 := NewPokerWebController(factory2)
		defer ctrl2.store.Stop()
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "xyz", "sessionId": "s-pk-unsupported"}`)
		ctrl2.Exec(fw, req)
		assert.Equal(t, http.StatusOK, fw.headerCode)
	})
}

// --- OldMaid WriteJson error tests ---

type mockOldMaidIF struct{ mock.Mock }

func (m *mockOldMaidIF) Reset() string       { return m.Called().String(0) }
func (m *mockOldMaidIF) Draw(idx int) string { return m.Called(idx).String(0) }

func TestOldMaidWebController_WriteJsonErrors(t *testing.T) {
	omMock := &mockOldMaidIF{}
	factory := func() usecase.OldMaidInteractorIF { return omMock }
	ctrl := NewOldMaidWebController(factory)
	defer ctrl.store.Stop()

	t.Run("param error WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "", "sessionId": "s1"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusBadRequest, fw.headerCode)
	})

	t.Run("quit WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "q", "sessionId": "s1"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusOK, fw.headerCode)
	})

	t.Run("session error WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "r", "sessionId": "` + strings.Repeat("a", SessionMaxIDLen+1) + `"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusBadRequest, fw.headerCode)
	})

	t.Run("unsupported command WriteJson fails", func(t *testing.T) {
		omMock2 := &mockOldMaidIF{}
		factory2 := func() usecase.OldMaidInteractorIF { return omMock2 }
		ctrl2 := NewOldMaidWebController(factory2)
		defer ctrl2.store.Stop()
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "xyz", "sessionId": "s-om-unsupported"}`)
		ctrl2.Exec(fw, req)
		assert.Equal(t, http.StatusOK, fw.headerCode)
	})
}

// --- Daifugo WriteJson error tests ---

type mockDaifugoIF struct{ mock.Mock }

func (m *mockDaifugoIF) Reset() string       { return m.Called().String(0) }
func (m *mockDaifugoIF) Play(i []int) string { return m.Called(i).String(0) }

func TestDaifugoWebController_WriteJsonErrors(t *testing.T) {
	dgMock := &mockDaifugoIF{}
	factory := func() usecase.DaifugoInteractorIF { return dgMock }
	ctrl := NewDaifugoWebController(factory)
	defer ctrl.store.Stop()

	t.Run("param error WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "", "sessionId": "s1"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusBadRequest, fw.headerCode)
	})

	t.Run("quit WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "q", "sessionId": "s1"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusOK, fw.headerCode)
	})

	t.Run("session error WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "r", "sessionId": "` + strings.Repeat("a", SessionMaxIDLen+1) + `"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusBadRequest, fw.headerCode)
	})

	t.Run("unsupported command WriteJson fails", func(t *testing.T) {
		dgMock2 := &mockDaifugoIF{}
		factory2 := func() usecase.DaifugoInteractorIF { return dgMock2 }
		ctrl2 := NewDaifugoWebController(factory2)
		defer ctrl2.store.Stop()
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "xyz", "sessionId": "s-dg-unsupported"}`)
		ctrl2.Exec(fw, req)
		assert.Equal(t, http.StatusOK, fw.headerCode)
	})
}

// --- Sevens WriteJson error tests ---

type mockSevensIF struct{ mock.Mock }

func (m *mockSevensIF) Reset() string { return m.Called().String(0) }
func (m *mockSevensIF) ResetWithConfig(t bool, j int, c bool) string {
	return m.Called(t, j, c).String(0)
}
func (m *mockSevensIF) Play(idx int) string { return m.Called(idx).String(0) }
func (m *mockSevensIF) PlayJoker(idx, suit, val int) string {
	return m.Called(idx, suit, val).String(0)
}

func TestSevensWebController_WriteJsonErrors(t *testing.T) {
	svMock := &mockSevensIF{}
	factory := func() usecase.SevensInteractorIF { return svMock }
	ctrl := NewSevensWebController(factory)
	defer ctrl.store.Stop()

	t.Run("param error WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "", "sessionId": "s1"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusBadRequest, fw.headerCode)
	})

	t.Run("quit WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "q", "sessionId": "s1"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusOK, fw.headerCode)
	})

	t.Run("session error WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "r", "sessionId": "` + strings.Repeat("a", SessionMaxIDLen+1) + `"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusBadRequest, fw.headerCode)
	})

	t.Run("unsupported command WriteJson fails", func(t *testing.T) {
		svMock2 := &mockSevensIF{}
		factory2 := func() usecase.SevensInteractorIF { return svMock2 }
		ctrl2 := NewSevensWebController(factory2)
		defer ctrl2.store.Stop()
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "xyz", "sessionId": "s-sv-unsupported"}`)
		ctrl2.Exec(fw, req)
		assert.Equal(t, http.StatusOK, fw.headerCode)
	})
}

// --- Doubt WriteJson error tests ---

type mockDoubtIF struct{ mock.Mock }

func (m *mockDoubtIF) Reset() string                 { return m.Called().String(0) }
func (m *mockDoubtIF) Play(i []int, v int) string    { return m.Called(i, v).String(0) }
func (m *mockDoubtIF) ResolveDoubt(idx []int) string { return m.Called(idx).String(0) }
func (m *mockDoubtIF) SkipDoubt() string             { return m.Called().String(0) }
func (m *mockDoubtIF) GetCpuDoubters() []int {
	ret := m.Called()
	if val, ok := ret.Get(0).([]int); ok {
		return val
	}
	return nil
}

func TestDoubtWebController_WriteJsonErrors(t *testing.T) {
	dwMock := &mockDoubtIF{}
	factory := func() usecase.DoubtInteractorIF { return dwMock }
	ctrl := NewDoubtWebController(factory)
	defer ctrl.store.Stop()

	t.Run("param error WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "", "sessionId": "s1"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusBadRequest, fw.headerCode)
	})

	t.Run("quit WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "q", "sessionId": "s1"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusOK, fw.headerCode)
	})

	t.Run("session error WriteJson fails", func(t *testing.T) {
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "r", "sessionId": "` + strings.Repeat("a", SessionMaxIDLen+1) + `"}`)
		ctrl.Exec(fw, req)
		assert.Equal(t, http.StatusBadRequest, fw.headerCode)
	})

	t.Run("unsupported command WriteJson fails", func(t *testing.T) {
		dwMock2 := &mockDoubtIF{}
		factory2 := func() usecase.DoubtInteractorIF { return dwMock2 }
		ctrl2 := NewDoubtWebController(factory2)
		defer ctrl2.store.Stop()
		fw := newFailWriter()
		req := makeRestRequest(`{"command": "xyz", "sessionId": "s-dw-unsupported"}`)
		ctrl2.Exec(fw, req)
		assert.Equal(t, http.StatusOK, fw.headerCode)
	})
}
