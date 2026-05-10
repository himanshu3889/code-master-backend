package codeRunner

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type SubmissionStatus string

const (
	StatusAccepted            SubmissionStatus = "AC"
	StatusCompilationError    SubmissionStatus = "CE"
	StatusTimeLimitExceeded   SubmissionStatus = "TLE"
	StatusMemoryLimitExceeded SubmissionStatus = "MLE"
	StatusWrongAnswer         SubmissionStatus = "WA"
	StatusRuntimeError        SubmissionStatus = "RE"
)

type CodeExecutionJob struct {
	ID         string
	LangExt    string
	Code       string
	Filename   string
	Stdin      string
	TestCases  []TestCase
	ResultChan chan<- CodeExecutionJobResult
}

type CodeExecutionJobResult struct {
	ID     string
	Result *CodeExecutionResult
	Error  string
}

type CodeExecutionResult struct {
	Status          SubmissionStatus
	TimeMs          int64
	MemoryBytes     int64
	ExitCode        int
	Stdout          string
	Stderr          string
	TestCasesResult []TestCaseResult
}

type TestCaseResult struct {
	TestIndex    int              `json:"testIndex"`
	Passed       bool             `json:"passed"`
	Status       SubmissionStatus `json:"status"`
	TimeMs       int64            `json:"timeMs"`
	MemoryBytes  int64            `json:"memoryBytes"`
	ActualOutput string           `json:"actualOutput"`
	Stderr       string           `json:"stderr,omitempty"`
}

type TestCase struct {
	Input    string
	Expected string
}

type CodeExecutionJobQueue struct {
	jobs   chan CodeExecutionJob
	closed bool
	mu     sync.Mutex
	wg     sync.WaitGroup
}

var jobQueueOnce sync.Once
var codeExecutionJobQueue *CodeExecutionJobQueue

func StartCodeExecutionJobQueue(workerCount int) {
	jobQueueOnce.Do(func() {
		codeExecutionJobQueue = &CodeExecutionJobQueue{
			jobs: make(chan CodeExecutionJob, 100),
		}
		for i := 0; i < workerCount; i++ {
			codeExecutionJobQueue.wg.Add(1)
			go worker(codeExecutionJobQueue.jobs, &codeExecutionJobQueue.wg)
		}
	})
}

func StopCodeExecutionJobQueue() {
	if codeExecutionJobQueue == nil {
		return
	}
	codeExecutionJobQueue.mu.Lock()
	if !codeExecutionJobQueue.closed {
		close(codeExecutionJobQueue.jobs)
		codeExecutionJobQueue.closed = true
	}
	codeExecutionJobQueue.mu.Unlock()
	codeExecutionJobQueue.wg.Wait()
}

func dedent(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) == 0 {
		return s
	}

	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}
	if minIndent <= 0 {
		return s
	}

	for i, line := range lines {
		if len(line) >= minIndent {
			lines[i] = line[minIndent:]
		}
	}
	return strings.Join(lines, "\n")
}

func SubmitCodeExecutionJob(job CodeExecutionJob) {
	if codeExecutionJobQueue == nil {
		StartCodeExecutionJobQueue(4)
	}
	job.Code = dedent(job.Code)
	codeExecutionJobQueue.jobs <- job
}

func worker(jobs <-chan CodeExecutionJob, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		result := executeJob(job)
		if job.ResultChan != nil {
			job.ResultChan <- result
		}
	}
}

func executeJob(job CodeExecutionJob) CodeExecutionJobResult {
	filename := job.Filename
	if filename == "" {
		filename = defaultFilename(job.LangExt)
	}
	dir := getWorkspaceDir(job.LangExt)

	var result *CodeExecutionResult
	if len(job.TestCases) > 0 {
		result = runJob(job.LangExt, job.Code, filename, "", job.TestCases, nil, dir)
	} else {
		result = runJob(job.LangExt, job.Code, filename, job.Stdin, nil, nil, dir)
	}

	os.Remove(filepath.Join(dir, filename))
	os.Remove(filepath.Join(dir, "run_bin"))
	os.Remove(filepath.Join(dir, "Main.class"))
	if files, err := filepath.Glob(filepath.Join(dir, "Main$*.class")); err == nil {
		for _, f := range files {
			os.Remove(f)
		}
	}

	return CodeExecutionJobResult{ID: job.ID, Result: result}
}

