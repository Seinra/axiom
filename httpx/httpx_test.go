package httpx_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/seinra/axiom/httpx"
	"github.com/seinra/axiom/repo"
)

// personPayload is the typed body every binding scenario round-trips.
type personPayload struct {
	Name string `json:"name"`
}

// newPostContext builds an Echo context carrying payload as a JSON POST body.
func newPostContext(engine *echo.Echo, payload string) (*echo.Context, *httptest.ResponseRecorder) {
	httpRequest := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload))
	httpRequest.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	httpContext := engine.NewContext(httpRequest, recorder)
	return httpContext, recorder
}

// newBareContext builds an Echo context for response-only scenarios.
func newBareContext(engine *echo.Echo) (*echo.Context, *httptest.ResponseRecorder) {
	httpRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	httpContext := engine.NewContext(httpRequest, recorder)
	return httpContext, recorder
}

func TestBindValidJSONRoundTrip(t *testing.T) {
	engine := echo.New()
	httpContext, _ := newPostContext(engine, `{"name":"Ada"}`)
	bound, bindFailure := httpx.Bind[personPayload](httpContext)
	if bindFailure != nil {
		t.Fatalf("Bind valid JSON: unexpected failure: %v", bindFailure)
	}
	if bound.Name != "Ada" {
		t.Fatalf("Bind valid JSON: Name = %q, want %q", bound.Name, "Ada")
	}
}

func TestBindInvalidJSONNamesType(t *testing.T) {
	engine := echo.New()
	httpContext, _ := newPostContext(engine, `{"name":}`)
	_, bindFailure := httpx.Bind[personPayload](httpContext)
	if bindFailure == nil {
		t.Fatal("Bind invalid JSON: expected failure, got nil")
	}
	if !strings.Contains(bindFailure.Error(), "personPayload") {
		t.Fatalf("Bind invalid JSON: failure %q does not name personPayload", bindFailure.Error())
	}
}

func TestJSONWritesStatusAndEnvelope(t *testing.T) {
	engine := echo.New()
	httpContext, recorder := newBareContext(engine)
	if writeFailure := httpx.JSON(httpContext, http.StatusCreated, personPayload{Name: "Ada"}); writeFailure != nil {
		t.Fatalf("JSON: unexpected failure: %v", writeFailure)
	}
	if recorder.Code != http.StatusCreated {
		t.Fatalf("JSON: status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	var envelope httpx.Response[personPayload]
	if decodeFailure := json.NewDecoder(recorder.Body).Decode(&envelope); decodeFailure != nil {
		t.Fatalf("JSON: decode failure: %v", decodeFailure)
	}
	if envelope.Data.Name != "Ada" {
		t.Fatalf("JSON: Data.Name = %q, want %q", envelope.Data.Name, "Ada")
	}
	if envelope.Error != "" {
		t.Fatalf("JSON: Error = %q, want empty", envelope.Error)
	}
}

func TestFailWritesStatusAndError(t *testing.T) {
	engine := echo.New()
	httpContext, recorder := newBareContext(engine)
	if writeFailure := httpx.Fail(httpContext, http.StatusBadRequest, errors.New("bad input")); writeFailure != nil {
		t.Fatalf("Fail: unexpected failure: %v", writeFailure)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("Fail: status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	var envelope httpx.Response[personPayload]
	if decodeFailure := json.NewDecoder(recorder.Body).Decode(&envelope); decodeFailure != nil {
		t.Fatalf("Fail: decode failure: %v", decodeFailure)
	}
	if envelope.Error != "bad input" {
		t.Fatalf("Fail: Error = %q, want %q", envelope.Error, "bad input")
	}
	if envelope.Data.Name != "" {
		t.Fatalf("Fail: Data.Name = %q, want empty", envelope.Data.Name)
	}
}

func TestFailNilNeverEmptyError(t *testing.T) {
	engine := echo.New()
	httpContext, recorder := newBareContext(engine)
	if writeFailure := httpx.Fail(httpContext, http.StatusInternalServerError, nil); writeFailure != nil {
		t.Fatalf("Fail(nil): unexpected failure: %v", writeFailure)
	}
	var envelope httpx.Response[personPayload]
	if decodeFailure := json.NewDecoder(recorder.Body).Decode(&envelope); decodeFailure != nil {
		t.Fatalf("Fail(nil): decode failure: %v", decodeFailure)
	}
	if envelope.Error == "" {
		t.Fatal("Fail(nil): Error is empty, want descriptive message")
	}
}

func TestListResponseCarriesItemsPageTotal(t *testing.T) {
	listing := httpx.ListResponse[personPayload]{
		Items: []personPayload{{Name: "Ada"}, {Name: "Grace"}},
		Page:  repo.Page{Number: 2, Size: 2},
		Total: 5,
	}
	encoded, encodeFailure := json.Marshal(listing)
	if encodeFailure != nil {
		t.Fatalf("ListResponse marshal: unexpected failure: %v", encodeFailure)
	}
	var decoded httpx.ListResponse[personPayload]
	if decodeFailure := json.Unmarshal(encoded, &decoded); decodeFailure != nil {
		t.Fatalf("ListResponse unmarshal: unexpected failure: %v", decodeFailure)
	}
	if len(decoded.Items) != 2 {
		t.Fatalf("ListResponse len(Items) = %d, want 2", len(decoded.Items))
	}
	if decoded.Items[0].Name != "Ada" || decoded.Items[1].Name != "Grace" {
		t.Fatalf("ListResponse Items = %+v, want Ada and Grace", decoded.Items)
	}
	if decoded.Page.Number != 2 || decoded.Page.Size != 2 {
		t.Fatalf("ListResponse Page = %+v, want {Number:2 Size:2}", decoded.Page)
	}
	if decoded.Total != 5 {
		t.Fatalf("ListResponse Total = %d, want 5", decoded.Total)
	}
}
