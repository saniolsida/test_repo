package light

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"tendermint-light/config"

	dbm "github.com/cometbft/cometbft-db"
	tmlight "github.com/cometbft/cometbft/light"
	provider "github.com/cometbft/cometbft/light/provider"
	"github.com/cometbft/cometbft/light/store/db"
	"github.com/cometbft/cometbft/types"
	"github.com/segmentio/kafka-go"
)

type TrustOptions = tmlight.TrustOptions

type Client struct {
	client *tmlight.Client
}

// ✅ KafkaProvider 구조체 정의
type KafkaProvider struct {
	chainID string
	broker  string
}

// ✅ KafkaProvider 생성자
func NewKafkaProvider(chainID, broker string) *KafkaProvider {
	return &KafkaProvider{chainID: chainID, broker: broker}
}

type LightBlockMessage struct {
	Height     int64            `json:"height"`
	LightBlock types.LightBlock `json:"light_block"`
}

func (kp *KafkaProvider) LightBlock(ctx context.Context, height int64) (*types.LightBlock, error) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{kp.broker},
		Topic:     "block.lightblock",
		Partition: 0,
		GroupID:   fmt.Sprintf("light-client-%s", kp.chainID), // GroupID 설정
		MinBytes:  1,
		MaxBytes:  10e6,
	})
	defer r.Close()

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for {
		msg, err := r.ReadMessage(timeoutCtx)
		if err != nil {
			return nil, fmt.Errorf("❌ Kafka 메시지 수신 실패: %w", err)
		}

		var data LightBlockMessage
		if err := json.Unmarshal(msg.Value, &data); err != nil {
			fmt.Println("⚠️ JSON 파싱 실패:", err)
			continue
		}

		if data.Height == height {
			return &data.LightBlock, nil
		}
	}
}

// ✅ interface 필수 구현: 체인 ID 반환
func (kp *KafkaProvider) ChainID() string {
	return kp.chainID
}

// ✅ interface 필수 구현: 증거 제출용 → 이 프로젝트에선 stub 처리 가능
func (kp *KafkaProvider) ReportEvidence(_ context.Context, _ types.Evidence) error {
	return nil
}

// ✅ Light Client 생성
func NewLightClient(opts TrustOptions) (*Client, error) {
	primary := NewKafkaProvider(config.ChainID, config.KafkaBroker)
	witness := NewKafkaProvider(config.ChainID, config.KafkaBroker2) // 필요시 다른 broker/토픽으로 변경

	memoryDB := dbm.NewMemDB()
	store := db.New(memoryDB, "light-client-store")

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

// ✅ LightBlock 검증
func (lc *Client) VerifyToHeight(ctx context.Context, height int64) error {
	_, err := lc.client.VerifyLightBlockAtHeight(ctx, height, time.Now())
	return err
}
