// Package extension defines the language-neutral boundary between Tempo and
// executable Universe mechanics.
package extension

import (
	"context"

	"github.com/aidanlsb/tempo-sdk-go/rng"
)

type ScopeKind string

const (
	ScopeUniverse ScopeKind = "universe"
	ScopeCampaign ScopeKind = "campaign"
	ScopeScenario ScopeKind = "scenario"
)

type Scope struct {
	Kind       ScopeKind `json:"kind,omitempty"`
	CampaignID string    `json:"campaign,omitempty"`
	ScenarioID string    `json:"scenario,omitempty"`
}

type Requirements struct {
	Abilities     []string         `json:"abilities,omitempty"`
	MinAttributes map[string]int64 `json:"min_attributes,omitempty"`
	// Knowledge lists Knowledge item IDs the acting Character must have
	// acquired for the Recipe to be available.
	Knowledge []string `json:"knowledge,omitempty"`
}

type Recipe struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Scope     Scope         `json:"scope,omitempty"`
	Requires  Requirements  `json:"requires,omitempty"`
	Arguments []FieldSchema `json:"arguments,omitempty"`
	Handler   string        `json:"handler,omitempty"`
	Disabled  bool          `json:"disabled,omitempty"`
}

type TriggerKind string

const (
	TriggerClock   TriggerKind = "clock"
	TriggerEvent   TriggerKind = "event"
	TriggerTurnEnd TriggerKind = "turn_end"
)

type Trigger struct {
	Kind      TriggerKind `json:"kind"`
	Every     int64       `json:"every,omitempty"`
	At        []int64     `json:"at,omitempty"`
	EventType string      `json:"event_type,omitempty"`
}

type System struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Scope   Scope   `json:"scope,omitempty"`
	Trigger Trigger `json:"trigger"`
	Handler string  `json:"handler"`
}

type PerceptionMode string

const (
	PerceptionEvent   PerceptionMode = "event"
	PerceptionCurrent PerceptionMode = "current"

	// PerceptionKindSuggestion marks optional current-state guidance for a
	// player. The Player Service exposes it separately from observations; it
	// never constrains Action availability.
	PerceptionKindSuggestion = "suggestion"
)

// PerceptionRule projects objective Action outcomes or current World state
// into Character-specific, player-safe Perceptions.
type PerceptionRule struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Scope     Scope          `json:"scope,omitempty"`
	Mode      PerceptionMode `json:"mode"`
	EventType string         `json:"event_type,omitempty"`
	Handler   string         `json:"handler"`
}

type ValueType string

const (
	ValueInteger ValueType = "integer"
	ValueBoolean ValueType = "boolean"
	ValueString  ValueType = "string"
	ValueEnum    ValueType = "enum"
	ValueEntity  ValueType = "entity"
)

type EntityKind string

const (
	EntityMap       EntityKind = "map"
	EntityLocation  EntityKind = "location"
	EntityPath      EntityKind = "path"
	EntityCharacter EntityKind = "character"
)

type FieldSchema struct {
	ID       string     `json:"id"`
	Type     ValueType  `json:"type"`
	Required bool       `json:"required,omitempty"`
	Min      *int64     `json:"min,omitempty"`
	Max      *int64     `json:"max,omitempty"`
	Values   []string   `json:"values,omitempty"`
	Target   EntityKind `json:"target,omitempty"`
}

type ComponentSchema struct {
	ID     string        `json:"id"`
	Scope  Scope         `json:"scope,omitempty"`
	Fields []FieldSchema `json:"fields,omitempty"`
}

type AbilitySchema struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CharacterSchema struct {
	Attributes []FieldSchema   `json:"attributes,omitempty"`
	Abilities  []AbilitySchema `json:"abilities,omitempty"`
}

type Descriptor struct {
	ID          string            `json:"id"`
	Character   *CharacterSchema  `json:"character,omitempty"`
	Components  []ComponentSchema `json:"components,omitempty"`
	Recipes     []Recipe          `json:"recipes,omitempty"`
	Systems     []System          `json:"systems,omitempty"`
	Perceptions []PerceptionRule  `json:"perceptions,omitempty"`
}

