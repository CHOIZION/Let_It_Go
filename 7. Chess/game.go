// game.go
package main

import (
	"errors"
)

type Game struct {
	Board       *Board
	CurrentTurn Color
	Finished    bool
}

// NewGame은 새로운 체스 게임을 생성합니다.
func NewGame() *Game {
	return &Game{
		Board:       NewBoard(),
		CurrentTurn: White,
		Finished:    false,
	}
}

// MakeMove는 기물을 이동하고 게임 상태를 업데이트합니다.
func (g *Game) MakeMove(fromX, fromY, toX, toY int) error {
	if g.Finished {
		return errors.New("게임이 종료되었습니다")
	}

	piece := g.Board[fromY][fromX]
	if piece == nil {
		return errors.New("선택한 위치에 기물이 없습니다")
	}
	if piece.Color != g.CurrentTurn {
		return errors.New("현재 턴의 플레이어가 아닙니다")
	}

	err := g.Board.MovePiece(fromX, fromY, toX, toY)
	if err != nil {
		return err
	}

	// 킹이 잡혔는지 확인하여 게임 종료 처리 (간단한 예시)
	if g.isKingCaptured() {
		g.Finished = true
	} else {
		// 턴 변경
		if g.CurrentTurn == White {
			g.CurrentTurn = Black
		} else {
			g.CurrentTurn = White
		}
	}
	return nil
}

// isKingCaptured는 킹이 잡혔는지 확인합니다.
func (g *Game) isKingCaptured() bool {
	hasWhiteKing := false
	hasBlackKing := false

	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			piece := g.Board[y][x]
			if piece != nil && piece.Type == King {
				if piece.Color == White {
					hasWhiteKing = true
				} else {
					hasBlackKing = true
				}
			}
		}
	}

	if !hasWhiteKing || !hasBlackKing {
		return true
	}
	return false
}
