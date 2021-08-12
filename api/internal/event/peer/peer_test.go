package peer

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage(t *testing.T) {
	t.Parallel()

	expected := Message{
		Type: TypeTrickle,
		Payload: Trickle{
			Target: 1,
		},
	}
	buf, err := json.Marshal(expected)
	require.NoError(t, err)
	var actual Message
	require.NoError(t, json.Unmarshal(buf, &actual))
	assert.Equal(t, expected, actual)
}
