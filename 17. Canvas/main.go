package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	player := NewPlayer()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n=== 터미널 음악 플레이어 ===")
		fmt.Println("1. 플레이리스트 생성")
		fmt.Println("2. 음악 재생")
		fmt.Println("3. 재생 제어")
		fmt.Println("4. 종료")
		fmt.Print("메뉴를 선택하세요: ")

		scanner.Scan()
		choice := scanner.Text()

		switch choice {
		case "1":
			fmt.Print("플레이리스트 이름을 입력하세요: ")
			scanner.Scan()
			name := scanner.Text()
			player.CreatePlaylist(name)
		case "2":
			if len(player.Playlists) == 0 {
				fmt.Println("플레이리스트가 없습니다. 먼저 플레이리스트를 생성하세요.")
				continue
			}
			fmt.Println("플레이리스트 목록:")
			for i, playlist := range player.Playlists {
				fmt.Printf("%d. %s\n", i+1, playlist.Name)
			}
			fmt.Print("재생할 플레이리스트 번호를 선택하세요: ")
			scanner.Scan()
			index := parseIndex(scanner.Text(), len(player.Playlists))
			if index == -1 {
				fmt.Println("잘못된 번호입니다.")
				continue
			}
			player.PlayPlaylist(index)
		case "3":
			if player.CurrentPlaylist == nil {
				fmt.Println("재생 중인 음악이 없습니다.")
				continue
			}
			fmt.Println("1. 일시 정지/재생")
			fmt.Println("2. 다음 곡")
			fmt.Println("3. 이전 곡")
			fmt.Print("선택하세요: ")
			scanner.Scan()
			controlChoice := scanner.Text()
			switch controlChoice {
			case "1":
				player.TogglePause()
			case "2":
				player.Next()
			case "3":
				player.Previous()
			default:
				fmt.Println("잘못된 선택입니다.")
			}
		case "4":
			player.Stop()
			fmt.Println("프로그램을 종료합니다.")
			os.Exit(0)
		default:
			fmt.Println("잘못된 선택입니다.")
		}
	}
}

func parseIndex(input string, max int) int {
	var index int
	_, err := fmt.Sscanf(input, "%d", &index)
	if err != nil || index < 1 || index > max {
		return -1
	}
	return index - 1
}
