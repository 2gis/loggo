package parsers

import "github.com/2gis/loggo/common"

func ParseStringPlain(line []byte) common.EntryMap {
	return common.EntryMap{LogKeyLog: string(line)}
}
