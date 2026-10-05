package mail

import "testing"

func TestGetDomain(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "default addr", input: "user@yellowhat.cz", want: "yellowhat.cz", wantErr: false},
		{name: "empty domain", input: "user@", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDomain(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetDomain(%q) error = %v, want error = %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("GetDomain(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
