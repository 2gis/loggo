package configuration

// SLA related
const (
	// AnnotationExporterEnableDefault name of the annotation for enable / disable parsing logs
	AnnotationExporterEnableDefault = "loggo.sla/enable" //

	// AnnotationExporterPathsDefault name of the annotation with label-regexpset definitions
	AnnotationExporterPathsDefault = "loggo.sla/paths"

	// AnnotationSLADomainsDefault name of the annotation with served domain list
	AnnotationSLADomainsDefault = "loggo.sla/domains"

	// AnnotationSLADomainsDeprecated deprecated name of the annotation with served domain list
	AnnotationSLADomainsDeprecated = "router.deis.io/domains"

	AnnotationExporterEnableTrueDeprecated = "enable"
	AnnotationExporterEnableTrue           = "true"


)