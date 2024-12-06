package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type LSystem struct {
	Axiom string
	Rules map[rune]string
	Angle float64
}

var builtInFractals = []struct {
	Name   string
	System LSystem
}{
	{
		Name: "Koch Curve",
		System: LSystem{
			Axiom: "F",
			Rules: map[rune]string{'F': "F+F-F-F+F"},
			Angle: 90.0,
		},
	},
	{
		Name: "Dragon Curve",
		System: LSystem{
			Axiom: "F",
			Rules: map[rune]string{'F': "F+G", 'G': "F-G"},
			Angle: 90.0,
		},
	},
	{
		Name: "Sierpinski Triangle",
		System: LSystem{
			Axiom: "F-G-G",
			Rules: map[rune]string{
				'F': "F-G+F+G-F",
				'G': "GG",
			},
			Angle: 120.0,
		},
	},
	{
		Name: "Hilbert Curve",
		System: LSystem{
			Axiom: "A",
			Rules: map[rune]string{
				'A': "-BF+AFA+FB-",
				'B': "+AF-BFB-FA+",
				// F, +, -는 동작 동일
			},
			Angle: 90.0,
		},
	},
}

// 전역 상태
var currentLSystem LSystem
var iterationCount int = 1
var stepSize float64 = 1.0
var drawChar rune = '#'
var lastRendered [][]rune

func (ls *LSystem) Generate(iterations int) string {
	str := ls.Axiom
	for i := 0; i < iterations; i++ {
		var newStr strings.Builder
		for _, ch := range str {
			if rep, ok := ls.Rules[ch]; ok {
				newStr.WriteString(rep)
			} else {
				newStr.WriteRune(ch)
			}
		}
		str = newStr.String()
	}
	return str
}

func DrawLSystem(str string, angle float64) {
	x, y := 0.0, 0.0
	dir := 0.0 // 0도 우방향
	points := []struct{ X, Y float64 }{{X: x, Y: y}}

	for _, ch := range str {
		switch ch {
		case 'F', 'G', 'A', 'B':
			// 전진
			rad := dir * math.Pi / 180.0
			x += math.Cos(rad) * stepSize
			y += math.Sin(rad) * stepSize
			points = append(points, struct{ X, Y float64 }{X: x, Y: y})
		case '+':
			// 왼회전
			dir += angle
		case '-':
			// 오른회전
			dir -= angle
		default:
			// 무시
		}
	}

	minX, maxX := points[0].X, points[0].X
	minY, maxY := points[0].Y, points[0].Y
	for _, p := range points {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	offsetX := 0 - int(math.Floor(minX))
	offsetY := 0 - int(math.Floor(minY))

	width := int(math.Ceil(maxX-minX)) + 1
	height := int(math.Ceil(maxY-minY)) + 1

	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	canvas := make([][]rune, height)
	for i := 0; i < height; i++ {
		canvas[i] = make([]rune, width)
		for j := 0; j < width; j++ {
			canvas[i][j] = ' '
		}
	}

	for _, p := range points {
		X := int(math.Round(p.X)) + offsetX
		Y := int(math.Round(p.Y)) + offsetY
		if X >= 0 && X < width && Y >= 0 && Y < height {
			canvas[height-1-Y][X] = drawChar
		}
	}

	// 출력
	for _, row := range canvas {
		fmt.Println(string(row))
	}

	// 마지막 렌더링 결과 저장
	lastRendered = canvas
}

func saveToFile(filename string) {
	if lastRendered == nil {
		fmt.Println("아직 렌더링된 패턴이 없습니다. 먼저 run 명령을 실행하세요.")
		return
	}
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("파일 생성 중 오류:", err)
		return
	}
	defer file.Close()

	for _, row := range lastRendered {
		file.WriteString(string(row) + "\n")
	}
	fmt.Println("파일에 저장되었습니다.")
}

func defineCustomLSystem(scanner *bufio.Scanner) {
	fmt.Println("사용자 정의 L-시스템을 설정합니다.")
	fmt.Print("Axiom을 입력하세요: ")
	scanner.Scan()
	axiom := scanner.Text()
	fmt.Print("각도(도 단위)를 입력하세요 (예: 90): ")
	scanner.Scan()
	angleStr := scanner.Text()
	angle, err := strconv.ParseFloat(angleStr, 64)
	if err != nil {
		fmt.Println("각도 설정 오류. 기본 90도로 설정합니다.")
		angle = 90.0
	}

	fmt.Print("규칙 개수를 입력하세요: ")
	scanner.Scan()
	ruleCountStr := scanner.Text()
	ruleCount, err := strconv.Atoi(ruleCountStr)
	if err != nil || ruleCount < 0 {
		fmt.Println("규칙 개수가 잘못되었습니다. 0개로 처리합니다.")
		ruleCount = 0
	}

	rules := make(map[rune]string)
	for i := 0; i < ruleCount; i++ {
		fmt.Printf("규칙 %d: <문자> <치환문자열> 형태로 입력: ", i+1)
		scanner.Scan()
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			fmt.Println("규칙 입력 실패. 이 규칙은 무시합니다.")
			continue
		}
		ch := []rune(parts[0])
		if len(ch) != 1 {
			fmt.Println("한 글자여야 합니다. 이 규칙 무시.")
			continue
		}
		rules[ch[0]] = parts[1]
	}

	currentLSystem = LSystem{
		Axiom: axiom,
		Rules: rules,
		Angle: angle,
	}
	fmt.Println("사용자 정의 L-시스템 설정 완료.")
}

