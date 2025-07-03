package controller

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"signature-server/types"
)

var signaturePool = make(map[string][]types.SignatureMessage)
var sigMu sync.Mutex

const requiredSignatures = 1 // 예시

func HandleSignature(sig types.SignatureMessage) {
	sigMu.Lock()
	defer sigMu.Unlock()

	signaturePool[sig.TxID] = append(signaturePool[sig.TxID], sig)
	count := len(signaturePool[sig.TxID])

	if count >= requiredSignatures {
		log.Printf("🎯 서명 %d개 확보 → 블록 생성 요청 시작 (tx_id: %s)", count, sig.TxID)
		go notifyFullNode(sig.TxID, signaturePool[sig.TxID])
		delete(signaturePool, sig.TxID)
	}
}

func notifyFullNode(txID string, sigs []types.SignatureMessage) {
	payload := map[string]interface{}{
		"tx_id":      txID,
		"signatures": sigs,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post("http://192.168.0.19:26657/tx_confirmed", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Println("❌ 풀노드 전달 실패:", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("✅ 블록 생성 요청 전송 완료 (tx_id: %s)", txID)
}
