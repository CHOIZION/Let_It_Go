package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"time"
)

type Vertex struct {
	X, Y, Z float64
}

type Face struct {
	A, B, C int
}

type Object struct {
	Vertices []Vertex
	Faces    []Face
	Position Vertex
	Rotation Vertex
	Scale    float64
}

var (
	width      = 80
	height     = 40
	frameDelay = time.Millisecond * 50
	zBuffer    [][]float64
	screen     [][]rune
	prevScreen [][]rune
	cameraPos  = Vertex{X: 0, Y: 0, Z: -5}
	lightDir   = Vertex{X: 0, Y: 0, Z: -1}
	scanner    = bufio.NewScanner(os.Stdin)
	rendering  = true
	autoRotate = true
)

func main() {
	cube := createCube()
	go renderLoop(&cube)
	inputLoop(&cube)
}

func createCube() Object {
	vertices := []Vertex{
		{-1, -1, -1},
		{1, -1, -1},
		{1, 1, -1},
		{-1, 1, -1},
		{-1, -1, 1},
		{1, -1, 1},
		{1, 1, 1},
		{-1, 1, 1},
	}

	faces := []Face{
		{0, 1, 2}, {0, 2, 3},
		{4, 5, 6}, {4, 6, 7},
		{0, 1, 5}, {0, 5, 4},
		{2, 3, 7}, {2, 7, 6},
		{1, 2, 6}, {1, 6, 5},
		{0, 3, 7}, {0, 7, 4},
	}

	return Object{
		Vertices: vertices,
		Faces:    faces,
		Position: Vertex{X: 0, Y: 0, Z: 0},
		Rotation: Vertex{X: 0, Y: 0, Z: 0},
		Scale:    1,
	}
}

func renderLoop(obj *Object) {
	for rendering {
		initBuffers()
		if autoRotate {
			obj.Rotation.Y += 0.05
			obj.Rotation.X += 0.03
		}
		renderObject(*obj)
		printScreen()
		time.Sleep(frameDelay)
	}
}

func initBuffers() {
	zBuffer = make([][]float64, height)
	screen = make([][]rune, height)
	if prevScreen == nil {
		prevScreen = make([][]rune, height)
	}
	for y := 0; y < height; y++ {
		zBuffer[y] = make([]float64, width)
		screen[y] = make([]rune, width)
		if prevScreen[y] == nil {
			prevScreen[y] = make([]rune, width)
			for x := 0; x < width; x++ {
				prevScreen[y][x] = ' '
			}
		}
		for x := 0; x < width; x++ {
			zBuffer[y][x] = math.Inf(1)
			screen[y][x] = ' '
		}
	}
}

func renderObject(obj Object) {
	transformedVertices := make([]Vertex, len(obj.Vertices))
	for i, v := range obj.Vertices {
		tv := transformVertex(v, obj)
		transformedVertices[i] = tv
	}

	for _, face := range obj.Faces {
		v1 := transformedVertices[face.A]
		v2 := transformedVertices[face.B]
		v3 := transformedVertices[face.C]

		normal := computeNormal(v1, v2, v3)
		if normal.Z >= 0 {
			continue
		}

		lightIntensity := computeLightIntensity(normal, lightDir)
		shadeChar := getShadeChar(lightIntensity)

		projectedV1 := projectVertex(v1)
		projectedV2 := projectVertex(v2)
		projectedV3 := projectVertex(v3)

		fillTriangle(projectedV1, projectedV2, projectedV3, shadeChar)
	}
}

func transformVertex(v Vertex, obj Object) Vertex {
	v.X *= obj.Scale
	v.Y *= obj.Scale
	v.Z *= obj.Scale
	v = rotateX(v, obj.Rotation.X)
	v = rotateY(v, obj.Rotation.Y)
	v = rotateZ(v, obj.Rotation.Z)
	v.X += obj.Position.X
	v.Y += obj.Position.Y
	v.Z += obj.Position.Z
	return v
}

func rotateX(v Vertex, angle float64) Vertex {
	cosA := math.Cos(angle)
	sinA := math.Sin(angle)
	return Vertex{
		X: v.X,
		Y: v.Y*cosA - v.Z*sinA,
		Z: v.Y*sinA + v.Z*cosA,
	}
}

func rotateY(v Vertex, angle float64) Vertex {
	cosA := math.Cos(angle)
	sinA := math.Sin(angle)
	return Vertex{
		X: v.X*cosA + v.Z*sinA,
		Y: v.Y,
		Z: -v.X*sinA + v.Z*cosA,
	}
}

