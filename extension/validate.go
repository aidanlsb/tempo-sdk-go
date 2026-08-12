package extension

import (
	"encoding/json"
	"fmt"
	"math"
)

// EntityResolver reports whether a core entity of the given kind and id exists.
// It is supplied by callers of ValidateComponentValue so the extension package
// stays free of engine World types.
type EntityResolver func(kind EntityKind, id string) bool

// ValidateComponentValue checks a component instance's field values against a
// ComponentSchema: required fields present, no unknown fields, and typed field
// constraints (integer bounds, boolean, string, enum membership, entity
// resolution). Entity-typed fields resolve through exists.
func ValidateComponentValue(
	schema ComponentSchema,
	value map[string]any,
	exists EntityResolver,
) error {
	fields := make(map[string]FieldSchema, len(schema.Fields))
	for _, field := range schema.Fields {
		fields[field.ID] = field
	}
	for id, field := range fields {
		if field.Required {
			if _, ok := value[id]; !ok {
				return fmt.Errorf("component %q requires field %q", schema.ID, id)
			}
		}
	}
	for id, fieldValue := range value {
		field, ok := fields[id]
		if !ok {
			return fmt.Errorf("component %q has unknown field %q", schema.ID, id)
		}
		switch field.Type {
		case ValueInteger:
			integer, ok := coerceInt64(fieldValue)
			if !ok {
				return fmt.Errorf("component %q field %q must be integer", schema.ID, id)
			}
			if field.Min != nil && integer < *field.Min {
				return fmt.Errorf("component %q field %q is below minimum", schema.ID, id)
			}
			if field.Max != nil && integer > *field.Max {
				return fmt.Errorf("component %q field %q exceeds maximum", schema.ID, id)
			}
		case ValueBoolean:
			if _, ok := fieldValue.(bool); !ok {
				return fmt.Errorf("component %q field %q must be boolean", schema.ID, id)
			}
		case ValueString:
			if _, ok := fieldValue.(string); !ok {
				return fmt.Errorf("component %q field %q must be string", schema.ID, id)
			}
		case ValueEnum:
			stringValue, ok := fieldValue.(string)
			if !ok {
				return fmt.Errorf("component %q field %q must be enum string", schema.ID, id)
			}
			found := false
			for _, allowed := range field.Values {
				if stringValue == allowed {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("component %q field %q has invalid value", schema.ID, id)
			}
		case ValueEntity:
			entityID, ok := fieldValue.(string)
			if !ok || exists == nil || !exists(field.Target, entityID) {
				return fmt.Errorf(
					"component %q field %q does not resolve to %s",
					schema.ID,
					id,
					field.Target,
				)
			}
		}
	}
	return nil
}

func coerceInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	case float64:
		if math.Trunc(typed) == typed &&
			typed >= -9007199254740991.0 &&
			typed <= 9007199254740991.0 {
			return int64(typed), true
		}
		return 0, false
	case json.Number:
		integer, err := typed.Int64()
		return integer, err == nil
	default:
		return 0, false
	}
}

func ValidateDescriptor(descriptor Descriptor) error {
	if descriptor.ID == "" {
		return fmt.Errorf("descriptor id is required")
	}
	seenRecipes := make(map[string]struct{})
	for _, recipe := range descriptor.Recipes {
		if recipe.ID == "" || recipe.Name == "" {
			return fmt.Errorf("Recipe id and name are required")
		}
		if recipe.Disabled {
			if recipe.Handler != "" {
				return fmt.Errorf("disabled Recipe %q cannot define a handler", recipe.ID)
			}
		} else if recipe.Handler == "" {
			return fmt.Errorf("Recipe %q handler is required", recipe.ID)
		}
		if err := validateScope(recipe.Scope); err != nil {
			return fmt.Errorf("Recipe %q: %w", recipe.ID, err)
		}
		if err := validateFields(recipe.Arguments); err != nil {
			return fmt.Errorf("Recipe %q arguments: %w", recipe.ID, err)
		}
		key := scopedID(recipe.Scope, recipe.ID)
		if _, exists := seenRecipes[key]; exists {
			return fmt.Errorf("duplicate Recipe %q in one scope", recipe.ID)
		}
		seenRecipes[key] = struct{}{}
	}
	seenSystems := make(map[string]struct{})
	for _, system := range descriptor.Systems {
		if system.ID == "" || system.Name == "" || system.Handler == "" {
			return fmt.Errorf("System id, name, and handler are required")
		}
		if err := validateScope(system.Scope); err != nil {
			return fmt.Errorf("System %q: %w", system.ID, err)
		}
		switch system.Trigger.Kind {
		case TriggerClock:
			if system.Trigger.Every < 0 {
				return fmt.Errorf("System %q Clock interval cannot be negative", system.ID)
			}
			if system.Trigger.Every <= 0 && len(system.Trigger.At) == 0 {
				return fmt.Errorf("System %q Clock trigger is unscheduled", system.ID)
			}
		case TriggerEvent:
			if system.Trigger.EventType == "" {
				return fmt.Errorf("System %q Event trigger requires a type", system.ID)
			}
		case TriggerTurnEnd:
		default:
			return fmt.Errorf(
				"System %q has unsupported trigger %q",
				system.ID,
				system.Trigger.Kind,
			)
		}
		key := scopedID(system.Scope, system.ID)
		if _, exists := seenSystems[key]; exists {
			return fmt.Errorf("duplicate System %q in one scope", system.ID)
		}
		seenSystems[key] = struct{}{}
	}
	seenPerceptions := make(map[string]struct{})
	for _, rule := range descriptor.Perceptions {
		if rule.ID == "" || rule.Name == "" || rule.Handler == "" {
			return fmt.Errorf("PerceptionRule id, name, and handler are required")
		}
		if err := validateScope(rule.Scope); err != nil {
			return fmt.Errorf("PerceptionRule %q: %w", rule.ID, err)
		}
		switch rule.Mode {
		case PerceptionEvent:
			if rule.EventType == "" {
				return fmt.Errorf(
					"PerceptionRule %q event mode requires event_type",
					rule.ID,
				)
			}
		case PerceptionCurrent:
			if rule.EventType != "" {
				return fmt.Errorf(
					"PerceptionRule %q current mode cannot define event_type",
					rule.ID,
				)
			}
		default:
			return fmt.Errorf(
				"PerceptionRule %q has unsupported mode %q",
				rule.ID,
				rule.Mode,
			)
		}
		key := scopedID(rule.Scope, rule.ID)
		if _, exists := seenPerceptions[key]; exists {
			return fmt.Errorf("duplicate PerceptionRule %q in one scope", rule.ID)
		}
		seenPerceptions[key] = struct{}{}
	}
	seenComponents := make(map[string]struct{})
	for _, component := range descriptor.Components {
		if component.ID == "" {
			return fmt.Errorf("Component id is required")
		}
		if err := validateScope(component.Scope); err != nil {
			return fmt.Errorf("Component %q: %w", component.ID, err)
		}
		key := scopedID(component.Scope, component.ID)
		if _, exists := seenComponents[key]; exists {
			return fmt.Errorf("duplicate Component %q in one scope", component.ID)
		}
		seenComponents[key] = struct{}{}
		if err := validateFields(component.Fields); err != nil {
			return fmt.Errorf("Component %q: %w", component.ID, err)
		}
	}
	if descriptor.Character != nil {
		if err := validateFields(descriptor.Character.Attributes); err != nil {
			return fmt.Errorf("Character schema: %w", err)
		}
		seenAbilities := make(map[string]struct{}, len(descriptor.Character.Abilities))
		for _, ability := range descriptor.Character.Abilities {
			if ability.ID == "" || ability.Name == "" {
				return fmt.Errorf("Character ability id and name are required")
			}
			if _, exists := seenAbilities[ability.ID]; exists {
				return fmt.Errorf("duplicate Character ability %q", ability.ID)
			}
			seenAbilities[ability.ID] = struct{}{}
		}
	}
	return nil
}

