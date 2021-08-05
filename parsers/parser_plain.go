package parsers

import "github.com/2gis/loggo/common"

func ParseStringPlain(line []byte) common.EntryMap {
	return common.EntryMap{"log": string(line)}
}
