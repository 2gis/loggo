package firehoseclient

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	fh "github.com/aws/aws-sdk-go/service/firehose"

	"github.com/2gis/loggo/configuration"
)

type FirehoseClient struct {
	client         *fh.Firehose
	deliveryStream string
}

func (c *FirehoseClient) DeliverMessages(strings []string) error {
	rb := &fh.PutRecordBatchInput{DeliveryStreamName: &c.deliveryStream}

	var records []*fh.Record

	for _, s := range strings {
		records = append(records, &fh.Record{Data: []byte(s)})
	}

	rb.SetRecords(records)

	if _, err := c.client.PutRecordBatch(rb); err != nil {
		return fmt.Errorf("unable to delive message to delivery stream %s, %w", c.deliveryStream, err)
	}

	return nil
}

func (c *FirehoseClient) Close() error {
	return nil
}

func NewFireHoseClient(config configuration.FirehoseTransportConfig) (*FirehoseClient, error) {
	s, err := session.NewSession()
	if err != nil {
		return &FirehoseClient{}, fmt.Errorf("unable to create new aws session, %w", err)
	}

	return &FirehoseClient{
		deliveryStream: config.DeliveryStream,
		client:         fh.New(s),
	}, nil
}
