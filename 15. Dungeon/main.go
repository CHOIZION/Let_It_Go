package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"
)

const (
	width       = 30
	height      = 15
	numMonsters = 10
	numItems    = 10
	maxLevel    = 5
)

type Tile int

const (
	Wall Tile = iota
	Floor
	StairsDown
	StairsUp
)

type Position struct {
	X, Y int
}

type Player struct {
	Pos       Position
	HP        int
	MaxHP     int
	Attack    int
	Defense   int
	Level     int
	Exp       int
	ExpToNext int
	Inventory []Item
}

type Monster struct {
	Pos      Position
	HP       int
	MaxHP    int
	Attack   int
	Defense  int
	Name     string
	ExpValue int
}

type Item struct {
	Pos   Position
	Name  string
	Type  string
	Value int
}

var (
	dungeon      [][][]Tile
	player       Player
	monsters     [][]Monster
	items        [][]Item
	currentFloor int
	rng          *rand.Rand
	scanner      *bufio.Scanner
	visible      [][]bool
	fovRange     = 5
)

func main() {
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	scanner = bufio.NewScanner(os.Stdin)

	generateDungeon()
	currentFloor = 0
	placePlayer()
	placeMonsters()
	placeItems()
	gameLoop()
}

func generateDungeon() {
	dungeon = make([][][]Tile, maxLevel)
	for level := 0; level < maxLevel; level++ {
		dungeon[level] = make([][]Tile, height)
		for y := 0; y < height; y++ {
			dungeon[level][y] = make([]Tile, width)
			for x := 0; x < width; x++ {
				if rng.Float64() < 0.2 {
					dungeon[level][y][x] = Wall
				} else {
					dungeon[level][y][x] = Floor
				}
			}
		}
		placeStairs(level)
	}
}

func placeStairs(level int) {
	for {
		x := rng.Intn(width)
		y := rng.Intn(height)
		if dungeon[level][y][x] == Floor {
			dungeon[level][y][x] = StairsDown
			break
		}
	}
	if level > 0 {
		for {
			x := rng.Intn(width)
			y := rng.Intn(height)
			if dungeon[level][y][x] == Floor {
				dungeon[level][y][x] = StairsUp
				break
			}
		}
	}
}

func placePlayer() {
	for {
		x := rng.Intn(width)
		y := rng.Intn(height)
		if dungeon[currentFloor][y][x] == Floor {
			player = Player{
				Pos:       Position{X: x, Y: y},
				HP:        30,
				MaxHP:     30,
				Attack:    5,
				Defense:   2,
				Level:     1,
				Exp:       0,
				ExpToNext: 20,
				Inventory: []Item{},
			}
			break
		}
	}
}

func placeMonsters() {
	monsters = make([][]Monster, maxLevel)
	for level := 0; level < maxLevel; level++ {
		for i := 0; i < numMonsters; i++ {
			for {
				x := rng.Intn(width)
				y := rng.Intn(height)
				if dungeon[level][y][x] == Floor && (x != player.Pos.X || y != player.Pos.Y) {
					monster := Monster{
						Pos:      Position{X: x, Y: y},
						HP:       15 + level*5,
						MaxHP:    15 + level*5,
						Attack:   4 + level*2,
						Defense:  2 + level,
						Name:     "Goblin",
						ExpValue: 10 + level*5,
					}
					monsters[level] = append(monsters[level], monster)
					break
				}
			}
		}
	}
	// 보스 몬스터 배치
	boss := Monster{
		Pos:      Position{X: width / 2, Y: height / 2},
		HP:       100,
		MaxHP:    100,
		Attack:   15,
		Defense:  5,
		Name:     "Dragon",
		ExpValue: 100,
	}
	monsters[maxLevel-1] = append(monsters[maxLevel-1], boss)
}

func placeItems() {
	items = make([][]Item, maxLevel)
	itemNames := []string{"Sword", "Shield", "Potion", "Armor", "Ring", "SpellBook", "Amulet"}
	for level := 0; level < maxLevel; level++ {
		for i := 0; i < numItems; i++ {
			for {
				x := rng.Intn(width)
				y := rng.Intn(height)
				if dungeon[level][y][x] == Floor && (x != player.Pos.X || y != player.Pos.Y) {
					item := Item{
						Pos:   Position{X: x, Y: y},
						Name:  itemNames[rng.Intn(len(itemNames))],
						Type:  "Equip",
						Value: rng.Intn(5) + 1 + level,
					}
					items[level] = append(items[level], item)
					break
				}
			}
		}
	}
}

