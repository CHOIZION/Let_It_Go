package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const uploadDir = "./uploads"

func main() {
	// 업로드 디렉토리 생성
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err := os.Mkdir(uploadDir, os.ModePerm)
		if err != nil {
			fmt.Printf("업로드 디렉토리 생성 중 오류 발생: %v\n", err)
			return
		}
		fmt.Println("업로드 디렉토리가 생성되었습니다.")
	}

	// HTTP 핸들러 설정
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/files", handleListFiles)
	http.HandleFunc("/download/", handleDownload)
	http.HandleFunc("/delete/", handleDelete)

	// 서버 시작
	fmt.Println("파일 서버가 시작되었습니다. http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("서버 시작 실패: %v\n", err)
	}
}

// 파일 업로드 핸들러
func handleUpload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("파일 업로드 요청을 받았습니다.")

	if r.Method != http.MethodPost {
		http.Error(w, "POST 요청만 가능합니다.", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("파일을 읽는 중 오류 발생: %v", err), http.StatusBadRequest)
		fmt.Printf("파일 읽기 오류: %v\n", err)
		return
	}
	defer file.Close()

	// 파일 저장
	dst, err := os.Create(filepath.Join(uploadDir, header.Filename))
	if err != nil {
		http.Error(w, fmt.Sprintf("파일 저장 중 오류 발생: %v", err), http.StatusInternalServerError)
		fmt.Printf("파일 저장 오류: %v\n", err)
		return
	}
	defer dst.Close()

	// 파일 복사
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, fmt.Sprintf("파일 복사 중 오류 발생: %v", err), http.StatusInternalServerError)
		fmt.Printf("파일 복사 오류: %v\n", err)
		return
	}

	fmt.Printf("파일 업로드 성공: %s\n", header.Filename)
	fmt.Fprintf(w, "파일 업로드 성공: %s\n", header.Filename)
}

// 파일 목록 핸들러
func handleListFiles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("파일 목록 조회 요청을 받았습니다.")

	files, err := os.ReadDir(uploadDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("파일 목록을 읽는 중 오류 발생: %v", err), http.StatusInternalServerError)
		fmt.Printf("파일 목록 읽기 오류: %v\n", err)
		return
	}

	var fileList []string
	for _, file := range files {
		if !file.IsDir() {
			fileList = append(fileList, file.Name())
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(strings.Join(fileList, "\n")))
	fmt.Println("파일 목록 조회 완료")
}

// 파일 다운로드 핸들러
func handleDownload(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/download/")
	filePath := filepath.Join(uploadDir, filename)

	fmt.Printf("파일 다운로드 요청: %s\n", filename)

	// 파일 존재 여부 확인
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("파일이 존재하지 않습니다: %s", filename), http.StatusNotFound)
		fmt.Printf("파일 존재하지 않음: %s\n", filename)
		return
	}

	http.ServeFile(w, r, filePath)
	fmt.Printf("파일 다운로드 완료: %s\n", filename)
}

// 파일 삭제 핸들러
func handleDelete(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/delete/")
	filePath := filepath.Join(uploadDir, filename)

	fmt.Printf("파일 삭제 요청: %s\n", filename)

	// 파일 존재 여부 확인
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("파일이 존재하지 않습니다: %s", filename), http.StatusNotFound)
		fmt.Printf("파일 존재하지 않음: %s\n", filename)
		return
	}

	// 파일 삭제
	if err := os.Remove(filePath); err != nil {
		http.Error(w, fmt.Sprintf("파일 삭제 중 오류 발생: %v", err), http.StatusInternalServerError)
		fmt.Printf("파일 삭제 오류: %v\n", err)
		return
	}

	fmt.Printf("파일 삭제 성공: %s\n", filename)
	fmt.Fprintf(w, "파일 삭제 성공: %s\n", filename)
}
