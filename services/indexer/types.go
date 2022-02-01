package indexer

import (
	"github.com/9spokes/go/types"
	"time"
)

// DatasourceRolling is an indexing data entry for a "rolling" datasource
type DatasourceRolling struct {
	Period  string    `json:"period"`
	Status  string    `json:"status"`
	Retry   bool      `json:"retry"`
	Updated time.Time `json:"updated,omit_empty"`
	Outcome string    `json:"outcome"`
	Index   string    `json:"index"`
}

// DatasourceAbsolute is an indexing data entry for an "absolute" datasource
type DatasourceAbsolute struct {
	Status  string    `json:"status"`
	Retry   bool      `json:"retry"`
	Updated time.Time `json:"updated,omit_empty"`
	Outcome string    `json:"outcome"`
	Index   string    `json:"index"`
	Expires time.Time `json:"expires"`
}

// Index is an index entry used to create a new Indexer document
type Index struct {
	Count        int64       `json:"count"`
	Cycle        string      `json:"cycle"`
	Connection   string      `json:"connection,omitempty"`
	Datasource   string      `json:"datasource"`
	Webhooks     bool        `json:"webhooks,omitempty"`
	Notify       bool        `json:"notify,omitempty"`
	OSP          string      `json:"osp"`
	Status       string      `json:"status,omitempty"`
	Storage      string      `json:"storage"`
	Type         string      `json:"type"`
	Dependencies []string    `json:"depends,omitempty"`
	Data         interface{} `json:"data,omitempty"`
}

// IndexStatus is the response type for getting index status
type IndexStatus struct {
	Total       int       `json:"total"`
	Completed   int       `json:"completed"`
	Percent     float64   `json:"percent"`
	LastUpdated time.Time `json:"last_updated"`
}

type ETLMessages struct {
	IndexerMessages []types.ETLMessage `json:"indexer_messages"`
	Outcome         string             `json:"outcome"`
	IsOK            bool               `json:"is_ok"`
	Retry           bool               `json:"retry"`
}
