package metadata

import (
	"context"

	"aether/internal/config"
	"aether/internal/player"
)

type MPVProvider struct {
	Player *player.Player
}

func (p MPVProvider) Fetch(ctx context.Context, station config.Station) (Metadata, error) {
	select {
	case <-ctx.Done():
		return Metadata{}, ctx.Err()
	default:
	}

	obj, err := p.Player.RawMetadata()
	if err != nil {
		return Metadata{}, err
	}

	md := Metadata{
		Artist: FirstString(obj, "artist", "ARTIST", "Artist"),
		Title:  FirstString(obj, "title", "TITLE", "Title"),
		Album:  FirstString(obj, "album", "ALBUM", "Album"),
		Raw:    FirstString(obj, "icy-title", "icy_title", "StreamTitle", "icy-title"),
		Source: "mpv",
	}

	if md.Title == "" && md.Raw != "" {
		md.Artist, md.Title = SplitICYTitle(md.Raw)
	}
	return md, nil
}
