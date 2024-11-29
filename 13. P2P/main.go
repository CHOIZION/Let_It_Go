package main

import (
	"bufio"
	"crypto/sha256"
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileMeta struct {
	FileName string
	FileSize int64
	FileHash [32]byte
	Category string
}

type Peer struct {
	Address string
	Files   []FileMeta
}

var (
	peers      = make(map[string]Peer)
	peersMutex sync.Mutex
	sharedDir  = "shared"
	myAddress  = ""
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("사용법: go run main.go [포트번호]")
		os.Exit(1)
	}

	port := os.Args[1]
	myAddress = GetLocalIP() + ":" + port

	if _, err := os.Stat(sharedDir); os.IsNotExist(err) {
		os.Mkdir(sharedDir, os.ModePerm)
	}

	go startServer(port)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("\n명령을 입력하세요 (connect, list, search, download, upload, delete, exit):")
		fmt.Print("> ")
		scanner.Scan()
		command := scanner.Text()
		args := strings.Split(command, " ")

		switch args[0] {
		case "connect":
			if len(args) != 2 {
				fmt.Println("사용법: connect [IP:포트]")
				continue
			}
			connectToPeer(args[1])
		case "list":
			listPeers()
		case "search":
			if len(args) < 2 {
				fmt.Println("사용법: search [키워드]")
				continue
			}
			keyword := strings.Join(args[1:], " ")
			searchFiles(keyword)
		case "download":
			if len(args) != 2 {
				fmt.Println("사용법: download [파일이름]")
				continue
			}
			downloadFile(args[1])
		case "upload":
			if len(args) != 2 {
				fmt.Println("사용법: upload [파일경로]")
				continue
			}
			uploadFile(args[1])
		case "delete":
			if len(args) != 2 {
				fmt.Println("사용법: delete [파일이름]")
				continue
			}
			deleteFile(args[1])
		case "exit":
			fmt.Println("프로그램을 종료합니다.")
			os.Exit(0)
		default:
			fmt.Println("알 수 없는 명령입니다.")
		}
	}
}

