package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Playlist struct {
	Name  string
	Songs []string
}

func NewPlaylist(name string) *Playlist {
	playlist := &Playlist{
		Name:  name,
		Songs: []string{},
	}
	playlist.addSongs()
	return playlist
}

func (p *Playlist) addSongs() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("플레이리스트 '%s'에 추가할 음악 파일 경로를 입력하세요 (완료하려면 엔터): ", p.Name)
		scanner.Scan()
		path := scanner.Text()
		if strings.TrimSpace(path) == "" {
			break
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Println("파일이 존재하지 않습니다.")
			continue
		}
		p.Songs = append(p.Songs, path)
		fmt.Println("음악이 추가되었습니다.")
	}
}
