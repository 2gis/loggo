package parsers

import "github.com/2gis/loggo/common"

// SLIStub is stub for testing purposes
type SLIStub struct{}

// NewParserSliStub is a constructor for SLIStub
func NewParserSliStub() *SLIStub {
	return &SLIStub{}
}

// Parse does nothing, stub
func (parser *SLIStub) Parse(_ common.EntryMap) {}