func startServer(port string) {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		fmt.Println("인증서 로드 실패:", err)
		os.Exit(1)
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	ln, err := tls.Listen("tcp", ":"+port, config)
	if err != nil {
		fmt.Println("서버 시작 실패:", err)
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Println("서버가 시작되었습니다. 포트:", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("연결 수락 실패:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	var request string
	if err := decoder.Decode(&request); err != nil {
		fmt.Println("요청 디코딩 실패:", err)
		return
	}

	switch request {
	case "PEER_INFO":
		sendPeerInfo(encoder)
	case "FILE_REQUEST":
		var fileName string
		if err := decoder.Decode(&fileName); err != nil {
			fmt.Println("파일 이름 디코딩 실패:", err)
			return
		}
		sendFile(conn, fileName)
	case "FILE_UPLOAD":
		receiveFile(conn)
	default:
		fmt.Println("알 수 없는 요청:", request)
	}
}

func sendPeerInfo(encoder *gob.Encoder) {
	files := getSharedFiles()
	peerInfo := Peer{
		Address: myAddress,
		Files:   files,
	}
	if err := encoder.Encode(peerInfo); err != nil {
		fmt.Println("피어 정보 인코딩 실패:", err)
	}
}

func getSharedFiles() []FileMeta {
	var files []FileMeta
	fileList, _ := filepath.Glob(filepath.Join(sharedDir, "*"))
	for _, file := range fileList {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		hash, _ := getFileHash(file)
		category := getCategory(file)
		files = append(files, FileMeta{
			FileName: filepath.Base(file),
			FileSize: info.Size(),
			FileHash: hash,
			Category: category,
		})
	}
	return files
}

func getFileHash(filePath string) ([32]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return [32]byte{}, err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return [32]byte{}, err
	}
	var hash [32]byte
	copy(hash[:], hasher.Sum(nil))
	return hash, nil
}

func getCategory(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".png", ".gif":
		return "Images"
	case ".mp4", ".avi", ".mkv":
		return "Videos"
	case ".mp3", ".wav", ".flac":
		return "Music"
	case ".txt", ".pdf", ".docx":
		return "Documents"
	default:
		return "Others"
	}
}

func connectToPeer(address string) {
	conn, err := tls.Dial("tcp", address, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		fmt.Println("피어 연결 실패:", err)
		return
	}
	defer conn.Close()

	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	if err := encoder.Encode("PEER_INFO"); err != nil {
		fmt.Println("요청 전송 실패:", err)
		return
	}

	var peerInfo Peer
	if err := decoder.Decode(&peerInfo); err != nil {
		fmt.Println("피어 정보 수신 실패:", err)
		return
	}

	peersMutex.Lock()
	peers[address] = peerInfo
	peersMutex.Unlock()

	fmt.Println("피어에 연결되었습니다:", address)
}

func listPeers() {
	peersMutex.Lock()
	defer peersMutex.Unlock()

	if len(peers) == 0 {
		fmt.Println("연결된 피어가 없습니다.")
		return
	}

	fmt.Println("\n연결된 피어 목록:")
	for addr, peer := range peers {
		fmt.Println("- 주소:", addr, "| 공유 파일 수:", len(peer.Files))
	}
}

func searchFiles(keyword string) {
	peersMutex.Lock()
	defer peersMutex.Unlock()

	if len(peers) == 0 {
		fmt.Println("연결된 피어가 없습니다.")
		return
	}

	fmt.Printf("\n'%s'에 대한 검색 결과:\n", keyword)
	found := false
	for _, peer := range peers {
		for _, file := range peer.Files {
			if strings.Contains(strings.ToLower(file.FileName), strings.ToLower(keyword)) {
				fmt.Printf("- %s [%s] (from %s)\n", file.FileName, file.Category, peer.Address)
				found = true
			}
		}
	}
	if !found {
		fmt.Println("검색 결과가 없습니다.")
	}
}

func downloadFile(fileName string) {
	peersMutex.Lock()
	defer peersMutex.Unlock()

	var targetPeer *Peer
	var targetFile *FileMeta
	for _, peer := range peers {
		for _, file := range peer.Files {
			if file.FileName == fileName {
				targetPeer = &peer
				targetFile = &file
				break
			}
		}
		if targetPeer != nil {
			break
		}
	}

	if targetPeer == nil || targetFile == nil {
		fmt.Println("파일을 찾을 수 없습니다.")
		return
	}

	conn, err := tls.Dial("tcp", targetPeer.Address, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		fmt.Println("피어 연결 실패:", err)
		return
	}
	defer conn.Close()

	encoder := gob.NewEncoder(conn)
	// decoder 선언 제거 (사용되지 않음)
	// decoder := gob.NewDecoder(conn)

	if err := encoder.Encode("FILE_REQUEST"); err != nil {
		fmt.Println("파일 요청 전송 실패:", err)
		return
	}

	if err := encoder.Encode(fileName); err != nil {
		fmt.Println("파일 이름 전송 실패:", err)
		return
	}

	tempFilePath := filepath.Join(sharedDir, fileName+".tmp")
	file, err := os.Create(tempFilePath)
	if err != nil {
		fmt.Println("파일 생성 실패:", err)
		return
	}
	defer file.Close()

	// 파일 다운로드 진행률 표시 기능 추가
	fmt.Println("다운로드 중...")
	buffer := make([]byte, 1024)
	var totalReceived int64 = 0
	for {
		n, err := conn.Read(buffer)
		if n > 0 {
			file.Write(buffer[:n])
			totalReceived += int64(n)
			percentage := float64(totalReceived) / float64(targetFile.FileSize) * 100
			fmt.Printf("\r진행률: %.2f%%", percentage)
		}
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println("\n데이터 수신 오류:", err)
				os.Remove(tempFilePath)
				return
			}
		}
	}
	fmt.Println("\n다운로드 완료")

	hash, err := getFileHash(tempFilePath)
	if err != nil {
		fmt.Println("해시 계산 실패:", err)
		os.Remove(tempFilePath)
		return
	}

	if hash != targetFile.FileHash {
		fmt.Println("파일 무결성 검증 실패")
		os.Remove(tempFilePath)
		return
	}

	finalFilePath := filepath.Join(sharedDir, fileName)
	os.Rename(tempFilePath, finalFilePath)

	fmt.Println("파일 다운로드 완료:", fileName)
}

func uploadFile(filePath string) {
	info, err := os.Stat(filePath)
	if err != nil {
		fmt.Println("파일 경로가 잘못되었습니다.")
		return
	}

	destPath := filepath.Join(sharedDir, info.Name())
	if filePath != destPath {
		_, err = copyFile(filePath, destPath)
		if err != nil {
			fmt.Println("파일 업로드 실패:", err)
			return
		}
	}

	fmt.Println("파일이 공유 디렉토리에 업로드되었습니다:", info.Name())
}

func deleteFile(fileName string) {
	filePath := filepath.Join(sharedDir, fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println("해당 파일이 존재하지 않습니다.")
		return
	}

	err := os.Remove(filePath)
	if err != nil {
		fmt.Println("파일 삭제 실패:", err)
		return
	}

	fmt.Println("파일이 삭제되었습니다:", fileName)
}

func sendFile(conn net.Conn, fileName string) {
	filePath := filepath.Join(sharedDir, fileName)
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("파일 열기 실패:", err)
		return
	}
	defer file.Close()

	_, err = io.Copy(conn, file)
	if err != nil {
		fmt.Println("파일 전송 실패:", err)
	}
}

func receiveFile(conn net.Conn) {
	decoder := gob.NewDecoder(conn)

	var fileName string
	if err := decoder.Decode(&fileName); err != nil {
		fmt.Println("파일 이름 디코딩 실패:", err)
		return
	}

	filePath := filepath.Join(sharedDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("파일 생성 실패:", err)
		return
	}
	defer file.Close()

	// conn을 사용하여 파일 수신
	_, err = io.Copy(file, conn)
	if err != nil {
		fmt.Println("파일 수신 실패:", err)
		os.Remove(filePath)
		return
	}

	fmt.Println("파일 업로드 완료:", fileName)
}

func copyFile(src, dst string) (int64, error) {
	sourceFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destFile.Close()

	nBytes, err := io.Copy(destFile, sourceFile)
	return nBytes, err
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}
