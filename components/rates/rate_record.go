package rates

// RateRecord is the struct to unmarshal yaml list item to
type RateRecord struct {
	Namespace string
	Pod       string
	Rate      float64
}
