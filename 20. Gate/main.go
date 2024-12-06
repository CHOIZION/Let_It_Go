package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type GateType int

const (
	GateAND GateType = iota
	GateOR
	GateNOT
	GateNAND
	GateXOR
	GateINPUT
	GateOUTPUT
)

type Gate struct {
	Name   string
	Type   GateType
	Inputs []string
}

type Circuit struct {
	Gates       map[string]*Gate
	InputValues map[string]bool
}

func parseGateType(s string) GateType {
	switch strings.ToUpper(s) {
	case "AND":
		return GateAND
	case "OR":
		return GateOR
	case "NOT":
		return GateNOT
	case "NAND":
		return GateNAND
	case "XOR":
		return GateXOR
	default:
		return GateOUTPUT // 임시, 후에 처리
	}
}

func isInputGate(name string) bool {
	return strings.HasPrefix(name, "INPUT_")
}

func isOutputGate(name string) bool {
	return strings.HasPrefix(name, "OUTPUT_")
}

func loadCircuit(filename string) (*Circuit, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	circuit := &Circuit{
		Gates:       make(map[string]*Gate),
		InputValues: make(map[string]bool),
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid line: %s", line)
		}

		name := parts[0]
		gtypeStr := parts[1]
		var gtype GateType

		if isInputGate(name) {
			// 입력 게이트: 입력 값은 이후 사용자 설정
			gtype = GateINPUT
		} else if isOutputGate(name) {
			// 출력 게이트: 출력 전용, 게이트 유형은 두 번째 필드로
			gtype = parseGateType(gtypeStr)
			if gtype == GateOUTPUT {
				// output 게이트 타입 명시 필요 없으면 기본 and로 두고 아래서 처리
				// 여기서는 output게이트도 특정게이트를 통해 나온다고 가정
				// 만약 OUTPUT_ 게이트가 유형을 명시하도록 한다면:
				gtype = GateOUTPUT
				// parts[1]이 실제 게이트유형이 아닌 경우 이 부분 조정 필요
				// 여기서는 output게이트도 하나의 gate처럼 처리
			}
		} else {
			gtype = parseGateType(gtypeStr)
		}

		inputs := []string{}
		if gtype == GateINPUT {
			// Input gate, no inputs needed
		} else {
			// 일반 게이트나 출력 게이트일 경우
			// Gate Type 뒤에 나오는 것은 모두 입력 게이트 이름
			startIdx := 2
			if isOutputGate(name) {
				// output 게이트도 입력 게이트를 가지므로
				// output게이트의 경우 parts[1]이 실제 게이트 유형을 나타내지 않는다고 했으니,
				// 여기서 reinterpretation 필요
				// 사용자 요구: 너무 복잡해지니 output게이트는 그냥 AND등으로 처리하자
				// 다만 output게이트는 하나의 게이트유형이 아니라 이름 접두사만 output_ 일뿐
				// output게이트도 실제 게이트유형이 주어진다고 가정
				gtype = parseGateType(gtypeStr)
				startIdx = 2
			}
			if len(parts) > startIdx {
				inputs = parts[startIdx:]
			}
		}

		gate := &Gate{
			Name:   name,
			Type:   gtype,
			Inputs: inputs,
		}
		circuit.Gates[name] = gate
	}

	return circuit, nil
}

func evalGate(c *Circuit, gateName string, memo map[string]bool) (bool, bool) {
	if val, ok := memo[gateName]; ok {
		return val, true
	}

	g, ok := c.Gates[gateName]
	if !ok {
		// 없는 게이트
		return false, false
	}

	switch g.Type {
	case GateINPUT:
		val, has := c.InputValues[gateName]
		if !has {
			return false, false
		}
		memo[gateName] = val
		return val, true
	case GateOUTPUT:
		// output 게이트도 사실상 다른 게이트와 동일하게 처리
		// output 게이트가 입력으로 받는 게이트의 출력을 반환
		if len(g.Inputs) == 0 {
			return false, false
		}
		// output게이트는 단일출력이라 가정
		inVal, ok := evalGate(c, g.Inputs[0], memo)
		if !ok {
			return false, false
		}
		memo[gateName] = inVal
		return inVal, true
	case GateAND, GateOR, GateNAND, GateXOR:
		// 2입력 이상 가능하지만 여기서는 2입력 게이트만 가정
		// NOT 게이트 제외 모든 게이트는 최소 2입력이라 가정
		if len(g.Inputs) < 2 {
			return false, false
		}
		valInputs := []bool{}
		for _, inG := range g.Inputs {
			inVal, ok := evalGate(c, inG, memo)
			if !ok {
				return false, false
			}
			valInputs = append(valInputs, inVal)
		}

		var result bool
		switch g.Type {
		case GateAND:
			result = true
			for _, v := range valInputs {
				result = result && v
			}
		case GateOR:
			result = false
			for _, v := range valInputs {
				result = result || v
			}
		case GateNAND:
			result = true
			for _, v := range valInputs {
				result = result && v
			}
			result = !result
		case GateXOR:
			// XOR for multiple inputs: 홀수 개 true 이면 true
			countTrue := 0
			for _, v := range valInputs {
				if v {
					countTrue++
				}
			}
			result = (countTrue%2 == 1)
		}

		memo[gateName] = result
		return result, true
	case GateNOT:
		// NOT 게이트는 단일 입력이라 가정
		if len(g.Inputs) < 1 {
			return false, false
		}
		inVal, ok := evalGate(c, g.Inputs[0], memo)
		if !ok {
			return false, false
		}
		result := !inVal
		memo[gateName] = result
		return result, true
	default:
		return false, false
	}
}

func runCircuit(c *Circuit) {
	// output게이트를 찾아서 값 계산
	// output게이트는 이름이 "OUTPUT_"로 시작하는 게이트
	memo := make(map[string]bool)
	for name := range c.Gates {
		if isOutputGate(name) {
			val, ok := evalGate(c, name, memo)
			if ok {
				fmt.Printf("%s = %v\n", name, val)
			} else {
				fmt.Printf("%s 계산 실패\n", name)
			}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("사용법: go run main.go <회로정의파일>")
		return
	}

	filename := os.Args[1]
	c, err := loadCircuit(filename)
	if err != nil {
		fmt.Println("회로 로드 중 오류 발생:", err)
		return
	}

	fmt.Println("디지털 논리회로 시뮬레이터입니다.")
	fmt.Println("명령어:")
	fmt.Println(" set [입력이름] [0|1] - 입력 신호 설정")
	fmt.Println(" run - 회로 평가 후 출력 표시")
	fmt.Println(" exit - 종료")

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
		case "set":
			if len(parts) < 3 {
				fmt.Println("사용법: set [입력이름] [0|1]")
				continue
			}
			inName := parts[1]
			valStr := parts[2]
			if !strings.HasPrefix(inName, "INPUT_") {
				fmt.Println("입력이름은 'INPUT_'로 시작해야 합니다.")
				continue
			}
			if valStr != "0" && valStr != "1" {
				fmt.Println("값은 0 또는 1이어야 합니다.")
				continue
			}
			c.InputValues[inName] = (valStr == "1")
			fmt.Printf("%s = %v 설정 완료\n", inName, c.InputValues[inName])
		case "run":
			runCircuit(c)
		case "exit":
			fmt.Println("프로그램을 종료합니다.")
			return
		default:
			fmt.Println("알 수 없는 명령어입니다.")
		}
	}
}
