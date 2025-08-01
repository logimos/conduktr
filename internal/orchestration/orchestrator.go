package orchestration

import (
        "context"
        "fmt"
        "log"
        "sync"
        "time"
)

// ExecutionContext holds workflow execution state
type ExecutionContext struct {
        WorkflowID    string                 `json:"workflowId"`
        ExecutionID   string                 `json:"executionId"`
        StartTime     time.Time              `json:"startTime"`
        Status        ExecutionStatus        `json:"status"`
        CurrentStep   int                    `json:"currentStep"`
        Variables     map[string]interface{} `json:"variables"`
        StepResults   map[string]interface{} `json:"stepResults"`
        ParentContext *ExecutionContext      `json:"parentContext,omitempty"`
        SubWorkflows  []*ExecutionContext    `json:"subWorkflows,omitempty"`
        ErrorDetails  *ErrorDetails          `json:"errorDetails,omitempty"`
        mu            sync.RWMutex           `json:"-"`
}

// ExecutionStatus represents the status of workflow execution
type ExecutionStatus string

const (
        StatusPending    ExecutionStatus = "pending"
        StatusRunning    ExecutionStatus = "running"
        StatusCompleted  ExecutionStatus = "completed"
        StatusFailed     ExecutionStatus = "failed"
        StatusCancelled  ExecutionStatus = "cancelled"
        StatusPaused     ExecutionStatus = "paused"
)

// ErrorDetails contains error information
type ErrorDetails struct {
        Message   string    `json:"message"`
        Step      string    `json:"step"`
        Timestamp time.Time `json:"timestamp"`
        Retries   int       `json:"retries"`
        Fatal     bool      `json:"fatal"`
}

// ParallelExecution defines parallel execution configuration
type ParallelExecution struct {
        Branches    []ExecutionBranch `json:"branches"`
        JoinType    JoinType          `json:"joinType"`
        Timeout     time.Duration     `json:"timeout"`
        FailFast    bool              `json:"failFast"`
        MaxParallel int               `json:"maxParallel"`
}

// ExecutionBranch represents a branch in parallel execution
type ExecutionBranch struct {
        ID        string                 `json:"id"`
        Name      string                 `json:"name"`
        Steps     []WorkflowStep         `json:"steps"`
        Condition string                 `json:"condition,omitempty"`
        Variables map[string]interface{} `json:"variables,omitempty"`
}

// JoinType defines how parallel branches are joined
type JoinType string

const (
        JoinAll    JoinType = "all"    // Wait for all branches
        JoinAny    JoinType = "any"    // Wait for any branch
        JoinFirst  JoinType = "first"  // Take first completed
        JoinCustom JoinType = "custom" // Custom join logic
)

// WorkflowStep represents a step in workflow execution
type WorkflowStep struct {
        ID            string                 `json:"id"`
        Name          string                 `json:"name"`
        Type          StepType               `json:"type"`
        Action        string                 `json:"action"`
        Config        map[string]interface{} `json:"config"`
        Condition     string                 `json:"condition,omitempty"`
        RetryPolicy   *RetryPolicy           `json:"retryPolicy,omitempty"`
        Timeout       time.Duration          `json:"timeout,omitempty"`
        Dependencies  []string               `json:"dependencies,omitempty"`
        OnSuccess     []string               `json:"onSuccess,omitempty"`
        OnFailure     []string               `json:"onFailure,omitempty"`
        Parallel      *ParallelExecution     `json:"parallel,omitempty"`
        SubWorkflow   *SubWorkflowConfig     `json:"subWorkflow,omitempty"`
}

// StepType defines the type of workflow step
type StepType string

const (
        StepTypeAction     StepType = "action"
        StepTypeCondition  StepType = "condition"
        StepTypeParallel   StepType = "parallel"
        StepTypeSubflow    StepType = "subflow"
        StepTypeDelay      StepType = "delay"
        StepTypeLoop       StepType = "loop"
        StepTypeGateway    StepType = "gateway"
)

// RetryPolicy defines retry behavior for steps
type RetryPolicy struct {
        MaxRetries   int           `json:"maxRetries"`
        RetryDelay   time.Duration `json:"retryDelay"`
        BackoffType  BackoffType   `json:"backoffType"`
        MaxDelay     time.Duration `json:"maxDelay"`
        RetryOn      []string      `json:"retryOn"`
        StopOn       []string      `json:"stopOn"`
}

