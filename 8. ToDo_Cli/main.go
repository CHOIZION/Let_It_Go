package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	ID        int
	Title     string
	Priority  string // "높음", "중간", "낮음"
	Category  string // 카테고리
	DueDate   string // 마감일 (YYYY-MM-DD)
	Repeat    string // 반복 주기 ("없음", "매일", "매주", "매월")
	Complete  bool
	CreatedAt time.Time // 생성일
	UpdatedAt time.Time // 업데이트일
}

type TaskList struct {
	Tasks           []Task
	NotificationDay int // 알림 기준 일수
}

const dataFile = "tasks.json"

func main() {
	taskList := &TaskList{NotificationDay: 3}

	// 기존 데이터 로드
	err := taskList.loadTasks()
	if err != nil {
		fmt.Println("데이터를 로드하는 중 오류 발생:", err)
	}

	taskList.generateRepeatingTasks()
	taskList.saveTasks()

	taskList.showNotifications()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n할 일 목록 애플리케이션")
		fmt.Println("1. 할 일 목록 보기")
		fmt.Println("2. 할 일 추가하기")
		fmt.Println("3. 할 일 완료 표시하기")
		fmt.Println("4. 할 일 삭제하기")
		fmt.Println("5. 할 일 편집하기")
		fmt.Println("6. 할 일 검색하기")
		fmt.Println("7. 통계 보기")
		fmt.Println("8. 종료하기")
		fmt.Print("선택하세요: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			taskList.viewTasks(reader)
		case "2":
			taskList.addTask(reader)
		case "3":
			taskList.completeTask(reader)
		case "4":
			taskList.deleteTask(reader)
		case "5":
			taskList.editTask(reader)
		case "6":
			taskList.searchTasks(reader)
		case "7":
			taskList.showStatistics()
		case "8":
			// 데이터 저장 후 종료
			err := taskList.saveTasks()
			if err != nil {
				fmt.Println("데이터를 저장하는 중 오류 발생:", err)
			}
			fmt.Println("프로그램을 종료합니다.")
			return
		default:
			fmt.Println("잘못된 입력입니다. 다시 선택해주세요.")
		}
	}
}

// loadTasks는 로컬 파일에서 할 일 목록을 로드합니다.
func (t *TaskList) loadTasks() error {
	file, err := os.Open(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 파일이 없으면 새로 생성
			t.Tasks = []Task{}
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&t.Tasks)
	if err != nil {
		return err
	}
	return nil
}

