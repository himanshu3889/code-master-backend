# Code Master Backend

This is the server for Code Master, a local tool for practicing coding interview problems. It stores your problems, runs your code in isolated Docker containers, measures performance, and keeps a history of every submission and code change.

It is not a replacement for LeetCode or HackerRank. It is a companion workspace. You browse problems on those platforms, import them automatically with the Competitive Companion browser extension, and then solve them here where you have your own editor, notes, and performance tracking.

---

## Core Features

### Problem Management

- **Automatic import**: Receives problem data from the Competitive Companion browser extension. Title, description, test cases, time limits, and memory limits are captured without manual copy-pasting.
- **Manual editing**: After import, you can edit descriptions, notes, tags, difficulty level, and status through the REST API.
- **Search and filter**: Full-text fuzzy search on problem names using PostgreSQL `pg_trgm`. Filter by status (Todo, In Progress, Solved, etc.).
- **Tagging**: Each problem has a JSONB tags array. Tags are auto-detected from the problem name and description on import, and can be edited later.

### Code Execution

- **Multi-language support**: Python, Go, Java, C, C++, Rust, JavaScript, and TypeScript.
- **Docker isolation**: Every execution happens inside a dedicated Docker container with no network access, read-only root filesystem, and strict memory and CPU limits.
- **Container pooling**: The system maintains a pool of long-running containers per language to avoid the overhead of starting a new container for every submission. Containers are recycled after 50 uses or on error.
- **Compilation strategy**: Compiled languages (Go, C, C++, Java, Rust) are built first inside the container. Interpreted languages (Python, Node) run directly. Compilation errors are returned to the user immediately.
- **Resource measurement**: Uses the `timeout` command and `/usr/bin/time -v` to measure wall-clock execution time and peak resident memory. The memory value is parsed from stderr.
- **Exit code mapping**:
  - `124` or `137`: Time Limit Exceeded
  - `139`: Runtime Error (segmentation fault)
  - `Killed` in stderr: Memory Limit Exceeded
  - Non-zero otherwise: Runtime Error
  - Zero: Accepted
- **Test case streaming**: When running against multiple test cases, results are streamed back to the frontend in real time over WebSockets so the user sees each case pass or fail as it happens.

### Interview Sessions

- **Timed practice**: Start a timed session for any problem (default 45 minutes). The server tracks the start time and time limit.
- **Auto-timeout**: A background worker checks every 30 seconds for sessions that have exceeded their time limit. When found, the session is marked as timed out, the current code is auto-submitted if possible, and the editor is locked.
- **Session states**: In Progress, Completed, Timed Out, Abandoned.
- **Scoped history**: All submissions, snapshots, and timeline entries made during a session are linked to that session. You can review your performance under time pressure separately from casual practice.

### Tracking and History

- **Submissions**: Every code submission is stored with its language, source input, stdout, stderr, execution time, memory usage, and per-test-case results.
- **Code snapshots**: Automatic or manual saves of your code at a point in time. Snapshots are linked to problems and optionally to interview sessions.
- **Timeline**: A chronological feed of all activity on a problem: submissions, snapshots, and status changes. The "story" endpoint returns a detailed timeline with full objects resolved.
- **Status workflow**: Problems move through statuses: Todo, In Progress, Solved, Review Needed, Stuck, Skipped, Redo List, and Archived.

---

## Architecture

### Server

The HTTP server is built with Go and the Gin framework. It listens on port `27122` by default. CORS is enabled for all origins to support the browser extension and local frontend development. Graceful shutdown is handled with a 5-second context timeout on SIGINT or SIGTERM.

### Database

PostgreSQL is used for all persistent storage. `sqlx` is used for structured queries and struct scanning. `golang-migrate` manages schema migrations in `internal/migrations`.

Key design decisions:
- **Snowflake IDs**: All primary keys are 64-bit Snowflake IDs generated locally. They are roughly time-sortable, which avoids the need for a separate `created_at` index for most queries.
- **JSONB fields**: Test cases, saved code per language, tags, and submission results are stored as JSONB. This keeps the schema flexible for problem data that varies in shape.
- **Enums**: PostgreSQL custom enum types are used for `problem_status`, `problem_difficulty_level`, `submission_status`, and `interview_session_status`. This enforces valid values at the database level.
- **Indexes**: Covering indexes on `problem_id + created_at` exist for all history tables (submissions, snapshots, timeline, interview sessions) to make per-problem history lookups fast.

### Code Runner

The execution engine lives in `codeRunner/executer.go`. It is not a microservice; it runs inside the main server process and spawns Docker child processes.

Pipeline for a single submission:
1. The user's code is written to a temp workspace directory on the host.
2. A Docker container is started (or reused from the pool) with the workspace mounted at `/workspace`.
3. If the language is compiled, a compilation step runs inside the container.
4. The compiled binary or interpreter command is executed with `timeout` and `/usr/bin/time -v`.
5. Stdout and stderr are captured. Memory is parsed from the time output written to stderr.
6. Exit codes are mapped to status strings (AC, WA, TLE, MLE, RE, CE).
7. Results are returned. If test cases were provided, each case result is sent individually.