// BackoffType defines retry backoff strategy
type BackoffType string

const (
        BackoffFixed       BackoffType = "fixed"
        BackoffLinear      BackoffType = "linear"
        BackoffExponential BackoffType = "exponential"
        BackoffRandom      BackoffType = "random"
)

// SubWorkflowConfig defines sub-workflow execution
type SubWorkflowConfig struct {
        WorkflowID    string                 `json:"workflowId"`
        Version       string                 `json:"version,omitempty"`
        InputMapping  map[string]string      `json:"inputMapping"`
        OutputMapping map[string]string      `json:"outputMapping"`
        Async         bool                   `json:"async"`
        Timeout       time.Duration          `json:"timeout"`
        Variables     map[string]interface{} `json:"variables"`
}

// Orchestrator manages advanced workflow execution
type Orchestrator struct {
        executions   map[string]*ExecutionContext
        stepExecutor StepExecutor
        mu           sync.RWMutex
        metrics      *ExecutionMetrics
}

// StepExecutor interface for executing individual steps
type StepExecutor interface {
        ExecuteStep(ctx context.Context, step WorkflowStep, execCtx *ExecutionContext) (interface{}, error)
}

// ExecutionMetrics tracks execution statistics
type ExecutionMetrics struct {
        TotalExecutions     int64         `json:"totalExecutions"`
        CompletedExecutions int64         `json:"completedExecutions"`
        FailedExecutions    int64         `json:"failedExecutions"`
        AverageTime         time.Duration `json:"averageTime"`
        ParallelExecutions  int64         `json:"parallelExecutions"`
        SubWorkflowCount    int64         `json:"subWorkflowCount"`
        mu                  sync.RWMutex  `json:"-"`
}

// NewOrchestrator creates a new orchestrator instance
func NewOrchestrator(stepExecutor StepExecutor) *Orchestrator {
        return &Orchestrator{
                executions:   make(map[string]*ExecutionContext),
                stepExecutor: stepExecutor,
                metrics:      &ExecutionMetrics{},
        }
}

// ExecuteWorkflow starts workflow execution with advanced orchestration
func (o *Orchestrator) ExecuteWorkflow(ctx context.Context, workflowID string, steps []WorkflowStep, variables map[string]interface{}) (*ExecutionContext, error) {
        executionID := fmt.Sprintf("exec_%d", time.Now().UnixNano())
        
        execCtx := &ExecutionContext{
                WorkflowID:  workflowID,
                ExecutionID: executionID,
                StartTime:   time.Now(),
                Status:      StatusRunning,
                Variables:   variables,
                StepResults: make(map[string]interface{}),
        }
        
        o.mu.Lock()
        o.executions[executionID] = execCtx
        o.mu.Unlock()
        
        o.metrics.mu.Lock()
        o.metrics.TotalExecutions++
        o.metrics.mu.Unlock()
        
        go o.executeSteps(ctx, execCtx, steps)
        
        return execCtx, nil
}

// executeSteps executes workflow steps with advanced orchestration
func (o *Orchestrator) executeSteps(ctx context.Context, execCtx *ExecutionContext, steps []WorkflowStep) {
        defer func() {
                if r := recover(); r != nil {
                        log.Printf("Panic in workflow execution: %v", r)
                        o.markExecutionFailed(execCtx, fmt.Errorf("execution panic: %v", r))
                }
        }()
        
        for i, step := range steps {
                if ctx.Err() != nil {
                        o.markExecutionCancelled(execCtx)
                        return
                }
                
                execCtx.mu.Lock()
                execCtx.CurrentStep = i
                execCtx.mu.Unlock()
                
                // Check step condition
                if step.Condition != "" && !o.evaluateCondition(step.Condition, execCtx) {
                        log.Printf("Step %s skipped due to condition: %s", step.Name, step.Condition)
                        continue
                }
                
                // Execute step based on type
                result, err := o.executeStep(ctx, step, execCtx)
                if err != nil {
                        if o.shouldRetry(step, err) {
                                result, err = o.retryStep(ctx, step, execCtx)
                        }
                        
                        if err != nil {
                                o.markExecutionFailed(execCtx, err)
                                return
                        }
                }
                
                // Store step result
                execCtx.mu.Lock()
                execCtx.StepResults[step.ID] = result
                execCtx.mu.Unlock()
        }
        
        o.markExecutionCompleted(execCtx)
}

