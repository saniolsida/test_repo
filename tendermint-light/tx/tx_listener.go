package tx

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"

	"tendermint-light/config"

	"github.com/gorilla/websocket"
)

// 채널 및 메시지 타입 정의는 동일
type SignatureMessage struct {
	TxID      string `json:"tx_id"`
	Signature string `json:"signature"`
	Sender    string `json:"sender"`
}

type TxPushMessage struct {
	TxID string `json:"tx_id"`
	Data string `json:"data"` // 트랜잭션 전체 raw JSON
}

// tx_id를 트랜잭션 원본에서 추출하는 함수
func extractTxIDFromData(raw string) string {
	var msg map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		log.Println("⚠️ 원본 트랜잭션 JSON 파싱 실패:", err)
		return "unknown"
	}

	result, ok := msg["result"].(map[string]interface{})
	if !ok {
		return "unknown"
	}
	events, ok := result["events"].(map[string]interface{})
	if !ok {
		return "unknown"
	}
	hashes, ok := events["tx.hash"].([]interface{})
	if !ok || len(hashes) == 0 {
		return "unknown"
	}
	txID, ok := hashes[0].(string)
	if !ok {
		return "unknown"
	}
	return txID
}

func StartWebSocketClient() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	conn, _, err := websocket.DefaultDialer.Dial(config.WebSocketRelay, nil)
	if err != nil {
		log.Fatal("❌ 중앙 WebSocket 서버 연결 실패:", err)
	}
	defer conn.Close()

	log.Println("🔌 중앙 서버 WebSocket 연결 성공")

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("❌ 메시지 수신 실패:", err)
				break
			}

			var tx TxPushMessage
			if err := json.Unmarshal(msg, &tx); err != nil {
				log.Println("⚠️ JSON 파싱 실패:", err)
				continue
			}

			// tx_id가 없으면 원본 데이터에서 추출
			txID := tx.TxID
			if txID == "" || txID == "unknown" {
				txID = extractTxIDFromData(tx.Data)
			}

			log.Printf("📦 트랜잭션 수신 및 서명 시작: tx_id=%s", txID)

			// 서명 생성 (임시값)
			signature := "SIGNATURE_BY_LIGHTNODE"

			// 중앙 서버로 서명 전송
			sigMsg := SignatureMessage{
				TxID:      txID,
				Signature: signature,
				Sender:    "light-node-1",
			}
			jsonSig, _ := json.Marshal(sigMsg)
			err = conn.WriteMessage(websocket.TextMessage, jsonSig)
			if err != nil {
				log.Println("❌ 서명 전송 실패:", err)
			} else {
				log.Printf("✅ 서명 전송 완료: %s", txID)
			}
		}
	}()

	<-interrupt
	log.Println("🛑 종료 신호 수신, 연결 종료")
}
