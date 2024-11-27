package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/crypto/scrypt"
)

type PasswordEntry struct {
	Website  string `json:"website"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type PasswordManager struct {
	Entries []PasswordEntry `json:"entries"`
	Key     []byte          `json:"-"`
}

const passwordFile = "passwords.enc"

func GenerateKey(password string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(password), salt, 16384, 8, 1, 32)
}

func Encrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func Decrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	return aesGCM.Open(nil, nonce, ciphertext, nil)
}

func LoadPasswords(manager *PasswordManager, masterPassword string) error {
	data, err := ioutil.ReadFile(passwordFile)
	if err != nil {
		if os.IsNotExist(err) {
			manager.Entries = []PasswordEntry{}
			return nil
		}
		return err
	}

	if len(data) < 32 {
		return errors.New("invalid password file")
	}

	salt := data[:32]
	ciphertext := data[32:]

	key, err := GenerateKey(masterPassword, salt)
	if err != nil {
		return err
	}
	manager.Key = key

	plaintext, err := Decrypt(ciphertext, key)
	if err != nil {
		return errors.New("failed to decrypt data. Check your master password")
	}

	return json.Unmarshal(plaintext, manager)
}

func SavePasswords(manager *PasswordManager, masterPassword string) error {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return err
	}

	key, err := GenerateKey(masterPassword, salt)
	if err != nil {
		return err
	}
	manager.Key = key

	plaintext, err := json.Marshal(manager)
	if err != nil {
		return err
	}

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		return err
	}

	data := append(salt, ciphertext...)
	return ioutil.WriteFile(passwordFile, data, 0600)
}

func (pm *PasswordManager) AddEntry(entry PasswordEntry) {
	pm.Entries = append(pm.Entries, entry)
}

func (pm *PasswordManager) RemoveEntry(index int) error {
	if index < 0 || index >= len(pm.Entries) {
		return errors.New("invalid index")
	}
	pm.Entries = append(pm.Entries[:index], pm.Entries[index+1:]...)
	return nil
}

func (pm *PasswordManager) SearchEntries(query string) []PasswordEntry {
	var results []PasswordEntry
	for _, entry := range pm.Entries {
		if strings.Contains(strings.ToLower(entry.Website), strings.ToLower(query)) {
			results = append(results, entry)
		}
	}
	return results
}

func GeneratePassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>/?"

	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

func promptInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func promptPassword(prompt string) string {
	return promptInput(prompt)
}

func displayMenu() int {
	fmt.Println("\n===== 터미널 비밀번호 관리자 =====")
	fmt.Println("1. 비밀번호 항목 추가")
	fmt.Println("2. 비밀번호 목록 조회")
	fmt.Println("3. 비밀번호 항목 삭제")
	fmt.Println("4. 비밀번호 검색")
	fmt.Println("5. 랜덤 비밀번호 생성")
	fmt.Println("6. 프로그램 종료")
	fmt.Print("원하는 작업의 번호를 입력하세요: ")

	choiceStr := promptInput("")
	choice, err := strconv.Atoi(choiceStr)
	if err != nil {
		return -1
	}
	return choice
}

func mainMenu() {
	fmt.Println("===== 터미널 비밀번호 관리자 =====")
	masterPassword := promptPassword("마스터 비밀번호를 입력하세요: ")

	manager := &PasswordManager{}
	if err := LoadPasswords(manager, masterPassword); err != nil {
		color.Red("오류: %v", err)
		os.Exit(1)
	}

	for {
		choice := displayMenu()
		switch choice {
		case 1:
			website := promptInput("Website: ")
			username := promptInput("Username: ")
			password := promptInput("Password (leave empty to generate): ")
			if password == "" {
				password = GeneratePassword(16)
				fmt.Println("Generated Password:", password)
			}
			entry := PasswordEntry{
				Website:  website,
				Username: username,
				Password: password,
			}
			manager.AddEntry(entry)
			if err := SavePasswords(manager, masterPassword); err != nil {
				color.Red("오류: %v", err)
			} else {
				color.Green("Password entry added successfully.")
			}
		case 2:
			if len(manager.Entries) == 0 {
				fmt.Println("저장된 비밀번호 항목이 없습니다.")
				continue
			}
			fmt.Println("\nStored Password Entries:")
			for i, entry := range manager.Entries {
				fmt.Printf("%d. Website: %s | Username: %s | Password: %s\n", i+1, entry.Website, entry.Username, entry.Password)
			}
		case 3:
			if len(manager.Entries) == 0 {
				fmt.Println("삭제할 비밀번호 항목이 없습니다.")
				continue
			}
			fmt.Println("\nStored Password Entries:")
			for i, entry := range manager.Entries {
				fmt.Printf("%d. Website: %s | Username: %s\n", i+1, entry.Website, entry.Username)
			}
			indexStr := promptInput("삭제할 항목의 번호를 입력하세요: ")
			index, err := strconv.Atoi(indexStr)
			if err != nil || index < 1 || index > len(manager.Entries) {
				color.Red("유효하지 않은 번호입니다.")
				continue
			}
			if err := manager.RemoveEntry(index-1); err != nil {
				color.Red("오류: %v", err)
				continue
			}
			if err := SavePasswords(manager, masterPassword); err != nil {
				color.Red("오류: %v", err)
			} else {
				color.Green("Password entry removed successfully.")
			}
		case 4:
			query := promptInput("검색할 웹사이트 이름을 입력하세요: ")
			results := manager.SearchEntries(query)
			if len(results) == 0 {
				fmt.Println("일치하는 비밀번호 항목이 없습니다.")
				continue
			}
			fmt.Println("\nSearch Results:")
			for i, entry := range results {
				fmt.Printf("%d. Website: %s | Username: %s | Password: %s\n", i+1, entry.Website, entry.Username, entry.Password)
			}
		case 5:
			lengthStr := promptInput("생성할 비밀번호의 길이를 입력하세요: ")
			length, err := strconv.Atoi(lengthStr)
			if err != nil || length <= 0 {
				color.Red("유효하지 않은 길이입니다.")
				continue
			}
			password := GeneratePassword(length)
			fmt.Println("Generated Password:", password)
		case 6:
			fmt.Println("프로그램을 종료합니다.")
			os.Exit(0)
		default:
			color.Red("유효하지 않은 선택입니다. 다시 시도해주세요.")
		}
	}
}

func main() {
	mainMenu()
}