// executeStep executes a single workflow step
func (o *Orchestrator) executeStep(ctx context.Context, step WorkflowStep, execCtx *ExecutionContext) (interface{}, error) {
        // Add timeout if specified
        if step.Timeout > 0 {
                var cancel context.CancelFunc
                ctx, cancel = context.WithTimeout(ctx, step.Timeout)
                defer cancel()
        }
        
        switch step.Type {
        case StepTypeParallel:
                return o.executeParallelStep(ctx, step, execCtx)
        case StepTypeSubflow:
                return o.executeSubWorkflow(ctx, step, execCtx)
        case StepTypeCondition:
                return o.executeConditionStep(ctx, step, execCtx)
        case StepTypeDelay:
                return o.executeDelayStep(ctx, step, execCtx)
        case StepTypeLoop:
                return o.executeLoopStep(ctx, step, execCtx)
        default:
                return o.stepExecutor.ExecuteStep(ctx, step, execCtx)
        }
}

// executeParallelStep executes steps in parallel with advanced coordination
func (o *Orchestrator) executeParallelStep(ctx context.Context, step WorkflowStep, execCtx *ExecutionContext) (interface{}, error) {
        if step.Parallel == nil {
                return nil, fmt.Errorf("parallel configuration missing for step %s", step.ID)
        }
        
        parallel := step.Parallel
        results := make(map[string]interface{})
        errors := make(map[string]error)
        var wg sync.WaitGroup
        var mu sync.Mutex
        
        // Create semaphore for limiting parallelism
        semaphore := make(chan struct{}, parallel.MaxParallel)
        if parallel.MaxParallel <= 0 {
                parallel.MaxParallel = len(parallel.Branches)
                semaphore = make(chan struct{}, parallel.MaxParallel)
        }
        
        // Create context with timeout
        branchCtx := ctx
        if parallel.Timeout > 0 {
                var cancel context.CancelFunc
                branchCtx, cancel = context.WithTimeout(ctx, parallel.Timeout)
                defer cancel()
        }
        
        o.metrics.mu.Lock()
        o.metrics.ParallelExecutions++
        o.metrics.mu.Unlock()
        
        // Execute branches in parallel
        for _, branch := range parallel.Branches {
                wg.Add(1)
                go func(b ExecutionBranch) {
                        defer wg.Done()
                        
                        // Acquire semaphore slot
                        semaphore <- struct{}{}
                        defer func() { <-semaphore }()
                        
                        // Check branch condition
                        if b.Condition != "" && !o.evaluateCondition(b.Condition, execCtx) {
                                mu.Lock()
                                results[b.ID] = "skipped"
                                mu.Unlock()
                                return
                        }
                        
                        // Create branch execution context
                        branchExecCtx := &ExecutionContext{
                                WorkflowID:    execCtx.WorkflowID + "_" + b.ID,
                                ExecutionID:   execCtx.ExecutionID + "_" + b.ID,
                                StartTime:     time.Now(),
                                Status:        StatusRunning,
                                Variables:     o.mergeVariables(execCtx.Variables, b.Variables),
                                StepResults:   make(map[string]interface{}),
                                ParentContext: execCtx,
                        }
                        
                        // Execute branch steps
                        branchResult := make(map[string]interface{})
                        for _, branchStep := range b.Steps {
                                if branchCtx.Err() != nil {
                                        mu.Lock()
                                        errors[b.ID] = branchCtx.Err()
                                        mu.Unlock()
                                        return
                                }
                                
                                stepResult, err := o.executeStep(branchCtx, branchStep, branchExecCtx)
                                if err != nil {
                                        if parallel.FailFast {
                                                mu.Lock()
                                                errors[b.ID] = err
                                                mu.Unlock()
                                                return
                                        }
                                        log.Printf("Branch %s step %s failed: %v", b.ID, branchStep.Name, err)
                                } else {
                                        branchResult[branchStep.ID] = stepResult
                                }
                        }
                        
                        mu.Lock()
                        results[b.ID] = branchResult
                        mu.Unlock()
                }(branch)
        }
        
        // Wait for completion based on join type
        switch parallel.JoinType {
        case JoinAll:
                wg.Wait()
        case JoinAny:
                done := make(chan struct{})
                go func() {
                        wg.Wait()
                        close(done)
                }()
                
                // Wait for first completion or all done
                ticker := time.NewTicker(100 * time.Millisecond)
                defer ticker.Stop()
                
                for {
                        select {
                        case <-done:
                                goto completed
                        case <-ticker.C:
                                mu.Lock()
                                if len(results) > 0 || len(errors) > 0 {
                                        mu.Unlock()
                                        goto completed
                                }
                                mu.Unlock()
                        case <-branchCtx.Done():
                                return nil, branchCtx.Err()
                        }
                }
        case JoinFirst:
                // Similar to JoinAny but returns immediately after first success
                done := make(chan struct{})
                go func() {
                        wg.Wait()
                        close(done)
                }()
                
                ticker := time.NewTicker(50 * time.Millisecond)
                defer ticker.Stop()
                
                for {
                        select {
                        case <-done:
                                goto completed
                        case <-ticker.C:
                                mu.Lock()
                                if len(results) > 0 {
                                        mu.Unlock()
                                        goto completed
                                }
                                mu.Unlock()
                        case <-branchCtx.Done():
                                return nil, branchCtx.Err()
                        }
                }
        default:
                wg.Wait()
        }
        
completed:
        // Check for errors
        if len(errors) > 0 && parallel.FailFast {
                for branchID, err := range errors {
                        return nil, fmt.Errorf("branch %s failed: %v", branchID, err)
                }
        }
        
        return map[string]interface{}{
                "results": results,
                "errors":  errors,
                "status":  "completed",
        }, nil
}

