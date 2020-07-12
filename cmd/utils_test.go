package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringBetween(t *testing.T) {
	source := "mypod [mynamespace]"
	actual := StringBetween(source, "[", "]")
	expected := "mynamespace"
	assert.Equal(t, expected, actual)
}

func TestInvalidStringBetween(t *testing.T) {
	source := "mypod [mynamespace]"
	actual := StringBetween(source, "##", "]")
	expected := ""
	assert.Equal(t, expected, actual)
}

func TestStringBefore(t *testing.T) {
	source := "mypod [mynamespace]"
	actual := StringBefore(source, " [")
	expected := "mypod"
	assert.Equal(t, expected, actual)
}

func TestInvalidStringBefore(t *testing.T) {
	source := "mypod [mynamespace]"
	actual := StringBefore(source, "#")
	expected := ""
	assert.Equal(t, expected, actual)
}
