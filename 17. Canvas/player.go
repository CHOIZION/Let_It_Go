package main

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v3"
)

type Player struct {
	Playlists       []*Playlist
	CurrentPlaylist *Playlist
	currentIndex    int
	context         *oto.Context
	player          oto.Player
	isPaused        bool
	isPlaying       bool
	mutex           sync.Mutex
}

func NewPlayer() *Player {
	return &Player{
		Playlists: []*Playlist{},
	}
}

func (p *Player) CreatePlaylist(name string) {
	playlist := NewPlaylist(name)
	p.Playlists = append(p.Playlists, playlist)
	fmt.Printf("플레이리스트 '%s'가 생성되었습니다.\n", name)
}

func (p *Player) PlayPlaylist(index int) {
	p.CurrentPlaylist = p.Playlists[index]
	if len(p.CurrentPlaylist.Songs) == 0 {
		fmt.Println("플레이리스트에 음악이 없습니다. 음악을 추가하세요.")
		return
	}
	p.currentIndex = 0
	p.playSong(p.currentIndex)
}

func (p *Player) playSong(index int) {
	p.Stop() // 이전 재생 중인 음악이 있으면 정지

	song := p.CurrentPlaylist.Songs[index]
	file, err := os.Open(song)
	if err != nil {
		fmt.Println("음악 파일을 열 수 없습니다:", err)
		return
	}

	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		fmt.Println("MP3 디코더를 생성할 수 없습니다:", err)
		return
	}

	sampleRate := decoder.SampleRate()
	context, ready, err := oto.NewContext(sampleRate, 2, 2)
	if err != nil {
		fmt.Println("오디오 컨텍스트를 생성할 수 없습니다:", err)
		return
	}
	<-ready

	p.context = context
	p.player = context.NewPlayer(decoder)
	p.isPaused = false
	p.isPlaying = true

	fmt.Printf("재생 중: %s\n", song)

	go func() {
		defer file.Close()
		defer p.player.Close()
		_, err := io.Copy(p.player, decoder)
		if err != nil {
			fmt.Println("음악 재생 중 오류가 발생했습니다:", err)
		}
		p.mutex.Lock()
		p.isPlaying = false
		p.mutex.Unlock()
		// 다음 곡으로 자동 이동
		p.Next()
	}()
}

func (p *Player) TogglePause() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if !p.isPlaying {
		fmt.Println("재생 중인 음악이 없습니다.")
		return
	}
	p.isPaused = !p.isPaused
	if p.isPaused {
		p.player.Pause()
		fmt.Println("일시 정지되었습니다.")
	} else {
		p.player.Play()
		fmt.Println("재생됩니다.")
	}
}

func (p *Player) Next() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.CurrentPlaylist == nil {
		fmt.Println("재생 중인 플레이리스트가 없습니다.")
		return
	}
	p.currentIndex++
	if p.currentIndex >= len(p.CurrentPlaylist.Songs) {
		fmt.Println("마지막 곡입니다.")
		p.currentIndex = len(p.CurrentPlaylist.Songs) - 1
		return
	}
	fmt.Println("다음 곡으로 이동합니다.")
	p.playSong(p.currentIndex)
}

func (p *Player) Previous() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.CurrentPlaylist == nil {
		fmt.Println("재생 중인 플레이리스트가 없습니다.")
		return
	}
	if p.currentIndex == 0 {
		fmt.Println("첫 번째 곡입니다.")
		return
	}
	p.currentIndex--
	fmt.Println("이전 곡으로 이동합니다.")
	p.playSong(p.currentIndex)
}

func (p *Player) Stop() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.player != nil {
		p.player.Close()
	}
	if p.context != nil {
		p.context.Close()
	}
	p.isPlaying = false
}
