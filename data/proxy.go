package data

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func Proxy(target string, writer io.Writer, request *http.Request) {
	targetUrl, _ := url.Parse(target) // replace with your target node

	request.URL.Scheme = targetUrl.Scheme
	request.URL.Host = targetUrl.Host
	resp, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		log.Printf("Error: %s", err)
		return
	}
	_, _ = io.Copy(writer, resp.Body)
	_ = resp.Body.Close()
	return
}

func ProxyEx(server string, writer http.ResponseWriter, request *http.Request) {
	copyHeader := func(dst, src http.Header) {
		for k, vv := range src {
			for _, v := range vv {
				dst.Add(k, v)
			}
		}
	}
	targetUrl, _ := url.Parse(server) // replace with your target node

	request.Method = "HEAD"
	request.URL.Scheme = targetUrl.Scheme
	request.URL.Host = targetUrl.Host
	resp, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(writer, "Error", http.StatusServiceUnavailable)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	copyHeader(writer.Header(), resp.Header)
	writer.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(writer, resp.Body)
	return
}
