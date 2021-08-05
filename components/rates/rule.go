package rates

import (
	"errors"
	"regexp"
)

// ErrRateInvalid is the error that signals about invalid rate value
var ErrRateInvalid = errors.New("rate must be greater that zero")

// Rule is the base class of rule
type Rule struct {
	Rate float64
}

// NewRule is the constructor for Rule
func NewRule(rate float64) (*Rule, error) {
	if rate <= 0 {
		return nil, ErrRateInvalid
	}

	return &Rule{Rate: rate}, nil
}

// NamespaceRule is the one of the rule types
type NamespaceRule struct {
	Rule
	namespace *regexp.Regexp
}

// NewNamespaceRule is the constructor for NamespaceRule
func NewNamespaceRule(record RateRecord) (*NamespaceRule, error) {
	ruleBase, err := NewRule(record.Rate)
	if err != nil {
		return nil, err
	}

	rule := &NamespaceRule{
		Rule: *ruleBase,
	}

	namespaceRegex, err := regexp.Compile(record.Namespace)
	if err != nil {
		return nil, err
	}

	rule.namespace = namespaceRegex

	return rule, nil
}

// Match tries to match specified string with namespace regular expression
func (rule *NamespaceRule) Match(namespace string) bool {
	return rule.namespace.MatchString(namespace)
}

// PodRule is the one of the rule types
type PodRule struct {
	Rule
	pod *regexp.Regexp
}

// NewPodRule is the constructor for PodRule
func NewPodRule(record RateRecord) (*PodRule, error) {
	ruleBase, err := NewRule(record.Rate)
	if err != nil {
		return nil, err
	}

	rule := &PodRule{
		Rule: *ruleBase,
	}

	podRegex, err := regexp.Compile(record.Pod)
	if err != nil {
		return nil, err
	}

	rule.pod = podRegex
	return rule, nil
}

// Match tries to match specified string with pod regular expression
func (rule *PodRule) Match(pod string) bool {
	return rule.pod.MatchString(pod)
}

// NamespacedPodRule is the one of the rule types
type NamespacedPodRule struct {
	Rule
	namespace *regexp.Regexp
	pod       *regexp.Regexp
}

// NewNamespacedPodRule is the constructor for NamespacedPodRule
func NewNamespacedPodRule(record RateRecord) (*NamespacedPodRule, error) {
	ruleBase, err := NewRule(record.Rate)
	if err != nil {
		return nil, err
	}

	rule := &NamespacedPodRule{
		Rule: *ruleBase,
	}

	namespaceRegex, err := regexp.Compile(record.Namespace)
	if err != nil {
		return nil, err
	}

	podRegex, err := regexp.Compile(record.Pod)
	if err != nil {
		return nil, err
	}

	rule.namespace = namespaceRegex
	rule.pod = podRegex
	return rule, nil
}

// Match tries to match specified pair with namespace and pod regular expressions
func (rule *NamespacedPodRule) Match(namespace, pod string) bool {
	return rule.namespace.MatchString(namespace) && rule.pod.MatchString(pod)
}
