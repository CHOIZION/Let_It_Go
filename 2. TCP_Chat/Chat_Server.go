package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {
	// TCP 서버 시작
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("서버 시작 실패:", err)
		return
	}
	defer listener.Close()
	fmt.Println("채팅 서버가 시작되었습니다. 포트: 8080")

	for {
		// 클라이언트 연결 대기
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("클라이언트 연결 실패:", err)
			continue
		}
		fmt.Println("클라이언트가 연결되었습니다.")

		// 고루틴을 사용해 클라이언트 핸들링
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// 클라이언트로부터 메시지 수신
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("클라이언트 연결 종료")
			break
		}

		// 메시지 출력
		message = strings.TrimSpace(message)
		fmt.Println("클라이언트:", message)

		// 에코 메시지 전송
		conn.Write([]byte("서버: " + message + "\n"))
	}
}