func rotateZ(v Vertex, angle float64) Vertex {
	cosA := math.Cos(angle)
	sinA := math.Sin(angle)
	return Vertex{
		X: v.X*cosA - v.Y*sinA,
		Y: v.X*sinA + v.Y*cosA,
		Z: v.Z,
	}
}

func computeNormal(v1, v2, v3 Vertex) Vertex {
	u := Vertex{
		X: v2.X - v1.X,
		Y: v2.Y - v1.Y,
		Z: v2.Z - v1.Z,
	}
	v := Vertex{
		X: v3.X - v1.X,
		Y: v3.Y - v1.Y,
		Z: v3.Z - v1.Z,
	}
	n := Vertex{
		X: u.Y*v.Z - u.Z*v.Y,
		Y: u.Z*v.X - u.X*v.Z,
		Z: u.X*v.Y - u.Y*v.X,
	}
	length := math.Sqrt(n.X*n.X + n.Y*n.Y + n.Z*n.Z)
	return Vertex{X: n.X / length, Y: n.Y / length, Z: n.Z / length}
}

func computeLightIntensity(normal, lightDir Vertex) float64 {
	dot := normal.X*lightDir.X + normal.Y*lightDir.Y + normal.Z*lightDir.Z
	return math.Max(0, dot)
}

func getShadeChar(intensity float64) rune {
	shades := []rune{' ', '.', ':', '-', '=', '+', '*', '#', '%', '@'}
	index := int(intensity * float64(len(shades)-1))
	if index < 0 {
		index = 0
	}
	if index >= len(shades) {
		index = len(shades) - 1
	}
	return shades[index]
}

func projectVertex(v Vertex) Vertex {
	fov := 1.0
	z := v.Z - cameraPos.Z
	if z == 0 {
		z = 0.001
	}
	x := (v.X-cameraPos.X)*(fov/z)*float64(width)/2 + float64(width)/2
	y := (v.Y-cameraPos.Y)*(fov/z)*float64(height)/2 + float64(height)/2
	return Vertex{X: x, Y: y, Z: z}
}

func fillTriangle(v1, v2, v3 Vertex, shadeChar rune) {
	x1, y1, z1 := int(v1.X), int(v1.Y), v1.Z
	x2, y2, z2 := int(v2.X), int(v2.Y), v2.Z
	x3, y3, z3 := int(v3.X), int(v3.Y), v3.Z

	drawLine(x1, y1, z1, x2, y2, z2, shadeChar)
	drawLine(x2, y2, z2, x3, y3, z3, shadeChar)
	drawLine(x3, y3, z3, x1, y1, z1, shadeChar)
}

func drawLine(x0, y0 int, z0 float64, x1, y1 int, z1 float64, shadeChar rune) {
	dx := math.Abs(float64(x1 - x0))
	dy := math.Abs(float64(y1 - y0))
	dz := z1 - z0
	n := int(math.Max(dx, dy))
	if n == 0 {
		plotPixel(x0, y0, z0, shadeChar)
		return
	}
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		x := int(float64(x0) + t*float64(x1-x0))
		y := int(float64(y0) + t*float64(y1-y0))
		z := z0 + t*dz
		plotPixel(x, y, z, shadeChar)
	}
}

func plotPixel(x, y int, z float64, shadeChar rune) {
	if x >= 0 && x < width && y >= 0 && y < height {
		if z < zBuffer[y][x] {
			zBuffer[y][x] = z
			screen[y][x] = shadeChar
		}
	}
}

func printScreen() {
	// 커서 숨기기
	fmt.Print("\033[?25l")
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if screen[y][x] != prevScreen[y][x] {
				// 커서 이동하여 변경된 문자 출력
				fmt.Printf("\033[%d;%dH%s", y+1, x+1, string(screen[y][x]))
			}
		}
	}
	// 커서 보이기
	fmt.Print("\033[?25h")
	// 표준 출력 버퍼 비우기
	os.Stdout.Sync()
	// 이전 프레임 업데이트
	for y := 0; y < height; y++ {
		copy(prevScreen[y], screen[y])
	}
}

func inputLoop(obj *Object) {
	for {
		fmt.Printf("\033[%d;1H", height+2) // 입력 위치로 커서 이동
		fmt.Println("명령을 입력하세요 (r: 자동회전 토글, q: 종료):")
		fmt.Print("> ")
		scanner.Scan()
		command := scanner.Text()
		switch command {
		case "q":
			rendering = false
			fmt.Println("프로그램을 종료합니다.")
			os.Exit(0)
		case "r":
			autoRotate = !autoRotate
			if autoRotate {
				fmt.Println("자동 회전이 활성화되었습니다.")
			} else {
				fmt.Println("자동 회전이 비활성화되었습니다.")
			}
		default:
			fmt.Println("알 수 없는 명령입니다.")
		}
	}
}
