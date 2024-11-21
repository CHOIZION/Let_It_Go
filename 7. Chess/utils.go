// utils.go
package main

import (
	"errors"
	"strconv"
	"strings"
)

// parsePosition은 체스 위치 문자열을 보드 인덱스로 변환합니다.
func parsePosition(pos string) (int, int, error) {
	if len(pos) != 2 {
		return 0, 0, errors.New("위치 입력이 올바르지 않습니다")
	}

	cols := "abcdefgh"
	x := strings.Index(cols, string(pos[0]))
	if x == -1 {
		return 0, 0, errors.New("열 값이 올바르지 않습니다")
	}

	y, err := strconv.Atoi(string(pos[1]))
	if err != nil || y < 1 || y > 8 {
		return 0, 0, errors.New("행 값이 올바르지 않습니다")
	}

	return x, y - 1, nil
}
