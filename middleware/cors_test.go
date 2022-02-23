package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorsMatchSubdomain(t *testing.T) {
	assert.True(t, matchSubdomain("*.example.com", "a.example.com"))
	assert.True(t, matchSubdomain("*.example.com", "b.example.com"))
	assert.False(t, matchSubdomain("*.example.com", "b.example1.com"))
	assert.True(t, matchSubdomain("a.*.com", "a.exmaple.com"))
	assert.True(t, matchSubdomain("a.*.com", "a.exmaple1.com"))
	assert.False(t, matchSubdomain("http://*.example.com", "https://a.example.com"))
}
