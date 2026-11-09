package linkcheck

import (
	"net/http"

	"github.com/amila-t-kumarasekara/web-analyzer/pkg/consts"
)

// NewHTTPClient returns a client suited to many concurrent checks against few hosts.
func NewHTTPClient() *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.MaxIdleConns = consts.DefaultLinkWorkers * 4
	tr.MaxIdleConnsPerHost = consts.DefaultLinkWorkers
	return &http.Client{
		Timeout:   consts.LinkHTTPTimeout,
		Transport: tr,
	}
}
