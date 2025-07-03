package subscriber

import (
	"encoding/json"
	"log"

	"signature-server/server"
	"signature-server/types"

	"github.com/gorilla/websocket"
)

const fullNodeWS = "ws://192.168.0.19:26657/websocket"

type RPCRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	ID      int         `json:"id"`
	Params  interface{} `json:"params,omitempty"`
}

type SubscribeParams struct {
	Query string `json:"query"`
}

func StartFullNodeSubscription() {
	conn, _, err := websocket.DefaultDialer.Dial(fullNodeWS, nil)
	if err != nil {
		log.Fatal("❌ 풀노드 WebSocket 연결 실패:", err)
	}
	log.Println("✅ 풀노드 WebSocket 연결 성공")

	defer conn.Close()

	req := RPCRequest{
		Jsonrpc: "2.0",
		Method:  "subscribe",
		ID:      1,
		Params:  SubscribeParams{Query: "tm.event='Tx'"},
	}
	if err := conn.WriteJSON(req); err != nil {
		log.Fatal("❌ 구독 요청 실패:", err)
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("❌ 풀노드 메시지 수신 실패:", err)
			break
		}

		var raw map[string]interface{}
		if err := json.Unmarshal(msg, &raw); err != nil {
			log.Println("⚠️ JSON 파싱 실패:", err)
			continue
		}

		txID := extractTxID(raw)
		server.BroadcastToClients(types.TxPushMessage{
			TxID: txID,
			Data: string(msg),
		})
	}
}

func extractTxID(data map[string]interface{}) string {
	result, ok := data["result"].(map[string]interface{})
	if !ok {
		return "unknown"
	}
	events, ok := result["events"].(map[string]interface{})
	if !ok {
		return "unknown"
	}
	txHashes, ok := events["tx.hash"].([]interface{})
	if !ok || len(txHashes) == 0 {
		return "unknown"
	}
	return txHashes[0].(string)
}
