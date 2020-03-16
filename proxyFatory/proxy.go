package proxyFatory

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// StatusClientClosedRequest non-standard HTTP status code for client disconnection
const StatusClientClosedRequest = 499

// StatusClientClosedRequestText non-standard HTTP status for client disconnection
const StatusClientClosedRequestText = "Client Closed Request"

type Override struct {
	Header string
	Match  string
	Host   string
	Path   string
}

type Config struct {
	Path     string
	Host     string
	Override Override
}

// generateProxy build a proxy to config.Host
func generateProxy(conf Config) *httputil.ReverseProxy {
	proxy := &httputil.ReverseProxy{Director: func(outReq *http.Request) {
		u := outReq.URL
		if outReq.RequestURI != "" {
			parsedURL, err := url.ParseRequestURI(outReq.RequestURI)
			if err == nil {
				u = parsedURL
			}
		}
		log.Println(outReq.RequestURI)
		// Do not pass client Host header unless want to.
		//if passHostHeader != nil && !*passHostHeader {
		//	outReq.Host = outReq.URL.Host
		//}
		originHost := conf.Host
		outReq.Header.Add("X-Forwarded-Host", outReq.Host)
		outReq.Header.Add("X-Origin-Host", originHost)
		outReq.Host = originHost
		outReq.URL.Host = originHost

		outReq.URL.Path = u.Path
		outReq.URL.RawPath = u.RawPath
		outReq.URL.RawQuery = u.RawQuery
		//outReq.URL.Scheme = u.Scheme
		outReq.RequestURI = "" // Outgoing request should not have RequestURI

		outReq.Proto = "HTTP/1.1"
		outReq.ProtoMajor = 1
		outReq.ProtoMinor = 1

		if _, ok := outReq.Header["User-Agent"]; !ok {
			outReq.Header.Set("User-Agent", "")
		}

		// Even if the websocket RFC says that headers should be case-insensitive,
		// some servers need Sec-WebSocket-Key to be case-sensitive.
		// https://tools.ietf.org/html/rfc6455#page-20
		outReq.Header["Sec-WebSocket-Key"] = outReq.Header["Sec-Websocket-Key"]
		delete(outReq.Header, "Sec-Websocket-Key")


		if conf.Override.Header != "" && conf.Override.Match != "" {
			if outReq.Header.Get(conf.Override.Header) == conf.Override.Match {
				outReq.URL.Path = conf.Override.Path
			}
		}
	},
		ErrorHandler: func(w http.ResponseWriter, request *http.Request, err error) {
			statusCode := http.StatusInternalServerError

			switch {
			case err == io.EOF:
				statusCode = http.StatusBadGateway
			case err == context.Canceled:
				statusCode = StatusClientClosedRequest
			default:
				if e, ok := err.(net.Error); ok {
					if e.Timeout() {
						statusCode = http.StatusGatewayTimeout
					} else {
						statusCode = http.StatusBadGateway
					}
				}
			}

			log.Printf("'%d %s' caused by: %v", statusCode, statusText(statusCode), err)
			w.WriteHeader(statusCode)
			_, werr := w.Write([]byte(statusText(statusCode)))
			if werr != nil {
				log.Println("Error while writing status code", werr)
			}
		},
	}

	return proxy
}

func statusText(statusCode int) string {
	if statusCode == StatusClientClosedRequest {
		return StatusClientClosedRequestText
	}
	return http.StatusText(statusCode)
}