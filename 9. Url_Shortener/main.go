package main

import (
    "html/template"
    "log"
    "math/rand"
    "net/http"
    "net/url"
    "regexp"
    "strconv"
    "sync"
    "time"
)

type URLData struct {
    LongURL     string
    CreatedAt   time.Time
    Expiration  time.Time
    AccessCount int
}

var (
    urlMap  = make(map[string]*URLData)
    mutex   = sync.RWMutex{}
    letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func main() {
    rand.Seed(time.Now().UnixNano())

    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/shorten", shortenHandler)
    http.HandleFunc("/s/", redirectHandler)
    http.HandleFunc("/stats", statsHandler)

    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    log.Println("서버가 포트 8080에서 시작되었습니다.")

    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("서버 시작 중 오류 발생:", err)
    }
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }

    tmpl := template.Must(template.ParseFiles("templates/index.html"))
    tmpl.Execute(w, nil)
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "잘못된 요청 방법입니다.", http.StatusMethodNotAllowed)
        return
    }

    longURL := r.FormValue("url")
    customCode := r.FormValue("custom_code")
    expirationDays := r.FormValue("expiration")

    if longURL == "" {
        http.Error(w, "URL을 입력해야 합니다.", http.StatusBadRequest)
        return
    }

    if !isValidURL(longURL) {
        http.Error(w, "유효한 URL이 아닙니다.", http.StatusBadRequest)
        return
    }

    var shortCode string
    if customCode != "" {
        if !isValidCode(customCode) {
            http.Error(w, "유효하지 않은 커스텀 코드입니다. 영숫자와 '-' 및 '_'만 사용할 수 있습니다.", http.StatusBadRequest)
            return
        }
        mutex.RLock()
        _, exists := urlMap[customCode]
        mutex.RUnlock()
        if exists {
            http.Error(w, "이미 사용 중인 커스텀 코드입니다.", http.StatusBadRequest)
            return
        }
        shortCode = customCode
    } else {
        shortCode = generateShortCode()
    }

    var expiration time.Time
    if expirationDays != "" {
        days, err := strconv.Atoi(expirationDays)
        if err != nil || days <= 0 {
            http.Error(w, "유효한 만료 일수를 입력해야 합니다.", http.StatusBadRequest)
            return
        }
        expiration = time.Now().AddDate(0, 0, days)
    }

    mutex.Lock()
    urlMap[shortCode] = &URLData{
        LongURL:     longURL,
        CreatedAt:   time.Now(),
        Expiration:  expiration,
        AccessCount: 0,
    }
    mutex.Unlock()

    shortURL := "http://" + r.Host + "/s/" + shortCode

    tmpl := template.Must(template.ParseFiles("templates/result.html"))
    tmpl.Execute(w, map[string]string{
        "ShortURL": shortURL,
    })
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
    shortCode := r.URL.Path[len("/s/"):]

    mutex.RLock()
    urlData, ok := urlMap[shortCode]
    mutex.RUnlock()

    if !ok {
        http.NotFound(w, r)
        return
    }

    if !urlData.Expiration.IsZero() && time.Now().After(urlData.Expiration) {
        http.Error(w, "이 URL은 만료되었습니다.", http.StatusGone)
        return
    }

    mutex.Lock()
    urlData.AccessCount++
    mutex.Unlock()

    http.Redirect(w, r, urlData.LongURL, http.StatusFound)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
    mutex.RLock()
    defer mutex.RUnlock()

    tmpl := template.Must(template.ParseFiles("templates/stats.html"))
    tmpl.Execute(w, urlMap)
}

func generateShortCode() string {
    length := 6
    for {
        b := make([]rune, length)
        for i := range b {
            b[i] = letters[rand.Intn(len(letters))]
        }
        code := string(b)

        mutex.RLock()
        _, exists := urlMap[code]
        mutex.RUnlock()

        if !exists {
            return code
        }
    }
}

func isValidURL(str string) bool {
    u, err := url.Parse(str)
    return err == nil && u.Scheme != "" && u.Host != ""
}

func isValidCode(code string) bool {
    re := regexp.MustCompile("^[a-zA-Z0-9_-]+$")
    return re.MatchString(code)
}
