// board.go
package main

import (
	"errors"
	"fmt"
	"strings"
)

// PieceType은 체스 기물의 종류를 나타냅니다.
type PieceType string

const (
	Pawn   PieceType = "Pawn"
	Knight PieceType = "Knight"
	Bishop PieceType = "Bishop"
	Rook   PieceType = "Rook"
	Queen  PieceType = "Queen"
	King   PieceType = "King"
)

// Color는 기물의 색상을 나타냅니다.
type Color string

const (
	White Color = "White"
	Black Color = "Black"
)

// Piece는 체스 기물을 나타냅니다.
type Piece struct {
	Type  PieceType
	Color Color
}

// Board는 체스 보드를 나타냅니다.
type Board [8][8]*Piece

// NewBoard는 초기 체스 보드를 생성합니다.
func NewBoard() *Board {
	board := &Board{}

	// 백색 주요 기물 배치
	board[0][0] = &Piece{Type: Rook, Color: White}
	board[0][1] = &Piece{Type: Knight, Color: White}
	board[0][2] = &Piece{Type: Bishop, Color: White}
	board[0][3] = &Piece{Type: Queen, Color: White}
	board[0][4] = &Piece{Type: King, Color: White}
	board[0][5] = &Piece{Type: Bishop, Color: White}
	board[0][6] = &Piece{Type: Knight, Color: White}
	board[0][7] = &Piece{Type: Rook, Color: White}

	// 백색 폰 배치
	for i := 0; i < 8; i++ {
		board[1][i] = &Piece{Type: Pawn, Color: White}
	}

	// 흑색 주요 기물 배치
	board[7][0] = &Piece{Type: Rook, Color: Black}
	board[7][1] = &Piece{Type: Knight, Color: Black}
	board[7][2] = &Piece{Type: Bishop, Color: Black}
	board[7][3] = &Piece{Type: Queen, Color: Black}
	board[7][4] = &Piece{Type: King, Color: Black}
	board[7][5] = &Piece{Type: Bishop, Color: Black}
	board[7][6] = &Piece{Type: Knight, Color: Black}
	board[7][7] = &Piece{Type: Rook, Color: Black}

	// 흑색 폰 배치
	for i := 0; i < 8; i++ {
		board[6][i] = &Piece{Type: Pawn, Color: Black}
	}

	return board
}

// MovePiece는 기물을 이동합니다.
func (b *Board) MovePiece(fromX, fromY, toX, toY int) error {
	piece := b[fromY][fromX]
	if piece == nil {
		return errors.New("선택한 위치에 기물이 없습니다")
	}

	// 이동 규칙 검사 (간단한 예시)
	valid, err := b.isValidMove(piece, fromX, fromY, toX, toY)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("유효하지 않은 이동입니다")
	}

	// 상대 기물 캡처
	targetPiece := b[toY][toX]
	if targetPiece != nil && targetPiece.Color == piece.Color {
		return errors.New("자신의 기물이 있는 위치로 이동할 수 없습니다")
	}

	// 이동 수행
	b[toY][toX] = piece
	b[fromY][fromX] = nil
	return nil
}

// isValidMove는 기물의 이동이 유효한지 검사합니다.
func (b *Board) isValidMove(piece *Piece, fromX, fromY, toX, toY int) (bool, error) {
	dx := toX - fromX
	dy := toY - fromY

	switch piece.Type {
	case Pawn:
		return b.isValidPawnMove(piece, fromX, fromY, toX, toY)
	case Knight:
		if (dx*dx+dy*dy == 5) && (dx != 0 && dy != 0) {
			return true, nil
		}
	case Bishop:
		if abs(dx) == abs(dy) {
			return b.isPathClear(fromX, fromY, toX, toY)
		}
	case Rook:
		if dx == 0 || dy == 0 {
			return b.isPathClear(fromX, fromY, toX, toY)
		}
	case Queen:
		if dx == 0 || dy == 0 || abs(dx) == abs(dy) {
			return b.isPathClear(fromX, fromY, toX, toY)
		}
	case King:
		if abs(dx) <= 1 && abs(dy) <= 1 {
			return true, nil
		}
	default:
		return false, errors.New("알 수 없는 기물입니다")
	}

	return false, nil
}

// isValidPawnMove는 폰의 이동이 유효한지 검사합니다.
func (b *Board) isValidPawnMove(piece *Piece, fromX, fromY, toX, toY int) (bool, error) {
	direction := 1
	startRow := 1
	if piece.Color == Black {
		direction = -1
		startRow = 6
	}

	dx := toX - fromX
	dy := toY - fromY

	// 전진 이동
	if dx == 0 {
		// 한 칸 전진
		if dy == direction && b[toY][toX] == nil {
			return true, nil
		}
		// 처음에 두 칸 전진
		if dy == 2*direction && fromY == startRow && b[fromY+direction][fromX] == nil && b[toY][toX] == nil {
			return true, nil
		}
	}

	// 대각선 공격
	if abs(dx) == 1 && dy == direction {
		targetPiece := b[toY][toX]
		if targetPiece != nil && targetPiece.Color != piece.Color {
			return true, nil
		}
	}

	return false, nil
}

// isPathClear는 두 위치 사이에 기물이 없는지 확인합니다.
func (b *Board) isPathClear(fromX, fromY, toX, toY int) (bool, error) {
	dx := sign(toX - fromX)
	dy := sign(toY - fromY)

	x, y := fromX+dx, fromY+dy
	for x != toX || y != toY {
		if b[y][x] != nil {
			return false, nil
		}
		x += dx
		y += dy
	}
	return true, nil
}

// abs는 정수의 절대값을 반환합니다.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// sign은 정수의 부호를 반환합니다.
func sign(n int) int {
	if n > 0 {
		return 1
	} else if n < 0 {
		return -1
	}
	return 0
}

// PrintBoard는 보드의 현재 상태를 출력합니다.
func (b *Board) PrintBoard() {
	fmt.Println("  a b c d e f g h")
	for y := 7; y >= 0; y-- {
		fmt.Printf("%d ", y+1)
		for x := 0; x < 8; x++ {
			piece := b[y][x]
			if piece == nil {
				fmt.Print(". ")
			} else {
				symbol := pieceSymbol(piece)
				fmt.Print(symbol + " ")
			}
		}
		fmt.Printf("%d\n", y+1)
	}
	fmt.Println("  a b c d e f g h")
}

// pieceSymbol은 기물의 심볼을 반환합니다.
func pieceSymbol(p *Piece) string {
	symbols := map[PieceType]string{
		Pawn:   "p",
		Knight: "n",
		Bishop: "b",
		Rook:   "r",
		Queen:  "q",
		King:   "k",
	}
	symbol := symbols[p.Type]
	if p.Color == White {
		symbol = strings.ToUpper(symbol)
	}
	return symbol
}