func validateFields(fields []FieldSchema) error {
	seen := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		if field.ID == "" {
			return fmt.Errorf("field id is required")
		}
		if _, exists := seen[field.ID]; exists {
			return fmt.Errorf("duplicate field %q", field.ID)
		}
		seen[field.ID] = struct{}{}
		switch field.Type {
		case ValueInteger:
			if len(field.Values) != 0 || field.Target != "" {
				return fmt.Errorf("integer field %q cannot define values or target", field.ID)
			}
			if field.Min != nil && field.Max != nil && *field.Min > *field.Max {
				return fmt.Errorf("field %q minimum exceeds maximum", field.ID)
			}
		case ValueBoolean, ValueString:
			if field.Min != nil || field.Max != nil ||
				len(field.Values) != 0 || field.Target != "" {
				return fmt.Errorf(
					"%s field %q cannot define bounds or values",
					field.Type,
					field.ID,
				)
			}
		case ValueEnum:
			if field.Min != nil || field.Max != nil || field.Target != "" {
				return fmt.Errorf("enum field %q cannot define bounds or target", field.ID)
			}
			if len(field.Values) == 0 {
				return fmt.Errorf("enum field %q requires values", field.ID)
			}
			seenValues := make(map[string]struct{}, len(field.Values))
			for _, value := range field.Values {
				if value == "" {
					return fmt.Errorf("enum field %q has an empty value", field.ID)
				}
				if _, exists := seenValues[value]; exists {
					return fmt.Errorf(
						"enum field %q repeats value %q",
						field.ID,
						value,
					)
				}
				seenValues[value] = struct{}{}
			}
		case ValueEntity:
			if field.Min != nil || field.Max != nil || len(field.Values) != 0 {
				return fmt.Errorf("entity field %q cannot define bounds or values", field.ID)
			}
			switch field.Target {
			case EntityMap, EntityLocation, EntityPath, EntityCharacter:
			default:
				return fmt.Errorf("entity field %q requires a supported target", field.ID)
			}
		default:
			return fmt.Errorf("field %q has unsupported type %q", field.ID, field.Type)
		}
	}
	return nil
}

func validateScope(scope Scope) error {
	switch scope.Kind {
	case "", ScopeUniverse:
		if scope.CampaignID != "" || scope.ScenarioID != "" {
			return fmt.Errorf("Universe scope cannot name a Campaign or Scenario")
		}
	case ScopeCampaign:
		if scope.CampaignID == "" || scope.ScenarioID != "" {
			return fmt.Errorf("Campaign scope requires only a Campaign")
		}
	case ScopeScenario:
		if scope.CampaignID == "" || scope.ScenarioID == "" {
			return fmt.Errorf("Scenario scope requires a Campaign and Scenario")
		}
	default:
		return fmt.Errorf("unsupported scope %q", scope.Kind)
	}
	return nil
}

func scopedID(scope Scope, id string) string {
	kind := scope.Kind
	if kind == "" {
		kind = ScopeUniverse
	}
	return string(kind) + "\x00" + scope.CampaignID + "\x00" +
		scope.ScenarioID + "\x00" + id
}
