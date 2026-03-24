package cli

import (
	"fmt"

	"github.com/DementorAK/photometa/internal/fake"
)

// ExampleNewServerCmd demonstrates how to inspect the server command.
func ExampleNewServerCmd() {
	mock := fake.NewMockImageAnalyzer()

	cmd := NewServerCmd(mock, &fake.MockLogger{})

	// Print command details to verify configuration
	fmt.Println("Command:", cmd.Use)
	fmt.Println("Short:", cmd.Short)
	fmt.Println("Alias:", cmd.Aliases[0])

	// Output:
	// Command: server
	// Short: Start the metadata analysis web server
	// Alias: s
}
