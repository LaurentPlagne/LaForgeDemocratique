package main

import (
	"encoding/json"
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
	Title string `json:"title"`
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
	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
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

	if err := runCmd("git", "add", "."); err != nil { http.Error(w, "Failed git add", 500); return }
	if err := runCmd("git", "commit", "-m", "Initial commit: create law"); err != nil { http.Error(w, "Failed git commit", 500); return }

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Law{ID: lawID, Title: req.Title})
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

func main() {
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		os.Mkdir(dataDir, 0755)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/laws", lawsHandler)
	handler := corsMiddleware(mux)
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
