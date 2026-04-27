package service

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// httpClientProxyFunc 为外部 HTTP 客户端统一生成代理选择逻辑。
func httpClientProxyFunc(httpProxy, httpsProxy, noProxy string) func(*http.Request) (*url.URL, error) {
	noProxyList := parseNoProxy(noProxy)
	return func(req *http.Request) (*url.URL, error) {
		if len(noProxyList) > 0 && matchesNoProxy(req.URL.Host, noProxyList) {
			return nil, nil
		}
		if req.URL.Scheme == "https" && httpsProxy != "" {
			return url.Parse(httpsProxy)
		}
		if httpProxy != "" {
			return url.Parse(httpProxy)
		}
		return http.ProxyFromEnvironment(req)
	}
}

// matchesNoProxy 判断目标主机是否命中 no_proxy 规则。
func matchesNoProxy(host string, noProxyList []string) bool {
	for _, pattern := range noProxyList {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if pattern[0] == '*' {
			suffix := pattern[1:]
			if strings.HasSuffix(host, suffix) {
				return true
			}
		} else if pattern == host {
			return true
		}
	}
	return false
}

// parseNoProxy 解析逗号分隔的 no_proxy 配置。
func parseNoProxy(noProxy string) []string {
	if noProxy == "" {
		return nil
	}
	parts := strings.Split(noProxy, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// newExternalHTTPClient 为漏洞增强和源码抓取提供统一外部客户端。
func newExternalHTTPClient(timeout time.Duration, proxyHTTP, proxyHTTPS, noProxy string) *http.Client {
	proxyFn := httpClientProxyFunc(proxyHTTP, proxyHTTPS, noProxy)
	transport := &http.Transport{
		Proxy:                 proxyFn,
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          10,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: timeout,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			dialer := &net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}
			return dialer.DialContext(ctx, "tcp4", address)
		},
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

var (
	globalProxyHTTP  string
	globalProxyHTTPS string
	globalNoProxy    string
)

// SetProxyConfig 初始化 service 包统一使用的外部代理配置。
func SetProxyConfig(httpProxy, httpsProxy, noProxy string) {
	globalProxyHTTP = httpProxy
	globalProxyHTTPS = httpsProxy
	globalNoProxy = noProxy
}

// extHTTP 返回基于全局代理配置的外部 HTTP 客户端。
func extHTTP(timeout time.Duration) *http.Client {
	return newExternalHTTPClient(timeout, globalProxyHTTP, globalProxyHTTPS, globalNoProxy)
}
