package jwts

import "testing"

func TestParseToken(t *testing.T) {
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTExNzc3NTUsInRva2VuIjoiMTAwMiJ9.yMjgj-69PFid3dXzSpjOMdyD_wJa-6X8KSLuzICMlM8"
	ParseToken(tokenString, "ms_project")
}
