package extension

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestContractRoundTrip(t *testing.T) {
	descriptor := Descriptor{
		ID: "test",
		Components: []ComponentSchema{{
			ID:    "test.state",
			Scope: Scope{Kind: ScopeChapter, CampaignID: "story", ChapterID: "opening"},
			Fields: []FieldSchema{{
				ID:   "active",
				Type: ValueBoolean,
			}},
		}},
		Relationships: []RelationshipSchema{{
			Kind: "suspicion",
			Fields: []FieldSchema{
				{ID: "level", Type: ValueInteger, Required: true},
				{ID: "confirmed", Type: ValueBoolean},
			},
		}},
		Recipes: []Recipe{{
			ID:      "test.perform",
			Name:    "Perform",
			Scope:   Scope{Kind: ScopeChapter, CampaignID: "story", ChapterID: "opening"},
			Handler: "handler:test.perform",
			Requires: Requirements{
				Locations: []string{"observatory"},
				Components: []ComponentRequirement{{
					CurrentLocation: true,
					Component:       "test.state",
					Values:          map[string]any{"active": true},
				}},
			},
		}},
		Systems: []System{{
			ID:      "test.tick",
			Name:    "Tick",
			Scope:   Scope{Kind: ScopeCampaign, CampaignID: "story"},
			Trigger: Trigger{Kind: TriggerTurnEnd},
			Handler: "handler:test.tick",
		}},
		Perceptions: []PerceptionRule{{
			ID:        "test.see-tick",
			Name:      "See Tick",
			Scope:     Scope{Kind: ScopeCampaign, CampaignID: "story"},
			Mode:      PerceptionEvent,
			EventType: "test.ticked",
			Handler:   "handler:test.see-tick",
		}},
		Resolvers: []DecisionResolver{{
			ID: "test.resolve", Name: "Resolve", Handler: "handler:test.resolve",
		}},
	}
	data, err := json.Marshal(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Descriptor
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, descriptor) {
		t.Fatalf("round trip = %#v, want %#v", decoded, descriptor)
	}
	if err := ValidateDescriptor(decoded); err != nil {
		t.Fatalf("ValidateDescriptor() error = %v", err)
	}
}

func TestDecisionInvocationRoundTrip(t *testing.T) {
	request := InvokeRequest{
		Kind: HandlerDecisionResolver, Handler: "resolve", Clock: 42,
		Decision: &DecisionWindowContext{
			ID: "exchange/1", Flow: "combat", Resolver: "test.resolve",
			Turn: "run/turn/1", Generation: 2, WorldRevision: 4,
			EligibleActors: []string{"a", "b"},
			AllowedRecipes: map[string][]string{
				"a": {"test.attack"}, "b": {"test.guard"},
			},
			CompletionPolicy: CompletionAllRequired,
			DisclosurePolicy: DisclosureSealedUntilResolution,
			Decisions: []Decision{
				{Actor: "a", Recipe: "test.attack", RequestID: "one"},
				{Actor: "b", Recipe: "test.guard", RequestID: "two"},
			},
		},
	}
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	var decoded InvokeRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, request) {
		t.Fatalf("round trip = %#v, want %#v", decoded, request)
	}

	result := InvokeResult{Transition: &DecisionTransition{
		Next: &DecisionWindowSpec{
			ID: "exchange/2", Resolver: "test.resolve",
			EligibleActors:   []string{"b"},
			AllowedRecipes:   map[string][]string{"b": {"test.guard"}},
			CompletionPolicy: CompletionAllRequired,
			DisclosurePolicy: DisclosureOnAcceptance,
		},
	}}
	data, err = json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var decodedResult InvokeResult
	if err := json.Unmarshal(data, &decodedResult); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decodedResult, result) {
		t.Fatalf("result round trip = %#v, want %#v", decodedResult, result)
	}
}

func TestValidateDescriptorRejectsInvalidDecisionResolver(t *testing.T) {
	err := ValidateDescriptor(Descriptor{
		ID:        "test",
		Resolvers: []DecisionResolver{{ID: "resolve", Name: "Resolve"}},
	})
	if err == nil {
		t.Fatal("ValidateDescriptor() accepted a resolver without a handler")
	}
}

func TestValidateDescriptorRejectsInvalidComponentRequirement(t *testing.T) {
	err := ValidateDescriptor(Descriptor{
		ID: "test",
		Recipes: []Recipe{{
			ID: "test.perform", Name: "Perform", Handler: "perform",
			Requires: Requirements{Components: []ComponentRequirement{{
				Target: "location:a", CurrentLocation: true, Component: "test.state",
			}}},
		}},
	})
	if err == nil {
		t.Fatal("ValidateDescriptor() accepted an ambiguous component requirement")
	}
}

