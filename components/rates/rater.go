package rates

import (
	"sync"
)

// RateRecordsProvider is the records provider interface
type RateRecordsProvider interface {
	RateRecords() ([]RateRecord, error)
}

// Rater is the structure that is used for getting rates according to regexp rules received from provider
type Rater struct {
	sync.RWMutex
	rateDefault        float64
	provider           RateRecordsProvider
	NamespacedPodRules []*NamespacedPodRule
	PodRules           []*PodRule
	NamespaceRules     []*NamespaceRule
}

// NewRater is the constructor for Rater
func NewRater(provider RateRecordsProvider, rateDefault float64) (*Rater, error) {
	if rateDefault <= 0 {
		return nil, ErrRateInvalid
	}

	return &Rater{
		rateDefault:        rateDefault,
		provider:           provider,
		NamespacedPodRules: []*NamespacedPodRule(nil),
		PodRules:           []*PodRule(nil),
		NamespaceRules:     []*NamespaceRule(nil),
	}, nil
}

// Retrieve may be called periodically to obtain changes in rules
func (rater *Rater) Retrieve() error {
	ruleRecords, err := rater.provider.RateRecords()
	if err != nil {
		return err
	}

	if len(ruleRecords) == 0 {
		return nil
	}

	namespacedPodRules := make([]*NamespacedPodRule, 0)
	podRules := make([]*PodRule, 0)
	namespaceRules := make([]*NamespaceRule, 0)

	for _, record := range ruleRecords {
		if record.Namespace == "" && record.Pod == "" {
			continue
		}

		if record.Namespace != "" && record.Pod != "" {
			rule, err := NewNamespacedPodRule(record)

			if err != nil {
				return err
			}

			namespacedPodRules = append(namespacedPodRules, rule)
			continue
		}

		if record.Pod != "" {
			rule, err := NewPodRule(record)

			if err != nil {
				return err
			}

			podRules = append(podRules, rule)
			continue
		}

		rule, err := NewNamespaceRule(record)

		if err != nil {
			return err
		}

		namespaceRules = append(namespaceRules, rule)
	}

	rater.Lock()
	rater.NamespacedPodRules = namespacedPodRules
	rater.PodRules = podRules
	rater.NamespaceRules = namespaceRules
	rater.Unlock()

	return nil
}

// Rate tries to find a rule that matches namespace and pod and returns corresponding rate
func (rater *Rater) Rate(namespace string, pod string) float64 {
	rater.RLock()
	defer rater.RUnlock()

	for _, rule := range rater.NamespacedPodRules {
		if rule.Match(namespace, pod) {
			return rule.Rate
		}
	}

	for _, rule := range rater.PodRules {
		if rule.Match(pod) {
			return rule.Rate
		}
	}

	for _, rule := range rater.NamespaceRules {
		if rule.Match(namespace) {
			return rule.Rate
		}
	}

	return rater.rateDefault
}
