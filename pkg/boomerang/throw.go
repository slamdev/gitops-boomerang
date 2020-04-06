package boomerang

import (
	"context"
	"errors"
	"fmt"
	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/kubedog/pkg/tracker"
	"github.com/flant/kubedog/pkg/trackers/rollout/multitrack"
	"github.com/flant/logboek"
	"io"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

type Config struct {
	Application string
	Namespace   string
	Image       string
	Timeout     time.Duration
}

func Throw(ctx context.Context, out io.Writer, cfg Config) error {
	if err := kube.Init(kube.InitOptions{}); err != nil {
		return fmt.Errorf("unable to initialize kube: %w", err)
	}
	kind, name, err := parseApplicationString(cfg.Application)
	if err != nil {
		return fmt.Errorf("unable parse application string: %w", err)
	}
	logboek.LogHighlightF("Polling image %s in %s \n", cfg.Image, name)
	if err := waitForImageUpdate(cfg.Namespace, kind, name, cfg.Image, cfg.Timeout); err != nil {
		return fmt.Errorf("failed to wait for image update: %w", err)
	}
	logboek.LogInfoLn("image was updated")
	logboek.LogHighlightF("Polling status of %s \n", name)
	if err := waitForStatusUpdate(ctx, cfg.Namespace, kind, name); err != nil {
		return fmt.Errorf("failed to wait for status update: %w", err)
	}
	return nil
}

func waitForStatusUpdate(ctx context.Context, namespace string, kind string, name string) error {
	options := multitrack.MultitrackOptions{
		StatusProgressPeriod: time.Second * 5,
		Options:              tracker.Options{ParentContext: ctx},
	}
	specs := multitrack.MultitrackSpecs{}
	spec := multitrack.MultitrackSpec{
		ResourceName: name,
		Namespace:    namespace,
	}
	switch kind {
	case "deployment":
		specs.Deployments = []multitrack.MultitrackSpec{spec}
	case "statefulset":
		specs.StatefulSets = []multitrack.MultitrackSpec{spec}
	case "daemonset":
		specs.DaemonSets = []multitrack.MultitrackSpec{spec}
	default:
		return fmt.Errorf("failed to map kind %s to MultitrackSpec", kind)
	}
	return multitrack.Multitrack(kube.Kubernetes, specs, options)
}

func doUntil(t time.Duration, f func() (bool, error)) error {
	timeout := time.After(t)
	ticker := time.Tick(5 * time.Second)
	for {
		select {
		case <-timeout:
			return errors.New("timed out")
		case <-ticker:
			ok, err := f()
			if err != nil {
				return err
			} else if ok {
				return nil
			}
		}
	}
}

func waitForImageUpdate(namespace string, kind string, name string, image string, timeout time.Duration) error {
	err := doUntil(timeout, func() (bool, error) {
		logboek.LogInfoLn("image was not updated yet")
		switch kind {
		case "deployment":
			return isDeploymentImageUpdated(namespace, name, image)
		case "statefulset":
			return isStatefulSetImageUpdated(namespace, name, image)
		case "daemonset":
			return isDaemonSetImageUpdated(namespace, name, image)
		default:
			return false, fmt.Errorf("failed to map kind %s to supported function", kind)
		}
	})
	return err
}

func isDaemonSetImageUpdated(namespace string, name string, image string) (bool, error) {
	d, err := kube.Kubernetes.AppsV1().DaemonSets(namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get daemonsets %s in %s", name, namespace)
	}
	for _, c := range d.Spec.Template.Spec.Containers {
		if c.Image == image {
			return true, nil
		}
	}
	return false, nil
}

func isStatefulSetImageUpdated(namespace string, name string, image string) (bool, error) {
	s, err := kube.Kubernetes.AppsV1().StatefulSets(namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get statefulset %s in %s", name, namespace)
	}
	for _, c := range s.Spec.Template.Spec.Containers {
		if c.Image == image {
			return true, nil
		}
	}
	return false, nil
}

func isDeploymentImageUpdated(namespace string, name string, image string) (bool, error) {
	d, err := kube.Kubernetes.AppsV1().Deployments(namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get deployment %s in %s", name, namespace)
	}
	for _, c := range d.Spec.Template.Spec.Containers {
		if c.Image == image {
			return true, nil
		}
	}
	return false, nil
}

func parseApplicationString(application string) (string, string, error) {
	parts := strings.Split(application, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("application should have a 'kind/name' form, got: %s", application)
	}
	name := parts[1]
	switch parts[0] {
	case "deployment", "deploy":
		return "deployment", name, nil
	case "sts", "statefulset":
		return "statefulset", name, nil
	case "ds", "daemonset":
		return "daemonset", name, nil
	default:
		return "", "", fmt.Errorf("kind should be one of deployment,statefulset,daemonset, got: %s", parts[0])
	}
}
