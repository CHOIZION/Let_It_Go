package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type ActivationFunction func(float64) float64
type ActivationFunctionDerivative func(float64) float64

func Sigmoid(x float64) float64 {
	return 1 / (1 + math.Exp(-x))
}

func SigmoidDerivative(x float64) float64 {
	sx := Sigmoid(x)
	return sx * (1 - sx)
}

func ReLU(x float64) float64 {
	if x > 0 {
		return x
	}
	return 0
}

func ReLUDerivative(x float64) float64 {
	if x > 0 {
		return 1
	}
	return 0
}

func Tanh(x float64) float64 {
	return math.Tanh(x)
}

func TanhDerivative(x float64) float64 {
	return 1 - math.Pow(math.Tanh(x), 2)
}

func LeakyReLU(x float64) float64 {
	if x > 0 {
		return x
	}
	return 0.01 * x
}

func LeakyReLUDerivative(x float64) float64 {
	if x > 0 {
		return 1
	}
	return 0.01
}

type NeuralNetwork struct {
	InputNeurons     int
	HiddenLayers     []int
	OutputNeurons    int
	LearningRate     float64
	ActivationFuncs  []ActivationFunction
	ActivationDerivs []ActivationFunctionDerivative
	Weights          [][][]float64
}

func NewNeuralNetwork(inputNeurons int, hiddenLayers []int, outputNeurons int, learningRate float64, activationFuncs []string) *NeuralNetwork {
	nn := &NeuralNetwork{
		InputNeurons:  inputNeurons,
		HiddenLayers:  hiddenLayers,
		OutputNeurons: outputNeurons,
		LearningRate:  learningRate,
	}

	totalLayers := len(hiddenLayers) + 1 // hidden layers + output layer

	nn.ActivationFuncs = make([]ActivationFunction, totalLayers)
	nn.ActivationDerivs = make([]ActivationFunctionDerivative, totalLayers)

	for i, act := range activationFuncs {
		switch act {
		case "sigmoid":
			nn.ActivationFuncs[i] = Sigmoid
			nn.ActivationDerivs[i] = SigmoidDerivative
		case "relu":
			nn.ActivationFuncs[i] = ReLU
			nn.ActivationDerivs[i] = ReLUDerivative
		case "tanh":
			nn.ActivationFuncs[i] = Tanh
			nn.ActivationDerivs[i] = TanhDerivative
		case "leakyrelu":
			nn.ActivationFuncs[i] = LeakyReLU
			nn.ActivationDerivs[i] = LeakyReLUDerivative
		default:
			nn.ActivationFuncs[i] = Sigmoid
			nn.ActivationDerivs[i] = SigmoidDerivative
		}
	}

	rand.Seed(time.Now().UnixNano())

	// Initialize weights
	nn.Weights = make([][][]float64, totalLayers)
	prevLayerNeurons := inputNeurons

	for idx := 0; idx < totalLayers; idx++ {
		var currentNeurons int
		if idx < len(hiddenLayers) {
			currentNeurons = hiddenLayers[idx]
		} else {
			currentNeurons = outputNeurons
		}
		nn.Weights[idx] = make([][]float64, prevLayerNeurons)
		for i := 0; i < prevLayerNeurons; i++ {
			nn.Weights[idx][i] = make([]float64, currentNeurons)
			for j := 0; j < currentNeurons; j++ {
				nn.Weights[idx][i][j] = rand.Float64()*2 - 1
			}
		}
		prevLayerNeurons = currentNeurons
	}

	return nn
}

func (nn *NeuralNetwork) Train(inputs [][]float64, targets [][]float64, epochs int, batchSize int) {
	losses := make([]float64, epochs)
	dataSize := len(inputs)

	for epoch := 0; epoch < epochs; epoch++ {
		totalLoss := 0.0

		// Shuffle data
		perm := rand.Perm(dataSize)
		for i := 0; i < dataSize; i += batchSize {
			end := i + batchSize
			if end > dataSize {
				end = dataSize
			}
			batchInputs := make([][]float64, end-i)
			batchTargets := make([][]float64, end-i)
			for j := i; j < end; j++ {
				batchInputs[j-i] = inputs[perm[j]]
				batchTargets[j-i] = targets[perm[j]]
			}

			// Batch training
			nn.updateBatch(batchInputs, batchTargets, &totalLoss)
		}

		avgLoss := totalLoss / float64(dataSize)
		losses[epoch] = avgLoss
		fmt.Printf("Epoch %d/%d, Loss: %.6f\n", epoch+1, epochs, avgLoss)
	}

	// 학습 곡선 시각화
	nn.plotLoss(losses)
}

