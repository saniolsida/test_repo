package rpc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"tendermint-light/config"

	"github.com/segmentio/kafka-go"
)

type CommitResponse struct {
	Result struct {
		SignedHeader struct {
			Header struct {
				Height string `json:"height"`
				Time   string `json:"time"`
			} `json:"header"`
			Commit struct {
				BlockID struct {
					Hash string `json:"hash"`
				} `json:"block_id"`
			} `json:"commit"`
		} `json:"signed_header"`
	} `json:"result"`
}

type LightBlockMessage struct {
	Height int64          `json:"height"`
	Commit CommitResponse `json:"commit"`
}

// Kafka에서 최신 블록 높이 받기
func FetchLatestHeightFromKafka() int64 {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{config.KafkaBroker},
		Topic:     "block.status",
		Partition: 0,
		MinBytes:  1,
		MaxBytes:  10e6,
	})
	defer r.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	m, err := r.ReadMessage(ctx)
	if err != nil {
		fmt.Println("❌ Kafka에서 block.status 읽기 실패:", err)
		return 0
	}

	var msg struct {
		LatestBlockHeight string `json:"latest_block_height"`
	}
	if err := json.Unmarshal(m.Value, &msg); err != nil {
		fmt.Println("❌ JSON 파싱 실패:", err)
		return 0
	}

	height, err := strconv.ParseInt(msg.LatestBlockHeight, 10, 64)
	if err != nil {
		fmt.Println("❌ height 변환 실패:", err)
		return 0
	}

	return height
}

// 단일 Kafka 토픽 block.lightblock에서 커밋 정보 가져오기
func FetchCommitFromKafka(height int64) CommitResponse {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{config.KafkaBroker},
		Topic:     "block.lightblock",
		Partition: 0,
		GroupID:   "light-client-commits",
		MinBytes:  1,
		MaxBytes:  10e6,
	})
	defer r.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			fmt.Println("❌ Kafka에서 block.lightblock 메시지 수신 실패:", err)
			return CommitResponse{}
		}

		var msg LightBlockMessage
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			fmt.Println("❌ LightBlockMessage JSON 파싱 실패:", err)
			continue
		}

		if msg.Height == height {
			return msg.Commit
		}
	}
}

// 해시를 []byte로 변환
func DecodeHexHash(hexStr string) []byte {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		fmt.Println("❌ 해시 디코딩 실패:", err)
		return nil
	}
	return bytes
}

// 투표권 문자열을 int64로 반환
func ParseVotingPower(powerStr string) int64 {
	power, err := strconv.ParseInt(powerStr, 10, 64)
	if err != nil {
		fmt.Println("⚠️  voting power 파싱 실패:", err)
		return 0
	}
	return power
}
