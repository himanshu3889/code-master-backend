# Code Master - Backend ⚡

A robust, secure Go-based execution engine designed to power a personal interview preparation workspace. It handles high-performance code compilation, runtime evaluation, and real-time data synchronization to facilitate smooth, local code iteration.

## 🎯 Project Aim
To provide a secure, extremely low-latency environment for compiling and testing DSA problem solutions locally. By leveraging Docker isolation, it acts as a dedicated execution backend ensuring accurate metrics (time/memory) during interview practice.

## 🌟 Key Features
- **Dockerized Code Execution**: Isolated container runtime to run untrusted code safely without polluting the host environment.
- **Real-time Updates**: WebSocket engine pushes live execution results and test case progress directly to the frontend IDE.
- **Multi-Language Support**: Custom pipeline ready for Python, Java, C++, Go, and Rust with auto-compilation.
- **Performance Benchmarking**: Measures execution time and peak memory usage for exact submission metrics.
- **Persistent Data Store**: Tracks code revisions, persistent notes, problem stats, and historical code attempts.

## 🔗 Frontend Repository
This backend is designed to integrate directly with the "Code Master" workspace frontend:
- **Frontend Repo**: [https://github.com/himanshu3889/code-master-frontend](https://github.com/himanshu3889/code-master-frontend)

## 🚀 Getting Started

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/himanshu3889/code-master-backend.git
   cd code-master-backend
   ```

2. **Environment Setup**:
   Ensure Docker is installed locally as the engine spins up lightweight temporary containers to execute code safely. 

3. **Start Supporting Services**:
   Use Docker Compose to run the necessary database stack:
   ```bash
   make up-db
   ```

4. **Run the Application**:
   ```bash
   go run main.go
   ```

*The server defaults to standard API/WS listening. Make sure your frontend environment is configured to point to this backend local listener.*
