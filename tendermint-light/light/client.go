package light

import (
	"context"
	"time"

	"tendermint-light/config"

	dbm "github.com/cometbft/cometbft-db"
	tmlight "github.com/cometbft/cometbft/light"
	provider "github.com/cometbft/cometbft/light/provider"
	httpProvider "github.com/cometbft/cometbft/light/provider/http"
	"github.com/cometbft/cometbft/light/store/db"
)

type TrustOptions = tmlight.TrustOptions

type Client struct {
	client *tmlight.Client
}

func NewLightClient(opts TrustOptions) (*Client, error) {
	// provider 설정
	primary, err := httpProvider.New(config.ChainID, config.NodeRPC)
	if err != nil {
		return nil, err
	}

	// witness provider 설정 (보통 primary와 동일하거나 다른 노드)
	witness, err := httpProvider.New(config.ChainID, config.NodeRPC) // ✅ 또는 다른 피어 주소
	if err != nil {
		return nil, err
	}

	// db 설정
	memoryDB := dbm.NewMemDB()
	store := db.New(memoryDB, "light-client-store")

	// light client 생성
	lc, err := tmlight.NewClient(
		context.Background(),
		config.ChainID,
		opts,
		primary,
		[]provider.Provider{witness},
		store,
	)
	if err != nil {
		return nil, err
	}

	return &Client{client: lc}, nil
}

func (lc *Client) VerifyToHeight(ctx context.Context, height int64) error {
	_, err := lc.client.VerifyLightBlockAtHeight(ctx, height, time.Now())
	return err
}
