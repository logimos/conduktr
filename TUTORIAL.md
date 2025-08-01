# Reactor Workflow Engine Tutorial

## Table of Contents
1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [Understanding Workflows](#understanding-workflows)
4. [Triggers and Events](#triggers-and-events)
5. [Creating Your First Workflow](#creating-your-first-workflow)
6. [Advanced Workflow Features](#advanced-workflow-features)
7. [AI Workflow Builder](#ai-workflow-builder)
8. [Marketplace Templates](#marketplace-templates)
9. [Best Practices](#best-practices)
10. [Troubleshooting](#troubleshooting)

## Introduction

Reactor is a powerful, event-driven workflow orchestration engine written in Go. It allows you to create automated workflows that respond to various events and perform complex sequences of actions.

### Key Concepts
- **Workflows**: YAML-defined sequences of steps that execute when triggered
- **Triggers**: Events that start a workflow (HTTP, file changes, schedules, etc.)
- **Actions**: Individual operations within a workflow (HTTP requests, database operations, etc.)
- **Events**: Data that flows through the workflow system

## Getting Started

### Installation
```bash
# Clone the repository
git clone <repository-url>
cd conduktr

# Build the application
make build

# Start the daemon
./conduktr run
```

### Accessing the Dashboard
Once running, you can access:
- **Main Dashboard**: http://localhost:5000/
- **AI Workflow Builder**: http://localhost:5000/ai/builder
- **Marketplace**: http://localhost:5000/marketplace

## Understanding Workflows

### Workflow Structure
A workflow consists of three main parts:

1. **Metadata**: Name, description, and version
2. **Trigger**: What event starts the workflow
3. **Steps**: The actions to perform

### Basic Workflow Example
```yaml
name: my-first-workflow
on:
  event: user.created

workflow:
  - name: log_event
    action: log.info
    message: "New user created: {{ .event.payload.name }}"
    level: info

  - name: send_welcome_email
    action: email.send
    to: "{{ .event.payload.email }}"
    subject: "Welcome!"
    body: "Welcome to our platform, {{ .event.payload.name }}!"

  - name: create_user_profile
    action: database.insert
    table: users
    data:
      id: "{{ .event.payload.id }}"
      name: "{{ .event.payload.name }}"
      email: "{{ .event.payload.email }}"
      created_at: "{{ .event.timestamp }}"
```

## Triggers and Events

### What are Triggers?
Triggers are events that start a workflow. Reactor supports multiple trigger types:

#### 1. HTTP Triggers
```yaml
on:
  event: http.request
  path: /webhook/user-created
```

**How it works:**
- Sends POST request to `http://localhost:5000/webhook/user-created`
- Workflow receives the request payload as event data
- Workflow executes with the payload

**Example HTTP request:**
```bash
curl -X POST http://localhost:5000/webhook/user-created \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "id": "12345"
  }'
```

#### 2. File Triggers
```yaml
on:
  event: file.created
  path: /uploads
```

**How it works:**
- Monitors the specified directory for new files
- Triggers workflow when files are created
- Passes file metadata as event data

#### 3. Scheduled Triggers
```yaml
on:
  event: scheduled
  cron: "0 */6 * * *"  # Every 6 hours
```

**Cron Syntax:**
- `* * * * *` = minute hour day month weekday
- `0 */6 * * *` = Every 6 hours
- `0 9 * * 1-5` = Weekdays at 9 AM

#### 4. Database Triggers
```yaml
on:
  event: database.change
  table: users
  operation: insert
```

#### 5. Custom Event Triggers
```yaml
on:
  event: custom.event.name
```

### Event Data Structure
Events contain structured data that flows through your workflow:

```json
{
  "event": {
    "type": "user.created",
    "payload": {
      "name": "John Doe",
      "email": "john@example.com",
      "id": "12345"
    },
    "metadata": {
      "timestamp": 1640995200,
      "source": "webhook",
      "user_agent": "curl/7.68.0"
    }
  },
  "variables": {
    "workflow_id": "wf_123",
    "execution_id": "exec_456"
  }
}
```

### Accessing Event Data
In your workflow steps, you can access event data using template syntax:

```yaml
workflow:
  - name: process_user
    action: log.info
    message: "Processing user: {{ .event.payload.name }}"
    
  - name: send_email
    action: email.send
    to: "{{ .event.payload.email }}"
    subject: "Welcome {{ .event.payload.name }}!"
```

## Creating Your First Workflow

### Step 1: Create a Workflow File
Create a new file in the `workflows/` directory:

```bash
touch workflows/my-workflow.yaml
```

### Step 2: Define the Workflow
```yaml
name: user-registration-workflow
on:
  event: user.created

workflow:
  - name: validate_user_data
    action: log.info
    message: "Validating user data for {{ .event.payload.name }}"
    level: info

  - name: send_welcome_email
    action: email.send
    to: "{{ .event.payload.email }}"
    subject: "Welcome to our platform!"
    body: |
      Hi {{ .event.payload.name }},
      
      Welcome to our platform! We're excited to have you on board.
      
      Best regards,
      The Team

  - name: create_user_profile
    action: database.insert
    table: user_profiles
    data:
      user_id: "{{ .event.payload.id }}"
      name: "{{ .event.payload.name }}"
      email: "{{ .event.payload.email }}"
      status: "active"
      created_at: "{{ .event.timestamp }}"

  - name: log_completion
    action: log.info
    message: "User registration completed for {{ .event.payload.name }}"
    level: info
```

### Step 3: Validate the Workflow
```bash
./conduktr validate workflows/my-workflow.yaml
```

### Step 4: Test the Workflow
```bash
# Execute with sample data
./conduktr execute workflows/my-workflow.yaml '{"name":"John Doe","email":"john@example.com","id":"12345"}'
```

## Advanced Workflow Features

### Conditional Execution
```yaml
workflow:
  - name: check_user_type
    action: log.info
    message: "Checking user type"
    
  - name: process_premium_user
    action: email.send
    to: "{{ .event.payload.email }}"
    subject: "Premium Welcome!"
    if: "{{ eq .event.payload.type \"premium\" }}"
    
  - name: process_regular_user
    action: email.send
    to: "{{ .event.payload.email }}"
    subject: "Welcome!"
    if: "{{ eq .event.payload.type \"regular\" }}"
```

### Retry Logic
```yaml
workflow:
  - name: send_notification
    action: http.request
    url: "https://api.slack.com/webhook"
    method: POST
    body:
      text: "New user registered: {{ .event.payload.name }}"
    retry:
      max: 3
      backoff: exponential
      delay: 5s
```

### Parallel Execution
```yaml
workflow:
  - name: parallel_tasks
    parallel: true
    steps:
      - name: send_email
        action: email.send
        to: "{{ .event.payload.email }}"
        
      - name: update_database
        action: database.update
        table: users
        where: "id = {{ .event.payload.id }}"
        
      - name: log_activity
        action: log.info
        message: "User processed"
```

### Error Handling
```yaml
workflow:
  - name: risky_operation
    action: http.request
    url: "https://external-api.com/endpoint"
    timeout: 30s
    on_error:
      - name: log_error
        action: log.error
        message: "External API call failed"
        
      - name: send_alert
        action: email.send
        to: "admin@company.com"
        subject: "Workflow Error Alert"
```

## AI Workflow Builder

### Using the AI Builder
1. Navigate to http://localhost:5000/ai/builder
2. Describe your workflow in plain English
3. Click "Generate Workflow"
4. Review and customize the generated workflow

### Example AI Prompts
```
"When a customer places an order, validate the payment, update inventory, and send confirmation email"

"When a file is uploaded, validate it, process the content, and store the results in the database"

"Every morning at 9 AM, check system health and send status report to the team"
```

### AI-Generated Workflow Example
Input: "When a user signs up, send welcome email and create profile"

Output:
```yaml
name: user-signup-workflow
on:
  event: user.created

workflow:
  - name: send_welcome_email
    action: email.send
    to: "{{ .event.payload.email }}"
    subject: "Welcome {{ .event.payload.name }}!"
    template: welcome-email
    
  - name: create_user_profile
    action: database.insert
    table: user_profiles
    data:
      user_id: "{{ .event.payload.id }}"
      name: "{{ .event.payload.name }}"
      email: "{{ .event.payload.email }}"
      
  - name: log_registration
    action: log.info
    message: "New user registered: {{ .event.payload.name }}"
```

## Marketplace Templates

### Browsing Templates
1. Visit http://localhost:5000/marketplace
2. Browse available templates by category
3. Filter by complexity, rating, or tags
4. Download templates that match your needs

### Available Template Categories
- **Data Processing**: Web scrapers, ETL pipelines
- **E-commerce**: Order processing, payment workflows
- **User Management**: Onboarding, profile management
- **File Management**: Upload processing, validation
- **Communication**: Notification systems, email workflows
- **Monitoring**: Health checks, alert systems

### Using Marketplace Templates
1. Find a template you like
2. Click "Download"
3. Customize the template for your needs
4. Save to your `workflows/` directory
5. Validate and test the workflow

## Best Practices

### Workflow Design
1. **Keep workflows focused**: One workflow per business process
2. **Use descriptive names**: Clear, meaningful workflow and step names
3. **Handle errors gracefully**: Always include error handling
4. **Log important events**: Track workflow execution for debugging
5. **Test thoroughly**: Validate workflows before production use

### Event Design
1. **Use consistent naming**: Follow a clear event naming convention
2. **Include necessary data**: Provide all required information in event payload
3. **Version your events**: Include version information for compatibility
4. **Validate input**: Always validate incoming event data

### Performance Considerations
1. **Use appropriate timeouts**: Set reasonable timeouts for external calls
2. **Implement retry logic**: Handle transient failures gracefully
3. **Monitor execution times**: Track workflow performance
4. **Use parallel execution**: When possible, run independent steps in parallel

## Troubleshooting

### Common Issues

#### Workflow Not Triggering
```bash
# Check if workflow is loaded
./conduktr validate workflows/my-workflow.yaml

# Check server logs
tail -f logs/reactor.log

# Verify event is being sent
curl -X POST http://localhost:5000/webhook/my-event \
  -H "Content-Type: application/json" \
  -d '{"test": "data"}'
```

#### Workflow Execution Failing
```bash
# Check workflow execution logs
./conduktr execute workflows/my-workflow.yaml '{"test":"data"}' --verbose

# Validate workflow syntax
./conduktr validate workflows/my-workflow.yaml
```

#### Template Variables Not Working
- Ensure event payload contains expected data
- Check template syntax: `{{ .event.payload.field_name }}`
- Verify variable names match exactly

### Debugging Tips
1. **Use log.info actions**: Add logging steps to track execution
2. **Check server logs**: Monitor application logs for errors
3. **Test with sample data**: Use the execute command with test data
4. **Validate workflows**: Always validate before deployment

### Getting Help
- Check the application logs for detailed error messages
- Use the health endpoint: `curl http://localhost:5000/health`
- Validate workflows before running: `./conduktr validate workflow.yaml`
- Test with sample data: `./conduktr execute workflow.yaml '{"test":"data"}'`

## Next Steps

1. **Explore the AI Builder**: Try creating workflows using natural language
2. **Browse the Marketplace**: Download and customize existing templates
3. **Create Custom Actions**: Extend the system with your own actions
4. **Set up Monitoring**: Configure alerts and monitoring for your workflows
5. **Scale Your Workflows**: Add more complex logic and integrations

## API Reference

### HTTP Endpoints
- `POST /webhook/{event}` - Trigger workflow by event type
- `POST /events` - Send event with payload
- `GET /workflows` - List registered workflows
- `GET /instances/{id}` - Get workflow execution details
- `GET /health` - Health check

### CLI Commands
- `./conduktr run` - Start the daemon
- `./conduktr validate <file>` - Validate workflow file
- `./conduktr execute <file> [data]` - Execute workflow with data

---

This tutorial covers the fundamentals of creating and managing workflows in Reactor. For more advanced features and integrations, refer to the main documentation and explore the AI Builder and Marketplace for inspiration. 