func (nn *NeuralNetwork) updateBatch(batchInputs [][]float64, batchTargets [][]float64, totalLoss *float64) {
	batchSize := len(batchInputs)
	nablaWs := make([][][]float64, len(nn.Weights))
	for idx := range nn.Weights {
		nablaWs[idx] = make([][]float64, len(nn.Weights[idx]))
		for i := range nn.Weights[idx] {
			nablaWs[idx][i] = make([]float64, len(nn.Weights[idx][i]))
		}
	}

	for idx, input := range batchInputs {
		target := batchTargets[idx]

		// Forward pass
		activations := make([][]float64, len(nn.Weights)+1)
		activations[0] = input
		zs := make([][]float64, len(nn.Weights))

		for layerIdx, weights := range nn.Weights {
			z := make([]float64, len(weights[0]))
			for i := 0; i < len(weights[0]); i++ {
				sum := 0.0
				for j := 0; j < len(weights); j++ {
					sum += activations[layerIdx][j] * weights[j][i]
				}
				z[i] = sum
			}
			zs[layerIdx] = z

			activation := make([]float64, len(z))
			for i, val := range z {
				activation[i] = nn.ActivationFuncs[layerIdx](val)
			}
			activations[layerIdx+1] = activation
		}

		// Calculate loss
		loss := 0.0
		outputActivations := activations[len(activations)-1]
		for i := 0; i < len(outputActivations); i++ {
			diff := target[i] - outputActivations[i]
			loss += diff * diff
		}
		*totalLoss += loss / float64(len(outputActivations))

		// Backward pass
		delta := make([][]float64, len(nn.Weights))
		// Output layer error
		delta[len(delta)-1] = make([]float64, len(outputActivations))
		for i := 0; i < len(outputActivations); i++ {
			delta[len(delta)-1][i] = (outputActivations[i] - target[i]) * nn.ActivationDerivs[len(nn.ActivationDerivs)-1](zs[len(zs)-1][i])
		}

		// Hidden layers error
		for l := len(nn.Weights) - 2; l >= 0; l-- {
			delta[l] = make([]float64, len(activations[l+1]))
			for i := 0; i < len(delta[l]); i++ {
				errorSum := 0.0
				for j := 0; j < len(delta[l+1]); j++ {
					errorSum += delta[l+1][j] * nn.Weights[l+1][i][j]
				}
				delta[l][i] = errorSum * nn.ActivationDerivs[l](zs[l][i])
			}
		}

		// Accumulate gradients
		for l := 0; l < len(nn.Weights); l++ {
			for i := 0; i < len(nn.Weights[l]); i++ {
				for j := 0; j < len(nn.Weights[l][i]); j++ {
					nablaWs[l][i][j] += delta[l][j] * activations[l][i]
				}
			}
		}
	}

	// Update weights
	eta := nn.LearningRate / float64(batchSize)
	for l := 0; l < len(nn.Weights); l++ {
		for i := 0; i < len(nn.Weights[l]); i++ {
			for j := 0; j < len(nn.Weights[l][i]); j++ {
				nn.Weights[l][i][j] -= eta * nablaWs[l][i][j]
			}
		}
	}
}

func (nn *NeuralNetwork) Predict(input []float64) []float64 {
	activations := input
	for layerIdx, weights := range nn.Weights {
		z := make([]float64, len(weights[0]))
		for i := 0; i < len(weights[0]); i++ {
			sum := 0.0
			for j := 0; j < len(weights); j++ {
				sum += activations[j] * weights[j][i]
			}
			z[i] = sum
		}
		activation := make([]float64, len(z))
		for i, val := range z {
			activation[i] = nn.ActivationFuncs[layerIdx](val)
		}
		activations = activation
	}
	return activations
}

func (nn *NeuralNetwork) SaveModel(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Save architecture
	arch := []string{strconv.Itoa(nn.InputNeurons)}
	for _, h := range nn.HiddenLayers {
		arch = append(arch, strconv.Itoa(h))
	}
	arch = append(arch, strconv.Itoa(nn.OutputNeurons))
	writer.Write(arch)

	// Save weights
	for _, layer := range nn.Weights {
		for _, weights := range layer {
			strWeights := make([]string, len(weights))
			for i, w := range weights {
				strWeights[i] = fmt.Sprintf("%f", w)
			}
			writer.Write(strWeights)
		}
	}

	return nil
}

