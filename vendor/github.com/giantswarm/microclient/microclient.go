// microclient provides opinionated helpers for REST clients.
package microclient

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-resty/resty"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/giantswarm/microerror"
)

const (
	prometheusNamespace = "microclient"
)

var (
	request = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: prometheusNamespace,
			Name:      "request",
			Help:      "Histogram for requests.",
		},
		[]string{"url"},
	)

	requestError = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusNamespace,
			Name:      "request_error_count",
			Help:      "Counters for errors during requests.",
		},
		[]string{"url", "error_type", "error_message"},
	)

	responseStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusNamespace,
			Name:      "response_status_count",
			Help:      "Counters for response status codes.",
		},
		[]string{"url", "code"},
	)
)

func init() {
	prometheus.MustRegister(request)
	prometheus.MustRegister(requestError)
	prometheus.MustRegister(responseStatus)
}

// Do takes a resty request function and executes it, adding metrics.
func Do(ctx context.Context, requestFunc func(string) (*resty.Response, error), url string) (*resty.Response, error) {
	requestTimer := prometheus.NewTimer(request.WithLabelValues(url))
	defer requestTimer.ObserveDuration()

	response, err := requestFunc(url)
	if err != nil {
		errorType := fmt.Sprintf("%T", err)
		errorMessage := err.Error()

		requestError.WithLabelValues(url, errorType, errorMessage).Inc()
		return nil, microerror.Mask(err)
	}

	responseStatus.WithLabelValues(url, strconv.Itoa(response.StatusCode())).Inc()

	return response, nil
}
