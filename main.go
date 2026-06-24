// Package main implements a simple web viewer for agmsg messages stored in a SQLite database.
package main

import (
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed templates/*.html
var templatesFS embed.FS

var db *sql.DB
var dbPath string

var (
	teamsMu     sync.Mutex
	knownTeams  = make(map[string]bool)
	lastLogTime time.Time
)

type Message struct {
	ID          int
	Team        string
	FromAgent   string
	ToAgent     string
	Body        string
	CreatedAt   string
	FullTime    string
	ShortTime   string
	IsRight     bool
	BubbleClass string // テンプレート側でフキダシ色を指定するためのフィールド
}

var tpl = template.Must(template.New("msg").Funcs(template.FuncMap{
	"upper":       strings.ToUpper,
	"avatarColor": avatarColor,
	"bubbleColor": bubbleColor, // 新しいテンプレート関数を追加
	"slice": func(s string, start, end int) string {
		if start >= len(s) {
			return ""
		}
		if end > len(s) {
			end = len(s)
		}
		return s[start:end]
	},
}).ParseFS(templatesFS, "templates/message.html"))

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Println("  -db string")
		fmt.Println("    \tPath to SQLite database file (default \"messages.db\")")
		fmt.Println("  -port int")
		fmt.Println("    \tPort to run the server on (default 8080)")
		fmt.Println("  -tail int")
		fmt.Println("    \tNumber of latest messages to display initially (0 for all) (default 40)")
		fmt.Println("  -team string")
		fmt.Println("    \tInitial team to select on load")
	}

	dbPathFlag := flag.String("db", "messages.db", "Path to SQLite database file")
	port := flag.Int("port", 8080, "Port to run the server on")
	tail := flag.Int("tail", 40, "Number of latest messages to display initially (0 for all) (default 40)")
	initialTeam := flag.String("team", "", "Initial team to select on load")
	flag.Parse()

	dbPath = *dbPathFlag
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)", dbPath)
	var err error
	db, err = sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		indexHandler(w, r, *initialTeam)
	})
	http.HandleFunc("/api/messages", func(w http.ResponseWriter, r *http.Request) {
		messagesHandler(w, r, *tail)
	})
	http.HandleFunc("/api/messages/partial", partialMessagesHandler)

	log.Printf("Server starting on http://localhost:%d (tail: %d)", *port, *tail)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func indexHandler(w http.ResponseWriter, _ *http.Request, initialTeam string) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	rows, err := db.Query("SELECT DISTINCT team FROM messages ORDER BY team ASC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = rows.Close()
	}()

	var teams []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err == nil {
			teams = append(teams, t)
		}
	}

	teamsMu.Lock()
	var newTeams []string
	for _, t := range teams {
		if !knownTeams[t] {
			newTeams = append(newTeams, t)
		}
	}
	if len(newTeams) > 0 && time.Since(lastLogTime) > 3*time.Minute {
		for _, t := range newTeams {
			log.Printf("Found new team: %s", t)
		}
		for _, t := range teams {
			knownTeams[t] = true
		}
		lastLogTime = time.Now()
	}
	teamsMu.Unlock()

	t, err := template.ParseFS(templatesFS, "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, struct {
		Teams       []string
		InitialTeam string
	}{Teams: teams, InitialTeam: initialTeam}); err != nil {
		log.Printf("Template execution failed: %v", err)
	}
}

func messagesHandler(w http.ResponseWriter, r *http.Request, limit int) {
	team := r.URL.Query().Get("team")
	if team == "" {
		return
	}

	messages := fetchMessages(team, 0, limit)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.Execute(w, messages); err != nil {
		log.Printf("Template execution failed: %v", err)
	}
}

func partialMessagesHandler(w http.ResponseWriter, r *http.Request) {
	team := r.URL.Query().Get("team")
	lastIDStr := r.URL.Query().Get("last_id")

	if team == "" {
		return
	}

	var lastID int
	_, _ = fmt.Sscanf(lastIDStr, "%d", &lastID)

	messages := fetchMessages(team, lastID, 0)
	if len(messages) > 0 {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tpl.Execute(w, messages); err != nil {
			log.Printf("Template execution failed: %v", err)
		}
	}
}

func fetchMessages(team string, lastID int, limit int) []Message {
	var query string
	var args []interface{}

	if lastID == 0 && limit > 0 {
		query = `
            SELECT id, team, from_agent, to_agent, body, created_at FROM (
                SELECT id, team, from_agent, to_agent, body, created_at
                FROM messages
                WHERE team = ?
                ORDER BY created_at DESC
                LIMIT ?
            ) AS sub
            ORDER BY created_at ASC
        `
		args = []interface{}{team, limit}
	} else {
		query = `
            SELECT id, team, from_agent, to_agent, body, created_at
            FROM messages
            WHERE team = ?
        `
		args = []interface{}{team}

		if lastID > 0 {
			query += " AND id > ?"
			args = append(args, lastID)
		}
		query += " ORDER BY created_at ASC"
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error querying messages for team %s: %v", team, err)
		return nil
	}
	defer func() {
		_ = rows.Close()
	}()

	var firstSpeaker string
	if lastID > 0 {
		_ = db.QueryRow("SELECT from_agent FROM messages WHERE team = ? ORDER BY created_at ASC LIMIT 1", team).Scan(&firstSpeaker)
	}

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.Team, &m.FromAgent, &m.ToAgent, &m.Body, &m.CreatedAt); err != nil {
			continue
		}
		m.FullTime, m.ShortTime = formatISOTime(m.CreatedAt)

		if firstSpeaker == "" {
			firstSpeaker = m.FromAgent
		}

		m.IsRight = (m.FromAgent == firstSpeaker)
		m.BubbleClass = bubbleColor(m.FromAgent, m.IsRight)

		messages = append(messages, m)
	}
	return messages
}

func formatISOTime(iso string) (string, string) {
	var t time.Time
	var err error

	if t, err = time.Parse(time.RFC3339, iso); err != nil {
		if t, err = time.Parse("2006-01-02 15:04:05", iso); err != nil {
			if t, err = time.Parse("2006-01-02T15:04:05", iso); err != nil {
				return iso, iso
			}
		}
	}

	local := t.Local()
	return local.Format("2006-01-02 15:04:05"), local.Format("15:04:05")
}

func avatarColor(name string) string {
	colors := []string{
		"bg-red-500", "bg-orange-500", "bg-amber-500", "bg-yellow-500",
		"bg-lime-500", "bg-green-500", "bg-emerald-500", "bg-teal-500",
		"bg-cyan-500", "bg-sky-500", "bg-blue-500", "bg-indigo-500",
		"bg-violet-500", "bg-purple-500", "bg-fuchsia-500", "bg-pink-500",
	}
	if len(name) == 0 {
		return colors[0]
	}
	sum := 0
	for _, c := range name {
		sum += int(c)
	}
	return colors[sum%len(colors)]
}

func bubbleColor(name string, isRight bool) string {
	if isRight {
		return "bg-[#85e249] text-black"
	}

	lightColors := []string{
		"bg-[#ffffff] text-black",
		"bg-[#eef5ff] text-black",
		"bg-[#f0f9eb] text-black",
		"bg-[#f5f0fa] text-black",
		"bg-[#fff0f2] text-black",
		"bg-[#fff7ec] text-black",
	}
	
	if len(name) == 0 {
		return lightColors[0]
	}
	sum := 0
	for _, c := range name {
		sum += int(c)
	}
	return lightColors[sum%len(lightColors)]
}
