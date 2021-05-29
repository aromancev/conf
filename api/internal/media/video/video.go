package video

import (
	"fmt"
	"path"

	"github.com/aromancev/confa/internal/platform/mkv"
)

type Converter struct {
	mediaDir string
}

func NewConverter(mediaDir string) *Converter {
	return &Converter{mediaDir: mediaDir}
}

func (p *Converter) Convert(mediaID string) error {
	err := mkv.BuildClues(
		path.Join(p.mediaDir, mediaID, "raw.webm"),
		path.Join(p.mediaDir, mediaID, "clued.webm"),
	)
	if err != nil {
		return fmt.Errorf("failed to build clues: %w", err)
	}

	err = mkv.Convert(
		path.Join(p.mediaDir, mediaID, "clued.webm"),
		path.Join(p.mediaDir, mediaID, "stream.webm"),
	)
	if err != nil {
		return fmt.Errorf("failed to convert: %w", err)
	}

	err = mkv.CreateManifest(
		path.Join(p.mediaDir, mediaID, "stream.webm"),
		path.Join(p.mediaDir, mediaID, "manifest.mpd"),
	)
	if err != nil {
		return fmt.Errorf("failed create manifest: %w", err)
	}

	//_ = os.Remove(path.Join(p.mediaDir, mediaID, "raw.webm"))
	//_ = os.Remove(path.Join(p.mediaDir, mediaID, "clued.webm"))
	return nil
}
