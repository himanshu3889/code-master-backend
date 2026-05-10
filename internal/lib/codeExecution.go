package lib

import (
	"fmt"

	"github.com/himanshu3889/code-master-backend/codeRunner"
	"github.com/himanshu3889/code-master-backend/internal/models"
	"github.com/himanshu3889/code-master-backend/internal/store"
	"github.com/sirupsen/logrus"
)

// Execute the code submission
func ExecuteSubmissionCode(dbStore *store.Store, submission *models.Submission) *codeRunner.CodeExecutionResult {
	// Create a result channel
	resultChan := make(chan codeRunner.CodeExecutionJobResult, 1)

	testCases := make([]codeRunner.TestCase, 0, len(submission.TestCases.Data))
	for _, testCase := range submission.TestCases.Data {
		testCases = append(testCases, codeRunner.TestCase{Input: testCase})
	}

	var langExt string
	switch submission.Language {
	case "python":
		langExt = ".py"
	case "go", "golang": // Handles both "go" and "golang"
		langExt = ".go"
	default:
		logrus.Errorf("Unsupported language: %s", submission.Language)
		return nil
	}

	codeExecutionJob := codeRunner.CodeExecutionJob{
		ID:         submission.ID.String(),
		LangExt:    langExt,
		Code:       submission.Stdin,
		TestCases:  testCases,
		ResultChan: resultChan,
	}

	codeRunner.SubmitCodeExecutionJob(codeExecutionJob)

	jobResult := <-resultChan

	close(resultChan)

	if jobResult.Error != "" {
		fmt.Printf("Error: %s\n", jobResult.Error)
		return nil
	}

	result := jobResult.Result

	fmt.Printf("Job ID: %s\n", jobResult.ID)
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Time: %d ms\n", result.TimeMs)
	fmt.Printf("Memory: %d bytes\n", result.MemoryBytes)
	fmt.Printf("Output: %s", result.Stdout)

	for _, tcResult := range result.TestCasesResult {
		fmt.Printf("\nTest %d:\n", tcResult.TestIndex)
		fmt.Printf("%s\n", tcResult.ActualOutput)
	}

	if result.Stderr != "" {
		fmt.Printf("Stderr: %s\n", result.Stderr)
	}

	return result

}
