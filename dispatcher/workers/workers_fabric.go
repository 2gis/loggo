package workers

import (
	"fmt"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/configuration"
	"github.com/2gis/loggo/logging"
	"github.com/2gis/loggo/readers"
)

type FollowersFabric struct {
	config    configuration.Config
	collector MetricsCollector
	storage   Storage
	rater     Rater
	logger    logging.Logger
}

func NewFollowersFabric(config configuration.Config,
	collector MetricsCollector, storage Storage, rater Rater, logger logging.Logger) *FollowersFabric {
	return &FollowersFabric{
		config:    config,
		collector: collector,
		storage:   storage,
		rater:     rater,
		logger:    logger,
	}
}

func (f *FollowersFabric) NewFollower(output chan<- *common.Entry, filePath, format string,
	containerExtends common.EntryMap) (Follower, error) {
	cursorString, _ := f.storage.Get(filePath)
	cursor, _ := readers.NewCursorFromString(cursorString)
	lineReader, err := readers.NewLineReader(
		filePath,
		f.config.FollowerConfig.ReaderBufferSize,
		cursor,
		f.config.FollowerConfig.FromTailFlag,
	)

	if err != nil {
		return nil, fmt.Errorf("unable to initiate LineReader for path %s, %s", filePath, err)
	}

	extends := f.config.K8SExtends.EntryMap()
	extends.Extend(containerExtends)

	worker := newFollower(
		output,
		filePath,
		format,
		lineReader,
		f.collector,
		f.storage,
		f.rater,
		extends,
		f.config.FollowerConfig.NoRecordsSleepIntervalSec,
		f.config.FollowerConfig.CursorCommitIntervalSec,
		f.config.FollowerConfig.ThrottlingLimitsUpdateIntervalSec,
		f.logger,
	)
	return worker, nil
}

// NewFollowerJournald constructor
func (f *FollowersFabric) NewFollowerJournald(output chan<- string, config configuration.ParserConfig,
	logger logging.Logger) (FollowerJournald, error) {
	journaldPath, err := readers.JournaldPath(
		f.config.JournaldConfig.MachineIDPath,
		f.config.JournaldConfig.JournaldPath,
	)

	if err != nil {
		return nil, fmt.Errorf("unable to obtain journald path, %s", err)
	}

	cursor, _ := f.storage.Get(common.FilenameJournald)
	readerJournald, err := readers.NewReaderJournald(journaldPath, cursor)

	if err != nil {
		f.logger.Errorf("workersDispatcher: unable to initialize journald reader, %s", err)
		return nil, err
	}

	return newFollowerJournald(
		output,
		readerJournald,
		config,
		common.EntryMap{
			common.LabelDataCenter:         f.config.K8SExtends.DataCenter,
			common.LabelPurpose:            f.config.K8SExtends.Purpose,
			common.LabelLogstashPrefix:     f.config.K8SExtends.LogstashPrefix,
			common.KubernetesNodeHostname:  f.config.K8SExtends.NodeHostname,
			common.KubernetesNamespaceName: common.NamespaceJournald,
		},
		f.storage,
		f.config.FollowerConfig.CursorCommitIntervalSec,
		f.config.FollowerConfig.NoRecordsSleepIntervalSec,
		logger,
	), nil
}
