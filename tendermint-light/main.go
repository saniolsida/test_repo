package main

import (
	"context"
	"fmt"
	"time"

	"os"
	"os/signal"
	"syscall"
	"tendermint-light/light"
	"tendermint-light/rpc"
	"tendermint-light/tx"
)

func main() {
	fmt.Println("🔌 Connecting to full node...")

	ctx := context.Background()

	// 트랜잭션 리스너를 별도 고루틴에서 실행
	go func() {
		tx.StartWebSocketClient()
	}()

	// 최초 블록 기준 설정
	initialHeight := rpc.FetchLatestHeight()
	trustedHeight := initialHeight - 10
	trustedCommit := rpc.FetchCommit(trustedHeight).Result.SignedHeader
	trustedHash := rpc.DecodeHexHash(trustedCommit.Commit.BlockID.Hash)

	// 라이트 클라이언트 생성
	client, err := light.NewLightClient(light.TrustOptions{
		Period: 7 * 24 * time.Hour,
		Height: trustedHeight,
		Hash:   trustedHash,
	})
	if err != nil {
		fmt.Printf("❌ 라이트 클라이언트 생성 실패: %v\n", err)
		return
	}

	// 블록 검증 루프를 별도 고루틴에서 실행
	go func() {
		ticker := time.NewTicker(5 * time.Second) // 5초마다 최신 블록 체크
		defer ticker.Stop()

		lastVerified := trustedHeight

		for range ticker.C {
			latestHeight := rpc.FetchLatestHeight()

			// 10블록마다 검증
			if latestHeight >= lastVerified+10 {
				if err := client.VerifyToHeight(ctx, latestHeight); err != nil {
					fmt.Printf("❌ 블록 검증 실패 (%d): %v\n", latestHeight, err)
				} else {
					fmt.Printf("✅ 블록 %d까지 검증 성공\n", latestHeight)
					lastVerified = latestHeight
				}
			}
		}
	}()

	// 종료 신호 감지 대기
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop // 신호 수신까지 대기
	fmt.Println("🛑 종료 신호 수신, Light node 종료")
}