func gameLoop() {
	for {
		updateVisibility()
		printDungeon()
		fmt.Println("명령을 입력하세요 (w/a/s/d, i: 인벤토리, q: 종료):")
		fmt.Print("> ")
		scanner.Scan()
		command := scanner.Text()
		switch command {
		case "w":
			movePlayer(0, -1)
		case "s":
			movePlayer(0, 1)
		case "a":
			movePlayer(-1, 0)
		case "d":
			movePlayer(1, 0)
		case "i":
			openInventory()
		case "q":
			fmt.Println("게임을 종료합니다.")
			os.Exit(0)
		default:
			fmt.Println("알 수 없는 명령입니다.")
		}

		if player.HP <= 0 {
			fmt.Println("플레이어가 사망했습니다. 게임 오버.")
			os.Exit(0)
		}

		if currentFloor == maxLevel-1 && len(monsters[currentFloor]) == 0 {
			fmt.Println("최종 보스를 처치했습니다. 승리!")
			os.Exit(0)
		}
	}
}

func movePlayer(dx, dy int) {
	newX := player.Pos.X + dx
	newY := player.Pos.Y + dy

	if newX < 0 || newX >= width || newY < 0 || newY >= height {
		fmt.Println("더 이상 이동할 수 없습니다.")
		return
	}

	tile := dungeon[currentFloor][newY][newX]
	if tile == Wall {
		fmt.Println("벽이 가로막고 있습니다.")
		return
	}

	// 몬스터와 조우
	for i, monster := range monsters[currentFloor] {
		if monster.Pos.X == newX && monster.Pos.Y == newY {
			fmt.Printf("%s와 전투가 시작됩니다!\n", monster.Name)
			battle(&monsters[currentFloor][i])
			return
		}
	}

	// 아이템 획득
	for i, item := range items[currentFloor] {
		if item.Pos.X == newX && item.Pos.Y == newY {
			fmt.Printf("%s을(를) 발견했습니다!\n", item.Name)
			player.Inventory = append(player.Inventory, item)
			applyItemEffect(item)
			// 아이템 제거
			items[currentFloor] = append(items[currentFloor][:i], items[currentFloor][i+1:]...)
			break
		}
	}

	// 계단 이동
	if tile == StairsDown {
		if currentFloor < maxLevel-1 {
			currentFloor++
			fmt.Println("아래 층으로 내려갑니다.")
			placePlayer()
			return
		}
	} else if tile == StairsUp {
		if currentFloor > 0 {
			currentFloor--
			fmt.Println("위 층으로 올라갑니다.")
			placePlayer()
			return
		}
	}

	player.Pos.X = newX
	player.Pos.Y = newY
}

func battle(monster *Monster) {
	for {
		// 플레이어 턴
		damage := player.Attack - monster.Defense + rng.Intn(2)
		if damage < 0 {
			damage = 0
		}
		monster.HP -= damage
		fmt.Printf("%s에게 %d의 피해를 입혔습니다! (%s HP: %d/%d)\n", monster.Name, damage, monster.Name, monster.HP, monster.MaxHP)
		if monster.HP <= 0 {
			fmt.Printf("%s를 처치했습니다!\n", monster.Name)
			gainExp(monster.ExpValue)
			// 몬스터 제거
			removeMonster(monster)
			break
		}

		// 몬스터 턴
		damage = monster.Attack - player.Defense + rng.Intn(2)
		if damage < 0 {
			damage = 0
		}
		player.HP -= damage
		fmt.Printf("%s로부터 %d의 피해를 입었습니다! (플레이어 HP: %d/%d)\n", monster.Name, damage, player.HP, player.MaxHP)
		if player.HP <= 0 {
			return
		}
	}
}

func removeMonster(monster *Monster) {
	for i, m := range monsters[currentFloor] {
		if m.Pos == monster.Pos {
			monsters[currentFloor] = append(monsters[currentFloor][:i], monsters[currentFloor][i+1:]...)
			break
		}
	}
}

