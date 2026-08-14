package conn

import "testing"

func TestReadSecret_commandTakesPriorityOverPrompt(t *testing.T) {
	secret, err := ReadSecret("unused prompt: ", "echo from-command")
	if err != nil {
		t.Fatalf("ReadSecret() error = %v", err)
	}
	if secret != "from-command" {
		t.Errorf("ReadSecret() = %q, want from-command", secret)
	}
}

func TestReadSecret_commandFailureIsReported(t *testing.T) {
	if _, err := ReadSecret("prompt: ", "false"); err == nil {
		t.Fatal("ReadSecret() error = nil, want an error for a failing command")
	}
}
