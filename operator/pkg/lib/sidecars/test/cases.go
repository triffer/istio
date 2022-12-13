package test

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/kyma-project/istio/operator/pkg/lib/sidecars"
	"github.com/kyma-project/istio/operator/pkg/lib/sidecars/pods"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

const annotationName = "kubectl.kubernetes.io/restartedAt"

func (s *scenario) aRestartHappens() error {
	return sidecars.RestartPodWithDifferentSidecarImage(context.TODO(),
		s.Client,
		pods.SidecarImage{Repository: "istio/proxyv2", Tag: s.istioVersion},
		&s.logger)
}

func (s *scenario) allRequiredResourcesAreDeleted() error {
	for _, v := range s.ToBeDeletedObjects {
		obj := v
		err := s.Client.Get(context.TODO(), types.NamespacedName{Name: v.GetName(), Namespace: v.GetNamespace()}, obj)
		if err == nil {
			return fmt.Errorf("the object %s/%s was deleted but shouldn't", v.GetNamespace(), v.GetName())
		}

		if !k8serrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (s *scenario) allRequiredResourcesAreRestarted() error {
	for _, v := range s.ToBeRestartedObjects {
		obj := v
		err := s.Client.Get(context.TODO(), types.NamespacedName{Name: v.GetName(), Namespace: v.GetNamespace()}, obj)
		if err != nil {
			return err
		}

		if _, ok := obj.GetAnnotations()[annotationName]; !ok {
			return fmt.Errorf("the annotation %s wasn't applied for object %s/%s", annotationName, obj.GetNamespace(), obj.GetName())
		}
	}
	return nil
}

func (s *scenario) thereArePodsWithNotYetInjectedSidecars() error {
	return s.WithNotYetInjectedPods()
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	var s scenario

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		scen, err := NewScenario()
		s = *scen
		return ctx, err
	})

	ctx.Step(`^there is cluster with Istio "([^"]*)"$`, s.WithIstioVersion)
	ctx.Step(`^a restart happens$`, s.aRestartHappens)
	ctx.Step(`^all required resources are deleted$`, s.allRequiredResourcesAreDeleted)
	ctx.Step(`^all required resources are restarted$`, s.allRequiredResourcesAreRestarted)
	ctx.Step(`^there are pods with not yet injected sidecars$`, s.thereArePodsWithNotYetInjectedSidecars)
}
