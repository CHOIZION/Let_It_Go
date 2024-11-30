package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html"
)

type Page struct {
	URL   string
	Links []Link
}

type Link struct {
	Text string
	URL  string
}

var (
	history      []Page
	currentIndex int = -1
	bookmarks        = make(map[string]string)
	tabs             = make([]Page, 0)
	currentTab   int = 0
	userAgent    string
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	loadHistory()
	loadBookmarks()
	userAgent = "TerminalBrowser/1.0"

	for {
		fmt.Println("\n명령을 입력하세요 (open, back, forward, tab, bookmark, search, agent, exit):")
		fmt.Print("> ")
		scanner.Scan()
		command := scanner.Text()
		args := strings.Split(command, " ")

		switch args[0] {
		case "open":
			if len(args) != 2 {
				fmt.Println("사용법: open [URL]")
				continue
			}
			openPage(args[1])
		case "back":
			back()
		case "forward":
			forward()
		case "tab":
			handleTabCommand(args[1:])
		case "bookmark":
			handleBookmarkCommand(args[1:])
		case "search":
			if len(args) != 2 {
				fmt.Println("사용법: search [키워드]")
				continue
			}
			searchInPage(args[1])
		case "agent":
			if len(args) != 2 {
				fmt.Println("사용법: agent [User-Agent 문자열]")
				continue
			}
			userAgent = args[1]
			fmt.Println("User-Agent가 설정되었습니다:", userAgent)
		case "exit":
			saveHistory()
			saveBookmarks()
			fmt.Println("프로그램을 종료합니다.")
			os.Exit(0)
		default:
			fmt.Println("알 수 없는 명령입니다.")
		}
	}
}

func openPage(pageURL string) {
	parsedURL, err := url.Parse(pageURL)
	if err != nil || !parsedURL.IsAbs() {
		fmt.Println("유효한 URL을 입력하세요.")
		return
	}

	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		fmt.Println("요청 생성 실패:", err)
		return
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("페이지를 불러올 수 없습니다:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("페이지를 불러올 수 없습니다. 상태 코드:", resp.StatusCode)
		return
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Println("페이지 파싱 실패:", err)
		return
	}

	links := []Link{}
	extractLinks(doc, &links, parsedURL)

	fmt.Println("\n===== 페이지 내용 =====")
	renderNode(doc)

	page := Page{
		URL:   pageURL,
		Links: links,
	}

	if currentIndex < len(history)-1 {
		history = history[:currentIndex+1]
	}
	history = append(history, page)
	currentIndex++

	if currentTab >= len(tabs) {
		tabs = append(tabs, page)
	} else {
		tabs[currentTab] = page
	}

	handleLinks(links)
}

func extractLinks(n *html.Node, links *[]Link, base *url.URL) {
	if n.Type == html.ElementNode && n.Data == "a" {
		var href, text string
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				href = attr.Val
				break
			}
		}
		if href != "" {
			text = getTextContent(n)
			resolvedURL := resolveURL(base, href)
			*links = append(*links, Link{Text: text, URL: resolvedURL})
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractLinks(c, links, base)
	}
}

func getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var result string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result += getTextContent(c)
	}
	return result
}

func resolveURL(base *url.URL, href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	parsedHref, err := url.Parse(href)
	if err != nil {
		return ""
	}
	resolvedURL := base.ResolveReference(parsedHref)
	return resolvedURL.String()
}

func renderNode(n *html.Node) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			fmt.Println(text)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		renderNode(c)
	}
}

