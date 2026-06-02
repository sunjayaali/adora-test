package mockcarrier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/mroth/weightedrand/v3"
)

func NewServer() (*httptest.Server, error) {
	chooser, err := weightedrand.NewChooser(
		weightedrand.NewChoice("active", 85),
		weightedrand.NewChoice("inactive", 10),
		weightedrand.NewChoice("api_error", 5),
	)
	if err != nil {
		return nil, err
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": chooser.Pick(),
		})
	})), nil
}
