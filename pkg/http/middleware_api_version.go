package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/vishenosik/gocherry/pkg/versions"
)

const (
	VersionParam  = "v"
	VersionHeader = "X-API-Version"
)

var (
	ErrFormat = errors.New("version must be in format MAJOR.MINOR")
)

// VersionHandler defines the interface for version handling
type VersionHandler interface {
	ParseRequest(r *http.Request) error
	WithContext(ctx context.Context) context.Context
}

type Middleware = func(http.Handler) http.Handler

type Handler = func(handlers HandlersMap) http.HandlerFunc

// apiVersionKey is an unexported type for context keys to prevent collisions
type apiVersionKey struct{}

// APIVersion represents a parsed semantic version
type APIVersion struct {
	Version versions.Interface
	param   string
	header  string
	minimal versions.Interface
	current versions.Interface
}

type ApiVersionOption = func(*APIVersion)

func defaultApiVersion(version versions.Interface) *APIVersion {
	return &APIVersion{
		Version: version,
		minimal: version,
		current: version,
		param:   VersionParam,
		header:  VersionHeader,
	}
}

func newApiVersion(version versions.Interface, opts ...ApiVersionOption) (*APIVersion, error) {
	if version == nil {
		return nil, errors.New("version interface is nil")
	}
	av := defaultApiVersion(version)
	for _, opt := range opts {
		opt(av)
	}
	return av, nil
}

func Min(version string) ApiVersionOption {
	return func(av *APIVersion) {
		min, err := av.Version.Parse_(version)
		if err == nil {
			av.minimal = min
		}
	}
}

// ParseFirst parses the version from request (query param or header)
func (av *APIVersion) ParseRequest(r *http.Request) error {

	versionStr := r.URL.Query().Get(VersionParam)
	if versionStr == "" {
		versionStr = r.Header.Get(VersionHeader)
	}

	if versionStr == "" {
		return errors.New("api version is not provided")
	}

	version, err := av.Version.Parse_(versionStr)
	if err != nil {
		return errors.Wrap(err, "invalid version format")
	}

	if !version.In_(av.minimal, av.current) {
		return fmt.Errorf("unsupported version. Min: %s, Max: %s", av.minimal, av.current)
	}

	av.Version = version
	return nil
}

// WithContext adds the APIVersion to the context
func (av *APIVersion) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, apiVersionKey{}, av.Version)
}

// ApiVersionFromContext retrieves the APIVersion from context
func TypedApiVersionFromContext[VersionType versions.Interface](ctx context.Context) (VersionType, error) {
	var empty VersionType
	val := ctx.Value(apiVersionKey{})
	if val == nil {
		return empty, errors.New("no API version in context")
	}

	version, ok := val.(VersionType)
	if !ok {
		return empty, errors.New("invalid API version type in context")
	}
	return version, nil
}

// ApiVersionFromContext retrieves the APIVersion from context
func ApiVersionFromContext(ctx context.Context) (versions.Interface, error) {
	val := ctx.Value(apiVersionKey{})
	if val == nil {
		return nil, errors.New("no API version in context")
	}

	version, ok := val.(versions.Interface)
	if !ok {
		return nil, errors.New("invalid API version type in context")
	}
	return version, nil
}

func ApiVersionMiddlewareHandler(version versions.Interface, opts ...ApiVersionOption) (Middleware, Handler) {
	return ApiVersionMiddleware(version, opts...), ApiVersionHandler
}

func DotVersionMiddlewareHandler(version string, opts ...ApiVersionOption) (Middleware, Handler) {
	return ApiVersionMiddlewareHandler(versions.NewDotVersion(version), opts...)
}

// ApiVersionMiddleware validates the API version from request
func ApiVersionMiddleware(
	version versions.Interface,
	opts ...ApiVersionOption,
) Middleware {

	av, err := newApiVersion(version, opts...)
	if err != nil {
		panic(err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := av.ParseRequest(r); err != nil {
				SendErrors(w, http.StatusBadRequest, errors.Wrap(err, "API version error"))
				return
			}
			ctx := context.WithValue(r.Context(), apiVersionKey{}, av.Version)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type HandlersMap = map[string]http.HandlerFunc

type HandlersSet = func(HandlersMap) http.HandlerFunc

func DotVersionHandler(handlers HandlersMap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiVersion, err := TypedApiVersionFromContext[versions.DotVersion](r.Context())
		if err != nil {
			SendErrors(w, http.StatusBadRequest, err)
			return
		}

		handler, ok := handlers[apiVersion.String()]
		if !ok || handler == nil {
			SendErrors(w, http.StatusNotImplemented, errors.New("api version unsupported"))
			return
		}

		handler(w, r)
	}
}

func ApiVersionHandler(handlers HandlersMap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiVersion, err := ApiVersionFromContext(r.Context())
		if err != nil {
			SendErrors(w, http.StatusBadRequest, err)
			return
		}

		handler, ok := handlers[apiVersion.String()]
		if !ok || handler == nil {
			SendErrors(w, http.StatusNotImplemented, errors.New("api version unsupported"))
			return
		}

		handler(w, r)
	}
}
