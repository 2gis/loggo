package mocks

import "github.com/2gis/loggo/common"

// SLIMock is stub for testing purposes
type SLIMock struct {
	parsed int
}

// NewSLIMock is a constructor for SLIMock
func NewSLIMock() *SLIMock {
	return &SLIMock{}
}

// Parse increments parsed count
func (parser *SLIMock) Parse(_ common.EntryMap) {
	parser.parsed++
}

// Parse increments parsed count
func (parser *SLIMock) Parsed() int {
	return parser.parsed
}
