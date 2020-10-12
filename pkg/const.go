package pkg

const (
	EnvVarResyncPeriod = "RESYNC_PERIOD"
	EnvVarDevMode = "DEV_MODE"
	// contains desired state that an event handler at 1 point applied ( could change if user decides to modify configmap )
	LastAppliedLabelsAnnotationKey = "agill.apps/last-applied-labels"
)
