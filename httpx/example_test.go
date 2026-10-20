package httpx_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/v5"

	"github.com/seinra/axiom/httpx"
)

// ExampleJSON shows the single-envelope success reply.
func ExampleJSON() {
	engine := echo.New()
	httpRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	httpContext := engine.NewContext(httpRequest, recorder)
	_ = httpx.JSON(httpContext, http.StatusOK, "ready")
	fmt.Println(recorder.Body.String())
	// Output: {"data":"ready"}
}
