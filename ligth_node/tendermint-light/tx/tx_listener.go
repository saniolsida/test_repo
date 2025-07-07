package tx

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"tendermint-light/config"

	"github.com/gorilla/websocket"
)

// 외부에서 트랜잭션을 수신할 수 있도록 채널 선언
var TxChan = make(chan map[string]interface{}, 100) // 버퍼 채널

type RPCRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	ID      int         `json:"id"`
	Params  interface{} `json:"params,omitempty"`
}

type SubscribeParams struct {
	Query string `json:"query"`
}

// PrintPretty 예쁘게 트랜잭션 출력
func PrintPretty(tx map[string]interface{}) {
	b, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		fmt.Println("⚠️ 트랜잭션 포맷 실패:", err)
		return
	}
	fmt.Println("📦 수신된 트랜잭션:")
	fmt.Println(string(b))
}

// StartTxListener connects to Tendermint WebSocket and subscribes to transaction events
func StartTxListener() {
	// 종료 감지 채널
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// WebSocket 연결
	conn, _, err := websocket.DefaultDialer.Dial(config.NodeWebSocketURL, nil)
	if err != nil {
		log.Fatal("❌ WebSocket 연결 실패:", err)
	}
	fmt.Println("🔌 WebSocket 연결 성공")

	// 구독 요청 전송
	sub := RPCRequest{
		Jsonrpc: "2.0",
		Method:  "subscribe",
		ID:      1,
		Params:  SubscribeParams{Query: "tm.event='Tx'"},
	}
	if err := conn.WriteJSON(sub); err != nil {
		log.Fatal("❌ 구독 요청 실패:", err)
	}
	fmt.Println("📡 트랜잭션 이벤트 구독 요청 보냄")

	// 메시지 수신 루프
	go func() {
		defer conn.Close()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("❌ 메시지 읽기 실패:", err)
				return
			}
			var data map[string]interface{}
			if err := json.Unmarshal(msg, &data); err == nil {
				TxChan <- data // ✅ 외부로 트랜잭션 이벤트 전달
			} else {
				log.Println("⚠️ JSON 파싱 실패:", err)
			}
		}
	}()

	// 종료 대기
	<-interrupt
	fmt.Println("\n🛑 종료 신호 수신, 트랜잭션 스트림 종료")
}
