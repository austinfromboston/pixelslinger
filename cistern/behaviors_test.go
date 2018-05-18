package cistern

import (
	"testing"
	"github.com/stretchr/testify/assert"
)
func TestVerticalHeight(t *testing.T) {
	assert := assert.New(t)

	t.Run("vertical height on first strip", func(t *testing.T) {
		strip1 := ZigZagStrip{100,200}
		assert.Equal(strip1.VerticalHeight(2), 2)
		assert.Equal(strip1.VerticalHeight(102), 98)
	})
}