type HandlerKind string

const (
	HandlerRecipe     HandlerKind = "recipe"
	HandlerSystem     HandlerKind = "system"
	HandlerPerception HandlerKind = "perception"
)

type Character struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Attributes  map[string]any `json:"attributes,omitempty"`
	Abilities   []string       `json:"abilities,omitempty"`
}

type Map struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type Location struct {
	ID          string `json:"id"`
	Map         string `json:"map"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type Path struct {
	ID          string `json:"id"`
	A           string `json:"a"`
	B           string `json:"b"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type Presence struct {
	Character string `json:"character"`
	Location  string `json:"location"`
}

// Relationship is the read-only view of a directed Character → Character edge.
type Relationship struct {
	ID          string `json:"id"`
	From        string `json:"from"`
	To          string `json:"to"`
	Kind        string `json:"kind"`
	Description string `json:"description,omitempty"`
}

// Knowledge is the read-only view of an authored memorizable item.
type Knowledge struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
	Prose   string `json:"prose,omitempty"`
	Target  string `json:"target,omitempty"`
}

// CharacterKnowledge is the read-only view of an acquired-memory relation.
type CharacterKnowledge struct {
	ID        string `json:"id"`
	Character string `json:"character"`
	Knowledge string `json:"knowledge"`
	Source    string `json:"source,omitempty"`
	Sequence  uint64 `json:"sequence,omitempty"`
}

type WorldView struct {
	Maps               map[string]Map                       `json:"maps,omitempty"`
	Locations          map[string]Location                  `json:"locations,omitempty"`
	Paths              map[string]Path                      `json:"paths,omitempty"`
	Characters         map[string]Character                 `json:"characters,omitempty"`
	Presences          map[string]Presence                  `json:"presences,omitempty"`
	Relationships      map[string]Relationship              `json:"relationships,omitempty"`
	Knowledge          map[string]Knowledge                 `json:"knowledge,omitempty"`
	CharacterKnowledge map[string]CharacterKnowledge        `json:"character_knowledge,omitempty"`
	Components         map[string]map[string]map[string]any `json:"components,omitempty"`
}

type Event struct {
	Type string            `json:"type"`
	Data map[string]string `json:"data,omitempty"`
}

type TriggerContext struct {
	Kind  TriggerKind `json:"kind,omitempty"`
	Event *Event      `json:"event,omitempty"`
}

// ActionContext is the speculative Action envelope visible to PerceptionRules.
type ActionContext struct {
	ID        string            `json:"id,omitempty"`
	Actor     string            `json:"actor,omitempty"`
	Verb      string            `json:"verb"`
	Trigger   string            `json:"trigger"`
	Arguments map[string]string `json:"arguments,omitempty"`
}

// PerceptionContext selects event-time or current-state projection. EventIndex
// is the one-based index of Event within the speculative Action.
type PerceptionContext struct {
	Mode       PerceptionMode `json:"mode,omitempty"`
	Action     ActionContext  `json:"action,omitempty"`
	Event      *Event         `json:"event,omitempty"`
	EventIndex int            `json:"event_index,omitempty"`
}

type InvokeRequest struct {
	Kind          HandlerKind       `json:"kind"`
	Handler       string            `json:"handler"`
	Campaign      string            `json:"campaign"`
	Scenario      string            `json:"scenario"`
	Actor         string            `json:"actor,omitempty"`
	Arguments     map[string]string `json:"arguments,omitempty"`
	Clock         int64             `json:"clock"`
	Trigger       TriggerContext    `json:"trigger,omitempty"`
	Perception    PerceptionContext `json:"perception,omitempty"`
	PreviousWorld *WorldView        `json:"previous_world,omitempty"`
	World         WorldView         `json:"world"`
	AttributeMax  map[string]int64  `json:"attribute_max,omitempty"`
	// Seed is the fresh Action-level RNG seed Tempo mints for this invocation
	// (from OS entropy on live turns, or injected by tests). Handlers draw
	// randomness by keying a local recordable PRNG on this value via the shared
	// extension/rng package; the host requires no ambient WASM randomness. The
	// seed is a portable fixed-width uint64 transported as a JSON number.
	Seed uint64 `json:"seed,omitempty"`
}

