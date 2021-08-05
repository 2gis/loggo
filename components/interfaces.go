package components

// Retriever represents component that can be retrieved in some way
type Retriever interface {
	Retrieve() error
}
