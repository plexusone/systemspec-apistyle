package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/plexusone/systemspec-apistyle/pkg/lint"
	"github.com/plexusone/systemspec-apistyle/pkg/profile"
	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

var (
	servePort   int
	serveHost   string
	serveCORS   bool
	serveWebDir string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the systemspec-apistyle web server",
	Long: `Start a web server that serves the systemspec-apistyle UI and provides
REST API endpoints for linting OpenAPI specifications.

The server provides:
  - Static file serving for the web UI
  - POST /api/lint - Lint an OpenAPI specification
  - GET /api/profiles - List available style profiles

Examples:
  api-style serve
  api-style serve --port 8080
  api-style serve --host 0.0.0.0 --port 3000
  api-style serve --web-dir ./web/dist`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 3000, "Port to listen on")
	serveCmd.Flags().StringVar(&serveHost, "host", "localhost", "Host to bind to")
	serveCmd.Flags().BoolVar(&serveCORS, "cors", true, "Enable CORS for development")
	serveCmd.Flags().StringVar(&serveWebDir, "web-dir", "", "Directory containing web UI files (default: embedded)")
}

// APIServer handles HTTP requests for the systemspec-apistyle service.
type APIServer struct {
	mux *http.ServeMux
}

// NewAPIServer creates a new API server.
func NewAPIServer(webFS fs.FS) *APIServer {
	s := &APIServer{
		mux: http.NewServeMux(),
	}

	// API routes
	s.mux.HandleFunc("/api/lint", s.handleLint)
	s.mux.HandleFunc("/api/profiles", s.handleProfiles)
	s.mux.HandleFunc("/api/health", s.handleHealth)

	// Static file serving
	if webFS != nil {
		fileServer := http.FileServer(http.FS(webFS))
		s.mux.Handle("/", s.spaHandler(fileServer))
	}

	return s
}

// ServeHTTP implements http.Handler.
func (s *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers if enabled
	if serveCORS {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	s.mux.ServeHTTP(w, r)
}

// spaHandler wraps a file server to support single-page app routing.
func (s *APIServer) spaHandler(fileServer http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Don't handle API routes
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Try to serve the file
		fileServer.ServeHTTP(w, r)
	}
}

// LintRequest is the request body for POST /api/lint.
type LintRequest struct {
	Spec    string `json:"spec"`
	Profile string `json:"profile"`
}

// LintResponse is the response body for POST /api/lint.
type LintResponse struct {
	Status     string         `json:"status"`
	Violations []ViolationDTO `json:"violations"`
	Summary    SummaryDTO     `json:"summary"`
	Profile    string         `json:"profile"`
	Duration   string         `json:"duration,omitempty"`
}

// ViolationDTO is the JSON representation of a violation.
type ViolationDTO struct {
	RuleID   string `json:"ruleId"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Path     string `json:"path"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
}

// SummaryDTO is the JSON representation of violation counts.
type SummaryDTO struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Infos    int `json:"infos"`
	Hints    int `json:"hints"`
	Total    int `json:"total"`
}

// ProfileDTO is the JSON representation of a profile.
type ProfileDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	RuleCount   int    `json:"ruleCount"`
}

