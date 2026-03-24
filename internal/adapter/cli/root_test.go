package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DementorAK/photometa/internal/domain"
	"github.com/DementorAK/photometa/internal/fake"

	"github.com/spf13/cobra"
)

// ============================================================================
// CLI COMMAND TESTS
// ============================================================================

func TestNewRootCmd_Creation(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()

	// Act
	cmd := NewRootCmd(mockAnalyzer, &fake.MockLogger{})

	// Assert
	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != "photometa [path]" {
		t.Errorf("expected Use 'photometa [path]', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
}

func TestNewRootCmd_HasFlags(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()
	cmd := NewRootCmd(mockAnalyzer, &fake.MockLogger{})

	// Act & Assert
	expectedFlags := []string{"path"}

	for _, flagName := range expectedFlags {
		t.Run(flagName, func(t *testing.T) {
			flag := cmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("expected flag --%s to exist", flagName)
			}
		})
	}
}

func TestNewRootCmd_Subcommands(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()
	cmd := NewRootCmd(mockAnalyzer, &fake.MockLogger{})

	// Act & Assert
	expectedCmds := []string{"server", "gui", "locales"}

	for _, cmdName := range expectedCmds {
		t.Run(cmdName, func(t *testing.T) {
			found := false
			for _, sub := range cmd.Commands() {
				if sub.Name() == cmdName {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected subcommand %q to exist", cmdName)
			}
		})
	}
}

func TestNewRootCmd_LocaleFlag(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()
	cmd := NewRootCmd(mockAnalyzer, &fake.MockLogger{})

	// Act
	localeFlag := cmd.PersistentFlags().Lookup("locale")

	// Assert
	if localeFlag == nil {
		t.Fatal("expected --locale persistent flag to exist")
	}

	if localeFlag.Shorthand != "l" {
		t.Errorf("expected shorthand 'l', got %q", localeFlag.Shorthand)
	}
}

func TestNewRootCmd_LocalesSubcommand(t *testing.T) {
	cmd := NewLocalesCmd()
	if cmd == nil {
		t.Fatal("expected command to be created")
	}
	if cmd.Use != "locales" {
		t.Errorf("expected Use 'locales', got %q", cmd.Use)
	}
}

func TestNewRootCmd_PathFlagShorthand(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()
	cmd := NewRootCmd(mockAnalyzer, &fake.MockLogger{})

	// Act
	pathFlag := cmd.Flags().Lookup("path")

	// Assert
	if pathFlag == nil {
		t.Fatal("expected --path flag to exist")
	}

	if pathFlag.Shorthand != "p" {
		t.Errorf("expected shorthand 'p', got %q", pathFlag.Shorthand)
	}
}

func TestNewRootCmd_PortDefaultValue(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()
	cmd := NewRootCmd(mockAnalyzer, &fake.MockLogger{})

	// Find server subcommand
	var serverCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "server" {
			serverCmd = sub
			break
		}
	}

	if serverCmd == nil {
		t.Fatal("expected server subcommand to exist")
	}

	// Act
	portFlag := serverCmd.Flags().Lookup("port")

	// Assert
	if portFlag == nil {
		t.Fatal("expected --port flag to exist on server command")
	}

	if portFlag.DefValue != "8080" {
		t.Errorf("expected default port '8080', got %q", portFlag.DefValue)
	}
}

func TestNewRootCmd_HelpOutput(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()
	cmd := NewRootCmd(mockAnalyzer, &fake.MockLogger{})

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	expectedParts := []string{
		"photometa",
		"server", // Subcommand
		"gui",    // Subcommand
		"--path", // Flag on root
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("expected help output to contain %q", part)
		}
	}
}

// ============================================================================
// runCLI FUNCTION TESTS
// ============================================================================

func TestRunCLI_WithPathFlag(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()
	mockAnalyzer.ScanDirectoryResult = []domain.ImageFile{
		{Name: "photo1.jpg", Path: "/test/photo1.jpg", Metadata: domain.Metadata{Format: "jpeg"}},
		{Name: "photo2.jpg", Path: "/test/photo2.jpg", Metadata: domain.Metadata{Format: "jpeg"}},
	}

	// Capture stdout
	oldStdout := captureStdout(t)
	defer oldStdout.Restore()

	// Act
	runCLI(&cobra.Command{}, mockAnalyzer, &fake.MockLogger{}, "/test/photos", nil, false)

	// Assert
	output := oldStdout.Output()

	// Verify ScanDirectory was called
	if len(mockAnalyzer.ScanDirectoryCalls) != 1 {
		t.Errorf("expected 1 ScanDirectory call, got %d", len(mockAnalyzer.ScanDirectoryCalls))
	}

	if mockAnalyzer.ScanDirectoryCalls[0] != "/test/photos" {
		t.Errorf("expected path '/test/photos', got %q", mockAnalyzer.ScanDirectoryCalls[0])
	}

	// Output should contain JSON
	if !strings.Contains(output, "photo1.jpg") {
		t.Errorf("expected output to contain 'photo1.jpg', got: %s", output)
	}
}

