package collector

import (
	"encoding/json"
	"net/http"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// HTTPHandler type
type HTTPHandler struct {
	Endpoint string
}

// Get method for HTTPHandler
func (h *HTTPHandler) Get() (http.Response, error) {
	response, err := http.Get(h.Endpoint)
	if err != nil {
		return http.Response{}, err
	}

	return *response, nil
}

// HTTPHandlerInterface interface
type HTTPHandlerInterface interface {
	Get() (http.Response, error)
}

func getMetrics(h HTTPHandlerInterface, target interface{}, logger log.Logger) error {
	response, err := h.Get()
	if err != nil {
		level.Error(logger).Log("msg", "Cannot retrieve metrics", "err", err)
		return err
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			level.Error(logger).Log("msg", "Cannot close response body", "err", err)
		}
	}()

	err = json.NewDecoder(response.Body).Decode(target)
	if err != nil {
		level.Error(logger).Log("msg", "Cannot parse Logstash response json", "err", err)
	}

	return err
}
