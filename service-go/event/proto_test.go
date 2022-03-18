package event

import (
	"testing"
	"time"

	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/require"
)

func TestProtoMarshalling(t *testing.T) {
	var event Event
	f := fuzz.New().NilChance(0).MaxDepth(100).NumElements(1, 3)
	// Time is passed as UNIX milliseconds.
	f = f.Funcs(func(val *time.Time, c fuzz.Continue) {
		*val = time.UnixMilli(val.UnixMilli())
	})
	f.Fuzz(&event)

	require.Equal(t, event, FromProto(ToProto(event)))
}
