package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	KafkaMessagesTotal        prometheus.Counter
	KafkaInvalidMessagesTotal prometheus.Counter
	KafkaDuplicateMessages    prometheus.Counter
	KafkaSaveErrorsTotal      prometheus.Counter

	CacheHitsTotal   prometheus.Counter
	CacheMissesTotal prometheus.Counter

	HTTPRequestTotal *prometheus.CounterVec
}

func New() *Metrics {
	m := &Metrics{
		KafkaMessagesTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "kafka_messages_total",
			Help: "Total number of Kafka messages received",
		}),
		KafkaInvalidMessagesTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "kafka_invalid_messages_total",
			Help: "Total number of invalid Kafka messages",
		}),
		KafkaDuplicateMessages: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "kafka_duplicate_messages_total",
			Help: "Total number of duplicate Kafka messages",
		}),
		KafkaSaveErrorsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "kafka_save_errors_total",
			Help: "Total number of Kafka message save errors",
		}),
		CacheHitsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		}),
		CacheMissesTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		}),
		HTTPRequestTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
	}

	prometheus.MustRegister(
		m.KafkaMessagesTotal,
		m.KafkaInvalidMessagesTotal,
		m.KafkaDuplicateMessages,
		m.KafkaSaveErrorsTotal,
		m.CacheHitsTotal,
		m.CacheMissesTotal,
		m.HTTPRequestTotal,
	)

	return m
}