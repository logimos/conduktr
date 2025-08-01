package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/logimos/conduktr/internal/ai"
	"github.com/logimos/conduktr/internal/analytics"
	"github.com/logimos/conduktr/internal/config"
	"github.com/logimos/conduktr/internal/engine"
	"github.com/logimos/conduktr/internal/integrations"
	"github.com/logimos/conduktr/internal/marketplace"
	"github.com/logimos/conduktr/internal/persistence"
	"github.com/logimos/conduktr/internal/triggers"
	"github.com/logimos/conduktr/internal/web"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	cfgFile     string
	workflowDir string
	port        int
	logger      *zap.Logger
)

var rootCmd = &cobra.Command{
	Use:   "reactor",
	Short: "Event-driven workflow engine",
	Long:  "Reactor is a Go-native workflow engine for defining, orchestrating, and executing asynchronous workflows based on incoming events.",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the reactor daemon",
	Long:  "Start the reactor daemon to listen for events and execute workflows",
	RunE:  runDaemon,
}

var validateCmd = &cobra.Command{
	Use:   "validate [workflow-file]",
	Short: "Validate a workflow YAML file",
	Args:  cobra.ExactArgs(1),
	RunE:  validateWorkflow,
}

var executeCmd = &cobra.Command{
	Use:   "execute [workflow-file] [event-data]",
	Short: "Execute a workflow with given event data",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  executeWorkflow,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.reactor.yaml)")
	rootCmd.PersistentFlags().StringVar(&workflowDir, "workflows", "./workflows", "directory containing workflow files")

	runCmd.Flags().IntVarP(&port, "port", "p", 5000, "HTTP server port")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(executeCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".reactor")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	// Initialize logger
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
}

func runDaemon(cmd *cobra.Command, args []string) error {
	cfg := &config.Config{
		WorkflowDir: workflowDir,
		HTTPPort:    port,
		LogLevel:    "info",
	}

	// Initialize persistence
	persist := persistence.NewJSONPersistence("./data")

	// Initialize workflow engine
	workflowEngine := engine.NewEngine(logger, persist)

	// Initialize advanced services
	_ = web.NewDesignerService()
	marketplaceService := marketplace.NewMarketplaceService()
	analyticsDashboard := analytics.NewAnalyticsDashboard(logger)
	aiBuilder := ai.NewAIWorkflowBuilder(logger)
	integrationHub := integrations.NewIntegrationHub(logger)

	// Load workflows from directory
	if err := loadWorkflows(workflowEngine, cfg.WorkflowDir); err != nil {
		return fmt.Errorf("failed to load workflows: %w", err)
	}

	// Start all trigger systems
	logger.Info("Starting trigger systems...")

	// Start HTTP trigger with advanced features
	httpTrigger := triggers.NewHTTPTrigger(logger, workflowEngine, cfg.HTTPPort)

	// Register advanced service routes
	httpTrigger.RegisterAdvancedRoutes(analyticsDashboard, aiBuilder, integrationHub)

	// Register marketplace routes
	httpTrigger.RegisterMarketplaceRoutes(marketplaceService)

	go func() {
		if err := httpTrigger.Start(); err != nil {
			logger.Error("HTTP trigger failed", zap.Error(err))
		}
	}()

	// Start file trigger
	fileTrigger := triggers.NewFileTrigger(logger, workflowEngine)
	go func() {
		if err := fileTrigger.Start(cfg.WorkflowDir); err != nil {
			logger.Error("File trigger failed", zap.Error(err))
		}
	}()

	logger.Info("Reactor daemon started with advanced features",
		zap.Int("port", cfg.HTTPPort),
		zap.String("workflow_dir", cfg.WorkflowDir))

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logger.Info("Shutting down reactor daemon...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Graceful shutdown
	httpTrigger.Stop(ctx)
	fileTrigger.Stop()

	return nil
}

func validateWorkflow(cmd *cobra.Command, args []string) error {
	workflowFile := args[0]

	workflow, err := engine.LoadWorkflowFromFile(workflowFile)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	fmt.Printf("✅ Workflow '%s' is valid\n", workflow.Name)
	fmt.Printf("   Trigger: %s\n", workflow.On.Event)
	fmt.Printf("   Steps: %d\n", len(workflow.Workflow))

	return nil
}

func executeWorkflow(cmd *cobra.Command, args []string) error {
	workflowFile := args[0]
	eventData := "{}"
	if len(args) > 1 {
		eventData = args[1]
	}

	persist := persistence.NewJSONPersistence("./data")
	workflowEngine := engine.NewEngine(logger, persist)

	workflow, err := engine.LoadWorkflowFromFile(workflowFile)
	if err != nil {
		return fmt.Errorf("failed to load workflow: %w", err)
	}

	// Create event context
	eventCtx := &persistence.EventContext{
		Event: &persistence.Event{
			Type:    workflow.On.Event,
			Payload: map[string]interface{}{},
		},
		Variables: make(map[string]interface{}),
	}

	// Parse event data if provided
	if eventData != "{}" {
		// Simple JSON-like parsing for demo
		eventCtx.Event.Payload["data"] = eventData
	}

	fmt.Printf("🚀 Executing workflow: %s\n", workflow.Name)

	ctx := context.Background()
	instanceID, err := workflowEngine.ExecuteWorkflow(ctx, workflow, eventCtx)
	if err != nil {
		return fmt.Errorf("workflow execution failed: %w", err)
	}

	fmt.Printf("✅ Workflow completed successfully (instance: %s)\n", instanceID)
	return nil
}

func loadWorkflows(workflowEngine *engine.Engine, workflowDir string) error {
	return filepath.Walk(workflowDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml") {
			workflow, err := engine.LoadWorkflowFromFile(path)
			if err != nil {
				logger.Warn("Failed to load workflow", zap.String("file", path), zap.Error(err))
				return nil
			}

			workflowEngine.RegisterWorkflow(workflow)
			logger.Info("Loaded workflow", zap.String("name", workflow.Name), zap.String("file", path))
		}

		return nil
	})
}

func Execute() error {
	return rootCmd.Execute()
}
