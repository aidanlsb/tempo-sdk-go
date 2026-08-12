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
			Scope: Scope{Kind: ScopeScenario, CampaignID: "story", ScenarioID: "opening"},
			Fields: []FieldSchema{{
				ID:   "active",
				Type: ValueBoolean,
			}},
		}},
		Recipes: []Recipe{{
			ID:      "test.perform",
			Name:    "Perform",
			Scope:   Scope{Kind: ScopeScenario, CampaignID: "story", ScenarioID: "opening"},
			Handler: "handler:test.perform",
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

func TestValidateDescriptorRejectsUnscopedScenario(t *testing.T) {
	err := ValidateDescriptor(Descriptor{
		ID: "test",
		Systems: []System{{
			ID:      "tick",
			Name:    "Tick",
			Scope:   Scope{Kind: ScopeScenario, ScenarioID: "opening"},
			Trigger: Trigger{Kind: TriggerTurnEnd},
			Handler: "tick",
		}},
	})
	if err == nil {
		t.Fatal("ValidateDescriptor() accepted Scenario scope without Campaign")
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
