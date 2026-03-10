package metrics

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var HttpRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	},
	[]string{"method", "route", "status"},
)

var HttpRequestDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	},
	[]string{"method", "route"},
)

var AuditQueueLength = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "audit_queue_length",
	Help: "Current number of items in the audit worker queue",
})

var AuditQueueCapacity = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "audit_queue_capacity",
	Help: "Total capacity of the audit worker queue",
})

var CacheHitsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total cache hits",
	},
	[]string{"cache"},
)

var CacheMissesTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total cache misses",
	},
	[]string{"cache"},
)

var RateLimitRejectsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "rate_limit_rejects_total",
		Help: "Total requests rejected by rate limiter",
	},
	[]string{"route"},
)

type dbPoolCollector struct {
	pool           *pgxpool.Pool
	activeConns    *prometheus.Desc
	idleConns      *prometheus.Desc
	maxConns       *prometheus.Desc
	totalConns     *prometheus.Desc
	acquireCount   *prometheus.Desc
	acquireDurDesc *prometheus.Desc
}

func NewDBPoolCollector(pool *pgxpool.Pool) prometheus.Collector {
	return &dbPoolCollector{
		pool:           pool,
		activeConns:    prometheus.NewDesc("db_pool_active_connections", "Number of active DB connections", nil, nil),
		idleConns:      prometheus.NewDesc("db_pool_idle_connections", "Number of idle DB connections", nil, nil),
		maxConns:       prometheus.NewDesc("db_pool_max_connections", "Maximum DB connections", nil, nil),
		totalConns:     prometheus.NewDesc("db_pool_total_connections", "Total DB connections (active + idle)", nil, nil),
		acquireCount:   prometheus.NewDesc("db_pool_acquire_count_total", "Total number of connection acquires", nil, nil),
		acquireDurDesc: prometheus.NewDesc("db_pool_acquire_duration_seconds_total", "Total time spent acquiring connections", nil, nil),
	}
}

func (c *dbPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.activeConns
	ch <- c.idleConns
	ch <- c.maxConns
	ch <- c.totalConns
	ch <- c.acquireCount
	ch <- c.acquireDurDesc
}

func (c *dbPoolCollector) Collect(ch chan<- prometheus.Metric) {
	stat := c.pool.Stat()
	ch <- prometheus.MustNewConstMetric(c.activeConns, prometheus.GaugeValue, float64(stat.AcquiredConns()))
	ch <- prometheus.MustNewConstMetric(c.idleConns, prometheus.GaugeValue, float64(stat.IdleConns()))
	ch <- prometheus.MustNewConstMetric(c.maxConns, prometheus.GaugeValue, float64(stat.MaxConns()))
	ch <- prometheus.MustNewConstMetric(c.totalConns, prometheus.GaugeValue, float64(stat.TotalConns()))
	ch <- prometheus.MustNewConstMetric(c.acquireCount, prometheus.CounterValue, float64(stat.AcquireCount()))
	ch <- prometheus.MustNewConstMetric(c.acquireDurDesc, prometheus.CounterValue, stat.AcquireDuration().Seconds())
}
