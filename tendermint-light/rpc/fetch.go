package rpc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const NodeRPC = "http://192.168.0.19:26657"

// 가장 최신 블록의 높이 반환
func FetchLatestHeight() int64 {
	url := fmt.Sprintf("%s/status", NodeRPC)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
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

	err = json.Unmarshal(body, &status)
	if err != nil {
		fmt.Println("❌ FetchLatestHeight JSON 파싱 실패:", err)
	}

	height, err := strconv.ParseInt(status.Result.SyncInfo.LatestBlockHeight, 10, 64)
	if err != nil {
		fmt.Println("❌ 블록 높이 변환 실패:", err)
		return 0
	}
	return height
}

// 커밋 정보를 가져온다 (신뢰 헤더용)
func FetchCommit(height int64) CommitResponse {
	url := fmt.Sprintf("%s/commit?height=%d", NodeRPC, height)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result CommitResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("❌ FetchCommit JSON 파싱 실패:", err)
	}
	return result
}

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
