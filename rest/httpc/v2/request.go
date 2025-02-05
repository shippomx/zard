package httpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	nurl "net/url"
	"strings"

	"github.com/shippomx/zard/core/lang"
	"github.com/shippomx/zard/core/mapping"
	"github.com/shippomx/zard/rest/internal/header"
)

func buildRequest(ctx context.Context, method, url string, data any) (*http.Request, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	var val map[string]map[string]any
	if data != nil {
		val, err = mapping.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	if err := fillPath(u, val[pathKey]); err != nil {
		return nil, err
	}

	var reader io.Reader
	jsonVars, hasJSONBody := val[jsonKey]
	if hasJSONBody {
		if method == http.MethodGet {
			return nil, ErrGetWithBody
		}

		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		if err := enc.Encode(jsonVars); err != nil {
			return nil, err
		}

		reader = &buf
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = buildFormQuery(u, val[formKey])
	fillHeader(req, val[headerKey])
	if hasJSONBody {
		req.Header.Set(header.ContentType, header.JsonContentType)
	}

	return req, nil
}

func fillHeader(r *http.Request, val map[string]any) {
	for k, v := range val {
		r.Header.Add(k, fmt.Sprint(v))
	}
}

func fillPath(u *nurl.URL, val map[string]any) error {
	used := make(map[string]lang.PlaceholderType)
	fields := strings.Split(u.Path, slash)

	for i := range fields {
		field := fields[i]
		if len(field) > 0 && field[0] == colon {
			name := field[1:]
			ival, ok := val[name]
			if !ok {
				return fmt.Errorf("missing path variable %q", name)
			}
			value := fmt.Sprint(ival)
			if len(value) == 0 {
				return fmt.Errorf("empty path variable %q", name)
			}
			fields[i] = value
			used[name] = lang.Placeholder
		}
	}

	if len(val) != len(used) {
		for key := range used {
			delete(val, key)
		}

		var unused []string
		for key := range val {
			unused = append(unused, key)
		}

		return fmt.Errorf("more path variables are provided: %q", strings.Join(unused, ", "))
	}

	u.Path = strings.Join(fields, slash)
	return nil
}

func buildFormQuery(u *nurl.URL, val map[string]any) string {
	query := u.Query()
	for k, v := range val {
		query.Add(k, fmt.Sprint(v))
	}

	return query.Encode()
}
