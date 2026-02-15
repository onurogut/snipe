package main

import "testing"

func TestParsePIDList(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []int
		wantErr bool
	}{
		{"single", "1234\n", []int{1234}, false},
		{"multiple", "100\n200\n300\n", []int{100, 200, 300}, false},
		{"duplicates", "100\n200\n100\n", []int{100, 200}, false},
		{"whitespace", "  100 \n  200\n", []int{100, 200}, false},
		{"empty", "", nil, true},
		{"garbage", "abc\nxyz\n", nil, true},
		{"mixed", "100\nabc\n200\n", []int{100, 200}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePIDList(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parsePIDList(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && !equalInts(got, tt.want) {
				t.Errorf("parsePIDList(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"single_command", "node", "node"},
		{"with_path", "node /app/server.js", "/app/server.js"},
		{"skip_flags", "node --inspect /app/server.js", "/app/server.js"},
		{"script_ext", "python app.py", "app.py"},
		{"backslash_path", `node C:\Users\app\server.js`, `C:\Users\app\server.js`},
		{"flags_only", "node --version", "node"},
		{"relative_path", "node src/index.js", "src/index.js"},
		{"multiple_paths", "node /first.js /second.js", "/first.js"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPath(tt.input)
			if got != tt.want {
				t.Errorf("extractPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLooksLikeFile(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"server.js", true},
		{"app.py", true},
		{"main.go", true},
		{"/usr/bin/node", false},  // no ext
		{"hello world.js", false}, // space
		{"file.toolongext", false},
		{"noext", false},
		{".hidden", false},
		{".env", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := looksLikeFile(tt.input)
			if got != tt.want {
				t.Errorf("looksLikeFile(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
