package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/openzerg/hydralisk/internal/api"
	"github.com/openzerg/hydralisk/internal/core/types"
	"github.com/openzerg/hydralisk/internal/db"
	"github.com/openzerg/hydralisk/internal/event-bus"
	"github.com/openzerg/hydralisk/internal/llm-client"
	"github.com/openzerg/hydralisk/internal/message-bus"
	"github.com/openzerg/hydralisk/internal/process-manager"
	"github.com/openzerg/hydralisk/internal/service"
	"github.com/openzerg/hydralisk/internal/tools"
)

var (
	host string
	port int
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "hydralisk",
	Short: "Hydralisk - OpenZerg Go Agent Implementation",
	Long:  "Hydralisk is a Go implementation of the OpenZerg AI Agent Platform.",
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Long:  "Start the OpenZerg agent server with Connect-RPC API.",
	Run:   runServe,
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Start the terminal UI",
	Long:  "Start the terminal UI for interacting with the agent.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TUI mode not yet implemented.")
	},
}

func init() {
	serveCmd.Flags().StringVarP(&host, "host", "H", "127.0.0.1", "Server hostname")
	serveCmd.Flags().IntVarP(&port, "port", "p", 15317, "Server port")

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(tuiCmd)
}

func runServe(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get home directory:", err)
		os.Exit(1)
	}

	dbPath := filepath.Join(homeDir, ".openzerg", "openzerg.db")
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create database directory:", err)
		os.Exit(1)
	}

	repository, err := db.NewRepository(dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create repository:", err)
		os.Exit(1)
	}

	outputDir := filepath.Join(homeDir, ".openzerg", "processes")
	eventBus := eventbus.NewEventBus()
	_ = messagebus.NewMessageBus()
	processMgr := processmanager.NewProcessManager(outputDir)
	llmClient := llmclient.NewClient(&types.LLMConfig{})
	toolRegistry := tools.NewToolRegistry()

	registerTools(toolRegistry)

	services := service.CreateServiceLayer(
		repository,
		eventBus,
		llmClient,
		toolRegistry,
		processMgr,
	)

	path, handler := api.NewAgentHandler(services)
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	addr := fmt.Sprintf("%s:%d", host, port)
	server := &http.Server{
		Addr:    addr,
		Handler: corsMiddleware(mux),
	}

	fmt.Printf("Starting Hydralisk server...\n")
	fmt.Printf("Database: %s\n", dbPath)
	fmt.Printf("\n[Connect] Server running at http://%s\n", addr)
	fmt.Printf("[Connect] API: http://%s/openzerg.Agent/\n", addr)
	fmt.Printf("\nServer ready. Press Ctrl+C to stop.\n")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		server.Shutdown(ctx)
		cancel()
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintln(os.Stderr, "Server error:", err)
		os.Exit(1)
	}
}

func registerTools(registry *tools.ToolRegistry) {
	registry.Register(&tools.ReadTool{})
	registry.Register(&tools.WriteTool{})
	registry.Register(&tools.EditTool{})
	registry.Register(&tools.GlobTool{})
	registry.Register(&tools.GrepTool{})
	registry.Register(&tools.LsTool{})
	registry.Register(&tools.JobTool{})
	registry.Register(&tools.QuestionTool{})
	registry.Register(&tools.TodoWriteTool{})
	registry.Register(&tools.TodoReadTool{})
	registry.Register(&tools.BatchTool{})
	registry.Register(&tools.TaskTool{})
	registry.Register(&tools.MessageTool{})
	registry.Register(&tools.MemorySearchTool{})
	registry.Register(&tools.MemoryGetTool{})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "content-type, connect-protocol-version")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
