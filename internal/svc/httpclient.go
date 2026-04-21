package svc

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/onlyLTY/dockerCopilot/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
)

func NewHTTPClient(cfg config.ProxyConfig, opts ...HTTPOption) *http.Client {
	o := defaultHTTPOptions()
	for _, opt := range opts {
		opt(&o)
	}

	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if o.insecureSkipVerify {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if o.timeout > 0 {
		tr.DialContext = (&net.Dialer{
			Timeout:   o.timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext
	}

	tr.Proxy = buildProxyFunc(cfg)

	return &http.Client{Transport: tr}
}

type httpOptions struct {
	insecureSkipVerify bool
	timeout            time.Duration
}

type HTTPOption func(*httpOptions)

func defaultHTTPOptions() httpOptions {
	return httpOptions{}
}

func WithInsecureSkipVerify(skip bool) HTTPOption {
	return func(o *httpOptions) {
		o.insecureSkipVerify = skip
	}
}

func WithTimeout(d time.Duration) HTTPOption {
	return func(o *httpOptions) {
		o.timeout = d
	}
}

func buildProxyFunc(cfg config.ProxyConfig) func(*http.Request) (*url.URL, error) {
	return func(req *http.Request) (*url.URL, error) {
		if !cfg.Enable {
			return http.ProxyFromEnvironment(req)
		}
		if req.URL.Scheme == "https" && cfg.Https != "" {
			return url.Parse(cfg.Https)
		}
		if cfg.Http != "" {
			return url.Parse(cfg.Http)
		}
		return http.ProxyFromEnvironment(req)
	}
}

func ApplyProxyEnv(cfg config.ProxyConfig) {
	if !cfg.Enable {
		return
	}
	if cfg.Http != "" {
		os.Setenv("HTTP_PROXY", cfg.Http)
		logx.Infof("已设置 HTTP_PROXY: %s", cfg.Http)
	}
	if cfg.Https != "" {
		os.Setenv("HTTPS_PROXY", cfg.Https)
		logx.Infof("已设置 HTTPS_PROXY: %s", cfg.Https)
	}
	if cfg.NoProxy != "" {
		os.Setenv("NO_PROXY", cfg.NoProxy)
		logx.Infof("已设置 NO_PROXY: %s", cfg.NoProxy)
	}
}
