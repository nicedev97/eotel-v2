package eotel

type Config struct {
	ServiceName   string
	JobName       string
	OtelCollector string

	EnableTracing bool
	EnableMetrics bool
	EnableSentry  bool
	EnableLoki    bool

	SentryDSN string
	LokiURL   string
}

var globalCfg Config
