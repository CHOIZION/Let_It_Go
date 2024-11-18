package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// 서버에 연결
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("서버 연결 실패:", err)
		return
	}
	defer conn.Close()
	fmt.Println("서버에 연결되었습니다. 메시지를 입력하세요.")

	// 고루틴으로 수신 메시지 처리
	go func() {
		reader := bufio.NewReader(conn)
		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("서버 연결 종료")
				os.Exit(0)
			}
			fmt.Print(message)
		}
	}()

	// 사용자 입력 처리
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "exit" {
			fmt.Println("연결을 종료합니다.")
			break
		}
		// 서버로 메시지 전송
		conn.Write([]byte(text + "\n"))
	}
}
