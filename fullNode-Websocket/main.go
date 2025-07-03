package main

import (
	"encoding/json"
	// "fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type SignatureMessage struct {
	TxID      string `json:"tx_id"`
	Signature string `json:"signature"`
	Sender    string `json:"sender"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("❌ WebSocket 업그레이드 실패:", err)
		return
	}
	defer conn.Close()

	log.Println("🔗 클라이언트 연결 수립됨")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("❌ 메시지 수신 오류:", err)
			break
		}

		var sm SignatureMessage
		if err := json.Unmarshal(msg, &sm); err != nil {
			log.Println("⚠️ JSON 파싱 실패:", err)
			continue
		}

		log.Printf("📨 서명 메시지 수신:\n- tx_id: %s\n- sender: %s\n- signature: %s\n",
			sm.TxID, sm.Sender, sm.Signature)
	}
}

func main() {
	http.HandleFunc("/ws", wsHandler)

	port := "8080"
	log.Printf("🚀 WebSocket 수신 서버 시작 (포트: %s)...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("서버 실행 실패:", err)
	}
}
