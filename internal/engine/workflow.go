package engine

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Workflow represents a complete workflow definition
type Workflow struct {
	Name     string         `yaml:"name"`
	On       TriggerConfig  `yaml:"on"`
	Workflow []WorkflowStep `yaml:"workflow"`
}

// TriggerConfig defines what triggers the workflow
type TriggerConfig struct {
	Event string `yaml:"event"`
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	Name   string                 `yaml:"name"`
	Action string                 `yaml:"action"`
	If     string                 `yaml:"if,omitempty"`
	Config map[string]interface{} `yaml:",inline"`
	Retry  *RetryConfig           `yaml:"retry,omitempty"`
}

// RetryConfig defines retry behavior for a step
type RetryConfig struct {
	Max     int    `yaml:"max"`
	Backoff string `yaml:"backoff"`
}

// LoadWorkflowFromFile loads a workflow definition from a YAML file
func LoadWorkflowFromFile(filename string) (*Workflow, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	return LoadWorkflowFromYAML(data)
}

// LoadWorkflowFromYAML loads a workflow definition from YAML data
func LoadWorkflowFromYAML(data []byte) (*Workflow, error) {
	var workflow Workflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow YAML: %w", err)
	}

	if err := validateWorkflow(&workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	return &workflow, nil
}

// validateWorkflow validates a workflow definition
func validateWorkflow(workflow *Workflow) error {
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if workflow.On.Event == "" {
		return fmt.Errorf("workflow trigger event is required")
	}

	if len(workflow.Workflow) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}

	for i, step := range workflow.Workflow {
		if step.Name == "" {
			return fmt.Errorf("step %d: name is required", i)
		}

		if step.Action == "" {
			return fmt.Errorf("step %d (%s): action is required", i, step.Name)
		}

		// Validate retry configuration
		if step.Retry != nil {
			if step.Retry.Max < 1 {
				return fmt.Errorf("step %d (%s): retry.max must be >= 1", i, step.Name)
			}

			if step.Retry.Backoff != "" && step.Retry.Backoff != "exponential" && step.Retry.Backoff != "linear" {
				return fmt.Errorf("step %d (%s): retry.backoff must be 'exponential' or 'linear'", i, step.Name)
			}
		}
	}

	return nil
}
