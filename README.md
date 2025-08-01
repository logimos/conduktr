# Reactor: Event-Driven Workflow Engine

## Overview

Reactor is a Go-native, event-driven workflow orchestration engine designed for high-performance asynchronous task execution. It combines the power of workflow engines like Temporal with the simplicity of automation platforms like Zapier, while maintaining the performance benefits of Go's concurrency model. The system is built to be lightweight, self-hostable, and deployable as a single binary.

## User Preferences

Preferred communication style: Simple, everyday language.

## System Architecture

### Core Design Philosophy
Reactor follows an event-driven architecture where workflows are triggered by various event sources and execute through a series of configurable steps. The system is designed around the principle: "When X happens, do Y — then Z — unless A — and only if B is true."

### Primary Architecture Components
- **Event Processing Engine**: Go-based asynchronous event handler with high concurrency support
- **Workflow Definition Layer**: YAML/DSL-based declarative flow definitions with embedded Go API support
- **State Management**: Stateful trigger system for managing workflow execution context
- **Event Router**: Structured routing system for directing events to appropriate workflows
- **Plugin System**: Extensible architecture supporting custom triggers via plugins or gRPC

## Key Components

### Event Trigger System
The system supports multiple event sources:
- **HTTP Endpoints**: REST API triggers for web-based events
- **File System Monitoring**: File change detection and processing
- **Scheduled Events**: CRON-based time triggers
- **Message Queue Integration**: Support for Kafka, NATS, Redis, and MQTT
- **Database Change Detection**: Polling and Change Data Capture (CDC) capabilities
- **Custom Triggers**: Plugin-based extensibility for specialized event sources

### Workflow Execution Engine
- **Declarative Configuration**: YAML-based workflow definitions for ease of use
- **Embedded API**: Go API for programmatic workflow creation and management
- **Asynchronous Processing**: Non-blocking workflow execution leveraging Go's goroutines
- **Conditional Logic**: Support for complex conditional branching and decision trees
- **State Persistence**: Workflow state management for long-running processes

### Performance Optimization
- **Single Binary Deployment**: Compiled workflows embedded directly into the executable
- **Go Concurrency**: Leverages Go's lightweight threading model for high throughput
- **Minimal Resource Footprint**: Designed for efficient resource utilization

## Data Flow

1. **Event Ingestion**: Events are received from various sources (HTTP, files, queues, etc.)
2. **Event Routing**: The structured router determines which workflows should be triggered
3. **Workflow Instantiation**: Matching workflows are created with appropriate context
4. **Step Execution**: Workflow steps execute asynchronously with state tracking
5. **Conditional Processing**: Decision points evaluate conditions to determine next steps
6. **Completion Handling**: Workflow results are processed and may trigger additional events

## External Dependencies

### Message Queue Systems
- **Kafka**: For high-throughput event streaming
- **NATS**: Lightweight messaging for microservices communication
- **Redis**: In-memory data structure store for caching and queuing
- **MQTT**: IoT device communication protocol support

### Database Integration
- **Change Data Capture**: Real-time database change monitoring
- **Polling Mechanisms**: Periodic database state checking
- **Multiple Database Support**: Flexible database connectivity options

### Communication Protocols
- **gRPC**: High-performance RPC for custom integrations
- **HTTP/REST**: Standard web API support
- **File System APIs**: OS-level file monitoring capabilities

## Deployment Strategy

### Single Binary Architecture
- **Embedded Workflows**: Custom flows compiled directly into the binary
- **Self-Contained**: No external runtime dependencies required
- **Cross-Platform**: Go's compilation capabilities enable multi-platform deployment

### Self-Hosting Capabilities
- **Minimal Infrastructure**: Can run on single servers or container environments
- **Scalability**: Horizontal scaling through multiple instances
- **Configuration Management**: Environment-based configuration for different deployment contexts

### Target Use Cases
- **DevOps Automation**: CI/CD pipeline orchestration and infrastructure management
- **Internal Business Automation**: Process automation and workflow management
- **IoT Coordination**: Device management and event processing
- **Real-time ETL**: Data pipeline orchestration and transformation
- **Reactive Systems**: Event-driven microservice coordination

