package e2e

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"
)

func waitForServer(port string) error {
	for i := 0; i < 120; i++ {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", port), 1*time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("server not ready")
}

func TestMain(m *testing.M) {

	// Build binary
	buildCmd := exec.Command("go", "build", "-o", "test-server", "../../cmd/api")
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("Failed to build server: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove("test-server")

	// We use an explicit command
	cmd := exec.Command("./test-server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "UNIFIED_MOCK=true")
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}

	if err := waitForServer("8080"); err != nil {
		fmt.Printf("Server failed to start in time: %v\n", err)
		if cmd.Process != nil {
			cmd.Process.Signal(os.Interrupt)
		}
		cmd.Wait()
		os.Exit(1)
	}

	code := m.Run()

	// Graceful shutdown
	if cmd.Process != nil {
		cmd.Process.Signal(os.Interrupt)
	}
	cmd.Wait()
	os.Exit(code)
}
