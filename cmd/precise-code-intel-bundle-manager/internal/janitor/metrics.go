package janitor

import "github.com/prometheus/client_golang/prometheus"

type JanitorMetrics struct {
	OldUploads    prometheus.Counter
	OrphanedDumps prometheus.Counter
	EvictedDumps  prometheus.Counter
	Errors        prometheus.Counter
}

func NewJanitorMetrics(r prometheus.Registerer) JanitorMetrics {
	oldUploads := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "precise_code_intel_bundle_manager",
		Name:      "janitor_old_uploads",
		Help:      "Total number of old upload removed",
	})
	r.MustRegister(oldUploads)

	orphanedDumps := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "precise_code_intel_bundle_manager",
		Name:      "janitor_orphaned_dumps",
		Help:      "Total number of orphaned dumps removed",
	})
	r.MustRegister(orphanedDumps)

	evictedDumps := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "precise_code_intel_bundle_manager",
		Name:      "janitor_old_dumps",
		Help:      "Total number of dumps evicted from disk",
	})
	r.MustRegister(evictedDumps)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "precise_code_intel_bundle_manager",
		Name:      "janitor_errors",
		Help:      "Total number of errors when running the janitor",
	})
	r.MustRegister(errors)

	return JanitorMetrics{
		OldUploads:    oldUploads,
		OrphanedDumps: orphanedDumps,
		EvictedDumps:  evictedDumps,
		Errors:        errors,
	}
}
