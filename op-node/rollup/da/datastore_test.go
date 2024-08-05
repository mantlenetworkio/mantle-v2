package da

import (
	"context"
	"fmt"
	"testing"

	"github.com/Layr-Labs/datalayr/common/graphView"
	"github.com/shurcooL/graphql"
)

func TestMantleDataStore_RetrievalFramesFromDaIndexer(t *testing.T) {
	type fields struct {
		Ctx           context.Context
		Cfg           *MantleDataStoreConfig
		GraphClient   *graphView.GraphClient
		GraphqlClient *graphql.Client
	}
	type args struct {
		dataStoreId uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "t1",
			fields: fields{
				Ctx: context.Background(),
				Cfg: &MantleDataStoreConfig{
					MantleDaIndexerSocket: "da-indexer-api-sepolia-qa6.qa.gomantle.org:80",
				},
				GraphClient:   &graphView.GraphClient{},
				GraphqlClient: &graphql.Client{},
			},
			args: args{
				dataStoreId: 10138,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mda := &MantleDataStore{
				Ctx:           tt.fields.Ctx,
				Cfg:           tt.fields.Cfg,
				GraphClient:   tt.fields.GraphClient,
				GraphqlClient: tt.fields.GraphqlClient,
			}
			got, err := mda.RetrievalFramesFromDaIndexer(tt.args.dataStoreId)
			if (err != nil) != tt.wantErr {
				t.Errorf("MantleDataStore.RetrievalFramesFromDaIndexer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println("got:", len(got))
		})
	}
}
