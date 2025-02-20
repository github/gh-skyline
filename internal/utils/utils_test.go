package utils

import "testing"

func TestParseYearRange(t *testing.T) {
	tests := []struct {
		name          string
		yearRange     string
		wantStart     int
		wantEnd       int
		wantErr       bool
		wantErrString string
	}{
		{
			name:      "single year",
			yearRange: "2024",
			wantStart: 2024,
			wantEnd:   2024,
			wantErr:   false,
		},
		{
			name:      "year range",
			yearRange: "2020-2024",
			wantStart: 2020,
			wantEnd:   2024,
			wantErr:   false,
		},
		{
			name:          "invalid format",
			yearRange:     "2020-2024-2025",
			wantErr:       true,
			wantErrString: "invalid year range format",
		},
		{
			name:      "invalid number",
			yearRange: "abc-2024",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := ParseYearRange(tt.yearRange)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseYearRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.wantErrString != "" && err.Error() != tt.wantErrString {
				t.Errorf("parseYearRange() error = %v, wantErrString %v", err, tt.wantErrString)
				return
			}
			if !tt.wantErr {
				if start != tt.wantStart {
					t.Errorf("parseYearRange() start = %v, want %v", start, tt.wantStart)
				}
				if end != tt.wantEnd {
					t.Errorf("parseYearRange() end = %v, want %v", end, tt.wantEnd)
				}
			}
		})
	}
}

func TestValidateYearRange(t *testing.T) {
	tests := []struct {
		name      string
		startYear int
		endYear   int
		wantErr   bool
	}{
		{
			name:      "valid range",
			startYear: 2020,
			endYear:   2024,
			wantErr:   false,
		},
		{
			name:      "invalid start year",
			startYear: 2007,
			endYear:   2024,
			wantErr:   true,
		},
		{
			name:      "start after end",
			startYear: 2024,
			endYear:   2020,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateYearRange(tt.startYear, tt.endYear)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateYearRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormatYearRange(t *testing.T) {
	tests := []struct {
		name      string
		startYear int
		endYear   int
		want      string
	}{
		{
			name:      "same year",
			startYear: 2024,
			endYear:   2024,
			want:      "2024",
		},
		{
			name:      "different years",
			startYear: 2020,
			endYear:   2024,
			want:      "2020-24",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatYearRange(tt.startYear, tt.endYear)
			if got != tt.want {
				t.Errorf("formatYearRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateOutputFilename(t *testing.T) {
	tests := []struct {
		name      string
		user      string
		startYear int
		endYear   int
		output    string
		want      string
	}{
		{
			name:      "single year",
			user:      "testuser",
			startYear: 2024,
			endYear:   2024,
			output:    "",
			want:      "testuser-2024-github-skyline.stl",
		},
		{
			name:      "year range",
			user:      "testuser",
			startYear: 2020,
			endYear:   2024,
			output:    "",
			want:      "testuser-2020-24-github-skyline.stl",
		},
		{
			name:      "override",
			user:      "testuser",
			startYear: 2020,
			endYear:   2024,
			output:    "myoutput.stl",
			want:      "myoutput.stl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateOutputFilename(tt.user, tt.startYear, tt.endYear, tt.output, "")
			if got != tt.want {
				t.Errorf("generateOutputFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}
