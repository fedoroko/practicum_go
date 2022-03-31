package agent

import (
	"github.com/stretchr/testify/assert"
	"testing"
	// "github.com/stretchr/testify/require"
	// "fmt"
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
