package process

// ElasticClientHandler defines the behaviour of an elastic search client handler
type ElasticClientHandler interface {
	DoScrollRequestAllDocuments(
		index string,
		body []byte,
		handlerFunc func(responseBytes []byte) error,
	) error
	IsInterfaceNil() bool
}