func TestRunCLI_WithFileArgs(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()

	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "image1.jpg")
	f2 := filepath.Join(tmpDir, "image2.png")
	os.WriteFile(f1, []byte("test"), 0644)
	os.WriteFile(f2, []byte("test"), 0644)

	oldStdout := captureStdout(t)
	defer oldStdout.Restore()

	files := []string{f1, f2}

	// Act
	runCLI(&cobra.Command{}, mockAnalyzer, &fake.MockLogger{}, "", files, false)

	// Assert
	if len(mockAnalyzer.AnalyzeFileCalls) != 2 {
		t.Errorf("expected 2 AnalyzeFile calls, got %d", len(mockAnalyzer.AnalyzeFileCalls))
	}

	for i, file := range files {
		if mockAnalyzer.AnalyzeFileCalls[i] != file {
			t.Errorf("expected call %d to be %q, got %q", i, file, mockAnalyzer.AnalyzeFileCalls[i])
		}
	}
}

func TestRunCLI_SingleFileArgReturnsObject(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()

	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "image1.jpg")
	os.WriteFile(f1, []byte("test"), 0644)

	oldStdout := captureStdout(t)
	defer oldStdout.Restore()

	// Act
	runCLI(&cobra.Command{}, mockAnalyzer, &fake.MockLogger{}, "", []string{f1}, false)

	// Assert
	output := oldStdout.Output()
	// Output should be a single JSON object, NOT an array of one
	// A JSON array starts with '[', a JSON object starts with '{'
	trimmed := strings.TrimSpace(output)
	if !strings.HasPrefix(trimmed, "{") {
		t.Errorf("expected single JSON object (starting with '{'), got: %s", output)
	}
}

func TestRunCLI_AnalyzeFileError(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()
	mockAnalyzer.AnalyzeFileError = errors.New("file not found")

	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "error.jpg")
	os.WriteFile(f1, []byte("test"), 0644)

	mockLogger := &fake.MockLogger{}
	// Act — should not panic, just log error and continue
	err := runCLI(&cobra.Command{}, mockAnalyzer, mockLogger, "", []string{f1}, false)

	// Assert
	if err != nil {
		t.Errorf("expected no error from runCLI (it should log and continue for positional args), got: %v", err)
	}
	if len(mockLogger.WarnCalls) == 0 {
		t.Error("expected at least one warning to be logged")
	}
}

func TestRunCLI_ScanDirectoryError(t *testing.T) {
	// Arrange
	mockAnalyzer := fake.NewMockImageAnalyzer()
	mockAnalyzer.ScanDirectoryError = errors.New("permission denied")

	// Act
	err := runCLI(&cobra.Command{}, mockAnalyzer, &fake.MockLogger{}, "some-path", nil, false)

	// Assert
	if err == nil {
		t.Fatal("expected error from runCLI, got nil")
	}

	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("expected error message to contain 'permission denied', got: %v", err)
	}
}

// ============================================================================
// runCLI EDGE CASE TESTS
// ============================================================================

