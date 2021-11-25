package fcm

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/yusufsyaifudin/ngendika/pkg/logger"
	"go.uber.org/multierr"
)

type RoundTripper struct {
	Base http.RoundTripper
}

var _ http.RoundTripper = (*RoundTripper)(nil)

func (r *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	t0 := time.Now()

	var (
		ctx  = req.Context() // request context
		resp *http.Response  // final response
		err  error           // final error
	)

	var (
		reqBody    []byte
		reqBodyErr error
	)
	if req.Body != nil {
		reqBody, reqBodyErr = ioutil.ReadAll(req.Body)
		if reqBodyErr != nil {
			err = multierr.Append(err, fmt.Errorf("error read request body: %w", reqBodyErr))
			reqBody = []byte("")
		}

		req.Body = io.NopCloser(bytes.NewReader(reqBody))
	}

	resp, err = r.Base.RoundTrip(req.WithContext(ctx))
	if err != nil {
		err = multierr.Append(err, fmt.Errorf("error doing actual request: %w", err))
	}

	if resp == nil {
		resp = &http.Response{}
	}

	var (
		respBody    []byte
		respErrBody error
	)
	if resp.Body != nil {
		respBody, respErrBody = ioutil.ReadAll(resp.Body)
		if respErrBody != nil {
			err = multierr.Append(err, fmt.Errorf("error read response body: %w", respErrBody))
			respBody = []byte{}
		}

		resp.Body = ioutil.NopCloser(bytes.NewBuffer(respBody))
	}

	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	// log outgoing request
	logger.Access(ctx, logger.AccessLogData{
		Path:        req.URL.String(),
		ReqBody:     string(reqBody),
		RespBody:    string(respBody),
		Error:       errStr,
		ElapsedTime: time.Since(t0).Nanoseconds(),
	})

	return resp, err
}
