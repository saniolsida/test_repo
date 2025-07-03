package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"signature-server/controller"
	"signature-server/types"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var clientsMu sync.Mutex

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func StartWebSocketServer() {
	http.HandleFunc("/ws", wsHandler)

	log.Println("🚀 중앙 WebSocket 서버 시작: 포트 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("❌ 서버 실패:", err)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("❌ 업그레이드 실패:", err)
		return
	}
	defer conn.Close()

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()
	log.Println("🔗 라이트노드 연결됨")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("🔌 연결 종료:", err)
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			break
		}

		var sig types.SignatureMessage
		if err := json.Unmarshal(msg, &sig); err != nil {
			log.Println("⚠️ JSON 파싱 실패 (서명):", err)
			continue
		}
		controller.HandleSignature(sig)
	}
}

// 라이트노드에게 메시지 전송
func BroadcastToClients(msg types.TxPushMessage) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	data, _ := json.Marshal(msg)
	for conn := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Println("❌ 전송 실패. 연결 해제:", err)
			conn.Close()
			delete(clients, conn)
		}
	}
}