// saveTasks는 현재 할 일 목록을 로컬 파일에 저장합니다.
func (t *TaskList) saveTasks() error {
	file, err := os.Create(dataFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(t.Tasks)
	if err != nil {
		return err
	}
	return nil
}

// generateRepeatingTasks는 반복할 일을 생성합니다.
func (t *TaskList) generateRepeatingTasks() {
	today := time.Now().Format("2006-01-02")
	for _, task := range t.Tasks {
		if task.Repeat != "없음" && task.DueDate != "" && task.Complete {
			dueDate, _ := time.Parse("2006-01-02", task.DueDate)
			nextDueDate := dueDate

			switch task.Repeat {
			case "매일":
				nextDueDate = dueDate.AddDate(0, 0, 1)
			case "매주":
				nextDueDate = dueDate.AddDate(0, 0, 7)
			case "매월":
				nextDueDate = dueDate.AddDate(0, 1, 0)
			}

			if nextDueDate.Format("2006-01-02") <= today {
				newTask := task
				newTask.ID = t.getNextID()
				newTask.DueDate = nextDueDate.Format("2006-01-02")
				newTask.Complete = false
				newTask.CreatedAt = time.Now()
				t.Tasks = append(t.Tasks, newTask)
			}
		}
	}
}

// showNotifications는 마감일이 임박한 할 일을 알립니다.
func (t *TaskList) showNotifications() {
	fmt.Println("\n[알림] 마감일이 임박한 할 일:")

	today := time.Now()
	notified := false

	for _, task := range t.Tasks {
		if task.DueDate != "" && !task.Complete {
			dueDate, _ := time.Parse("2006-01-02", task.DueDate)
			daysLeft := int(dueDate.Sub(today).Hours() / 24)
			if daysLeft >= 0 && daysLeft <= t.NotificationDay {
				fmt.Printf("- %s (마감일: %s, 남은 일수: %d일)\n", task.Title, task.DueDate, daysLeft)
				notified = true
			}
		}
	}

	if !notified {
		fmt.Println("마감일이 임박한 할 일이 없습니다.")
	}
}

// viewTasks는 현재 할 일 목록을 출력합니다.
func (t *TaskList) viewTasks(reader *bufio.Reader) {
	if len(t.Tasks) == 0 {
		fmt.Println("할 일이 없습니다.")
		return
	}

	fmt.Println("\n정렬 기준을 선택하세요:")
	fmt.Println("1. 우선순위")
	fmt.Println("2. 마감일")
	fmt.Println("3. 생성일")
	fmt.Println("4. 카테고리")
	fmt.Print("선택하세요: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "1":
		t.sortByPriority()
	case "2":
		t.sortByDueDate()
	case "3":
		t.sortByCreatedAt()
	case "4":
		t.sortByCategory()
	default:
		fmt.Println("잘못된 입력입니다. 기본 정렬로 표시합니다.")
	}

	fmt.Println("\n현재 할 일 목록:")
	for _, task := range t.Tasks {
		status := "[ ]"
		if task.Complete {
			status = "[X]"
		}

		dueDateStr := ""
		if task.DueDate != "" {
			dueDate, _ := time.Parse("2006-01-02", task.DueDate)
			daysLeft := int(dueDate.Sub(time.Now()).Hours() / 24)
			dueDateStr = fmt.Sprintf(" (마감일: %s, 남은 일수: %d일)", task.DueDate, daysLeft)
		}

		fmt.Printf("%d. %s %s - 우선순위: %s, 카테고리: %s%s\n", task.ID, status, task.Title, task.Priority, task.Category, dueDateStr)
	}
}

// addTask는 새로운 할 일을 추가합니다.
func (t *TaskList) addTask(reader *bufio.Reader) {
	fmt.Print("새로운 할 일 제목을 입력하세요: ")
	title, _ := reader.ReadString('\n')
	title = strings.TrimSpace(title)

	fmt.Print("우선순위를 선택하세요 (높음, 중간, 낮음): ")
	priority, _ := reader.ReadString('\n')
	priority = strings.TrimSpace(priority)
	if priority != "높음" && priority != "중간" && priority != "낮음" {
		fmt.Println("잘못된 우선순위입니다. '중간'으로 설정됩니다.")
		priority = "중간"
	}

	fmt.Print("카테고리를 입력하세요: ")
	category, _ := reader.ReadString('\n')
	category = strings.TrimSpace(category)
	if category == "" {
		category = "기타"
	}

	fmt.Print("마감일을 입력하세요 (YYYY-MM-DD, 없으면 엔터): ")
	dueDate, _ := reader.ReadString('\n')
	dueDate = strings.TrimSpace(dueDate)
	if dueDate != "" {
		_, err := time.Parse("2006-01-02", dueDate)
		if err != nil {
			fmt.Println("잘못된 날짜 형식입니다. 마감일이 설정되지 않습니다.")
			dueDate = ""
		}
	}

	fmt.Print("반복 주기를 선택하세요 (없음, 매일, 매주, 매월): ")
	repeat, _ := reader.ReadString('\n')
	repeat = strings.TrimSpace(repeat)
	if repeat != "없음" && repeat != "매일" && repeat != "매주" && repeat != "매월" {
		fmt.Println("잘못된 입력입니다. '없음'으로 설정됩니다.")
		repeat = "없음"
	}

	id := t.getNextID()
	newTask := Task{
		ID:        id,
		Title:     title,
		Priority:  priority,
		Category:  category,
		DueDate:   dueDate,
		Repeat:    repeat,
		Complete:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Tasks = append(t.Tasks, newTask)
	fmt.Println("할 일이 추가되었습니다.")
}

// completeTask는 할 일을 완료 처리합니다.
func (t *TaskList) completeTask(reader *bufio.Reader) {
	fmt.Print("완료할 할 일의 번호를 입력하세요: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	id, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("잘못된 입력입니다.")
		return
	}

	for i, task := range t.Tasks {
		if task.ID == id {
			if t.Tasks[i].Complete {
				fmt.Println("이미 완료된 할 일입니다.")
				return
			}
			t.Tasks[i].Complete = true
			t.Tasks[i].UpdatedAt = time.Now()
			fmt.Println("할 일을 완료 처리했습니다.")
			return
		}
	}
	fmt.Println("해당 번호의 할 일을 찾을 수 없습니다.")
}

// deleteTask는 할 일을 삭제합니다.
func (t *TaskList) deleteTask(reader *bufio.Reader) {
	fmt.Print("삭제할 할 일의 번호를 입력하세요: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	id, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("잘못된 입력입니다.")
		return
	}

	for i, task := range t.Tasks {
		if task.ID == id {
			t.Tasks = append(t.Tasks[:i], t.Tasks[i+1:]...)
			fmt.Println("할 일을 삭제했습니다.")
			return
		}
	}
	fmt.Println("해당 번호의 할 일을 찾을 수 없습니다.")
}

// editTask는 기존 할 일의 정보를 수정합니다.
func (t *TaskList) editTask(reader *bufio.Reader) {
	fmt.Print("수정할 할 일의 번호를 입력하세요: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	id, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("잘못된 입력입니다.")
		return
	}

	for i, task := range t.Tasks {
		if task.ID == id {
			fmt.Printf("현재 제목: %s\n", task.Title)
			fmt.Print("새로운 제목을 입력하세요 (변경하지 않으려면 엔터): ")
			newTitle, _ := reader.ReadString('\n')
			newTitle = strings.TrimSpace(newTitle)
			if newTitle != "" {
				t.Tasks[i].Title = newTitle
			}

			fmt.Printf("현재 우선순위: %s\n", task.Priority)
			fmt.Print("새로운 우선순위를 입력하세요 (높음, 중간, 낮음, 변경하지 않으려면 엔터): ")
			newPriority, _ := reader.ReadString('\n')
			newPriority = strings.TrimSpace(newPriority)
			if newPriority != "" {
				if newPriority != "높음" && newPriority != "중간" && newPriority != "낮음" {
					fmt.Println("잘못된 우선순위입니다. 변경되지 않습니다.")
				} else {
					t.Tasks[i].Priority = newPriority
				}
			}

			fmt.Printf("현재 카테고리: %s\n", task.Category)
			fmt.Print("새로운 카테고리를 입력하세요 (변경하지 않으려면 엔터): ")
			newCategory, _ := reader.ReadString('\n')
			newCategory = strings.TrimSpace(newCategory)
			if newCategory != "" {
				t.Tasks[i].Category = newCategory
			}

			fmt.Printf("현재 마감일: %s\n", task.DueDate)
			fmt.Print("새로운 마감일을 입력하세요 (YYYY-MM-DD, 변경하지 않으려면 엔터): ")
			newDueDate, _ := reader.ReadString('\n')
			newDueDate = strings.TrimSpace(newDueDate)
			if newDueDate != "" {
				_, err := time.Parse("2006-01-02", newDueDate)
				if err != nil {
					fmt.Println("잘못된 날짜 형식입니다. 마감일이 변경되지 않습니다.")
				} else {
					t.Tasks[i].DueDate = newDueDate
				}
			}

			fmt.Printf("현재 반복 주기: %s\n", task.Repeat)
			fmt.Print("새로운 반복 주기를 입력하세요 (없음, 매일, 매주, 매월, 변경하지 않으려면 엔터): ")
			newRepeat, _ := reader.ReadString('\n')
			newRepeat = strings.TrimSpace(newRepeat)
			if newRepeat != "" {
				if newRepeat != "없음" && newRepeat != "매일" && newRepeat != "매주" && newRepeat != "매월" {
					fmt.Println("잘못된 입력입니다. 변경되지 않습니다.")
				} else {
					t.Tasks[i].Repeat = newRepeat
				}
			}

			t.Tasks[i].UpdatedAt = time.Now()
			fmt.Println("할 일이 수정되었습니다.")
			return
		}
	}
	fmt.Println("해당 번호의 할 일을 찾을 수 없습니다.")
}

// searchTasks는 제목을 기반으로 할 일을 검색합니다.
func (t *TaskList) searchTasks(reader *bufio.Reader) {
	fmt.Print("검색할 키워드를 입력하세요: ")
	keyword, _ := reader.ReadString('\n')
	keyword = strings.TrimSpace(keyword)

	found := false
	fmt.Println("\n검색 결과:")
	for _, task := range t.Tasks {
		if strings.Contains(task.Title, keyword) {
			status := "[ ]"
			if task.Complete {
				status = "[X]"
			}
			fmt.Printf("%d. %s %s - 우선순위: %s, 카테고리: %s\n", task.ID, status, task.Title, task.Priority, task.Category)
			found = true
		}
	}
	if !found {
		fmt.Println("일치하는 할 일이 없습니다.")
	}
}

// showStatistics는 할 일 통계를 출력합니다.
func (t *TaskList) showStatistics() {
	total := len(t.Tasks)
	completed := 0
	pending := 0

	for _, task := range t.Tasks {
		if task.Complete {
			completed++
		} else {
			pending++
		}
	}

	fmt.Println("\n할 일 통계:")
	fmt.Printf("- 전체 할 일 수: %d\n", total)
	fmt.Printf("- 완료된 할 일 수: %d\n", completed)
	fmt.Printf("- 남은 할 일 수: %d\n", pending)
}

// getNextID는 새로운 할 일에 사용할 다음 ID를 반환합니다.
func (t *TaskList) getNextID() int {
	maxID := 0
	for _, task := range t.Tasks {
		if task.ID > maxID {
			maxID = task.ID
		}
	}
	return maxID + 1
}

// 정렬 함수들
func (t *TaskList) sortByPriority() {
	sort.SliceStable(t.Tasks, func(i, j int) bool {
		return priorityValue(t.Tasks[i].Priority) < priorityValue(t.Tasks[j].Priority)
	})
}

func (t *TaskList) sortByDueDate() {
	sort.SliceStable(t.Tasks, func(i, j int) bool {
		if t.Tasks[i].DueDate == "" {
			return false
		}
		if t.Tasks[j].DueDate == "" {
			return true
		}
		return t.Tasks[i].DueDate < t.Tasks[j].DueDate
	})
}

func (t *TaskList) sortByCreatedAt() {
	sort.SliceStable(t.Tasks, func(i, j int) bool {
		return t.Tasks[i].CreatedAt.Before(t.Tasks[j].CreatedAt)
	})
}

func (t *TaskList) sortByCategory() {
	sort.SliceStable(t.Tasks, func(i, j int) bool {
		return t.Tasks[i].Category < t.Tasks[j].Category
	})
}

func priorityValue(p string) int {
	switch p {
	case "높음":
		return 1
	case "중간":
		return 2
	case "낮음":
		return 3
	default:
		return 4
	}
}
