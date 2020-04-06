package boomerang

import (
	"context"
	"github.com/sirupsen/logrus"
	"testing"
	"time"
)

func Test_Throw(t *testing.T) {
	ctx := context.Background()
	out := logrus.StandardLogger().Out

	cfg := Config{
		Application: "deploy/nginx",
		Image:       "some/image",
		Namespace:   "web",
		Timeout:     30 * time.Second,
	}

	if err := Throw(ctx, out, cfg); err != nil {
		t.Fatalf("%+v", err)
	}
}
