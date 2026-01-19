package mock

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorScenario_Apply(t *testing.T) {
	scenario := &ErrorScenario{
		StatusCode: http.StatusNotFound,
		RC:         "error",
		Message:    "Not found",
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/test", nil)

	applied := scenario.Apply(w, r)
	if !applied {
		t.Error("Apply() = false, want true")
	}

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestErrorScenario_ApplyWithPath(t *testing.T) {
	scenario := &ErrorScenario{
		Path:       "/api/specific",
		StatusCode: http.StatusBadRequest,
		RC:         "error",
		Message:    "Bad request",
	}

	// Request to matching path
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("GET", "/api/specific", nil)

	if !scenario.Apply(w1, r1) {
		t.Error("Should apply to matching path")
	}

	// Request to non-matching path
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/api/other", nil)

	if scenario.Apply(w2, r2) {
		t.Error("Should not apply to non-matching path")
	}
}

func TestPredefinedScenarios(t *testing.T) {
	scenarios := map[string]Scenario{
		"SessionExpired": ScenarioSessionExpired,
		"CSRFFailure":    ScenarioCSRFFailure,
		"RateLimit":      ScenarioRateLimit,
		"ServerError":    ScenarioServerError,
	}

	for name, scenario := range scenarios {
		t.Run(name, func(t *testing.T) {
			if scenario == nil {
				t.Errorf("%s scenario is nil", name)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/api/test", nil)

			if !scenario.Apply(w, r) {
				t.Errorf("%s scenario did not apply", name)
			}

			if w.Code < 400 {
				t.Errorf("%s scenario status = %d, want >= 400", name, w.Code)
			}
		})
	}
}
