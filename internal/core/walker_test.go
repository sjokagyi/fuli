package core_test

//import (
//"os"
//"path/filepath"
//"strings"
//"testing"
//"time"
//
//tea "github.com/charmbracelet/bubbletea"
//
//"github.com/sjokagyi/fuli/internal/config"
//"github.com/sjokagyi/fuli/internal/core"
//"github.com/sjokagyi/fuli/internal/ignore"
//)
//
//// mockProgram simulates a tea.Program for testing by channeling messages.
//type mockProgram struct {
//msgs chan tea.Msg
//}
//
//func (mp *mockProgram) Send(msg tea.Msg) {
//mp.msgs <- msg
//}
//
//func TestWalker(t *testing.T) {
//// 1. Setup Temp Directory mirroring scalable structure
//tmpDir, err := os.MkdirTemp("", "fuli-test-*")
//if err != nil {
//t.Fatal(err)
//}
//defer os.RemoveAll(tmpDir)
//
//// Create subdirs and files
//dirs := []string{
//"services/message-service/application",
//"services/message-service/cmd/message-service",
//"services/message-service/internal/adapters/realtime",
//"services/message-service/internal/core",
//"services/message-service/internal/domain",
//"services/message-service/internal/transport/rest",
//"services/websocket-manager/application",
//"services/websocket-manager/cmd/websocket-manager",
//"services/websocket-manager/internal/adapters/realtime",
//"services/websocket-manager/internal/core",
//"services/websocket-manager/internal/domain",
//"services/websocket-manager/internal/transport/websocket",
//}
//for _, d := range dirs {
//if err := os.MkdirAll(filepath.Join(tmpDir, d), 0755); err != nil {
//t.Fatal(err)
//}
//}
//
//// Create sample files
//files := map[string]string{
//".env":               "SECRET=123",
//"go.mod":             "module test",
//"go.sum":             "sums",
//"docker-compose.yml": "services: []",
//"README.md":          "Read me",
//"services/message-service/application/app.go":          "package app",
//"services/message-service/cmd/message-service/main.go": "func main() {}",
//// Add more as needed
//}
//for file, content := range files {
//if err := os.WriteFile(filepath.Join(tmpDir, file), []byte(content), 0644); err != nil {
//t.Fatal(err)
//}
//}
//
//// Create .contextignore in root
//ignoreContent := `# Ignore build artifacts
///build
//*.exe
//# Ignore dependencies
//node_modules/
//# Ignore secrets
//.env
//.gitignore
///.git
///gen
//go.mod
//go.sum
//go.work.sum
//!/database/querier.go
///database
//`
//os.WriteFile(filepath.Join(tmpDir, ".contextignore"), []byte(ignoreContent), 0644)
//
//// Create a binary file
//os.WriteFile(filepath.Join(tmpDir, "test.exe"), []byte{0x00, 0xFF}, 0644)
//
//// 2. Config
//cfg := config.RunConfig{
//SourceDir:      tmpDir,
//OutputPath:     filepath.Join(t.TempDir(), "output.txt"),
//IgnoreFiles:    []string{},
//IsVerbose:      true,
//DryRun:         false,
//FollowSymlinks: false,
//}
//
//// 3. Matcher
//matcher := ignore.NewMatcher(cfg.SourceDir)
//patterns, _ := ignore.ParseFile(filepath.Join(tmpDir, ".contextignore"))
//matcher.AddPatterns(patterns)
//
//// 4. Mock Program
//mp := &mockProgram{msgs: make(chan tea.Msg, 100)}
//
//// 5. Run Walker
//walker := core.NewWalker(cfg, matcher)
//walker.Start(mp)
//
//// 6. Wait for Completion
//var completion core.CompletionMsg
//select {
//case msg := <-mp.msgs:
//if c, ok := msg.(core.CompletionMsg); ok {
//completion = c
//} else {
//t.Fatalf("Unexpected msg: %v", msg)
//}
//case <-time.After(5 * time.Second):
//t.Fatal("Timeout waiting for completion")
//}
//
//// 7. Assertions
//expectedProcessed := 7 // Adjust based on files minus ignored/binary (e.g., README.md, docker-compose.yml, go files, etc.)
//if completion.TotalFiles != expectedProcessed {
//t.Errorf("Expected %d files processed, got %d", expectedProcessed, completion.TotalFiles)
//}
//if completion.Skipped < 5 { // .env, go.mod, go.sum, test.exe, etc.
//t.Errorf("Expected at least 5 skipped, got %d", completion.Skipped)
//}
//
//// 8. Output Verification
//output, err := os.ReadFile(cfg.OutputPath)
//if err != nil {
//t.Fatal(err)
//}
//outputStr := string(output)
//if !strings.Contains(outputStr, "This is the content of app.go in the scalable_websocket_service_tutorial/services/message-service/application directory") {
//t.Error("Output missing expected header/content")
//}
//// Add more regex/asserts for specific headers
//}
