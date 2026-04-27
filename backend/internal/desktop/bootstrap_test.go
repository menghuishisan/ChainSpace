package desktop

import "testing"

func TestFileExistsRejectsMissingFile(t *testing.T) {
	if fileExists("Z:\\definitely-missing-file") {
		t.Fatal("expected missing file to return false")
	}
}

func TestWaitForTCPBuilderReturnsFunction(t *testing.T) {
	check := waitForTCP("127.0.0.1:65535")
	if check == nil {
		t.Fatal("expected non-nil check function")
	}
}

func TestComposeFilePathUsesDesktopCompose(t *testing.T) {
	got := composeFilePath(`E:\code\ChainPrac`)
	want := `E:\code\ChainPrac\deploy\docker-compose.desktop.yml`
	if got != want {
		t.Fatalf("compose path mismatch: got %q want %q", got, want)
	}
}
