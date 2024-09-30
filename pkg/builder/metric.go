package builder

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "paymaster"

// subsystem
const (
	system = "builder"
)

var (
	ErrorCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: system,
		Name:      "error",
	}, []string{"url"})
)
