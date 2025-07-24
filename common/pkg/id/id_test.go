package id

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToLLMarinerModelID(t *testing.T) {
	tcs := []struct {
		id   string
		want string
	}{
		{
			id:   "meta-llama/Llama-3.3-70B-Instruct",
			want: "meta-llama-Llama-3.3-70B-Instruct",
		},
	}

	for _, tc := range tcs {
		got := ToLLMarinerModelID(tc.id)
		assert.Equal(t, tc.want, got)
	}
}