// executeSubWorkflow executes a sub-workflow
func (o *Orchestrator) executeSubWorkflow(ctx context.Context, step WorkflowStep, execCtx *ExecutionContext) (interface{}, error) {
        if step.SubWorkflow == nil {
                return nil, fmt.Errorf("subworkflow configuration missing for step %s", step.ID)
        }
        
        subConfig := step.SubWorkflow
        
        // Map input variables
        subVariables := make(map[string]interface{})
        for subVar, parentVar := range subConfig.InputMapping {
                if value, exists := execCtx.Variables[parentVar]; exists {
                        subVariables[subVar] = value
                }
        }
        
        // Add sub-workflow specific variables
        for k, v := range subConfig.Variables {
                subVariables[k] = v
        }
        
        o.metrics.mu.Lock()
        o.metrics.SubWorkflowCount++
        o.metrics.mu.Unlock()
        
        // Create sub-workflow execution context
        subExecCtx := &ExecutionContext{
                WorkflowID:    subConfig.WorkflowID,
                ExecutionID:   execCtx.ExecutionID + "_sub_" + step.ID,
                StartTime:     time.Now(),
                Status:        StatusRunning,
                Variables:     subVariables,
                StepResults:   make(map[string]interface{}),
                ParentContext: execCtx,
        }
        
        // Add to parent's sub-workflows
        execCtx.mu.Lock()
        execCtx.SubWorkflows = append(execCtx.SubWorkflows, subExecCtx)
        execCtx.mu.Unlock()
        
        // Execute sub-workflow (placeholder - would load actual workflow definition)
        // For now, simulate sub-workflow execution
        log.Printf("Executing sub-workflow: %s", subConfig.WorkflowID)
        
        if subConfig.Async {
                // Start async execution
                go func() {
                        time.Sleep(1 * time.Second) // Simulate execution
                        subExecCtx.Status = StatusCompleted
                }()
                
                return map[string]interface{}{
                        "status":      "started",
                        "executionId": subExecCtx.ExecutionID,
                        "async":       true,
                }, nil
        } else {
                // Synchronous execution
                time.Sleep(500 * time.Millisecond) // Simulate execution
                subExecCtx.Status = StatusCompleted
                
                // Map output variables back to parent
                result := make(map[string]interface{})
                for parentVar, subVar := range subConfig.OutputMapping {
                        if value, exists := subExecCtx.StepResults[subVar]; exists {
                                execCtx.Variables[parentVar] = value
                        }
                }
                
                result["status"] = "completed"
                result["executionId"] = subExecCtx.ExecutionID
                result["variables"] = subExecCtx.Variables
                
                return result, nil
        }
}

