package collector

import (
	"github.com/go-kit/log"
)

// NodeInfoResponse type
type NodeInfoResponse struct {
	Host        string `json:"host"`
	Version     string `json:"version"`
	HTTPAddress string `json:"http_address"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Pipeline    struct {
		Workers               int  `json:"workers"`
		BatchSize             int  `json:"batch_size"`
		BatchDelay            int  `json:"batch_delay"`
		ConfigReloadAutomatic bool `json:"config_reload_automatic"`
		ConfigReloadInterval  int  `json:"config_reload_interval"`
	} `json:"pipeline"`
	Os struct {
		Name                string `json:"name"`
		Arch                string `json:"arch"`
		Version             string `json:"version"`
		AvailableProcessors int    `json:"available_processors"`
	} `json:"os"`
	Jvm struct {
		Pid               int    `json:"pid"`
		Version           string `json:"version"`
		VMName            string `json:"vm_name"`
		VMVersion         string `json:"vm_version"`
		VMVendor          string `json:"vm_vendor"`
		StartTimeInMillis int64  `json:"start_time_in_millis"`
		Mem               struct {
			HeapInitInBytes    int `json:"heap_init_in_bytes"`
			HeapMaxInBytes     int `json:"heap_max_in_bytes"`
			NonHeapInitInBytes int `json:"non_heap_init_in_bytes"`
			NonHeapMaxInBytes  int `json:"non_heap_max_in_bytes"`
		} `json:"mem"`
		GcCollectors []string `json:"gc_collectors"`
	} `json:"jvm"`
}

// NodeInfo function
func NodeInfo(endpoint string, logger log.Logger) (NodeInfoResponse, error) {
	var response NodeInfoResponse

	handler := &HTTPHandler{
		Endpoint: endpoint + "/_node",
	}

	err := getMetrics(handler, &response, logger)

	return response, err
}
