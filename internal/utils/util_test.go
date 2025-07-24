package utils_test

import (
	"testing"
	"urlShortner/internal/utils"
)
func TestValidateURL(t *testing.T){
	test := []struct{
		url string
		expected bool
		}{
		{"https://www.google.com",true},
		{"https://www.google.com/search?q=golang",true},
		{"https://www.google.com/search?q=golang&oq=golang",true},
		{"http://www.google@com",false},
		{"https://www.*.com",false},
	}

	for _,test := range test{
		got := utils.ValidateURL(test.url)
		if(got!=test.expected){
			t.Errorf("ValidateURL(%s) = %v, want true",test.url,got)
		}
		
	}
}