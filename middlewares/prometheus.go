package middlewares

import (
	"time"

	"github.com/matroskin13/stepper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
)

type Prometheus struct {
	total    *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func NewPrometheus() *Prometheus {
	return &Prometheus{
		total: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "stepper_task_execution",
			Help: "Count of all task executions",
		}, []string{"task", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "stepper_task_duration_seconds",
			Help:    "Duration of all executions",
			Buckets: []float64{.025, .05, .1, .25, .5, 1, 2.5, 5, 10, 20, 30},
		}, []string{"task", "status"}),
	}
}

func (p *Prometheus) GetRegistry() *prometheus.Registry {
	r := prometheus.NewRegistry()

	r.MustRegister(p.total)
	r.MustRegister(p.duration)

	return r
}

func (p *Prometheus) GetMiddleware() stepper.MiddlewareHandler {
	return func(next stepper.MiddlewareFunc) stepper.MiddlewareFunc {
		return func(ctx stepper.Context, t *stepper.Task) error {
			startTime := time.Now()

			err := next(ctx, t)

			status := lo.Ternary(err == nil, "success", "fail")
			duration := time.Since(startTime)

			p.total.WithLabelValues(t.Name, status).Inc()
			p.duration.WithLabelValues(t.Name, status).Observe(duration.Seconds())

			return err
		}
	}
}