// handleLint handles POST /api/lint requests.
func (s *APIServer) handleLint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.jsonError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	var req LintRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.jsonError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Spec == "" {
		s.jsonError(w, "spec is required", http.StatusBadRequest)
		return
	}

	// Default profile
	if req.Profile == "" {
		req.Profile = "default"
	}

	// Perform linting
	start := time.Now()
	ctx := context.Background()
	opts := &lint.Options{
		FileName: "spec.yaml",
		Profile:  req.Profile,
	}

	var report *types.LintReport

	// Load profile if specified
	if req.Profile != "" && req.Profile != "vacuum" {
		styleSpec, loadErr := profile.Load(req.Profile)
		if loadErr != nil {
			// Fall back to vacuum defaults
			report, err = lint.WithDefaults(ctx, []byte(req.Spec), opts)
		} else {
			linter := lint.NewVacuumLinter(styleSpec)
			report, err = linter.Lint(ctx, []byte(req.Spec), opts)
		}
	} else {
		report, err = lint.WithDefaults(ctx, []byte(req.Spec), opts)
	}

	if err != nil {
		s.jsonError(w, "Linting failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to response
	resp := s.convertReport(report, req.Profile, time.Since(start))
	s.jsonResponse(w, resp)
}

// handleProfiles handles GET /api/profiles requests.
func (s *APIServer) handleProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Return available profiles
	profiles := []ProfileDTO{
		{Name: "default", Description: "Common REST API best practices", Version: "1.0.0"},
		{Name: "azure", Description: "Microsoft Azure API guidelines", Version: "1.0.0"},
		{Name: "google", Description: "Google API Design Guide", Version: "1.0.0"},
		{Name: "zalando", Description: "Zalando RESTful API Guidelines", Version: "1.0.0"},
		{Name: "vacuum", Description: "Vacuum built-in recommended rules", Version: "1.0.0"},
	}

	// Try to get actual rule counts
	for i := range profiles {
		if profiles[i].Name == "vacuum" {
			continue
		}
		spec, err := profile.Load(profiles[i].Name)
		if err == nil {
			profiles[i].RuleCount = len(spec.Rules)
			profiles[i].Version = spec.Version
			if spec.Description != "" {
				profiles[i].Description = spec.Description
			}
		}
	}

	s.jsonResponse(w, profiles)
}

// handleHealth handles GET /api/health requests.
func (s *APIServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	s.jsonResponse(w, map[string]string{
		"status":  "ok",
		"version": "0.1.0",
	})
}

// convertReport converts a LintReport to a LintResponse.
func (s *APIServer) convertReport(report *types.LintReport, profileName string, duration time.Duration) LintResponse {
	violations := make([]ViolationDTO, 0, len(report.Violations))
	for _, v := range report.Violations {
		violations = append(violations, ViolationDTO{
			RuleID:   v.RuleID,
			Severity: string(v.Severity),
			Message:  v.Message,
			Path:     v.Path,
			Line:     v.Line,
			Column:   v.Column,
		})
	}

	return LintResponse{
		Status:     string(report.Status),
		Violations: violations,
		Summary: SummaryDTO{
			Errors:   report.Summary.Errors,
			Warnings: report.Summary.Warnings,
			Infos:    report.Summary.Infos,
			Hints:    report.Summary.Hints,
			Total:    report.Summary.Total,
		},
		Profile:  profileName,
		Duration: duration.String(),
	}
}

// jsonResponse writes a JSON response.
func (s *APIServer) jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// jsonError writes a JSON error response.
func (s *APIServer) jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func runServe(_ *cobra.Command, _ []string) error {
	// Determine web file source
	var webFS fs.FS
	var err error

	if serveWebDir != "" {
		// Use specified directory
		webFS = os.DirFS(serveWebDir)
		fmt.Printf("Serving web UI from: %s\n", serveWebDir)
	} else {
		// Try to find web/dist directory relative to working directory
		candidates := []string{
			"web/dist",
			"../web/dist",
			"../../web/dist",
		}

		for _, dir := range candidates {
			absPath, _ := filepath.Abs(dir)
			if info, statErr := os.Stat(absPath); statErr == nil && info.IsDir() {
				webFS = os.DirFS(absPath)
				fmt.Printf("Serving web UI from: %s\n", absPath)
				break
			}
		}

		if webFS == nil {
			fmt.Println("Warning: Web UI directory not found. API-only mode.")
			fmt.Println("Run 'pnpm build' in the web/ directory to build the UI.")
		}
	}

	// Create server
	server := NewAPIServer(webFS)

	addr := fmt.Sprintf("%s:%d", serveHost, servePort)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      server,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Graceful shutdown
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		fmt.Println("\nShutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
		close(done)
	}()

	fmt.Printf("\nsystemspec-apistyle Server\n")
	fmt.Printf("=====================\n")
	fmt.Printf("Listening on: http://%s\n", addr)
	fmt.Printf("\nAPI Endpoints:\n")
	fmt.Printf("  POST /api/lint      - Lint an OpenAPI specification\n")
	fmt.Printf("  GET  /api/profiles  - List available profiles\n")
	fmt.Printf("  GET  /api/health    - Health check\n")
	fmt.Printf("\nPress Ctrl+C to stop\n\n")

	if err = httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	<-done
	fmt.Println("Server stopped.")
	return nil
}
