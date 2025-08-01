# Reactor Quick Reference

## Quick Start
```bash
# Build and run
make build
./conduktr run

# Access dashboards
open http://localhost:5000/          # Main dashboard
open http://localhost:5000/ai/builder # AI Builder
open http://localhost:5000/marketplace # Marketplace
```

## CLI Commands
```bash
# Start daemon
./conduktr run

# Validate workflow
./conduktr validate workflows/my-workflow.yaml

# Execute workflow with data
./conduktr execute workflows/my-workflow.yaml '{"name":"John","email":"john@example.com"}'

# Get help
./conduktr --help
```

## HTTP Triggers
```bash
# Trigger workflow by event type
curl -X POST http://localhost:5000/webhook/user-created \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com"}'

# Send generic event
curl -X POST http://localhost:5000/events \
  -H "Content-Type: application/json" \
  -d '{"event":"user.created","data":{"name":"John","email":"john@example.com"}}'
```

## Workflow Examples

### Basic User Registration
```yaml
name: user-registration
on:
  event: user.created

workflow:
  - name: send_welcome
    action: email.send
    to: "{{ .event.payload.email }}"
    subject: "Welcome {{ .event.payload.name }}!"

  - name: create_profile
    action: database.insert
    table: users
    data:
      name: "{{ .event.payload.name }}"
      email: "{{ .event.payload.email }}"
```

### File Processing
```yaml
name: file-processor
on:
  event: file.created

workflow:
  - name: validate_file
    action: log.info
    message: "Processing file: {{ .event.payload.file_name }}"

  - name: backup_file
    action: shell.exec
    command: "cp {{ .event.payload.file_path }} {{ .event.payload.file_path }}.backup"
```

### Scheduled Task
```yaml
name: daily-report
on:
  event: scheduled
  cron: "0 9 * * *"  # Daily at 9 AM

workflow:
  - name: generate_report
    action: http.request
    url: "https://api.company.com/reports/daily"
    method: GET

  - name: send_report
    action: email.send
    to: "team@company.com"
    subject: "Daily Report"
    body: "{{ .response.body }}"
```

## Event Data Structure
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
      "source": "webhook"
    }
  },
  "variables": {
    "workflow_id": "wf_123",
    "execution_id": "exec_456"
  }
}
```

## Template Variables
```yaml
# Access event data
{{ .event.payload.name }}
{{ .event.payload.email }}
{{ .event.timestamp }}

# Access workflow variables
{{ .variables.workflow_id }}

# Conditional logic
{{ if eq .event.payload.type "premium" }}
{{ if gt .event.payload.amount 100 }}
```

## Common Actions

### HTTP Request
```yaml
- name: call_api
  action: http.request
  url: "https://api.example.com/endpoint"
  method: POST
  headers:
    Authorization: "Bearer {{ .token }}"
  body:
    data: "{{ .event.payload }}"
```

### Database Operations
```yaml
- name: insert_user
  action: database.insert
  table: users
  data:
    name: "{{ .event.payload.name }}"
    email: "{{ .event.payload.email }}"

- name: update_user
  action: database.update
  table: users
  where: "id = {{ .event.payload.id }}"
  data:
    status: "active"
```

### Email Sending
```yaml
- name: send_notification
  action: email.send
  to: "{{ .event.payload.email }}"
  subject: "Welcome!"
  body: "Hi {{ .event.payload.name }}, welcome to our platform!"
```

### Logging
```yaml
- name: log_event
  action: log.info
  message: "Processing user: {{ .event.payload.name }}"
  level: info
```

### Shell Commands
```yaml
- name: backup_file
  action: shell.exec
  command: "cp {{ .event.payload.file_path }} /backups/"
  timeout: 30
```

## Advanced Features

### Conditional Execution
```yaml
- name: premium_welcome
  action: email.send
  to: "{{ .event.payload.email }}"
  subject: "Premium Welcome!"
  if: "{{ eq .event.payload.type \"premium\" }}"
```

### Retry Logic
```yaml
- name: api_call
  action: http.request
  url: "https://api.example.com/endpoint"
  retry:
    max: 3
    backoff: exponential
    delay: 5s
```

### Parallel Execution
```yaml
- name: parallel_tasks
  parallel: true
  steps:
    - name: send_email
      action: email.send
      to: "{{ .event.payload.email }}"
    - name: update_db
      action: database.update
      table: users
      where: "id = {{ .event.payload.id }}"
```

## AI Builder Prompts
```
"When a customer places an order, validate payment and send confirmation"

"When a file is uploaded, validate it and process the content"

"Every morning at 9 AM, check system health and send status report"

"When a user signs up, send welcome email and create their profile"
```

## Marketplace Categories
- **Data Processing**: Web scrapers, ETL pipelines
- **E-commerce**: Order processing, payment workflows  
- **User Management**: Onboarding, profile management
- **File Management**: Upload processing, validation
- **Communication**: Notification systems, email workflows
- **Monitoring**: Health checks, alert systems

## Troubleshooting

### Check if server is running
```bash
curl http://localhost:5000/health
```

### Validate workflow syntax
```bash
./conduktr validate workflows/my-workflow.yaml
```

### Test workflow execution
```bash
./conduktr execute workflows/my-workflow.yaml '{"test":"data"}'
```

### Check logs
```bash
tail -f logs/reactor.log
```

## API Endpoints
- `GET /` - Main dashboard
- `GET /ai/builder` - AI Workflow Builder
- `GET /marketplace` - Template marketplace
- `POST /webhook/{event}` - Trigger workflow
- `POST /events` - Send event
- `GET /workflows` - List workflows
- `GET /health` - Health check
- `GET /metrics` - System metrics
- `GET /logs` - Execution logs 