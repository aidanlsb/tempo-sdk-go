package extension

import (
	"context"
	"testing"
)

type fixedInvoker struct {
	descriptor Descriptor
	name       string
}

func (i *fixedInvoker) Describe(context.Context) (Descriptor, error) {
	return i.descriptor, nil
}

func (i *fixedInvoker) Invoke(
	_ context.Context,
	_ InvokeRequest,
) (InvokeResult, error) {
	return InvokeResult{Events: []Event{{Type: i.name}}}, nil
}

func TestCompositeInvokerRoutesHigherPrecedenceOverride(t *testing.T) {
	stock := &fixedInvoker{
		name: "stock",
		descriptor: Descriptor{
			ID: "stock",
			Recipes: []Recipe{{
				ID: "tempo.move", Name: "Move", Handler: "move",
			}},
		},
	}
	pack := &fixedInvoker{
		name: "pack",
		descriptor: Descriptor{
			ID: "pack",
			Recipes: []Recipe{{
				ID: "tempo.move", Name: "Custom Move", Handler: "custom",
			}},
		},
	}
	composite, err := Compose(context.Background(), stock, pack)
	if err != nil {
		t.Fatal(err)
	}
	descriptor, err := composite.Describe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(descriptor.Recipes) != 1 ||
		descriptor.Recipes[0].Name != "Custom Move" {
		t.Fatalf("composite descriptor = %#v", descriptor)
	}
	result, err := composite.Invoke(context.Background(), InvokeRequest{
		Kind:    HandlerRecipe,
		Handler: descriptor.Recipes[0].Handler,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Events) != 1 || result.Events[0].Type != "pack" {
		t.Fatalf("composite result = %#v", result)
	}
}

func TestCompositeInvokerRoutesDecisionResolver(t *testing.T) {
	pack := &fixedInvoker{
		name: "resolver",
		descriptor: Descriptor{
			ID: "pack",
			Resolvers: []DecisionResolver{{
				ID: "test.resolve", Name: "Resolve", Handler: "resolve",
			}},
		},
	}
	composite, err := Compose(context.Background(), pack)
	if err != nil {
		t.Fatal(err)
	}
	descriptor, err := composite.Describe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(descriptor.Resolvers) != 1 ||
		descriptor.Resolvers[0].Handler == "resolve" {
		t.Fatalf("composite resolver descriptor = %#v", descriptor.Resolvers)
	}
	result, err := composite.Invoke(context.Background(), InvokeRequest{
		Kind: HandlerDecisionResolver, Handler: descriptor.Resolvers[0].Handler,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Events) != 1 || result.Events[0].Type != "resolver" {
		t.Fatalf("composite resolver result = %#v", result)
	}
}

func TestCompositeInvokerKeepsNarrowerScope(t *testing.T) {
	stock := &fixedInvoker{
		descriptor: Descriptor{
			ID: "stock",
			Recipes: []Recipe{{
				ID: "tempo.move", Name: "Move", Handler: "move",
			}},
		},
	}
	pack := &fixedInvoker{
		descriptor: Descriptor{
			ID: "pack",
			Recipes: []Recipe{{
				ID: "tempo.move", Name: "Scenario Move",
				Scope: Scope{
					Kind: ScopeScenario, CampaignID: "story", ScenarioID: "opening",
				},
				Handler: "scenario-move",
			}},
		},
	}
	composite, err := Compose(context.Background(), stock, pack)
	if err != nil {
		t.Fatal(err)
	}
	descriptor, _ := composite.Describe(context.Background())
	if len(descriptor.Recipes) != 2 {
		t.Fatalf("composite descriptor = %#v", descriptor)
	}
}

func TestCompositeInvokerRetainsDisableTombstone(t *testing.T) {
	stock := &fixedInvoker{
		descriptor: Descriptor{
			ID: "stock",
			Recipes: []Recipe{{
				ID: "tempo.move", Name: "Move", Handler: "move",
			}},
		},
	}
	pack := &fixedInvoker{
		descriptor: Descriptor{
			ID: "pack",
			Recipes: []Recipe{{
				ID: "tempo.move", Name: "Move", Disabled: true,
			}},
		},
	}
	composite, err := Compose(context.Background(), stock, pack)
	if err != nil {
		t.Fatal(err)
	}
	descriptor, _ := composite.Describe(context.Background())
	if len(descriptor.Recipes) != 1 || !descriptor.Recipes[0].Disabled {
		t.Fatalf("disable descriptor = %#v", descriptor)
	}
}
