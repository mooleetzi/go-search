package cache

import "testing"

func TestRedis(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "t",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Redis(false)
		})
	}
}