type EffectKind string

const (
	EffectCharacterAttributeSet EffectKind = "character.attribute.set"
	EffectPresenceSet           EffectKind = "presence.set"
	EffectComponentSet          EffectKind = "component.set"
	EffectRelationshipSet       EffectKind = "relationship.set"
	EffectRelationshipRemove    EffectKind = "relationship.remove"
	EffectKnowledgeGrant        EffectKind = "knowledge.grant"
	EffectKnowledgeRevoke       EffectKind = "knowledge.revoke"
)

type Effect struct {
	Kind         EffectKind             `json:"kind"`
	Attribute    *CharacterAttributeSet `json:"attribute,omitempty"`
	Presence     *PresenceSet           `json:"presence,omitempty"`
	Component    *ComponentSet          `json:"component,omitempty"`
	Relationship *RelationshipSet       `json:"relationship,omitempty"`
	RemoveEdge   *RelationshipRemove    `json:"remove_relationship,omitempty"`
	Grant        *KnowledgeGrant        `json:"grant,omitempty"`
	Revoke       *KnowledgeRevoke       `json:"revoke,omitempty"`
}

type CharacterAttributeSet struct {
	Character string `json:"character"`
	Attribute string `json:"attribute"`
	Previous  any    `json:"previous,omitempty"`
	Value     any    `json:"value"`
}

type PresenceSet struct {
	Character string `json:"character"`
	From      string `json:"from,omitempty"`
	To        string `json:"to"`
}

type ComponentSet struct {
	Target    string         `json:"target"`
	Component string         `json:"component"`
	Value     map[string]any `json:"value"`
}

// RelationshipSet creates or replaces a directed edge. Scope is "campaign" or
// "scenario"; a Scenario-scoped edge requires IntroducedBy.
type RelationshipSet struct {
	ID           string `json:"id"`
	From         string `json:"from"`
	To           string `json:"to"`
	Kind         string `json:"kind"`
	Description  string `json:"description,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IntroducedBy string `json:"introduced_by,omitempty"`
}

// RelationshipRemove deletes a directed edge by stable id.
type RelationshipRemove struct {
	ID string `json:"id"`
}

// KnowledgeGrant records that a Character remembers a Knowledge item.
type KnowledgeGrant struct {
	Character string `json:"character"`
	Knowledge string `json:"knowledge"`
	Source    string `json:"source,omitempty"`
}

// KnowledgeRevoke removes a CharacterKnowledge relation.
type KnowledgeRevoke struct {
	Character string `json:"character"`
	Knowledge string `json:"knowledge"`
}

// Perception is a Character-specific, player-safe result. Action-time
// Perceptions are persisted; current-state Perceptions are recomputed.
type Perception struct {
	Character  string            `json:"character"`
	Kind       string            `json:"kind"`
	Content    string            `json:"content"`
	Data       map[string]string `json:"data,omitempty"`
	EventIndex int               `json:"event_index,omitempty"`
}

type InvokeResult struct {
	Effects     []Effect     `json:"effects,omitempty"`
	Events      []Event      `json:"events,omitempty"`
	Perceptions []Perception `json:"perceptions,omitempty"`
	// Rejected is an authored, player-safe explanation for declining a Recipe.
	// Hosts may return it to players; internal errors must use the invocation
	// error path instead.
	Rejected string `json:"rejected,omitempty"`
	// Draws is the ordered audit trail of RNG pulls the handler made from the
	// request Seed, in draw order. Tempo records it (with the Seed) on the
	// committed Action so a turn's randomness is reconstructable. Replay never
	// reads Draws to recompute state — it folds recorded Effects only.
	Draws []rng.Draw `json:"draws,omitempty"`
}

type Invoker interface {
	Describe(context.Context) (Descriptor, error)
	Invoke(context.Context, InvokeRequest) (InvokeResult, error)
}

type Closer interface {
	Close(context.Context) error
}
