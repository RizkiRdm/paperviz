package services

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestChartValuesUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    chartValues
		wantErr bool
	}{
		{
			name: "standard number array",
			json: `[72, 89]`,
			want: chartValues{72, 89},
		},
		{
			name: "single number",
			json: `[42]`,
			want: chartValues{42},
		},
		{
			name:    "empty array",
			json:    `[]`,
			want:    chartValues{},
		},
		{
			name: "comma-separated string",
			json: `"72, 89"`,
			want: chartValues{72, 89},
		},
		{
			name: "comma-separated no space",
			json: `"72,89"`,
			want: chartValues{72, 89},
		},
		{
			name: "space-separated string",
			json: `"72 89"`,
			want: chartValues{72, 89},
		},
		{
			name: "single value string",
			json: `"42"`,
			want: chartValues{42},
		},
		{
			name: "percentages",
			json: `"72%"`,
			want: chartValues{72},
		},
		{
			name: "percentages with comma",
			json: `"72%, 89%"`,
			want: chartValues{72, 89},
		},
		{
			name: "mixed whitespace and commas",
			json: `"72,89  90"`,
			want: chartValues{72, 89, 90},
		},
		{
			name:    "invalid string",
			json:    `"abc"`,
			wantErr: true,
		},
		{
			name:    "invalid mixed",
			json:    `"72, abc"`,
			wantErr: true,
		},
		{
			name: "decimals",
			json: `[72.5, 89.1]`,
			want: chartValues{72.5, 89.1},
		},
		{
			name: "string decimals",
			json: `"72.5, 89.1"`,
			want: chartValues{72.5, 89.1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v chartValues
			err := json.Unmarshal([]byte(tt.json), &v)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(v, tt.want) {
				t.Errorf("UnmarshalJSON() = %v, want %v", v, tt.want)
			}
		})
	}
}

// TestChartValuesRoundtrip ensures chartValues marshals back to standard JSON.
func TestChartValuesRoundtrip(t *testing.T) {
	v := chartValues{1, 2, 3}
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var got chartValues
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, v) {
		t.Errorf("roundtrip = %v, want %v", got, v)
	}
}

func TestChartValuesInStruct(t *testing.T) {
	input := `{"chart_type":"bar","labels":["A","B"],"values":"72, 89","title":"test"}`
	var elem textChartElem
	if err := json.Unmarshal([]byte(input), &elem); err != nil {
		t.Fatal(err)
	}
	if elem.ChartType != "bar" {
		t.Errorf("chart_type = %q, want bar", elem.ChartType)
	}
	if !reflect.DeepEqual([]float64(elem.Values), []float64{72, 89}) {
		t.Errorf("values = %v, want [72 89]", elem.Values)
	}
	if elem.Title != "test" {
		t.Errorf("title = %q, want test", elem.Title)
	}
}
