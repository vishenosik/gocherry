package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/vishenosik/gocherry/pkg/errors"
	"github.com/vishenosik/gocherry/pkg/versions"
)

const (
	VersionParam  = "v"
	VersionHeader = "X-Api-Version"
)

// apiVersionKey is an unexported type for context keys to prevent collisions
type apiVersionKey struct{}

// ParseFirst parses the version from request (query param or header)
func ParseDotVersion(r *http.Request) (versions.Interface, error) {

	versionStr := r.URL.Query().Get(VersionParam)
	if versionStr == "" {
		versionStr = r.Header.Get(VersionHeader)
	}

	if versionStr == "" {
		return nil, errors.New("api version is not provided")
	}

	version, err := versions.ParseDotVersion(versionStr)
	if err != nil {
		return nil, errors.Wrap(err, "invalid version format")
	}

	return version, nil
}

// WithContext adds the APIVersion to the context
func WithContext(ctx context.Context, v versions.Interface) context.Context {
	return context.WithValue(ctx, apiVersionKey{}, v)
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

type HandlersMap = map[string]http.Handler

func ApiVersionHandler(handlers HandlersMap) http.Handler {
	latest := latestVersion(handlers)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		handler := func() http.Handler {

			version, err := ParseDotVersion(r)
			if err != nil {
				return handlers[latest]
			}

			handler, ok := handlers[version.String()]
			if !ok {
				return handlers[latest]
			}
			return handler
		}

		handler().ServeHTTP(w, r)
	})
}

func latestVersion(handlers HandlersMap) string {

	if len(handlers) == 0 {
		panic("handlers map can't be nil")
	}

	mulerr := new(errors.MultiError)
	_versions_ := make([]versions.DotVersion, 0, len(handlers))
	for v, handler := range handlers {
		if handler == nil {
			mulerr.Append(fmt.Errorf("handler with version %s can't be nil", v))
		}

		version, err := versions.ParseDotVersion(v)
		if err != nil {
			mulerr.Append(errors.Wrapf(err, "version %s can't be parsed", v))
		}

		_versions_ = append(_versions_, version)

	}

	if err := mulerr.ErrorOrNil(); err != nil {
		panic(err)
	}

	// latest version listed in handlers map
	return versions.LatestDotVersion(_versions_...).String()
}