func main() {
	rand.Seed(time.Now().UnixNano())
	// 초기값: Koch Curve로 시작
	currentLSystem = builtInFractals[0].System

	fmt.Println("=== 터미널 기반 L-시스템 프랙탈 ASCII 생성기 ===")
	fmt.Println("명령어:")
	fmt.Println(" list : 사용 가능한 내장 프랙탈 목록 표시")
	fmt.Println(" fractal <번호> : 해당 번호의 내장 프랙탈로 변경")
	fmt.Println(" custom : 사용자 정의 L-시스템 설정 모드")
	fmt.Println(" angle <float> : 각도 변경")
	fmt.Println(" iter <int> : 반복 횟수 변경")
	fmt.Println(" set step <int> : 전진 거리 변경(기본 1)")
	fmt.Println(" set char <c> : 그릴 문자 변경(기본 '#')")
	fmt.Println(" run : 현재 설정으로 프랙탈 생성 및 렌더링")
	fmt.Println(" save <filename> : 마지막 렌더링 결과 파일로 저장")
	fmt.Println(" exit : 종료")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(">> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]
		switch cmd {
		case "list":
			fmt.Println("내장 프랙탈 목록:")
			for i, f := range builtInFractals {
				fmt.Printf("%d. %s\n", i+1, f.Name)
			}
		case "fractal":
			if len(parts) < 2 {
				fmt.Println("사용법: fractal <번호>")
				continue
			}
			idx, err := strconv.Atoi(parts[1])
			if err != nil || idx < 1 || idx > len(builtInFractals) {
				fmt.Println("유효한 번호를 입력하세요.")
				continue
			}
			currentLSystem = builtInFractals[idx-1].System
			fmt.Printf("%s 프랙탈로 변경되었습니다.\n", builtInFractals[idx-1].Name)
		case "custom":
			defineCustomLSystem(scanner)
		case "angle":
			if len(parts) < 2 {
				fmt.Println("사용법: angle <float>")
				continue
			}
			a, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				fmt.Println("유효한 실수를 입력하세요.")
				continue
			}
			currentLSystem.Angle = a
			fmt.Printf("각도를 %.2f도로 설정.\n", a)
		case "iter":
			if len(parts) < 2 {
				fmt.Println("사용법: iter <int>")
				continue
			}
			it, err := strconv.Atoi(parts[1])
			if err != nil || it < 0 {
				fmt.Println("유효한 0 이상의 정수를 입력하세요.")
				continue
			}
			iterationCount = it
			fmt.Printf("반복 횟수를 %d로 설정.\n", it)
		case "set":
			if len(parts) < 3 {
				fmt.Println("사용법: set step <int> 또는 set char <c>")
				continue
			}
			if parts[1] == "step" {
				st, err := strconv.Atoi(parts[2])
				if err != nil || st <= 0 {
					fmt.Println("유효한 양의 정수를 입력하세요.")
					continue
				}
				stepSize = float64(st)
				fmt.Printf("전진 거리를 %d로 설정.\n", st)
			} else if parts[1] == "char" {
				chRunes := []rune(parts[2])
				if len(chRunes) != 1 {
					fmt.Println("한 글자 문자를 입력하세요.")
					continue
				}
				drawChar = chRunes[0]
				fmt.Printf("그릴 문자를 '%c'로 설정.\n", drawChar)
			} else {
				fmt.Println("사용법: set step <int> 또는 set char <c>")
			}
		case "run":
			fmt.Println("패턴을 생성 중...")
			str := currentLSystem.Generate(iterationCount)
			fmt.Println("ASCII로 렌더링 중...")
			DrawLSystem(str, currentLSystem.Angle)
			fmt.Println("완료!")
		case "save":
			if len(parts) < 2 {
				fmt.Println("사용법: save <filename>")
				continue
			}
			saveToFile(parts[1])
		case "exit":
			fmt.Println("프로그램을 종료합니다.")
			return
		default:
			fmt.Println("알 수 없는 명령어입니다.")
		}
	}
}
