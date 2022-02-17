package double

import "github.com/aromancev/confa/proto/rtc"

// Checking that doubles implement the interface.
var _ rtc.RTC = NewMemory()
