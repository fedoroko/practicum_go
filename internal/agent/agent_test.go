package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectMemStats(t *testing.T) {
	t.Run("collected", func(t *testing.T) {
		count := int64(0)
		var empty []gStat

		nonEmpty, newCount := collectMemStats(count)

		assert.NotEqual(t, empty, nonEmpty)
		assert.NotEqual(t, count, newCount)
	})
}
