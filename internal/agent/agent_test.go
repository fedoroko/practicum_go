package agent

import (
	"testing"
	"github.com/stretchr/testify/assert"
    // "github.com/stretchr/testify/require"
    "fmt"
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

// func TestSendMemStats(t *testing.T) {
// 	t.Run("sent", func (t *testing.T) {

// 		err := sendMemStats()
// 		fmt.Println(err)
// 		// require.NoError(t, err)	
// 	})
// }

