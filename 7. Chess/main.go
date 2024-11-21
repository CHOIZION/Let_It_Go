// main.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	game := NewGame()
	reader := bufio.NewReader(os.Stdin)

	for {
		game.Board.PrintBoard()
		if game.Finished {
			fmt.Printf("게임이 종료되었습니다. 승자: %s\n", game.CurrentTurn)
			break
		}

		fmt.Printf("%s의 턴입니다. 이동할 위치를 입력하세요 (예: e2 e4): ", game.CurrentTurn)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		parts := strings.Split(input, " ")
		if len(parts) != 2 {
			fmt.Println("입력이 올바르지 않습니다. 다시 시도하세요.")
			continue
		}

		fromX, fromY, err := parsePosition(parts[0])
		if err != nil {
			fmt.Println("출발 위치 오류:", err)
			continue
		}

		toX, toY, err := parsePosition(parts[1])
		if err != nil {
			fmt.Println("도착 위치 오류:", err)
			continue
		}

		err = game.MakeMove(fromX, fromY, toX, toY)
		if err != nil {
			fmt.Println("오류:", err)
			continue
		}
	}
}