func (nn *NeuralNetwork) LoadModel(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Load architecture
	arch, err := reader.Read()
	if err != nil {
		return err
	}
	nn.InputNeurons, _ = strconv.Atoi(arch[0])
	nn.HiddenLayers = []int{}
	for _, h := range arch[1 : len(arch)-1] {
		neurons, _ := strconv.Atoi(h)
		nn.HiddenLayers = append(nn.HiddenLayers, neurons)
	}
	nn.OutputNeurons, _ = strconv.Atoi(arch[len(arch)-1])

	// Reinitialize weights
	totalLayers := len(nn.HiddenLayers) + 1
	nn.Weights = make([][][]float64, totalLayers)

	// Load weights
	for idx := 0; idx < totalLayers; idx++ {
		var currentNeurons int
		if idx < len(nn.HiddenLayers) {
			currentNeurons = nn.HiddenLayers[idx]
		} else {
			currentNeurons = nn.OutputNeurons
		}

		var prevNeurons int
		if idx == 0 {
			prevNeurons = nn.InputNeurons
		} else {
			prevNeurons = nn.HiddenLayers[idx-1]
		}

		nn.Weights[idx] = make([][]float64, prevNeurons)
		for i := 0; i < prevNeurons; i++ {
			record, err := reader.Read()
			if err != nil {
				return err
			}
			nn.Weights[idx][i] = make([]float64, currentNeurons)
			for j, val := range record {
				nn.Weights[idx][i][j], _ = strconv.ParseFloat(val, 64)
			}
		}
	}

	return nil
}

func (nn *NeuralNetwork) plotLoss(losses []float64) {
	pts := make(plotter.XYs, len(losses))
	for i := range pts {
		pts[i].X = float64(i + 1)
		pts[i].Y = losses[i]
	}

	p := plot.New()
	p.Title.Text = "Training Loss"
	p.X.Label.Text = "Epoch"
	p.Y.Label.Text = "Loss"

	line, err := plotter.NewLine(pts)
	if err != nil {
		fmt.Println("그래프를 그리는 중 오류가 발생했습니다:", err)
		return
	}
	p.Add(line)

	if err := p.Save(6*vg.Inch, 4*vg.Inch, "loss.png"); err != nil {
		fmt.Println("그래프를 저장하는 중 오류가 발생했습니다:", err)
	} else {
		fmt.Println("학습 곡선 그래프가 'loss.png'로 저장되었습니다.")
	}
}

