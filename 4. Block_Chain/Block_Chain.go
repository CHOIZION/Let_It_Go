package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// 블록 구조체 정의
type Block struct {
	Index        int           `json:"index"`        // 블록 번호
	Timestamp    string        `json:"timestamp"`    // 블록 생성 시간
	Transactions []Transaction `json:"transactions"` // 거래 목록
	PrevHash     string        `json:"prev_hash"`    // 이전 블록의 해시
	Hash         string        `json:"hash"`         // 현재 블록의 해시
	Nonce        int           `json:"nonce"`        // 작업 증명에 사용된 논스
}

// 거래 구조체 정의
type Transaction struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    int    `json:"amount"`
}

// 블록체인 정의 (전역 변수)
var Blockchain []Block
var mutex = &sync.Mutex{} // 동시성 제어를 위한 뮤텍스

const difficulty = 4 // 작업 증명의 난이도 (해시 앞에 0의 개수)

// P2P 네트워킹을 위한 피어 목록
var peers []string
var peersMutex = &sync.Mutex{}

// 웹소켓 업그레이더 설정
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 모든 도메인에서의 연결을 허용
	},
}

// 해시 계산 함수
func calculateHash(block Block) string {
	record := fmt.Sprintf("%d%s%v%s%d", block.Index, block.Timestamp, block.Transactions, block.PrevHash, block.Nonce)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// 작업 증명 함수 (Proof-of-Work)
func proofOfWork(block Block) Block {
	for {
		hash := calculateHash(block)
		if strings.HasPrefix(hash, strings.Repeat("0", difficulty)) {
			block.Hash = hash
			return block
		}
		block.Nonce++
	}
}

// 새로운 블록 생성 함수
func generateBlock(prevBlock Block, transactions []Transaction) Block {
	newBlock := Block{
		Index:        prevBlock.Index + 1,
		Timestamp:    time.Now().Format(time.RFC3339),
		Transactions: transactions,
		PrevHash:     prevBlock.Hash,
		Nonce:        0,
		Hash:         "",
	}
	newBlock = proofOfWork(newBlock)
	return newBlock
}

// 블록 유효성 검사 함수
func isBlockValid(newBlock, prevBlock Block) bool {
	if prevBlock.Index+1 != newBlock.Index {
		return false
	}
	if prevBlock.Hash != newBlock.PrevHash {
		return false
	}
	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}
	if !strings.HasPrefix(newBlock.Hash, strings.Repeat("0", difficulty)) {
		return false
	}
	return true
}

// 블록체인에 새로운 블록 추가
func addBlock(newBlock Block) bool {
	mutex.Lock()
	defer mutex.Unlock()

	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		Blockchain = append(Blockchain, newBlock)
		fmt.Println("블록이 추가되었습니다:", newBlock)
		// 피어에게 블록 전파
		broadcastBlock(newBlock)
		// 웹소켓을 통해 클라이언트에 알림
		notifyConnections(fmt.Sprintf("새 블록이 추가되었습니다: Index %d", newBlock.Index))
		saveBlockchain()
		return true
	}
	fmt.Println("블록 추가 실패: 유효하지 않은 블록입니다.")
	return false
}

// 제네시스 블록 생성 함수
func createGenesisBlock() Block {
	genesisBlock := Block{
		Index:        0,
		Timestamp:    time.Now().Format(time.RFC3339),
		Transactions: []Transaction{},
		PrevHash:     "",
		Nonce:        0,
		Hash:         "",
	}
	genesisBlock = proofOfWork(genesisBlock)
	return genesisBlock
}

// 블록체인 출력 함수 (콘솔용)
func printBlockchain() {
	mutex.Lock()
	defer mutex.Unlock()

	fmt.Println("\n현재 블록체인:")
	for _, block := range Blockchain {
		fmt.Printf("Index: %d, Timestamp: %s, Transactions: %v, Hash: %s, PrevHash: %s, Nonce: %d\n",
			block.Index, block.Timestamp, block.Transactions, block.Hash, block.PrevHash, block.Nonce)
	}
	fmt.Println()
}

// REST API 핸들러

