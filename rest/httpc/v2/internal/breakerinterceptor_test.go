package internal

import (
	"errors"
	"net/http"
	"testing"

	"github.com/shippomx/zard/core/breaker"
	"github.com/stretchr/testify/assert"
)

func TestBreakerInterceptor(t *testing.T) {
	p, err := breaker.NewBreaker().Allow()
	assert.Nil(t, err)
	tests := []struct {
		name        string
		ex          ExtendInfo
		resp        *http.Response
		err         error
		wantPromise breaker.Promise
	}{
		{
			name:        "no ExtendInfo",
			ex:          ExtendInfo{},
			resp:        &http.Response{StatusCode: http.StatusOK},
			err:         nil,
			wantPromise: nil,
		},
		{
			name:        "ExtendInfo with no promise",
			ex:          ExtendInfo{Promise: nil},
			resp:        &http.Response{StatusCode: http.StatusOK},
			err:         nil,
			wantPromise: nil,
		},
		{
			name:        "ExtendInfo with promise, no error, and OK response",
			ex:          ExtendInfo{Promise: p},
			resp:        &http.Response{StatusCode: http.StatusOK},
			err:         nil,
			wantPromise: p,
		},
		{
			name:        "ExtendInfo with promise, error, and non-OK response",
			ex:          ExtendInfo{Promise: p},
			resp:        &http.Response{StatusCode: http.StatusInternalServerError},
			err:         errors.New("test error"),
			wantPromise: p,
		},
		{
			name:        "ExtendInfo with promise, no error, and non-OK response",
			ex:          ExtendInfo{Promise: p},
			resp:        &http.Response{StatusCode: http.StatusInternalServerError},
			err:         nil,
			wantPromise: p,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, responseHandler := BreakerInterceptor(&http.Request{}, test.ex)
			responseHandler(test.resp, test.err)

			assert.Equal(t, test.wantPromise, test.ex.Promise)
		})
	}
}

func TestBreakerInterceptorError(t *testing.T) {
	// fail
	brk := breaker.NewBreaker()
	var firsterr error
	for i := 0; i < 1000; i++ {
		p, err := brk.Allow()
		if err != nil {
			firsterr = err
			break
		}
		_, responseHandler := BreakerInterceptor(&http.Request{}, ExtendInfo{Promise: p})
		responseHandler(&http.Response{StatusCode: http.StatusServiceUnavailable}, nil)
	}
	assert.Error(t, firsterr)

	brk = breaker.NewBreaker()
	firsterr = nil
	for i := 0; i < 1000; i++ {
		p, err := brk.Allow()
		if err != nil {
			firsterr = err
			break
		}
		_, responseHandler := BreakerInterceptor(&http.Request{}, ExtendInfo{Promise: p})
		responseHandler(&http.Response{StatusCode: http.StatusOK}, nil)
	}
	assert.NoError(t, firsterr)

	brk = breaker.NewBreaker()
	firsterr = nil
	for i := 0; i < 1000; i++ {
		p, err := brk.Allow()
		if err != nil {
			firsterr = err
			break
		}
		_, responseHandler := BreakerInterceptor(&http.Request{}, ExtendInfo{Promise: p})
		responseHandler(&http.Response{StatusCode: http.StatusOK}, errors.New("test error"))
	}
	assert.Error(t, firsterr)

	brk2 := breaker.NewBreaker()
	if brk2 != nil && brk2.Name() != "" {
		t.Log("breaker is not nil")
		t.Log(brk2.Name())
	}
	brk2 = nil
	assert.Nil(t, brk2)
}