func applyItemEffect(item Item) {
	switch item.Name {
	case "Sword":
		player.Attack += item.Value
		fmt.Printf("공격력이 %d만큼 증가했습니다! (현재 공격력: %d)\n", item.Value, player.Attack)
	case "Shield":
		player.Defense += item.Value
		fmt.Printf("방어력이 %d만큼 증가했습니다! (현재 방어력: %d)\n", item.Value, player.Defense)
	case "Potion":
		player.HP += item.Value * 2
		if player.HP > player.MaxHP {
			player.HP = player.MaxHP
		}
		fmt.Printf("체력이 회복되었습니다! (현재 HP: %d/%d)\n", player.HP, player.MaxHP)
	case "Armor":
		player.MaxHP += item.Value * 2
		player.HP += item.Value * 2
		fmt.Printf("최대 체력이 증가했습니다! (현재 Max HP: %d)\n", player.MaxHP)
	case "Ring":
		player.Attack += item.Value
		player.Defense += item.Value
		fmt.Printf("공격력과 방어력이 증가했습니다! (공격력: %d, 방어력: %d)\n", player.Attack, player.Defense)
	case "SpellBook":
		player.Attack += item.Value * 2
		fmt.Printf("마법 공격력이 증가했습니다! (현재 공격력: %d)\n", player.Attack)
	case "Amulet":
		player.Defense += item.Value * 2
		fmt.Printf("마법 방어력이 증가했습니다! (현재 방어력: %d)\n", player.Defense)
	}
}

func printDungeon() {
	fmt.Println("\n===== 던전 맵 =====")
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if !visible[y][x] {
				fmt.Print(" ")
				continue
			}
			if player.Pos.X == x && player.Pos.Y == y {
				fmt.Print("@")
			} else if isMonsterAt(x, y) {
				fmt.Print("M")
			} else if isItemAt(x, y) {
				fmt.Print("I")
			} else {
				switch dungeon[currentFloor][y][x] {
				case Wall:
					fmt.Print("#")
				case Floor:
					fmt.Print(".")
				case StairsDown:
					fmt.Print(">")
				case StairsUp:
					fmt.Print("<")
				}
			}
		}
		fmt.Println()
	}
	fmt.Printf("HP: %d/%d, 공격력: %d, 방어력: %d, 레벨: %d, 경험치: %d/%d\n", player.HP, player.MaxHP, player.Attack, player.Defense, player.Level, player.Exp, player.ExpToNext)
}

func isMonsterAt(x, y int) bool {
	for _, monster := range monsters[currentFloor] {
		if monster.Pos.X == x && monster.Pos.Y == y {
			return true
		}
	}
	return false
}

func isItemAt(x, y int) bool {
	for _, item := range items[currentFloor] {
		if item.Pos.X == x && item.Pos.Y == y {
			return true
		}
	}
	return false
}

func gainExp(amount int) {
	player.Exp += amount
	fmt.Printf("경험치를 %d만큼 획득했습니다! (현재 경험치: %d/%d)\n", amount, player.Exp, player.ExpToNext)
	if player.Exp >= player.ExpToNext {
		levelUp()
	}
}

func levelUp() {
	player.Level++
	player.Exp = 0
	player.ExpToNext += player.Level * 20
	player.MaxHP += 10
	player.HP = player.MaxHP
	player.Attack += 3
	player.Defense += 2
	fmt.Printf("레벨 업! 현재 레벨: %d\n", player.Level)
}

func openInventory() {
	fmt.Println("\n===== 인벤토리 =====")
	for i, item := range player.Inventory {
		fmt.Printf("[%d] %s\n", i+1, item.Name)
	}
	fmt.Println("사용할 아이템 번호를 입력하세요 (취소하려면 Enter):")
	fmt.Print("> ")
	scanner.Scan()
	input := scanner.Text()
	if input == "" {
		return
	}
	var index int
	_, err := fmt.Sscanf(input, "%d", &index)
	if err != nil || index < 1 || index > len(player.Inventory) {
		fmt.Println("유효한 번호를 입력하세요.")
		return
	}
	item := player.Inventory[index-1]
	if item.Type == "Equip" {
		fmt.Printf("%s을(를) 사용했습니다.\n", item.Name)
		applyItemEffect(item)
		player.Inventory = append(player.Inventory[:index-1], player.Inventory[index:]...)
	}
}

func updateVisibility() {
	visible = make([][]bool, height)
	for y := 0; y < height; y++ {
		visible[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			distance := abs(player.Pos.X-x) + abs(player.Pos.Y-y)
			if distance <= fovRange {
				visible[y][x] = true
			}
		}
	}
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