func getWorkspaceDir(lang string) string {
	dir := filepath.Join(os.TempDir(), "coderunner", strings.TrimPrefix(lang, "."))
	os.MkdirAll(dir, 0755)

	os.Remove(filepath.Join(dir, "main.go"))
	os.Remove(filepath.Join(dir, "run_bin"))
	os.Remove(filepath.Join(dir, "Main.java"))
	os.Remove(filepath.Join(dir, "Main.class"))
	return dir
}

func defaultFilename(lang string) string {
	switch lang {
	case ".py":
		return "main.py"
	case ".go":
		return "main.go"
	case ".js":
		return "main.js"
	case ".c":
		return "main.c"
	case ".cpp", ".cc", ".cxx":
		return "main.cpp"
	case ".java":
		return "Main.java"
	case ".rs":
		return "main.rs"
	default:
		return "main" + lang
	}
}

var (
	poolMu  sync.Mutex
	runners = make(map[string]*runner)
)

type runner struct {
	name     string
	mu       sync.Mutex
	useCount int
}

func containerName(lang, dir string) string {
	h := fnv.New32a()
	h.Write([]byte(dir))
	return fmt.Sprintf("runner_%s_%08x", strings.TrimPrefix(lang, "."), h.Sum32())
}

func (r *runner) isAlive() bool {
	return isContainerRunning(r.name)
}

func getOrCreateRunner(lang, dir string) (*runner, error) {
	poolMu.Lock()
	defer poolMu.Unlock()

	key := lang + ":" + dir
	name := containerName(lang, dir)

	if r, ok := runners[key]; ok && r.isAlive() {
		return r, nil
	}

	if isContainerRunning(name) {
		r := &runner{name: name}
		runners[key] = r
		return r, nil
	}

	exec.Command("docker", "rm", "-f", name).Run()

	image := fmt.Sprintf("%s:%s", GetSetting("DOCKER_IMAGE_NAME"), GetSetting("DOCKER_IMAGE_VERSION"))
	memLimit := GetSetting("DEFAULT_MEMORY_LIMIT")

	args := []string{
		"run", "-d",
		"--name", name,
		"--init",
		"-v", fmt.Sprintf("%s:/workspace", dir),
		"--memory", memLimit,
		"--memory-swap", memLimit,
		"--cpus", "1.0",
		"--network", "none",
		"--read-only",
		"--tmpfs", "/tmp:noexec,nosuid,size=500m",
		image,
		"sleep", "infinity",
	}

	if out, err := exec.Command("docker", args...).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("start container failed: %v, output: %s", err, out)
	}

	r := &runner{name: name}
	runners[key] = r
	return r, nil
}

func removeRunner(name string) {
	poolMu.Lock()
	defer poolMu.Unlock()

	exec.Command("docker", "stop", "-t", "1", name).Run()
	exec.Command("docker", "rm", "-f", name).Run()

	for k, r := range runners {
		if r.name == name {
			delete(runners, k)
			return
		}
	}
}

func isContainerRunning(name string) bool {
	out, err := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", name).Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

func cleanContainerWorkspace(containerName string) {
	exec.Command("docker", "exec", containerName, "sh", "-c",
		"find /workspace -maxdepth 1 -type f -delete 2>/dev/null; true").Run()
	exec.Command("docker", "exec", containerName, "sh", "-c",
		"rm -rf /tmp/* /tmp/.??* 2>/dev/null; true").Run()
}

func compileCode(langExt, filename, dir string) (string, error) {
	image := fmt.Sprintf("%s:%s", GetSetting("DOCKER_IMAGE_NAME"), GetSetting("DOCKER_IMAGE_VERSION"))
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	switch langExt {
	case ".go":
		cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
			"-v", fmt.Sprintf("%s:/workspace", dir),
			"-v", "go-build-cache:/root/.cache/go-build",
			"--workdir", "/workspace",
			"--memory", "1g",
			"--memory-swap", "1g",
			"--network", "none",
			image,
			"sh", "-c", fmt.Sprintf("CGO_ENABLED=0 go build -o run_bin '%s'", filename))
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("go build failed: %v, output: %s", err, out)
		}
		return "./run_bin", nil
	case ".c":
		cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
			"-v", fmt.Sprintf("%s:/workspace", dir),
			"--workdir", "/workspace",
			"--memory", "1g",
			"--memory-swap", "1g",
			"--network", "none",
			image,
			"sh", "-c", fmt.Sprintf("gcc -o run_bin '%s'", filename))
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("gcc failed: %v, output: %s", err, out)
		}
		return "./run_bin", nil
	case ".cpp", ".cc", ".cxx":
		cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
			"-v", fmt.Sprintf("%s:/workspace", dir),
			"--workdir", "/workspace",
			"--memory", "1g",
			"--memory-swap", "1g",
			"--network", "none",
			image,
			"sh", "-c", fmt.Sprintf("g++ -o run_bin '%s'", filename))
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("g++ failed: %v, output: %s", err, out)
		}
		return "./run_bin", nil
	case ".java":
		cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
			"-v", fmt.Sprintf("%s:/workspace", dir),
			"--workdir", "/workspace",
			"--memory", "1g",
			"--memory-swap", "1g",
			"--network", "none",
			image,
			"sh", "-c", fmt.Sprintf("javac '%s'", filename))
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("javac failed: %v, output: %s", err, out)
		}
		className := strings.TrimSuffix(filename, ".java")
		return fmt.Sprintf("java -cp . %s", className), nil
	case ".rs":
		cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
			"-v", fmt.Sprintf("%s:/workspace", dir),
			"--workdir", "/workspace",
			"--memory", "1g",
			"--memory-swap", "1g",
			"--network", "none",
			image,
			"sh", "-c", fmt.Sprintf("rustc -o run_bin '%s'", filename))
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("rustc failed: %v, output: %s", err, out)
		}
		return "./run_bin", nil
	default:
		runCmd := GetDockerCommand(langExt, filename)
		if runCmd == "" {
			return "", fmt.Errorf("unsupported language: %s", langExt)
		}
		return runCmd, nil
	}
}

