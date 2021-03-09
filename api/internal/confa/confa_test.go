package confa

import (
	"os"
	"testing"

	"github.com/aromancev/confa/internal/platform/psql/double"
)

func TestMain(m *testing.M) {
	code := m.Run()
	double.Purge()
	os.Exit(code)
}