func TestRunCLI_EmptyArgs(t *testing.T) {
	mockAnalyzer := fake.NewMockImageAnalyzer()
	mockLogger := &fake.MockLogger{}

	oldStdout := captureStdout(t)
	defer oldStdout.Restore()

	err := runCLI(&cobra.Command{}, mockAnalyzer, mockLogger, "", nil, false)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestRunCLI_DirectoryAsArg(t *testing.T) {
	mockAnalyzer := fake.NewMockImageAnalyzer()
	mockAnalyzer.ScanDirectoryResult = []domain.ImageFile{
		{Name: "file1.jpg", Path: "/test/file1.jpg"},
	}

	tmpDir := t.TempDir()

	oldStdout := captureStdout(t)
	defer oldStdout.Restore()

	err := runCLI(&cobra.Command{}, mockAnalyzer, &fake.MockLogger{}, "", []string{tmpDir}, false)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if len(mockAnalyzer.ScanDirectoryCalls) != 1 {
		t.Errorf("expected 1 ScanDirectory call, got %d", len(mockAnalyzer.ScanDirectoryCalls))
	}
}

func TestRunCLI_MixedFilesAndDirs(t *testing.T) {
	mockAnalyzer := fake.NewMockImageAnalyzer()

	tmpDir := t.TempDir()
	subDir := tmpDir + "/subdir"
	os.Mkdir(subDir, 0755)

	f1 := tmpDir + "/image1.jpg"
	f2 := subDir + "/image2.jpg"
	os.WriteFile(f1, []byte("test"), 0644)
	os.WriteFile(f2, []byte("test"), 0644)

	oldStdout := captureStdout(t)
	defer oldStdout.Restore()

	err := runCLI(&cobra.Command{}, mockAnalyzer, &fake.MockLogger{}, "", []string{tmpDir, f1}, false)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if len(mockAnalyzer.ScanDirectoryCalls) != 1 {
		t.Errorf("expected 1 ScanDirectory call, got %d", len(mockAnalyzer.ScanDirectoryCalls))
	}
	if len(mockAnalyzer.AnalyzeFileCalls) != 1 {
		t.Errorf("expected 1 AnalyzeFile call, got %d", len(mockAnalyzer.AnalyzeFileCalls))
	}
}

func TestRunCLI_WithPipe(t *testing.T) {
	mockAnalyzer := fake.NewMockImageAnalyzer()
	mockAnalyzer.AnalyzeStreamResult = &domain.ImageFile{
		Name: "stream.jpg",
		Metadata: domain.Metadata{
			Format:   "jpeg",
			FileSize: 1024,
		},
	}

	oldStdout := captureStdout(t)
	defer oldStdout.Restore()

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	err := runCLI(cmd, mockAnalyzer, &fake.MockLogger{}, "", nil, true)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if len(mockAnalyzer.AnalyzeStreamCalls) != 1 {
		t.Errorf("expected 1 AnalyzeStream call, got %d", len(mockAnalyzer.AnalyzeStreamCalls))
	}
}

func TestRunCLI_WithPipeAndOtherArgs(t *testing.T) {
	mockAnalyzer := fake.NewMockImageAnalyzer()
	mockAnalyzer.AnalyzeStreamResult = &domain.ImageFile{
		Name:     "stream.jpg",
		Metadata: domain.Metadata{Format: "jpeg"},
	}

	tmpDir := t.TempDir()
	f1 := tmpDir + "/file.jpg"
	os.WriteFile(f1, []byte("test"), 0644)

	oldStdout := captureStdout(t)
	defer oldStdout.Restore()

	err := runCLI(&cobra.Command{}, mockAnalyzer, &fake.MockLogger{}, "", []string{f1}, true)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if len(mockAnalyzer.AnalyzeStreamCalls) != 1 {
		t.Errorf("expected 1 AnalyzeStream call for pipe, got %d", len(mockAnalyzer.AnalyzeStreamCalls))
	}
}

func TestRunCLI_Stats(t *testing.T) {
	mockAnalyzer := fake.NewMockImageAnalyzer()
	mockAnalyzer.ScanDirectoryResult = []domain.ImageFile{
		{Name: "file1.jpg"},
		{Name: "file2.jpg"},
		{Name: "file3.jpg"},
	}

	mockLogger := &fake.MockLogger{}

	oldStdout := captureStdout(t)
	defer oldStdout.Restore()

	runCLI(&cobra.Command{}, mockAnalyzer, mockLogger, "/test", nil, false)

	// The mock logger should have recorded Info calls from the service
	// The exact messages depend on service implementation
	if len(mockLogger.InfoCalls) == 0 {
		t.Error("expected Info logging calls from service")
	}
}

func TestRunCLI_MultipleFileArgs(t *testing.T) {
	mockAnalyzer := fake.NewMockImageAnalyzer()

	tmpDir := t.TempDir()
	files := make([]string, 5)
	for i := 0; i < 5; i++ {
		files[i] = tmpDir + "/file" + string(rune('0'+i)) + ".jpg"
		os.WriteFile(files[i], []byte("test"), 0644)
	}

	oldStdout := captureStdout(t)
	defer oldStdout.Restore()

	err := runCLI(&cobra.Command{}, mockAnalyzer, &fake.MockLogger{}, "", files, false)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if len(mockAnalyzer.AnalyzeFileCalls) != 5 {
		t.Errorf("expected 5 AnalyzeFile calls, got %d", len(mockAnalyzer.AnalyzeFileCalls))
	}
}

func TestRunCLI_StatsWithErrors(t *testing.T) {
	mockAnalyzer := fake.NewMockImageAnalyzer()
	mockAnalyzer.AnalyzeFileError = errors.New("some error")

	tmpDir := t.TempDir()
	files := []string{tmpDir + "/error1.jpg", tmpDir + "/error2.jpg"}
	for _, f := range files {
		os.WriteFile(f, []byte("test"), 0644)
	}

	mockLogger := &fake.MockLogger{}

	runCLI(&cobra.Command{}, mockAnalyzer, mockLogger, "", files, false)

	// Should log warnings for errors
	if len(mockLogger.WarnCalls) == 0 {
		t.Error("expected Warn calls for file errors")
	}
}

func TestRunCLI_WithPathFlag_SingleDir(t *testing.T) {
	mockAnalyzer := fake.NewMockImageAnalyzer()
	mockAnalyzer.ScanDirectoryResult = []domain.ImageFile{
		{Name: "result1.jpg"},
	}

	oldStdout := captureStdout(t)
	defer oldStdout.Restore()

	err := runCLI(&cobra.Command{}, mockAnalyzer, &fake.MockLogger{}, "/my/photos", nil, false)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if len(mockAnalyzer.ScanDirectoryCalls) != 1 {
		t.Errorf("expected 1 ScanDirectory call, got %d", len(mockAnalyzer.ScanDirectoryCalls))
	}

	if mockAnalyzer.ScanDirectoryCalls[0] != "/my/photos" {
		t.Errorf("expected path '/my/photos', got %q", mockAnalyzer.ScanDirectoryCalls[0])
	}
}

func TestRunCLI_PathFlagTakesPrecedence(t *testing.T) {
	mockAnalyzer := fake.NewMockImageAnalyzer()

	tmpDir := t.TempDir()
	file := tmpDir + "/file.jpg"
	os.WriteFile(file, []byte("test"), 0644)

	oldStdout := captureStdout(t)
	defer oldStdout.Restore()

	runCLI(&cobra.Command{}, mockAnalyzer, &fake.MockLogger{}, tmpDir, []string{file}, false)

	// Path flag should trigger ScanDirectory, not AnalyzeFile
	if len(mockAnalyzer.ScanDirectoryCalls) != 1 {
		t.Errorf("expected 1 ScanDirectory call (path flag takes precedence), got %d", len(mockAnalyzer.ScanDirectoryCalls))
	}
}

// ============================================================================
// NewLocalesCmd TESTS
// ============================================================================

func TestNewLocalesCmd_Execute(t *testing.T) {
	cmd := NewLocalesCmd()
	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != "locales" {
		t.Errorf("expected Use 'locales', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Execute the command - it writes JSON to stdout
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewLocalesCmd_ReturnsJSON(t *testing.T) {
	cmd := NewLocalesCmd()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Execute()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []struct {
		Code        string `json:"code"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r).Decode(&result); err != nil {
		t.Fatalf("command should output valid JSON: %v", err)
	}

	if len(result) == 0 {
		t.Error("expected at least one locale")
	}

	// Verify "en" locale is present
	found := false
	for _, loc := range result {
		if loc.Code == "en" && loc.Description == "English" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'en' locale with description 'English'")
	}
}

// ============================================================================
// HELPER: Output Capturing
// ============================================================================

type capturedOutput struct {
	t       *testing.T
	target  **os.File
	oldFile *os.File
	reader  *os.File
	writer  *os.File
	outChan chan string
}

func captureStdout(t *testing.T) *capturedOutput {
	t.Helper()
	return captureFile(t, &os.Stdout)
}

func captureFile(t *testing.T, file **os.File) *capturedOutput {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	old := *file
	*file = w

	outChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outChan <- buf.String()
	}()

	return &capturedOutput{
		t:       t,
		target:  file,
		oldFile: old,
		reader:  r,
		writer:  w,
		outChan: outChan,
	}
}

func (c *capturedOutput) Restore() {
	c.writer.Close()
	*c.target = c.oldFile
	c.reader.Close()
}

func (c *capturedOutput) Output() string {
	c.writer.Close()
	result := <-c.outChan
	*c.target = c.oldFile
	c.reader.Close()
	return result
}
