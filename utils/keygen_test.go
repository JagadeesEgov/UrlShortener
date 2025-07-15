package utils_test

import (
	"testing"
	"urlShortner/utils"

	"github.com/stretchr/testify/assert"
)

func TestGenerateShortKey(t *testing.T) {
	tests := []struct {
		length int
	}{
		{6},
		{5},
		{3},
	}
	for _, test := range tests {
		got, err := utils.GenerateShortKey(test.length)
		assert.NoError(t, err)
		
		if len(got) != test.length {
			t.Errorf("GenerateShortKey(%d) = %s, want length %d", test.length, got, test.length)
		}
		got1,err1 :=utils.GenerateShortKey(test.length)
		assert.NoError(t,err1)

		if(got1==got){
			t.Errorf("GenerateShortKey(%d) = %s, want different key", test.length, got)
		}
	}
	
}
