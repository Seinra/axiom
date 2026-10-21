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

func TestDecodeValidJSONRoundTrip(t *testing.T) {
	engine := echo.New()
	httpContext, _ := newPostContext(engine, `{"name":"Ada"}`)
	decoded, decodeFailure := httpx.Decode[personPayload](httpContext)
	if decodeFailure != nil {
		t.Fatalf("Decode valid JSON: unexpected failure: %v", decodeFailure)
	}
	if decoded.Name != "Ada" {
		t.Fatalf("Decode valid JSON: Name = %q, want %q", decoded.Name, "Ada")
	}
}

func TestDecodeInvalidJSONNamesType(t *testing.T) {
	engine := echo.New()
	httpContext, _ := newPostContext(engine, `{"name":}`)
	_, decodeFailure := httpx.Decode[personPayload](httpContext)
	if decodeFailure == nil {
		t.Fatal("Decode invalid JSON: expected failure, got nil")
	}
	if !strings.Contains(decodeFailure.Error(), "personPayload") {
		t.Fatalf("Decode invalid JSON: failure %q does not name personPayload", decodeFailure.Error())
	}
}

func TestWriteJSONStatusAndEnvelope(t *testing.T) {
	engine := echo.New()
	httpContext, recorder := newBareContext(engine)
	if writeFailure := httpx.WriteJSON(httpContext, http.StatusCreated, personPayload{Name: "Ada"}); writeFailure != nil {
		t.Fatalf("WriteJSON: unexpected failure: %v", writeFailure)
	}
	if recorder.Code != http.StatusCreated {
		t.Fatalf("WriteJSON: status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	var envelope httpx.Response[personPayload]
	if decodeFailure := json.NewDecoder(recorder.Body).Decode(&envelope); decodeFailure != nil {
		t.Fatalf("WriteJSON: decode failure: %v", decodeFailure)
	}
	if envelope.Data.Name != "Ada" {
		t.Fatalf("WriteJSON: Data.Name = %q, want %q", envelope.Data.Name, "Ada")
	}
	if envelope.Error != "" {
		t.Fatalf("WriteJSON: Error = %q, want empty", envelope.Error)
	}
}

func TestWriteErrorWritesStatusAndError(t *testing.T) {
	engine := echo.New()
	httpContext, recorder := newBareContext(engine)
	if writeFailure := httpx.WriteError(httpContext, http.StatusBadRequest, errors.New("bad input")); writeFailure != nil {
		t.Fatalf("WriteError: unexpected failure: %v", writeFailure)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("WriteError: status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	var envelope httpx.Response[personPayload]
	if decodeFailure := json.NewDecoder(recorder.Body).Decode(&envelope); decodeFailure != nil {
		t.Fatalf("WriteError: decode failure: %v", decodeFailure)
	}
	if envelope.Error != "bad input" {
		t.Fatalf("WriteError: Error = %q, want %q", envelope.Error, "bad input")
	}
	if envelope.Data.Name != "" {
		t.Fatalf("WriteError: Data.Name = %q, want empty", envelope.Data.Name)
	}
}

func TestWriteErrorNilNeverEmptyError(t *testing.T) {
	engine := echo.New()
	httpContext, recorder := newBareContext(engine)
	if writeFailure := httpx.WriteError(httpContext, http.StatusInternalServerError, nil); writeFailure != nil {
		t.Fatalf("WriteError(nil): unexpected failure: %v", writeFailure)
	}
	var envelope httpx.Response[personPayload]
	if decodeFailure := json.NewDecoder(recorder.Body).Decode(&envelope); decodeFailure != nil {
		t.Fatalf("WriteError(nil): decode failure: %v", decodeFailure)
	}
	if envelope.Error == "" {
		t.Fatal("WriteError(nil): Error is empty, want descriptive message")
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

// TestResponseStructZeroDataIsRetained pins the envelope contract: a zero
// struct Body still emits the data key, so struct Bodies never collapse.
func TestResponseStructZeroDataIsRetained(t *testing.T) {
	encoded, encodeFailure := json.Marshal(httpx.Response[personPayload]{Data: personPayload{}})
	if encodeFailure != nil {
		t.Fatalf("marshal struct-zero envelope: unexpected failure: %v", encodeFailure)
	}
	if !strings.Contains(string(encoded), `"data"`) {
		t.Fatalf("struct-zero envelope = %s, want a data key", string(encoded))
	}
}

// TestResponseScalarZeroDataOmitsKey pins the envelope contract: a scalar
// zero Body (0, "", false) encodes with no data key under omitempty.
func TestResponseScalarZeroDataOmitsKey(t *testing.T) {
	encoded, encodeFailure := json.Marshal(httpx.Response[int]{Data: 0})
	if encodeFailure != nil {
		t.Fatalf("marshal int-zero envelope: unexpected failure: %v", encodeFailure)
	}
	if string(encoded) != "{}" {
		t.Fatalf("int-zero envelope = %s, want {} (keyless zero value)", string(encoded))
	}
}

// TestWriteErrorEnvelopeOmitsDataKey pins that failure envelopes carry no data
// key: a nil any Data is omitted, never encoded as null.
func TestWriteErrorEnvelopeOmitsDataKey(t *testing.T) {
	encoded, encodeFailure := json.Marshal(httpx.Response[any]{Error: "bad input"})
	if encodeFailure != nil {
		t.Fatalf("marshal fail envelope: unexpected failure: %v", encodeFailure)
	}
	if string(encoded) != `{"error":"bad input"}` {
		t.Fatalf("fail envelope = %s, want exactly an error key", string(encoded))
	}
}

func TestValidationWriteErrorJoinsMultipleFailures(t *testing.T) {
	engine := echo.New()
	httpContext, recorder := newBareContext(engine)
	failures := []error{errors.New("name is required"), errors.New("age must be positive")}
	if writeFailure := httpx.ValidationWriteError(httpContext, failures); writeFailure != nil {
		t.Fatalf("ValidationWriteError multi: unexpected failure: %v", writeFailure)
	}
	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ValidationWriteError multi: status = %d, want %d", recorder.Code, http.StatusUnprocessableEntity)
	}
	var envelope httpx.Response[personPayload]
	if decodeFailure := json.NewDecoder(recorder.Body).Decode(&envelope); decodeFailure != nil {
		t.Fatalf("ValidationWriteError multi: decode failure: %v", decodeFailure)
	}
	if envelope.Error != "name is required; age must be positive" {
		t.Fatalf("ValidationWriteError multi: Error = %q, want joined messages", envelope.Error)
	}
}

func TestValidationWriteErrorSingleFailure(t *testing.T) {
	engine := echo.New()
	httpContext, recorder := newBareContext(engine)
	failures := []error{errors.New("name is required")}
	if writeFailure := httpx.ValidationWriteError(httpContext, failures); writeFailure != nil {
		t.Fatalf("ValidationWriteError single: unexpected failure: %v", writeFailure)
	}
	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ValidationWriteError single: status = %d, want %d", recorder.Code, http.StatusUnprocessableEntity)
	}
	var envelope httpx.Response[personPayload]
	if decodeFailure := json.NewDecoder(recorder.Body).Decode(&envelope); decodeFailure != nil {
		t.Fatalf("ValidationWriteError single: decode failure: %v", decodeFailure)
	}
	if envelope.Error != "name is required" {
		t.Fatalf("ValidationWriteError single: Error = %q, want %q", envelope.Error, "name is required")
	}
}

func TestValidationWriteErrorEmptySliceGuard(t *testing.T) {
	for _, scenario := range []struct {
		scenarioName string
		failures     []error
	}{
		{scenarioName: "nil failures", failures: nil},
		{scenarioName: "empty failures", failures: []error{}},
	} {
		engine := echo.New()
		httpContext, recorder := newBareContext(engine)
		if writeFailure := httpx.ValidationWriteError(httpContext, scenario.failures); writeFailure != nil {
			t.Fatalf("ValidationWriteError %s: unexpected failure: %v", scenario.scenarioName, writeFailure)
		}
		if recorder.Code != http.StatusUnprocessableEntity {
			t.Fatalf("ValidationWriteError %s: status = %d, want %d", scenario.scenarioName, recorder.Code, http.StatusUnprocessableEntity)
		}
		var envelope httpx.Response[personPayload]
		if decodeFailure := json.NewDecoder(recorder.Body).Decode(&envelope); decodeFailure != nil {
			t.Fatalf("ValidationWriteError %s: decode failure: %v", scenario.scenarioName, decodeFailure)
		}
		if envelope.Error == "" {
			t.Fatalf("ValidationWriteError %s: Error is empty, want descriptive message", scenario.scenarioName)
		}
	}
}

// TestValidationWriteErrorSkipsNilEntries pins that nil entries inside failures
// are skipped instead of panicking or contributing empty messages.
func TestValidationWriteErrorSkipsNilEntries(t *testing.T) {
	engine := echo.New()
	httpContext, recorder := newBareContext(engine)
	failures := []error{nil, errors.New("name is required"), nil}
	if writeFailure := httpx.ValidationWriteError(httpContext, failures); writeFailure != nil {
		t.Fatalf("ValidationWriteError nil entries: unexpected failure: %v", writeFailure)
	}
	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ValidationWriteError nil entries: status = %d, want %d", recorder.Code, http.StatusUnprocessableEntity)
	}
	var envelope httpx.Response[personPayload]
	if decodeFailure := json.NewDecoder(recorder.Body).Decode(&envelope); decodeFailure != nil {
		t.Fatalf("ValidationWriteError nil entries: decode failure: %v", decodeFailure)
	}
	if envelope.Error != "name is required" {
		t.Fatalf("ValidationWriteError nil entries: Error = %q, want %q", envelope.Error, "name is required")
	}
}

// TestWriteJSONScalarZeroEndToEnd pins the envelope contract over the wire:
// a scalar zero Body encodes keyless ({}), decodes back as the zero value,
// and never reads as failure.
func TestWriteJSONScalarZeroEndToEnd(t *testing.T) {
	engine := echo.New()
	httpContext, recorder := newBareContext(engine)
	if writeFailure := httpx.WriteJSON(httpContext, http.StatusOK, 0); writeFailure != nil {
		t.Fatalf("WriteJSON scalar zero: unexpected failure: %v", writeFailure)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("WriteJSON scalar zero: status = %d, want %d", recorder.Code, http.StatusOK)
	}
	rawBody := strings.TrimSpace(recorder.Body.String())
	if rawBody != "{}" {
		t.Fatalf("WriteJSON scalar zero: body = %s, want {} (keyless zero value)", rawBody)
	}
	var envelope httpx.Response[int]
	if decodeFailure := json.NewDecoder(strings.NewReader(rawBody)).Decode(&envelope); decodeFailure != nil {
		t.Fatalf("WriteJSON scalar zero: decode failure: %v", decodeFailure)
	}
	if envelope.Data != 0 {
		t.Fatalf("WriteJSON scalar zero: Data = %d, want 0", envelope.Data)
	}
	if envelope.Error != "" {
		t.Fatalf("WriteJSON scalar zero: Error = %q, want empty (zero value reads as success)", envelope.Error)
	}
}

// TestValidationWriteErrorAllNilGuard pins the guard message: an all-nil
// failures slice carries no message, so the writer reports the descriptive
// guard instead of an empty envelope.
func TestValidationWriteErrorAllNilGuard(t *testing.T) {
	engine := echo.New()
	httpContext, recorder := newBareContext(engine)
	if writeFailure := httpx.ValidationWriteError(httpContext, []error{nil, nil}); writeFailure != nil {
		t.Fatalf("ValidationWriteError all-nil: unexpected failure: %v", writeFailure)
	}
	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ValidationWriteError all-nil: status = %d, want %d", recorder.Code, http.StatusUnprocessableEntity)
	}
	var envelope httpx.Response[personPayload]
	if decodeFailure := json.NewDecoder(recorder.Body).Decode(&envelope); decodeFailure != nil {
		t.Fatalf("ValidationWriteError all-nil: decode failure: %v", decodeFailure)
	}
	if !strings.Contains(envelope.Error, "needs at least one failure") {
		t.Fatalf("ValidationWriteError all-nil: Error = %q, want guard message", envelope.Error)
	}
}

// TestFieldFailureMatchesSentinel pins per-field matchability: each field
// failure wraps ErrValidationFailed, so errors.Is matches the sentinel and
// errors.As recovers the field in-process (the 422 envelope keeps carrying
// only text, where the stdlib contract cannot reach).
func TestFieldFailureMatchesSentinel(t *testing.T) {
	failure := httpx.FieldFailure{Field: "name", Reason: "name is required"}
	if !errors.Is(failure, httpx.ErrValidationFailed) {
		t.Fatalf("errors.Is(%v, ErrValidationFailed) = false, want true", failure)
	}
	var matched httpx.FieldFailure
	if !errors.As(failure, &matched) {
		t.Fatalf("errors.As(%v) = false, want FieldFailure match", failure)
	}
	if matched.Field != "name" || matched.Reason != "name is required" {
		t.Fatalf("errors.As matched = %+v, want name/name is required", matched)
	}
}

// TestFieldFailureWrapsCause pins that an underlying cause stays matchable
// alongside the sentinel through the multi-unwrap chain.
func TestFieldFailureWrapsCause(t *testing.T) {
	cause := errors.New("too short")
	failure := httpx.FieldFailure{Field: "name", Reason: "name is too short", Cause: cause}
	if !errors.Is(failure, httpx.ErrValidationFailed) {
		t.Fatalf("errors.Is(%v, ErrValidationFailed) = false, want true", failure)
	}
	if !errors.Is(failure, cause) {
		t.Fatalf("errors.Is(%v, cause) = false, want true", failure)
	}
}

// TestValidationWriteErrorKeepsShapeForFieldFailures pins the HTTP contract:
// FieldFailure inputs still write 422 with "; "-joined messages, so the
// envelope shape never changes while failures stay matchable caller-side.
func TestValidationWriteErrorKeepsShapeForFieldFailures(t *testing.T) {
	engine := echo.New()
	httpContext, recorder := newBareContext(engine)
	failures := []error{
		httpx.FieldFailure{Field: "name", Reason: "name is required"},
		httpx.FieldFailure{Field: "age", Reason: "age must be positive"},
	}
	if writeFailure := httpx.ValidationWriteError(httpContext, failures); writeFailure != nil {
		t.Fatalf("ValidationWriteError field failures: unexpected failure: %v", writeFailure)
	}
	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ValidationWriteError field failures: status = %d, want %d", recorder.Code, http.StatusUnprocessableEntity)
	}
	var envelope httpx.Response[personPayload]
	if decodeFailure := json.NewDecoder(recorder.Body).Decode(&envelope); decodeFailure != nil {
		t.Fatalf("ValidationWriteError field failures: decode failure: %v", decodeFailure)
	}
	want := failures[0].Error() + "; " + failures[1].Error()
	if envelope.Error != want {
		t.Fatalf("ValidationWriteError field failures: Error = %q, want %q", envelope.Error, want)
	}
}
