package extension

import (
	"context"
	"fmt"
	"sort"
	"strconv"
)

type compositeRoute struct {
	invoker Invoker
	handler string
}

type CompositeInvoker struct {
	descriptor Descriptor
	routes     map[string]compositeRoute
	children   []Invoker
}

// Compose merges Invokers from lowest to highest precedence. Later Invokers
// replace same-ID registrations in the same scope; narrower runtime scopes are
// still selected by the host from the merged descriptor.
func Compose(ctx context.Context, invokers ...Invoker) (*CompositeInvoker, error) {
	composite := &CompositeInvoker{
		descriptor: Descriptor{ID: "tempo.composite"},
		routes:     make(map[string]compositeRoute),
		children:   append([]Invoker(nil), invokers...),
	}
	recipes := make(map[string]Recipe)
	systems := make(map[string]System)
	perceptions := make(map[string]PerceptionRule)
	resolvers := make(map[string]DecisionResolver)
	components := make(map[string]ComponentSchema)
	for index, invoker := range invokers {
		if invoker == nil {
			continue
		}
		descriptor, err := invoker.Describe(ctx)
		if err != nil {
			return nil, fmt.Errorf("describe Invoker %d: %w", index, err)
		}
		if err := ValidateDescriptor(descriptor); err != nil {
			return nil, fmt.Errorf("invalid Invoker %d descriptor: %w", index, err)
		}
		for _, recipe := range descriptor.Recipes {
			if !recipe.Disabled {
				handler := compositeHandler(index, recipe.Handler)
				composite.routes[handler] = compositeRoute{
					invoker: invoker,
					handler: recipe.Handler,
				}
				recipe.Handler = handler
			}
			recipes[scopedID(recipe.Scope, recipe.ID)] = recipe
		}
		for _, system := range descriptor.Systems {
			handler := compositeHandler(index, system.Handler)
			composite.routes[handler] = compositeRoute{
				invoker: invoker,
				handler: system.Handler,
			}
			system.Handler = handler
			systems[scopedID(system.Scope, system.ID)] = system
		}
		for _, rule := range descriptor.Perceptions {
			handler := compositeHandler(index, rule.Handler)
			composite.routes[handler] = compositeRoute{
				invoker: invoker,
				handler: rule.Handler,
			}
			rule.Handler = handler
			perceptions[scopedID(rule.Scope, rule.ID)] = rule
		}
		for _, resolver := range descriptor.Resolvers {
			handler := compositeHandler(index, resolver.Handler)
			composite.routes[handler] = compositeRoute{
				invoker: invoker,
				handler: resolver.Handler,
			}
			resolver.Handler = handler
			resolvers[scopedID(resolver.Scope, resolver.ID)] = resolver
		}
		for _, component := range descriptor.Components {
			components[scopedID(component.Scope, component.ID)] = component
		}
		if descriptor.Character != nil {
			value := *descriptor.Character
			composite.descriptor.Character = &value
		}
	}
	composite.descriptor.Recipes = sortedValues(recipes)
	composite.descriptor.Systems = sortedValues(systems)
	composite.descriptor.Perceptions = sortedValues(perceptions)
	composite.descriptor.Resolvers = sortedValues(resolvers)
	composite.descriptor.Components = sortedValues(components)
	if err := ValidateDescriptor(composite.descriptor); err != nil {
		return nil, err
	}
	return composite, nil
}

func (c *CompositeInvoker) Describe(context.Context) (Descriptor, error) {
	return c.descriptor, nil
}

func (c *CompositeInvoker) Invoke(
	ctx context.Context,
	request InvokeRequest,
) (InvokeResult, error) {
	route, ok := c.routes[request.Handler]
	if !ok {
		return InvokeResult{}, fmt.Errorf(
			"composite handler %q is not registered",
			request.Handler,
		)
	}
	request.Handler = route.handler
	return route.invoker.Invoke(ctx, request)
}

func (c *CompositeInvoker) Close(ctx context.Context) error {
	var first error
	for _, child := range c.children {
		if closer, ok := child.(Closer); ok {
			if err := closer.Close(ctx); err != nil && first == nil {
				first = err
			}
		}
	}
	return first
}

func compositeHandler(index int, handler string) string {
	return strconv.Itoa(index) + ":" + handler
}

func sortedValues[T any](values map[string]T) []T {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]T, 0, len(keys))
	for _, key := range keys {
		out = append(out, values[key])
	}
	return out
}