func TestValidateDescriptorRejectsInvalidPerceptionRule(t *testing.T) {
	err := ValidateDescriptor(Descriptor{
		ID: "test",
		Perceptions: []PerceptionRule{{
			ID:        "current",
			Name:      "Current",
			Mode:      PerceptionCurrent,
			EventType: "not-allowed",
			Handler:   "current",
		}},
	})
	if err == nil {
		t.Fatal("ValidateDescriptor() accepted event type on current PerceptionRule")
	}
}

func TestValidateDescriptorRejectsUnscopedChapter(t *testing.T) {
	err := ValidateDescriptor(Descriptor{
		ID: "test",
		Systems: []System{{
			ID:      "tick",
			Name:    "Tick",
			Scope:   Scope{Kind: ScopeChapter, ChapterID: "opening"},
			Trigger: Trigger{Kind: TriggerTurnEnd},
			Handler: "tick",
		}},
	})
	if err == nil {
		t.Fatal("ValidateDescriptor() accepted Chapter scope without Campaign")
	}
}

func TestValidateDescriptorNormalizesUniverseScopeForDuplicates(t *testing.T) {
	err := ValidateDescriptor(Descriptor{
		ID: "test",
		Recipes: []Recipe{
			{ID: "same", Name: "Same", Handler: "one"},
			{
				ID:      "same",
				Name:    "Same",
				Scope:   Scope{Kind: ScopeUniverse},
				Handler: "two",
			},
		},
	})
	if err == nil {
		t.Fatal("ValidateDescriptor() accepted duplicate Universe Recipe")
	}
}

func TestValidateDescriptorRejectsContradictoryFieldSchema(t *testing.T) {
	err := ValidateDescriptor(Descriptor{
		ID: "test",
		Components: []ComponentSchema{{
			ID: "test.state",
			Fields: []FieldSchema{{
				ID:     "active",
				Type:   ValueBoolean,
				Values: []string{"yes", "no"},
			}},
		}},
	})
	if err == nil {
		t.Fatal("ValidateDescriptor() accepted values on a boolean field")
	}
}

func TestValidateRelationshipValue(t *testing.T) {
	minimum, maximum := int64(0), int64(100)
	schema := RelationshipSchema{
		Kind: "suspicion",
		Fields: []FieldSchema{
			{ID: "level", Type: ValueInteger, Required: true, Min: &minimum, Max: &maximum},
			{ID: "confirmed", Type: ValueBoolean},
			{ID: "basis", Type: ValueEnum, Values: []string{"instinct", "evidence"}},
		},
	}
	if err := ValidateRelationshipValue(schema, map[string]any{
		"level": int64(72), "confirmed": false, "basis": "evidence",
	}, nil); err != nil {
		t.Fatalf("ValidateRelationshipValue() error = %v", err)
	}
	for name, values := range map[string]map[string]any{
		"missing required": {"confirmed": false},
		"above maximum":    {"level": int64(101)},
		"invalid enum":     {"level": int64(50), "basis": "rumor"},
		"unknown field":    {"level": int64(50), "motive": "unknown"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := ValidateRelationshipValue(schema, values, nil); err == nil {
				t.Fatal("ValidateRelationshipValue() accepted invalid values")
			}
		})
	}
}

func TestValidateDescriptorRejectsDuplicateRelationshipKind(t *testing.T) {
	err := ValidateDescriptor(Descriptor{
		ID: "test",
		Relationships: []RelationshipSchema{
			{Kind: "suspicion"},
			{Kind: "suspicion"},
		},
	})
	if err == nil {
		t.Fatal("ValidateDescriptor() accepted duplicate Relationship schema")
	}
}

func TestValidateDescriptorRejectsNegativeClockInterval(t *testing.T) {
	err := ValidateDescriptor(Descriptor{
		ID: "test",
		Systems: []System{{
			ID: "tick", Name: "Tick", Handler: "tick",
			Trigger: Trigger{Kind: TriggerClock, Every: -1, At: []int64{10}},
		}},
	})
	if err == nil {
		t.Fatal("ValidateDescriptor() accepted a negative Clock interval")
	}
}

func TestValidateDescriptorRejectsDuplicateAbilities(t *testing.T) {
	err := ValidateDescriptor(Descriptor{
		ID: "test",
		Character: &CharacterSchema{Abilities: []AbilitySchema{
			{ID: "focus", Name: "Focus"},
			{ID: "focus", Name: "Again"},
		}},
	})
	if err == nil {
		t.Fatal("ValidateDescriptor() accepted duplicate Character abilities")
	}
}
