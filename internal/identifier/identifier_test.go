package identifier

import "testing"

func TestUsername(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"zuck", "zuck"},
		{"@zuck", "zuck"},
		{"https://www.threads.com/@zuck", "zuck"},
		{" https://www.threads.net/@zuck ", "zuck"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := Username(tc.in)
			if err != nil || got != tc.want {
				t.Fatalf("Username(%q) = %q, %v; want %q", tc.in, got, err, tc.want)
			}
		})
	}
}

func TestUsernameRejectsInvalidInput(t *testing.T) {
	for _, input := range []string{"", "https://example.com/@zuck", "https://www.threads.com/@zuck/post/ABC123"} {
		t.Run(input, func(t *testing.T) {
			if _, err := Username(input); err == nil {
				t.Fatalf("Username(%q) succeeded", input)
			}
		})
	}
}

func TestPost(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"https://www.threads.com/@zuck/post/ABC123", "https://www.threads.com/@zuck/post/ABC123", false},
		{"https://www.threads.net/@zuck/post/ABC123", "https://www.threads.com/@zuck/post/ABC123", false},
		{"ABC123", "", true},
		{"123456789", "", true},
		{"https://example.com/@zuck/post/ABC123", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := Post(tc.in)
			if (err != nil) != tc.wantErr || got != tc.want {
				t.Fatalf("Post(%q) = %q, %v; want %q, error=%v", tc.in, got, err, tc.want, tc.wantErr)
			}
		})
	}
}
