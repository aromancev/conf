package confa

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestXxx(t *testing.T) {
	payload, err := proto.Marshal(&RecordingUpdate{
		RoomId:      []byte{1},
		RecordingId: []byte{2},
		UpdatedAt:   time.Now().UnixMilli(),
		Update: &RecordingUpdate_Update{
			Update: &RecordingUpdate_Update_ProcessingFinished{},
		},
	})
	require.NoError(t, err)
	update := &RecordingUpdate{}
	err = proto.Unmarshal(payload, update)
	require.NoError(t, err)
	assert.NotNil(t, update.Update)
}
