package osrm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	osrmgenerated "github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/infrastructure/osrm/generated"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	ErrUnavailable     = errors.New("osrm unavailable")
	ErrNoRouteFound    = errors.New("osrm no route found")
	ErrInvalidResponse = errors.New("osrm invalid response")
	errIncompleteAuth  = errors.New("both auth header name and value must be set")
)

type RouteResult struct {
	DistanceMeters float64
	Duration       time.Duration
	Geometry       string
}

type Client struct {
	api *osrmgenerated.ClientWithResponses
}

func NewClient(baseURL string, timeout time.Duration, authHeaderName, authHeaderValue string) (*Client, error) {
	httpClient := &http.Client{
		Timeout: timeout,
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
			otelhttp.WithSpanNameFormatter(func(_ string, req *http.Request) string {
				return "osrm.http " + req.Method + " " + req.URL.Path
			}),
		),
	}

	options := []osrmgenerated.ClientOption{
		osrmgenerated.WithHTTPClient(httpClient),
	}

	if authHeaderName != "" || authHeaderValue != "" {
		if authHeaderName == "" || authHeaderValue == "" {
			return nil, errIncompleteAuth
		}

		options = append(options, osrmgenerated.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			req.Header.Set(authHeaderName, authHeaderValue)
			return nil
		}))
	}

	apiClient, err := osrmgenerated.NewClientWithResponses(baseURL, options...)
	if err != nil {
		return nil, fmt.Errorf("create OSRM client: %w", err)
	}

	return &Client{
		api: apiClient,
	}, nil
}

func (c *Client) Route(ctx context.Context, originCoordinates, destinationCoordinates string) (RouteResult, error) {
	coordinates := originCoordinates + ";" + destinationCoordinates
	overview := osrmgenerated.RouteParamsOverviewFull
	geometries := osrmgenerated.RouteParamsGeometriesPolyline

	response, err := c.api.RouteWithResponse(
		ctx,
		"driving",
		coordinates,
		&osrmgenerated.RouteParams{
			Overview:   &overview,
			Geometries: &geometries,
		},
	)
	if err != nil {
		return RouteResult{}, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}

	if response == nil {
		return RouteResult{}, fmt.Errorf("%w: empty route response", ErrInvalidResponse)
	}

	if response.StatusCode() != http.StatusOK {
		return RouteResult{}, fmt.Errorf("%w: status code %d", ErrUnavailable, response.StatusCode())
	}

	if response.JSON200 == nil || response.JSON200.Routes == nil || len(*response.JSON200.Routes) == 0 {
		return RouteResult{}, fmt.Errorf("%w: empty routes", ErrNoRouteFound)
	}

	if response.JSON200.Code != "Ok" {
		return RouteResult{}, fmt.Errorf("%w: code %s", ErrNoRouteFound, response.JSON200.Code)
	}

	route := (*response.JSON200.Routes)[0]
	if route.Distance == nil || route.Duration == nil || route.Geometry == nil {
		return RouteResult{}, fmt.Errorf("%w: route payload is incomplete", ErrInvalidResponse)
	}

	geometry, err := route.Geometry.AsGeometry0()
	if err != nil || geometry == "" {
		return RouteResult{}, fmt.Errorf("%w: route geometry is empty", ErrInvalidResponse)
	}

	return RouteResult{
		DistanceMeters: *route.Distance,
		Duration:       time.Duration(*route.Duration) * time.Second,
		Geometry:       geometry,
	}, nil
}