// 블록체인 전체 조회
func getBlockchain(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	bytes, err := json.MarshalIndent(Blockchain, "", "  ")
	if err != nil {
		http.Error(w, "블록체인 데이터를 JSON으로 변환 중 오류 발생", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

// 특정 블록 조회
func getBlock(w http.ResponseWriter, r *http.Request) {
	indexStr := strings.TrimPrefix(r.URL.Path, "/blocks/")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "잘못된 블록 인덱스", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	if index < 0 || index >= len(Blockchain) {
		http.Error(w, "블록을 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	bytes, err := json.MarshalIndent(Blockchain[index], "", "  ")
	if err != nil {
		http.Error(w, "블록 데이터를 JSON으로 변환 중 오류 발생", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

// 새로운 블록 추가 (거래 포함)
func createBlock(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Transactions []Transaction `json:"transactions"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil || len(data.Transactions) == 0 {
		http.Error(w, "유효한 거래 데이터가 필요합니다", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	prevBlock := Blockchain[len(Blockchain)-1]
	mutex.Unlock()

	newBlock := generateBlock(prevBlock, data.Transactions)
	success := addBlock(newBlock)

	if success {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newBlock)
	} else {
		http.Error(w, "블록 추가 실패", http.StatusInternalServerError)
	}
}

// 피어 목록 저장을 위한 슬라이스 및 뮤텍스
var connections = make([]*websocket.Conn, 0)
var connectionsMutex = &sync.Mutex{}

// 모든 연결에 메시지 보내기
func notifyConnections(message string) {
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()

	for _, conn := range connections {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			fmt.Println("웹소켓 메시지 전송 실패:", err)
			conn.Close()
			// 연결 제거
			for i, c := range connections {
				if c == conn {
					connections = append(connections[:i], connections[i+1:]...)
					break
				}
			}
		}
	}
}

// 웹소켓 핸들러
func handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("웹소켓 업그레이드 실패:", err)
		return
	}
	defer conn.Close()

	// 연결 저장
	connectionsMutex.Lock()
	connections = append(connections, conn)
	connectionsMutex.Unlock()

	for {
		// 클라이언트로부터 메시지를 읽지 않음 (단방향)
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("웹소켓 연결 종료:", err)
			break
		}
	}
}

// 템플릿 파일 경로 확인 및 수정
func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "페이지를 찾을 수 없습니다", http.StatusNotFound)
		return
	}
	tmpl, err := template.ParseFiles("index.html") // 경로 수정
	if err != nil {
		http.Error(w, fmt.Sprintf("템플릿 로드 중 오류 발생: %v", err), http.StatusInternalServerError)
		fmt.Printf("템플릿 로드 오류: %v\n", err)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	tmpl.Execute(w, Blockchain)
}

// 블록체인 무결성 검사 함수
func isBlockchainValid() bool {
	mutex.Lock()
	defer mutex.Unlock()

	for i := 1; i < len(Blockchain); i++ {
		currentBlock := Blockchain[i]
		prevBlock := Blockchain[i-1]

		if currentBlock.Index != prevBlock.Index+1 {
			return false
		}
		if currentBlock.PrevHash != prevBlock.Hash {
			return false
		}
		if calculateHash(currentBlock) != currentBlock.Hash {
			return false
		}
		if !strings.HasPrefix(currentBlock.Hash, strings.Repeat("0", difficulty)) {
			return false
		}
	}
	return true
}

// 무결성 검사 및 동기화 모니터링
func monitorBlockchain() {
	for {
		time.Sleep(10 * time.Second)
		if isBlockchainValid() {
			fmt.Println("블록체인 무결성 검사: 유효합니다.")
		} else {
			fmt.Println("블록체인 무결성 검사: 무결성이 깨졌습니다!")
			syncBlockchain()
		}
	}
}

// 블록체인 저장 함수
func saveBlockchain() {
	mutex.Lock()
	defer mutex.Unlock()

	data, err := json.MarshalIndent(Blockchain, "", "  ")
	if err != nil {
		fmt.Println("블록체인 저장 중 오류 발생:", err)
		return
	}
	err = ioutil.WriteFile("blockchain.json", data, 0644)
	if err != nil {
		fmt.Println("블록체인 파일 쓰기 오류:", err)
	}
}

// 블록체인 로드 함수
func loadBlockchain() {
	if _, err := os.Stat("blockchain.json"); err == nil {
		data, err := ioutil.ReadFile("blockchain.json")
		if err != nil {
			fmt.Println("블록체인 파일 읽기 오류:", err)
			return
		}
		var loadedBlockchain []Block
		if err := json.Unmarshal(data, &loadedBlockchain); err != nil {
			fmt.Println("블록체인 데이터 파싱 오류:", err)
			return
		}
		mutex.Lock()
		Blockchain = loadedBlockchain
		mutex.Unlock()
		fmt.Println("블록체인이 로드되었습니다.")
	}
}

// 피어 목록 로드 함수
func loadPeers() {
	if _, err := os.Stat("peers.json"); err == nil {
		data, err := ioutil.ReadFile("peers.json")
		if err != nil {
			fmt.Println("피어 파일 읽기 오류:", err)
			return
		}
		if err := json.Unmarshal(data, &peers); err != nil {
			fmt.Println("피어 데이터 파싱 오류:", err)
			return
		}
		fmt.Println("피어 목록이 로드되었습니다:", peers)
	}
}

// 피어 목록 저장 함수
func savePeers() {
	peersMutex.Lock()
	defer peersMutex.Unlock()

	data, err := json.MarshalIndent(peers, "", "  ")
	if err != nil {
		fmt.Println("피어 저장 중 오류 발생:", err)
		return
	}
	err = ioutil.WriteFile("peers.json", data, 0644)
	if err != nil {
		fmt.Println("피어 파일 쓰기 오류:", err)
	}
}

// 피어 추가 함수
func addPeer(peer string) {
	peersMutex.Lock()
	defer peersMutex.Unlock()

	for _, p := range peers {
		if p == peer {
			return // 이미 존재하는 피어
		}
	}
	peers = append(peers, peer)
	savePeers()
	fmt.Println("피어가 추가되었습니다:", peer)
}

// 피어 제거 함수
func removePeer(peer string) {
	peersMutex.Lock()
	defer peersMutex.Unlock()

	for i, p := range peers {
		if p == peer {
			peers = append(peers[:i], peers[i+1:]...)
			break
		}
	}
	savePeers()
	fmt.Println("피어가 제거되었습니다:", peer)
}

// 피어 간 블록체인 동기화
func syncBlockchain() {
	for _, peer := range peers {
		resp, err := http.Get(peer + "/blocks")
		if err != nil {
			fmt.Println("블록체인 동기화 실패:", err)
			continue
		}
		defer resp.Body.Close()

		var peerBlockchain []Block
		if err := json.NewDecoder(resp.Body).Decode(&peerBlockchain); err != nil {
			fmt.Println("피어 블록체인 데이터 파싱 실패:", err)
			continue
		}

		mutex.Lock()
		if len(peerBlockchain) > len(Blockchain) && isBlockchainValidChain(peerBlockchain) {
			Blockchain = peerBlockchain
			fmt.Println("블록체인이 동기화되었습니다. 새로운 체인:", Blockchain)
			saveBlockchain()
		}
		mutex.Unlock()
	}
}

// 피어 블록체인이 유효한지 검사
func isBlockchainValidChain(chain []Block) bool {
	for i := 1; i < len(chain); i++ {
		if !isBlockValid(chain[i], chain[i-1]) {
			return false
		}
	}
	return true
}

// 피어에게 블록 전파
func broadcastBlock(block Block) {
	for _, peer := range peers {
		go func(peer string) {
			url := peer + "/blocks/create"

			// 요청 바디를 구조체로 정의하여 JSON으로 마샬링
			payload := map[string][]Transaction{
				"transactions": block.Transactions,
			}
			jsonData, err := json.Marshal(payload)
			if err != nil {
				fmt.Println("블록 전파 중 JSON 마샬링 오류:", err)
				return
			}

			resp, err := http.Post(url, "application/json", bytes.NewReader(jsonData)) // jsonData 사용
			if err != nil {
				fmt.Println("블록 전파 실패:", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				body, _ := io.ReadAll(resp.Body)
				fmt.Printf("블록 전파 실패: %s\n", string(body))
			}
		}(peer)
	}
}

func main() {
	// 블록체인 로드
	loadBlockchain()

	// 피어 목록 로드
	loadPeers()

	// 블록체인에 제네시스 블록이 없다면 생성
	if len(Blockchain) == 0 {
		genesisBlock := createGenesisBlock()
		Blockchain = append(Blockchain, genesisBlock)
		fmt.Println("제네시스 블록이 생성되었습니다.")
		printBlockchain()
		saveBlockchain()
	}

	// 블록체인 출력 (콘솔용)
	printBlockchain()

	// 무결성 검사 고루틴 시작
	go monitorBlockchain()

	// REST API 라우팅 설정
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/blocks", getBlockchain)
	http.HandleFunc("/blocks/", getBlock)
	http.HandleFunc("/blocks/create", createBlock)
	http.HandleFunc("/ws", handleWebSocketConnection)

	// 피어 추가 엔드포인트 (간단한 예)
	http.HandleFunc("/peers/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST 요청만 가능합니다.", http.StatusMethodNotAllowed)
			return
		}
		var peer struct {
			Peer string `json:"peer"`
		}
		if err := json.NewDecoder(r.Body).Decode(&peer); err != nil || peer.Peer == "" {
			http.Error(w, "유효한 피어 주소가 필요합니다", http.StatusBadRequest)
			return
		}
		addPeer(peer.Peer)
		// 피어에게 블록체인 요청
		go syncBlockchain()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("피어 추가 성공"))
	})

	// 웹 인터페이스용 정적 파일 서빙
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// 서버 시작
	fmt.Println("블록체인 서버가 시작되었습니다. http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("서버 시작 실패:", err)
	}
}
