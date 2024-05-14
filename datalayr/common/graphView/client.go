package graphView

import (
	"context"
	"sync"
	"time"

	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/shurcooL/graphql"
)

type GraphClient struct {
	Endpoint string
	Logger   *logging.Logger
	mu       sync.Mutex
}

func NewGraphClient(e string, logger *logging.Logger) *GraphClient {
	return &GraphClient{
		Endpoint: e,
		Logger:   logger,
	}
}

func (g *GraphClient) GetEndpoint() string {
	g.mu.Lock()
	endpoint := g.Endpoint
	g.mu.Unlock()
	return endpoint
}

func (g *GraphClient) SetEndpoint(e string) {
	g.mu.Lock()
	g.Endpoint = e
	g.mu.Unlock()
}

func (g *GraphClient) WaitForSubgraph() error {

	check := func() error {
		var query struct {
			DataStores []DataStoreGql `graphql:"dataStores(first:1)"`
		}
		variables := map[string]interface{}{}

		client := graphql.NewClient(g.GetEndpoint(), nil)
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			g.Logger.Warn().Err(err).Msg("Error connecting to subgraph")
			return err
		}
		return nil
	}

	interval := time.Second
	maxInterval := 5 * time.Minute

	for {
		err := check()
		if err != nil {
			interval *= 2
			if interval > maxInterval {
				return ErrSubgraphTimeout
			}
			t := time.NewTicker(interval)
			<-t.C
		} else {
			return nil
		}
	}
}
