package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const dataDir = "data"

type Law struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type CreateLawRequest struct {
	Title          string `json:"title"`
	ArticleTitle   string `json:"article_title"`
	ArticleContent string `json:"article_content"`
	Author         string `json:"author"`
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getLaws(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	files, err := os.ReadDir(dataDir)
	if err != nil {
		http.Error(w, "Could not list laws", http.StatusInternalServerError)
		return
	}

	var laws []Law
	for _, file := range files {
		if file.IsDir() {
			lawID := file.Name()
			titlePath := filepath.Join(dataDir, lawID, "title.txt")
			titleBytes, err := os.ReadFile(titlePath)
			if err != nil {
				continue
			}
			laws = append(laws, Law{ID: lawID, Title: strings.TrimSpace(string(titleBytes))})
		}
	}
	if laws == nil {
		laws = []Law{}
	}
	json.NewEncoder(w).Encode(laws)
}

func createLaw(w http.ResponseWriter, r *http.Request) {
	var req CreateLawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Title == "" || req.ArticleTitle == "" || req.ArticleContent == "" || req.Author == "" {
		http.Error(w, "Title, ArticleTitle, ArticleContent, and Author are required", http.StatusBadRequest)
		return
	}

	lawID := uuid.New().String()
	repoPath := filepath.Join(dataDir, lawID)

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		http.Error(w, "Failed to create law", http.StatusInternalServerError)
		return
	}

	runCmd := func(command string, args ...string) error {
		cmd := exec.Command(command, args...)
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Printf("Failed to run '%s': %v\nOutput: %s", command, err, string(output))
			return err
		}
		return nil
	}

	if err := runCmd("git", "init"); err != nil { http.Error(w, "Failed git init", 500); return }

	titlePath := filepath.Join(repoPath, "title.txt")
	if err := os.WriteFile(titlePath, []byte(req.Title), 0644); err != nil {
		http.Error(w, "Failed to create law metadata", 500)
		return
	}

	// Create the first article file
	articleFilename := "article-1.md" // Simple name for the first article
	articlePath := filepath.Join(repoPath, articleFilename)
	if err := os.WriteFile(articlePath, []byte(req.ArticleContent), 0644); err != nil {
		http.Error(w, "Failed to create article file", 500)
		return
	}

	if err := runCmd("git", "add", "."); err != nil { http.Error(w, "Failed git add", 500); return }

	commitMessage := "Initial commit: create law and add article: " + req.ArticleTitle
	author := req.Author + " <" + req.Author + "@example.com>" // Create a dummy email for the author
	if err := runCmd("git", "commit", "--author", author, "-m", commitMessage); err != nil {
		http.Error(w, "Failed git commit", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Law{ID: lawID, Title: req.Title})
}

type CreateArticleRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

func createArticle(w http.ResponseWriter, r *http.Request) {
	lawID := strings.TrimPrefix(r.URL.Path, "/api/laws/")
	lawID = strings.TrimSuffix(lawID, "/articles")

	repoPath := filepath.Join(dataDir, lawID)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		http.Error(w, "Law not found", http.StatusNotFound)
		return
	}

	var req CreateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Title == "" || req.Content == "" || req.Author == "" {
		http.Error(w, "Title, Content, and Author are required", http.StatusBadRequest)
		return
	}

	// Determine the next article number
	files, err := os.ReadDir(repoPath)
	if err != nil {
		http.Error(w, "Could not read law directory", 500)
		return
	}
	articleCount := 0
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "article-") && strings.HasSuffix(file.Name(), ".md") {
			articleCount++
		}
	}
	articleFilename := "article-" + fmt.Sprintf("%d", articleCount+1) + ".md"
	articlePath := filepath.Join(repoPath, articleFilename)

	if err := os.WriteFile(articlePath, []byte(req.Content), 0644); err != nil {
		http.Error(w, "Failed to create article file", 500)
		return
	}

	runCmd := func(command string, args ...string) error {
		cmd := exec.Command(command, args...)
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Printf("Failed to run '%s': %v\nOutput: %s", command, err, string(output))
			return err
		}
		return nil
	}

	if err := runCmd("git", "add", "."); err != nil { http.Error(w, "Failed git add", 500); return }

	commitMessage := "Add article: " + req.Title
	author := req.Author + " <" + req.Author + "@example.com>"
	if err := runCmd("git", "commit", "--author", author, "-m", commitMessage); err != nil {
		http.Error(w, "Failed git commit", 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func lawsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getLaws(w, r)
	case http.MethodPost:
		createLaw(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func lawRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Route for creating an article within a law: POST /api/laws/{law_id}/articles
	if strings.HasSuffix(path, "/articles") {
		if r.Method == http.MethodPost {
			createArticle(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Route for listing laws (GET) or creating a law (POST): /api/laws
	// The handler is registered with a trailing slash, so we might get /api/laws/
	// We check for both exact match and match with trailing slash.
	if path == "/api/laws" || path == "/api/laws/" {
		lawsHandler(w, r)
		return
	}

	http.Error(w, "Not Found", http.StatusNotFound)
}

func main() {
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		os.Mkdir(dataDir, 0755)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/laws/", lawRouter) // Note the trailing slash
	handler := corsMiddleware(mux)
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
