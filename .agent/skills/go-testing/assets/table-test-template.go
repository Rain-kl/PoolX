package example_test

import "testing"

func TestExample(t *testing.T) {
	tests := []struct {
		name string
		// TODO: add input fields
		// TODO: add expected output fields
	}{
		{
			name: "basic case",
			// TODO: fill in
		},
		{
			name: "edge case",
			// TODO: fill in
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: call function under test
			// TODO: compare got vs want
			// if diff := cmp.Diff(want, got); diff != "" {
			// 	t.Errorf("Example() mismatch (-want +got):\n%s", diff)
			// }
		})
	}
}
