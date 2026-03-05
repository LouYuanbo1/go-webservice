package elasticsearchx

import (
	"errors"
)

var (
	WarnInvalidDocument    = "elasticsearchx: document is invalid"
	WarnIndexExist         = "elasticsearchx: index exist"
	WarnEmptyDocumentSlice = "elasticsearchx: empty document slice"
	WarnEmptyIDSlice       = "elasticsearchx: empty ids slice"
	WarnDocumentNotFound   = "elasticsearchx: document not found"
)

var (
	ErrCheckExistence      = errors.New("elasticsearchx: check index existence failed")
	ErrGetMapping          = errors.New("elasticsearchx: get index mapping failed")
	ErrMarshalMapping      = errors.New("elasticsearchx: marshal index mapping failed")
	ErrCreateIndex         = errors.New("elasticsearchx: create index failed")
	ErrDeleteIndex         = errors.New("elasticsearchx: delete index failed")
	ErrIndexDocument       = errors.New("elasticsearchx: index document failed")
	ErrBulkIndexDocuments  = errors.New("elasticsearchx: bulk index documents failed")
	ErrNewBulkIndexer      = errors.New("elasticsearchx: new bulk indexer failed")
	ErrMarshalDocument     = errors.New("elasticsearchx: marshal document failed")
	ErrGetIndices          = errors.New("elasticsearchx: get indices failed")
	ErrGetDocument         = errors.New("elasticsearchx: get document failed")
	ErrFindDocuments       = errors.New("elasticsearchx: find documents failed")
	ErrUnmarshalDocument   = errors.New("elasticsearchx: unmarshal document failed")
	ErrCountDocument       = errors.New("elasticsearchx: count document failed")
	ErrUpdateDocument      = errors.New("elasticsearchx: update document failed")
	ErrDeleteDocument      = errors.New("elasticsearchx: delete document failed")
	ErrBulkDeleteDocuments = errors.New("elasticsearchx: bulk delete documents failed")
)