func runSingleTest(r *runner, command, stdin, timeLimit string) TestCaseResult {
	result := TestCaseResult{Status: StatusAccepted, Passed: true}

	if command == "" {
		result.Status = StatusCompilationError
		result.Passed = false
		return result
	}

	// FIXED: Use /proc/self/fd/2 to write cat output to stderr without
	// accidentally redirecting it into /dev/null due to shell order.
	wrapperCmd := fmt.Sprintf(`
cd /workspace
rm -f .time_log .stderr_log
/usr/bin/time -v -o .time_log timeout -s KILL %s %s 2>.stderr_log
EXIT=$?
cat .stderr_log > /proc/self/fd/2 2>/dev/null
awk '/Maximum resident set size/{print "__MEM_KB__:" $6}' .time_log >&2
exit $EXIT
`, timeLimit, command)

	timeLimitSec, _ := strconv.Atoi(strings.TrimSuffix(timeLimit, "s"))
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeLimitSec+3)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "exec", "-i", r.name, "sh", "-c", wrapperCmd)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	result.TimeMs = elapsed.Milliseconds()
	result.ActualOutput = strings.TrimSpace(stdout.String())

	stderrStr := stderr.String()

	const memPrefix = "__MEM_KB__:"
	if idx := strings.LastIndex(stderrStr, memPrefix); idx != -1 {
		line := stderrStr[idx:]
		if nl := strings.Index(line, "\n"); nl != -1 {
			line = line[:nl]
		}
		if valStr := strings.TrimPrefix(line, memPrefix); valStr != "" {
			if val, parseErr := strconv.ParseInt(strings.TrimSpace(valStr), 10, 64); parseErr == nil {
				result.MemoryBytes = val * 1024
			}
		}
		// Safely strip the marker line
		endIdx := idx + len(line)
		if endIdx < len(stderrStr) && stderrStr[endIdx] == '\n' {
			endIdx++
		}
		stderrStr = stderrStr[:idx] + stderrStr[endIdx:]
	}

	result.Stderr = strings.TrimSpace(stderrStr)

	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode := exitErr.ExitCode()
		switch exitCode {
		case 124, 137:
			result.Status = StatusTimeLimitExceeded
		case 139:
			result.Status = StatusRuntimeError
		default:
			// FIXED: During execution, non-zero exit means runtime error.
			// Compilation errors are handled in compileCode, not here.
			if strings.Contains(result.Stderr, "Killed") || strings.Contains(result.Stderr, "killed") {
				result.Status = StatusMemoryLimitExceeded
			} else {
				result.Status = StatusRuntimeError
			}
		}
		result.Passed = false
	} else if ctx.Err() == context.DeadlineExceeded {
		result.Status = StatusTimeLimitExceeded
		result.Passed = false
	} else if err != nil {
		result.Status = StatusRuntimeError
		result.Passed = false
	}

	return result
}

