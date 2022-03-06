package krill

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnumStringsIntersection(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		i := enumStringsIntersection("PROPERTY_FORMAT_STRING", "PROPERTY_FORMAT_INT")

		a := assert.New(t)
		a.Equal(i, "PROPERTY_FORMAT_")
	})
}