### Job Queue

Code execution runs through a buffered Go channel with a fixed worker pool. The default size is 2-4 workers. This prevents the server from being overwhelmed by submission bursts and keeps Docker container count bounded. Jobs are structs containing the code, language, stdin, test cases, and a result channel.

### WebSockets

A WebSocket endpoint at `/ws` streams live test case progress. When a submission with multiple test cases is executed, each case result is pushed to the frontend as soon as it finishes, rather than waiting for the full batch to complete.

### Background Workers

Two workers run on independent tickers:

1. **Submission Retry Worker**: Runs every 5 minutes. Finds submissions still in `PENDING` status that are older than 5 minutes and re-executes them. This handles cases where the Docker runner crashed or the server restarted mid-execution.
2. **Session Timeout Worker**: Runs every 30 seconds. Finds interview sessions whose `started_at + time_limit_seconds` has passed. Marks them as timed out and triggers auto-submission.

---

## API Overview

All API routes are prefixed with `/api` except the root `POST /` which is reserved for the Competitive Companion browser extension.

### Problems
- `GET /api/problems` - List problems with pagination and status filter
- `GET /api/problems/:id` - Single problem with full details
- `GET /api/problems/latest` - Most recently imported problem
- `GET /api/problems/after/:afterId` - Cursor pagination forward
- `GET /api/problems/before/:beforeId` - Cursor pagination backward
- `GET /api/problems/search?q=...` - Fuzzy name search
- `POST /` - Receive problem from Competitive Companion
- `PATCH /api/problems/:id/status` - Update status
- `PATCH /api/problems/:id/difficulty` - Update difficulty
- `PATCH /api/problems/:id/tags` - Update tags
- `PATCH /api/problems/:id/notes` - Update notes
- `PATCH /api/problems/:id/description` - Update description

### Submissions
- `POST /api/problems/:id/submissions` - Submit code for execution
- `GET /api/problems/:id/submissions` - Get submissions for a problem
- `GET /api/problems/:id/with-submissions` - Problem + submissions together

### Snapshots
- `GET /api/snapshots/:id` - Get a single snapshot
- `GET /api/problems/:id/snapshots` - List snapshots for a problem

### Timeline
- `GET /api/problems/:id/timeline` - Activity timeline
- `GET /api/problems/:id/story` - Detailed timeline with resolved objects

### Interview Sessions
- `POST /api/interview-sessions` - Create a session
- `GET /api/interview-sessions/:id` - Get session with remaining time
- `GET /api/problems/:id/active-session` - Active session for problem
- `GET /api/problems/:id/interview-sessions` - All sessions for problem
- `PATCH /api/interview-sessions/:id/complete` - Mark completed
- `PATCH /api/interview-sessions/:id/abandon` - Mark abandoned
- `POST /api/interview-sessions/:id/timeout` - Trigger timeout manually

### Languages
- `GET /api/languages` - List supported languages
- `POST /api/languages` - Add a language
- `PATCH /api/languages/:code/template` - Update starter template

### Tags
- `GET /api/problems/tags` - List all known pattern tags

### WebSocket
- `GET /ws` - Connect for live test case streaming

---

## Project Layout

```
code-master-backend/
├── cmd/server/              # Server bootstrap, DI, and graceful shutdown
├── internal/
│   ├── apiHandler/          # HTTP handlers (Gin route functions)
│   ├── store/               # Database queries and transactions
│   ├── models/              # Structs mapping to DB tables
│   ├── migrations/          # Up/down SQL migration files
│   ├── database/            # Postgres connection setup
│   ├── jobs/                # Background workers (retry, timeout)
│   ├── lib/                 # Internal helpers (tagging, execution)
│   └── websocket/           # WebSocket connection manager
├── codeRunner/              # Docker execution engine
│   ├── executer.go          # Job queue, container pool, compilation, running
│   └── config.go            # Environment and Docker command mapping
├── base/                    # Shared utilities
│   ├── lib/                 # SQL JSONB helpers, pagination, app errors
│   └── utils/               # Snowflake generation, ID validation
├── configs/                 # Viper config initialization
├── main.go                  # Entry point: starts queue, server, workers
├── Makefile                 # Common tasks (db up, migrate, run)
├── docker-compose.yml       # Postgres and Redis services
└── Dockerfile               # Build image for deployment
```

---

## Running Locally

Requirements: Go 1.25+, Docker, Make.

1. Clone the repository and enter the directory.
2. Copy `.env.example` to `.env` and set values if needed.
3. Start Postgres and Redis:
   ```bash
   make up-db
   ```
4. Run migrations:
   ```bash
   make migrate-up
   ```
5. Start the server:
   ```bash
   go run main.go
   ```

The server starts on `:27122`. The frontend expects it at `http://localhost:27122`.

Docker must be running. The first submission may take a few seconds while the runner image is built.