func runJob(langExt, code, filename, stdin string, testCases []TestCase, onResult func(TestCaseResult), dir string) *CodeExecutionResult {
	if dir == "" {
		dir = getWorkspaceDir(langExt)
	}

	if err := buildDockerImage(); err != nil {
		return &CodeExecutionResult{Status: StatusCompilationError, Stderr: err.Error()}
	}

	r, err := getOrCreateRunner(langExt, dir)
	if err != nil {
		return &CodeExecutionResult{Status: StatusCompilationError, Stderr: err.Error()}
	}

	r.mu.Lock()

	if r.useCount >= 50 {
		name := r.name
		r.mu.Unlock()
		removeRunner(name)
		r, err = getOrCreateRunner(langExt, dir)
		if err != nil {
			return &CodeExecutionResult{Status: StatusCompilationError, Stderr: err.Error()}
		}
		r.mu.Lock()
	}
	r.useCount++
	defer r.mu.Unlock()

	cleanContainerWorkspace(r.name)

	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		return &CodeExecutionResult{Status: StatusCompilationError, Stderr: fmt.Sprintf("write file failed: %v", err)}
	}

	timeLimit := GetSetting("DEFAULT_TIME_LIMIT")
	command, err := compileCode(langExt, filename, dir)
	if err != nil {
		return &CodeExecutionResult{Status: StatusCompilationError, Stderr: err.Error()}
	}

	if len(testCases) == 0 {
		tc := runSingleTest(r, command, stdin, timeLimit)
		return &CodeExecutionResult{
			Status:      tc.Status,
			TimeMs:      tc.TimeMs,
			MemoryBytes: tc.MemoryBytes,
			Stdout:      tc.ActualOutput,
			Stderr:      tc.Stderr,
		}
	}

	overall := &CodeExecutionResult{
		Status:          StatusAccepted,
		TestCasesResult: make([]TestCaseResult, 0, len(testCases)),
	}

	for i, tc := range testCases {
		tr := runSingleTest(r, command, tc.Input, timeLimit)
		tr.TestIndex = i

		// Compare output for test cases
		if tr.Status == StatusAccepted && strings.TrimSpace(tr.ActualOutput) != strings.TrimSpace(tc.Expected) {
			tr.Status = StatusWrongAnswer
			tr.Passed = false
		}

		overall.TestCasesResult = append(overall.TestCasesResult, tr)
		overall.TimeMs += tr.TimeMs
		if tr.MemoryBytes > overall.MemoryBytes {
			overall.MemoryBytes = tr.MemoryBytes
		}
		if !tr.Passed && overall.Status == StatusAccepted {
			overall.Status = tr.Status
		}

		if onResult != nil {
			onResult(tr)
		}
	}

	return overall
}

func ExecuteCode(fileExtension, filePath, stdin string) *CodeExecutionResult {
	code, err := os.ReadFile(filePath)
	if err != nil {
		return &CodeExecutionResult{Status: StatusCompilationError, Stderr: err.Error()}
	}
	return runJob(fileExtension, string(code), filepath.Base(filePath), stdin, nil, nil, filepath.Dir(filePath))
}

func ExecuteCodeWithTestCases(fileExtension, filePath string, testCases []TestCase, onResult func(TestCaseResult)) *CodeExecutionResult {
	code, err := os.ReadFile(filePath)
	if err != nil {
		return &CodeExecutionResult{Status: StatusCompilationError, Stderr: err.Error()}
	}
	return runJob(fileExtension, string(code), filepath.Base(filePath), "", testCases, onResult, filepath.Dir(filePath))
}

func ShutdownCodeRunners() {
	poolMu.Lock()
	defer poolMu.Unlock()
	logrus.Info("Shutting down code runners...")

	for _, r := range runners {
		exec.Command("docker", "stop", "-t", "1", r.name).Run()
		exec.Command("docker", "rm", "-f", r.name).Run()
	}
	runners = make(map[string]*runner)
}

func buildDockerImage() error {
	imageName := fmt.Sprintf("%s:%s", GetSetting("DOCKER_IMAGE_NAME"), GetSetting("DOCKER_IMAGE_VERSION"))

	checkCmd := exec.Command("docker", "images", "-q", imageName)
	out, _ := checkCmd.Output()
	if len(out) > 0 {
		return nil
	}

	cmd := exec.Command("docker", "build", "-t", imageName, ".")
	buildOut, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker build failed: %v, output: %s", err, buildOut)
	}
	return nil
}
