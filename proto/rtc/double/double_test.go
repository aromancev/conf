package double

import "github.com/aromancev/proto/rtc"

// Checking that doubles implements the interface.
var _ rtc.RTC = NewMemory()
