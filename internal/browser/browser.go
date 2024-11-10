package browser

import (
	"fmt"

	"github.com/google/wire"
	"github.com/pkg/browser"
)

var Set = wire.NewSet(
	wire.Struct(new(Browser), "*"),
	wire.Bind(new(Interface), new(*Browser)),
)

type Interface interface {
	Open(url string) error
}

type Browser struct{}

// Open opens the default browser.
func (*Browser) Open(url string) error {
	if err := browser.OpenURL(url); err != nil {
		return fmt.Errorf("could not open the browser: %w", err)
	}
	return nil
}
