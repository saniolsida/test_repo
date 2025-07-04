package kafkamodule

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	nodeRPC     = "http://localhost:26657" // 풀노드 RPC 주소
	kafkaBroker = "localhost:9092"         // Kafka 브로커 주소
	checkPeriod = 5 * time.Second          // 몇 초마다 체크할지
)

func main() {
	fmt.Println("🛰️ Kafka 중계 서버 시작")

	latestHeight := int64(0)

	for {
		height := fetchLatestHeight()
		if height == 0 {
			time.Sleep(checkPeriod)
			continue
		}

		// 새로운 블록이 생긴 경우만 처리
		if height > latestHeight {
			for h := latestHeight + 1; h <= height; h++ {
				sendCommit(h)
				sendValidatorSet(h)
				sendLightBlock(h)
			}
			latestHeight = height
		}

		sendStatus()
		time.Sleep(checkPeriod)
	}
}

func fetchLatestHeight() int64 {
	resp, err := http.Get(fmt.Sprintf("%s/status", nodeRPC))
	if err != nil {
		fmt.Println("❌ 상태 요청 실패:", err)
		return 0
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var status struct {
		Result struct {
			SyncInfo struct {
				LatestBlockHeight string `json:"latest_block_height"`
			} `json:"sync_info"`
		} `json:"result"`
	}
	json.Unmarshal(body, &status)

	height, err := strconv.ParseInt(status.Result.SyncInfo.LatestBlockHeight, 10, 64)
	if err != nil {
		fmt.Println("❌ 블록 높이 파싱 실패:", err)
		return 0
	}
	return height
}

func sendStatus() {
	data := fetchRPC("/status")
	if data != nil {
		sendToKafka("block.status", data)
	}
}

func sendCommit(height int64) {
	endpoint := fmt.Sprintf("/commit?height=%d", height)
	data := fetchRPC(endpoint)
	if data != nil {
		sendToKafka(fmt.Sprintf("block.commit.%d", height), data)
	}
}

func sendValidatorSet(height int64) {
	endpoint := fmt.Sprintf("/validators?height=%d", height)
	data := fetchRPC(endpoint)
	if data != nil {
		sendToKafka(fmt.Sprintf("block.validator.%d", height), data)
	}
}

func sendLightBlock(height int64) {
	endpoint := fmt.Sprintf("/block?height=%d", height)
	data := fetchRPC(endpoint)
	if data != nil {
		sendToKafka("block.lightblock", data)
	}
}

func fetchRPC(path string) []byte {
	url := nodeRPC + path
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ RPC 요청 실패 (%s): %v\n", path, err)
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return body
}

func sendToKafka(topic string, value []byte) {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	err := writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(fmt.Sprintf("key-%d", time.Now().UnixNano())),
		Value: value,
	})
	if err != nil {
		fmt.Printf("❌ Kafka 전송 실패 (%s): %v\n", topic, err)
	} else {
		fmt.Printf("✅ Kafka 전송 성공 (%s)\n", topic)
	}
}