func loadCSVData(filename string) ([][]float64, [][]float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	inputs := [][]float64{}
	targets := [][]float64{}

	for _, record := range records {
		input := []float64{}
		for _, val := range record[:len(record)-1] {
			fval, _ := strconv.ParseFloat(val, 64)
			input = append(input, fval)
		}
		targetVal, _ := strconv.ParseFloat(record[len(record)-1], 64)
		target := []float64{targetVal}
		inputs = append(inputs, input)
		targets = append(targets, target)
	}

	return inputs, targets, nil
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("=== 신경망 학습 시뮬레이터 ===")

	// 신경망 구조 설정
	var inputNeurons int
	for {
		fmt.Print("입력 뉴런 수를 입력하세요: ")
		scanner.Scan()
		inputNeurons, _ = strconv.Atoi(scanner.Text())
		if inputNeurons > 0 {
			break
		}
		fmt.Println("입력 뉴런 수는 0보다 커야 합니다.")
	}

	var hiddenLayers []int
	fmt.Print("은닉층의 뉴런 수를 쉼표로 구분하여 입력하세요 (예: 4,3): ")
	scanner.Scan()
	hiddenStrs := strings.Split(scanner.Text(), ",")
	for _, s := range hiddenStrs {
		val, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil || val <= 0 {
			fmt.Println("유효한 숫자를 입력하세요.")
			return
		}
		hiddenLayers = append(hiddenLayers, val)
	}

	var outputNeurons int
	for {
		fmt.Print("출력 뉴런 수를 입력하세요: ")
		scanner.Scan()
		outputNeurons, _ = strconv.Atoi(scanner.Text())
		if outputNeurons > 0 {
			break
		}
		fmt.Println("출력 뉴런 수는 0보다 커야 합니다.")
	}

	var learningRate float64
	for {
		fmt.Print("학습률(Learning Rate)을 입력하세요 (예: 0.1): ")
		scanner.Scan()
		learningRate, _ = strconv.ParseFloat(scanner.Text(), 64)
		if learningRate > 0 {
			break
		}
		fmt.Println("학습률은 0보다 커야 합니다.")
	}

	fmt.Print("각 층의 활성화 함수를 선택하세요 (sigmoid/relu/tanh/leakyrelu), 쉼표로 구분 (예: sigmoid,relu,tanh): ")
	scanner.Scan()
	activationStrs := strings.Split(scanner.Text(), ",")
	if len(activationStrs) != len(hiddenLayers)+1 {
		fmt.Println("활성화 함수의 수는 은닉층 수 + 1 이어야 합니다.")
		return
	}
	activationFuncs := []string{}
	for _, act := range activationStrs {
		activationFuncs = append(activationFuncs, strings.TrimSpace(act))
	}

	nn := NewNeuralNetwork(inputNeurons, hiddenLayers, outputNeurons, learningRate, activationFuncs)
	fmt.Println("신경망이 생성되었습니다.")

	// 데이터셋 로드 또는 선택
	fmt.Println("\n데이터셋을 선택하세요:")
	fmt.Println("1. XOR 문제")
	fmt.Println("2. AND 게이트")
	fmt.Println("3. OR 게이트")
	fmt.Println("4. CSV 파일에서 로드")
	fmt.Print("선택: ")
	scanner.Scan()
	choice := scanner.Text()

	var inputs [][]float64
	var targets [][]float64

	switch choice {
	case "1":
		inputs = [][]float64{
			{0, 0},
			{0, 1},
			{1, 0},
			{1, 1},
		}
		targets = [][]float64{
			{0},
			{1},
			{1},
			{0},
		}
	case "2":
		inputs = [][]float64{
			{0, 0},
			{0, 1},
			{1, 0},
			{1, 1},
		}
		targets = [][]float64{
			{0},
			{0},
			{0},
			{1},
		}
	case "3":
		inputs = [][]float64{
			{0, 0},
			{0, 1},
			{1, 0},
			{1, 1},
		}
		targets = [][]float64{
			{0},
			{1},
			{1},
			{1},
		}
	case "4":
		fmt.Print("CSV 파일의 경로를 입력하세요: ")
		scanner.Scan()
		filename := scanner.Text()
		var err error
		inputs, targets, err = loadCSVData(filename)
		if err != nil {
			fmt.Println("데이터를 로드하는 중 오류가 발생했습니다:", err)
			return
		}
	default:
		fmt.Println("잘못된 선택입니다.")
		return
	}

	// 학습 수행
	var epochs int
	for {
		fmt.Print("\n에포크(epoch) 수를 입력하세요: ")
		scanner.Scan()
		epochs, _ = strconv.Atoi(scanner.Text())
		if epochs > 0 {
			break
		}
		fmt.Println("에포크 수는 0보다 커야 합니다.")
	}

	var batchSize int
	for {
		fmt.Print("배치 크기를 입력하세요: ")
		scanner.Scan()
		batchSize, _ = strconv.Atoi(scanner.Text())
		if batchSize > 0 {
			break
		}
		fmt.Println("배치 크기는 0보다 커야 합니다.")
	}

	fmt.Println("\n학습을 시작합니다...")
	nn.Train(inputs, targets, epochs, batchSize)
	fmt.Println("학습이 완료되었습니다.")

	// 모델 저장
	fmt.Print("\n학습된 모델을 저장하시겠습니까? (y/n): ")
	scanner.Scan()
	saveChoice := scanner.Text()
	if strings.ToLower(saveChoice) == "y" {
		fmt.Print("저장할 파일 이름을 입력하세요: ")
		scanner.Scan()
		filename := scanner.Text()
		err := nn.SaveModel(filename)
		if err != nil {
			fmt.Println("모델을 저장하는 중 오류가 발생했습니다:", err)
		} else {
			fmt.Println("모델이 저장되었습니다.")
		}
	}

	// 예측 수행
	for {
		fmt.Print("\n예측할 입력 값을 공백으로 구분하여 입력하세요 (종료하려면 'exit'): ")
		scanner.Scan()
		line := scanner.Text()
		if strings.TrimSpace(line) == "exit" {
			fmt.Println("프로그램을 종료합니다.")
			break
		}
		inputStrs := strings.Fields(line)
		if len(inputStrs) != inputNeurons {
			fmt.Printf("입력 값의 수는 %d이어야 합니다.\n", inputNeurons)
			continue
		}
		inputValues := make([]float64, inputNeurons)
		validInput := true
		for i, s := range inputStrs {
			val, err := strconv.ParseFloat(s, 64)
			if err != nil {
				fmt.Println("유효한 숫자를 입력하세요.")
				validInput = false
				break
			}
			inputValues[i] = val
		}
		if !validInput {
			continue
		}
		output := nn.Predict(inputValues)
		fmt.Printf("예측 결과: %v\n", output)
	}
}
