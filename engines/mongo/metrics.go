package mongo

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	overallUnreleasedMetric = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "stepper_mongo_count_all_unreleased",
		Help: "Can be used to detect overall unreleased tasks",
	})
)