func handleLinks(links []Link) {
	if len(links) == 0 {
		return
	}
	fmt.Println("\n===== 링크 목록 =====")
	for i, link := range links {
		fmt.Printf("[%d] %s\n", i+1, link.Text)
	}
	fmt.Println("\n이동할 링크 번호를 입력하세요 (무시하려면 Enter):")
	fmt.Print("> ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()
	if input == "" {
		return
	}
	var index int
	_, err := fmt.Sscanf(input, "%d", &index)
	if err != nil || index < 1 || index > len(links) {
		fmt.Println("유효한 번호를 입력하세요.")
		return
	}
	openPage(links[index-1].URL)
}

func back() {
	if currentIndex <= 0 {
		fmt.Println("이전 페이지가 없습니다.")
		return
	}
	currentIndex--
	page := history[currentIndex]
	fmt.Println("\n===== 이전 페이지로 이동 =====")
	fmt.Println("URL:", page.URL)
	openPage(page.URL)
}

func forward() {
	if currentIndex >= len(history)-1 {
		fmt.Println("다음 페이지가 없습니다.")
		return
	}
	currentIndex++
	page := history[currentIndex]
	fmt.Println("\n===== 다음 페이지로 이동 =====")
	fmt.Println("URL:", page.URL)
	openPage(page.URL)
}

func handleTabCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("사용법: tab [new|switch|list]")
		return
	}
	switch args[0] {
	case "new":
		tabs = append(tabs, Page{})
		currentTab = len(tabs) - 1
		fmt.Println("새 탭이 열렸습니다. 탭 번호:", currentTab)
	case "switch":
		if len(args) != 2 {
			fmt.Println("사용법: tab switch [탭 번호]")
			return
		}
		var tabIndex int
		_, err := fmt.Sscanf(args[1], "%d", &tabIndex)
		if err != nil || tabIndex < 0 || tabIndex >= len(tabs) {
			fmt.Println("유효한 탭 번호를 입력하세요.")
			return
		}
		currentTab = tabIndex
		fmt.Println("탭이 전환되었습니다. 현재 탭 번호:", currentTab)
		if tabs[currentTab].URL != "" {
			openPage(tabs[currentTab].URL)
		}
	case "list":
		fmt.Println("\n===== 탭 목록 =====")
		for i, tab := range tabs {
			status := ""
			if i == currentTab {
				status = "(현재 탭)"
			}
			fmt.Printf("[%d] %s %s\n", i, tab.URL, status)
		}
	default:
		fmt.Println("알 수 없는 탭 명령입니다.")
	}
}

func handleBookmarkCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("사용법: bookmark [add|list|open]")
		return
	}
	switch args[0] {
	case "add":
		if currentIndex < 0 || currentIndex >= len(history) {
			fmt.Println("현재 페이지가 없습니다.")
			return
		}
		page := history[currentIndex]
		bookmarks[page.URL] = page.URL
		fmt.Println("북마크가 추가되었습니다:", page.URL)
	case "list":
		fmt.Println("\n===== 북마크 목록 =====")
		i := 1
		for url := range bookmarks {
			fmt.Printf("[%d] %s\n", i, url)
			i++
		}
	case "open":
		if len(args) != 2 {
			fmt.Println("사용법: bookmark open [번호]")
			return
		}
		var index int
		_, err := fmt.Sscanf(args[1], "%d", &index)
		if err != nil || index < 1 || index > len(bookmarks) {
			fmt.Println("유효한 번호를 입력하세요.")
			return
		}
		i := 1
		for url := range bookmarks {
			if i == index {
				openPage(url)
				return
			}
			i++
		}
	default:
		fmt.Println("알 수 없는 북마크 명령입니다.")
	}
}

func searchInPage(keyword string) {
	if currentIndex < 0 || currentIndex >= len(history) {
		fmt.Println("현재 페이지가 없습니다.")
		return
	}
	page := history[currentIndex]
	req, err := http.NewRequest("GET", page.URL, nil)
	if err != nil {
		fmt.Println("요청 생성 실패:", err)
		return
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("페이지를 불러올 수 없습니다:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("페이지를 불러올 수 없습니다. 상태 코드:", resp.StatusCode)
		return
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Println("페이지 파싱 실패:", err)
		return
	}

	fmt.Println("\n===== 검색 결과 =====")
	found := searchNode(doc, strings.ToLower(keyword))
	if !found {
		fmt.Println("검색 결과가 없습니다.")
	}
}

func searchNode(n *html.Node, keyword string) bool {
	found := false
	if n.Type == html.TextNode {
		text := strings.ToLower(strings.TrimSpace(n.Data))
		if strings.Contains(text, keyword) {
			fmt.Println(n.Data)
			found = true
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if searchNode(c, keyword) {
			found = true
		}
	}
	return found
}

func saveHistory() {
	file, err := os.Create("history.txt")
	if err != nil {
		fmt.Println("히스토리 저장 실패:", err)
		return
	}
	defer file.Close()

	for _, page := range history {
		file.WriteString(page.URL + "\n")
	}
}

func loadHistory() {
	file, err := os.Open("history.txt")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := scanner.Text()
		history = append(history, Page{URL: url})
		currentIndex++
	}
}

func saveBookmarks() {
	file, err := os.Create("bookmarks.txt")
	if err != nil {
		fmt.Println("북마크 저장 실패:", err)
		return
	}
	defer file.Close()

	for url := range bookmarks {
		file.WriteString(url + "\n")
	}
}

func loadBookmarks() {
	file, err := os.Open("bookmarks.txt")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := scanner.Text()
		bookmarks[url] = url
	}
}
