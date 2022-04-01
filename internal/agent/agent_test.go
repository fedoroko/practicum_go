package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectMemStats(t *testing.T) {
	t.Run("collected", func(t *testing.T) {
		count := pollCount

		var emptyStats []memStat
		collectMemStats()
		assert.NotEqual(t, stats, emptyStats)
		assert.NotEqual(t, count, pollCount)
	})
}
