package browser

import (
	"github.com/google/wire"
	"github.com/pkg/browser"
	"golang.org/x/xerrors"
)

var Set = wire.NewSet(
	wire.Struct(new(Browser), "*"),
	wire.Bind(new(Interface), new(*Browser)),
)

//go:generate mockgen -destination mock_browser/mock_browser.go github.com/int128/kauthproxy/pkg/browser Interface

type Interface interface {
	Open(url string) error
}

type Browser struct{}

// Open opens the default browser.
func (*Browser) Open(url string) error {
	if err := browser.OpenURL(url); err != nil {
		return xerrors.Errorf("could not open the browser: %w", err)
	}
	return nil
}