// executeConditionStep evaluates conditions and routes execution
func (o *Orchestrator) executeConditionStep(ctx context.Context, step WorkflowStep, execCtx *ExecutionContext) (interface{}, error) {
        condition := step.Config["condition"].(string)
        result := o.evaluateCondition(condition, execCtx)
        
        return map[string]interface{}{
                "condition": condition,
                "result":    result,
                "nextSteps": o.getNextSteps(step, result),
        }, nil
}

// executeDelayStep introduces a delay in execution
func (o *Orchestrator) executeDelayStep(ctx context.Context, step WorkflowStep, execCtx *ExecutionContext) (interface{}, error) {
        durationStr, ok := step.Config["duration"].(string)
        if !ok {
                return nil, fmt.Errorf("duration not specified for delay step %s", step.ID)
        }
        
        duration, err := time.ParseDuration(durationStr)
        if err != nil {
                return nil, fmt.Errorf("invalid duration format: %s", durationStr)
        }
        
        select {
        case <-time.After(duration):
                return map[string]interface{}{
                        "delayed": duration.String(),
                        "status":  "completed",
                }, nil
        case <-ctx.Done():
                return nil, ctx.Err()
        }
}

// executeLoopStep executes steps in a loop
func (o *Orchestrator) executeLoopStep(ctx context.Context, step WorkflowStep, execCtx *ExecutionContext) (interface{}, error) {
        loopConfig := step.Config["loop"].(map[string]interface{})
        condition := loopConfig["condition"].(string)
        maxIterations := loopConfig["maxIterations"].(int)
        
        results := make([]interface{}, 0)
        iteration := 0
        
        for iteration < maxIterations {
                if !o.evaluateCondition(condition, execCtx) {
                        break
                }
                
                // Execute loop body (would contain sub-steps)
                result := map[string]interface{}{
                        "iteration": iteration,
                        "timestamp": time.Now(),
                }
                
                results = append(results, result)
                iteration++
                
                // Check for context cancellation
                if ctx.Err() != nil {
                        return nil, ctx.Err()
                }
        }
        
        return map[string]interface{}{
                "iterations": iteration,
                "results":    results,
                "status":     "completed",
        }, nil
}

// Helper methods

func (o *Orchestrator) evaluateCondition(condition string, execCtx *ExecutionContext) bool {
        // Simplified condition evaluation - in production, use a proper expression evaluator
        // For now, return true for demo purposes
        return true
}

func (o *Orchestrator) getNextSteps(step WorkflowStep, conditionResult bool) []string {
        if conditionResult {
                return step.OnSuccess
        }
        return step.OnFailure
}

func (o *Orchestrator) shouldRetry(step WorkflowStep, err error) bool {
        return step.RetryPolicy != nil && step.RetryPolicy.MaxRetries > 0
}

func (o *Orchestrator) retryStep(ctx context.Context, step WorkflowStep, execCtx *ExecutionContext) (interface{}, error) {
        policy := step.RetryPolicy
        
        for attempt := 1; attempt <= policy.MaxRetries; attempt++ {
                // Calculate delay based on backoff type
                delay := o.calculateRetryDelay(policy, attempt)
                
                select {
                case <-time.After(delay):
                        result, err := o.stepExecutor.ExecuteStep(ctx, step, execCtx)
                        if err == nil {
                                return result, nil
                        }
                        
                        log.Printf("Retry %d/%d failed for step %s: %v", attempt, policy.MaxRetries, step.Name, err)
                        
                        if attempt == policy.MaxRetries {
                                return nil, fmt.Errorf("step %s failed after %d retries: %v", step.Name, policy.MaxRetries, err)
                        }
                case <-ctx.Done():
                        return nil, ctx.Err()
                }
        }
        
        return nil, fmt.Errorf("retry attempts exhausted")
}

func (o *Orchestrator) calculateRetryDelay(policy *RetryPolicy, attempt int) time.Duration {
        switch policy.BackoffType {
        case BackoffLinear:
                return policy.RetryDelay * time.Duration(attempt)
        case BackoffExponential:
                delay := policy.RetryDelay * time.Duration(1<<uint(attempt-1))
                if policy.MaxDelay > 0 && delay > policy.MaxDelay {
                        delay = policy.MaxDelay
                }
                return delay
        case BackoffRandom:
                // Simple random between retry delay and 2x retry delay
                return policy.RetryDelay + time.Duration(time.Now().UnixNano()%int64(policy.RetryDelay))
        default:
                return policy.RetryDelay
        }
}

