package double

import "github.com/aromancev/confa/internal/proto/rtc"

// Checking that doubles implements the interface.
var _ rtc.RTC = NewMemory()
