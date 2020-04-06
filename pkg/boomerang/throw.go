package boomerang

import (
	"context"
	"github.com/sirupsen/logrus"
	"io"
	"time"
)

type Config struct {
	Application string
	Namespace   string
	Image       string
	Timeout     time.Duration
}

func Throw(ctx context.Context, out io.Writer, cfg Config) error {
	logrus.Infof("%+v", cfg)
	return nil
}