func (o *Orchestrator) mergeVariables(parent, child map[string]interface{}) map[string]interface{} {
        result := make(map[string]interface{})
        
        // Copy parent variables
        for k, v := range parent {
                result[k] = v
        }
        
        // Override with child variables
        for k, v := range child {
                result[k] = v
        }
        
        return result
}

func (o *Orchestrator) markExecutionCompleted(execCtx *ExecutionContext) {
        execCtx.mu.Lock()
        execCtx.Status = StatusCompleted
        execCtx.mu.Unlock()
        
        o.metrics.mu.Lock()
        o.metrics.CompletedExecutions++
        executionTime := time.Since(execCtx.StartTime)
        o.metrics.AverageTime = (o.metrics.AverageTime + executionTime) / 2
        o.metrics.mu.Unlock()
        
        log.Printf("Workflow execution completed: %s", execCtx.ExecutionID)
}

func (o *Orchestrator) markExecutionFailed(execCtx *ExecutionContext, err error) {
        execCtx.mu.Lock()
        execCtx.Status = StatusFailed
        execCtx.ErrorDetails = &ErrorDetails{
                Message:   err.Error(),
                Timestamp: time.Now(),
                Fatal:     true,
        }
        execCtx.mu.Unlock()
        
        o.metrics.mu.Lock()
        o.metrics.FailedExecutions++
        o.metrics.mu.Unlock()
        
        log.Printf("Workflow execution failed: %s - %v", execCtx.ExecutionID, err)
}

func (o *Orchestrator) markExecutionCancelled(execCtx *ExecutionContext) {
        execCtx.mu.Lock()
        execCtx.Status = StatusCancelled
        execCtx.mu.Unlock()
        
        log.Printf("Workflow execution cancelled: %s", execCtx.ExecutionID)
}

// GetExecution returns execution context by ID
func (o *Orchestrator) GetExecution(executionID string) (*ExecutionContext, bool) {
        o.mu.RLock()
        defer o.mu.RUnlock()
        
        execCtx, exists := o.executions[executionID]
        return execCtx, exists
}

// GetMetrics returns execution metrics
func (o *Orchestrator) GetMetrics() *ExecutionMetrics {
        o.metrics.mu.RLock()
        defer o.metrics.mu.RUnlock()
        
        return &ExecutionMetrics{
                TotalExecutions:     o.metrics.TotalExecutions,
                CompletedExecutions: o.metrics.CompletedExecutions,
                FailedExecutions:    o.metrics.FailedExecutions,
                AverageTime:         o.metrics.AverageTime,
                ParallelExecutions:  o.metrics.ParallelExecutions,
                SubWorkflowCount:    o.metrics.SubWorkflowCount,
        }
}

// CancelExecution cancels a running execution
func (o *Orchestrator) CancelExecution(executionID string) error {
        o.mu.RLock()
        execCtx, exists := o.executions[executionID]
        o.mu.RUnlock()
        
        if !exists {
                return fmt.Errorf("execution not found: %s", executionID)
        }
        
        o.markExecutionCancelled(execCtx)
        return nil
}

// PauseExecution pauses a running execution
func (o *Orchestrator) PauseExecution(executionID string) error {
        o.mu.RLock()
        execCtx, exists := o.executions[executionID]
        o.mu.RUnlock()
        
        if !exists {
                return fmt.Errorf("execution not found: %s", executionID)
        }
        
        execCtx.mu.Lock()
        execCtx.Status = StatusPaused
        execCtx.mu.Unlock()
        
        return nil
}

// ResumeExecution resumes a paused execution
func (o *Orchestrator) ResumeExecution(executionID string) error {
        o.mu.RLock()
        execCtx, exists := o.executions[executionID]
        o.mu.RUnlock()
        
        if !exists {
                return fmt.Errorf("execution not found: %s", executionID)
        }
        
        execCtx.mu.Lock()
        if execCtx.Status == StatusPaused {
                execCtx.Status = StatusRunning
        }
        execCtx.mu.Unlock()
        
        return nil
}