package proxyFatory

import (
	"time"
)

const (
	defaultHealthCheckInterval = 30 * time.Second
	defaultHealthCheckTimeout  = 5 * time.Second
)

const defaultMaxBodySize int64 = -1

// ServiceManager according to service generate handler to handle request.
type ServiceManager struct {
	ServiceConfig
}

type ServiceConfig struct {
	ResponseForwardingConfig
}

type ResourceService struct {

}