The architecture prioritizes performance, simplicity, and operational efficiency while providing the flexibility needed for complex workflow orchestration scenarios.

## Recent Changes: Latest modifications with dates

### July 29, 2025 - Advanced Trigger Systems & Enhanced Dashboard
- **Added Redis Trigger System**: Pub/sub and stream-based event processing with real-time Redis integration
- **Added Kafka Trigger System**: High-throughput message processing with consumer groups and topic monitoring  
- **Added Database Trigger System**: Change data capture for PostgreSQL/MySQL with polling and WAL support
- **Added Scheduler Trigger System**: Cron-based workflow scheduling with second-precision timing
- **Enhanced Web Dashboard**: Advanced visualization with real-time metrics, flow diagrams, and execution traceability
- **Improved Architecture**: Modular trigger system with shared helper functions for workflow execution
- **Added Comprehensive Logging**: Detailed execution tracking with structured logging and error handling
- **Enhanced Error Handling**: Retry mechanisms with exponential backoff across all trigger systems

### July 29, 2025 - Enterprise-Grade Advanced Features Added
- **Real-Time Analytics Dashboard**: Live workflow metrics, performance monitoring, and alert system
  - Live execution metrics with success rates, latencies, and throughput monitoring
  - Real-time event stream visualization with Server-Sent Events
  - Intelligent alert system with configurable thresholds and notifications
  - System resource monitoring (CPU, memory, active workflows)
  - Beautiful dashboard with charts, graphs, and live data updates

- **AI-Powered Workflow Builder**: Natural language workflow creation with intelligent suggestions
  - Convert plain English descriptions into complete YAML workflows
  - Smart pattern matching with curated workflow templates library
  - Intelligent step suggestions and auto-completion capabilities
  - Workflow validation engine with error detection and fix suggestions
  - Pre-built patterns for common use cases (user onboarding, file processing, notifications)

- **Advanced Integration Hub**: Comprehensive third-party service connectivity
  - Pre-built connectors for Slack, GitHub, Stripe, and popular services
  - OAuth authentication flow management for secure integrations
  - Webhook management system with endpoint creation and testing
  - API client framework with rate limiting and error handling
  - Visual integration wizard with connection testing and validation

### July 29, 2025 - Visual Workflow Designer & Marketplace Added
- **Visual Workflow Designer**: Complete drag-and-drop interface for creating workflows visually
  - Node palette with pre-built components (triggers, actions, conditions, parallel execution)
  - Canvas with real-time visual connections and property editing
  - YAML generation from visual designs with validation
  - Save/load workflows with import/export functionality
- **Workflow Marketplace**: Template sharing platform with enterprise-grade catalog
  - Curated template library (web scrapers, e-commerce processors, API monitors, ETL pipelines)  
  - Search and filtering by category, complexity, and rating
  - Template downloads with usage statistics and reviews
  - Featured templates and verified publisher system
- **Advanced Orchestration Engine**: Enterprise workflow execution with sophisticated controls
  - Parallel execution with configurable join types (all, any, first, custom)
  - Sub-workflow support with input/output mapping and async execution
  - Advanced retry policies with exponential backoff and conditional logic
  - Loop and conditional execution with complex branching
  - Execution context management with state persistence and monitoring

### System Capabilities Expanded
- **Multi-Protocol Support**: HTTP, File System, Redis, Kafka, Database, and Scheduled triggers
- **Real-time Monitoring**: Live dashboard with execution flow visualization and performance metrics
- **Enterprise Features**: Database change detection, message queue integration, and distributed event processing
- **Scalability Features**: Asynchronous processing, concurrent workflow execution, and resource monitoring
- **Visual Development**: Drag-and-drop workflow designer with professional-grade interface
- **Template Ecosystem**: Marketplace with curated workflow templates and community sharing
- **Advanced Execution**: Parallel processing, sub-workflows, complex conditions, and enterprise orchestration