// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package manager

import (
	"strconv"
	"strings"
	"sync"

	"github.com/mmonterroca/docxgo/v2/internal/xml"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
	"github.com/mmonterroca/docxgo/v2/pkg/errors"
)

// Relationship represents an OOXML relationship.
type Relationship struct {
	ID         string // Relationship ID (e.g., "rId1")
	Type       string // Relationship type (e.g., "http://schemas.openxmlformats.org/.../image")
	Target     string // Target path (e.g., "media/image1.png")
	TargetMode string // "Internal" or "External"
}

// RelationshipManager manages relationships for a document part.
// It is thread-safe.
type RelationshipManager struct {
	mu            sync.RWMutex
	relationships map[string]*Relationship // key is ID
	idGen         *IDGenerator
}

// NewRelationshipManager creates a new relationship manager.
func NewRelationshipManager(idGen *IDGenerator) *RelationshipManager {
	return &RelationshipManager{
		relationships: make(map[string]*Relationship, constants.DefaultRelCapacity),
		idGen:         idGen,
	}
}

// Add adds a new relationship and returns its ID.
func (rm *RelationshipManager) Add(relType, target, targetMode string) (string, error) {
	if relType == "" {
		return "", errors.InvalidArgument("RelationshipManager.Add", "relType", relType, "relationship type cannot be empty")
	}
	if target == "" {
		return "", errors.InvalidArgument("RelationshipManager.Add", "target", target, "target cannot be empty")
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Generate new ID
	id := rm.idGen.NextRelID()

	// Only "External" requires TargetMode attribute. For internal
	// relationships Word expects the attribute to be omitted entirely.
	mode := strings.TrimSpace(targetMode)
	if strings.EqualFold(mode, "internal") || mode == "" {
		mode = ""
	}

	rel := &Relationship{
		ID:         id,
		Type:       relType,
		Target:     target,
		TargetMode: mode,
	}

	rm.relationships[id] = rel
	return id, nil
}

// Get retrieves a relationship by ID.
func (rm *RelationshipManager) Get(id string) (*Relationship, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rel, exists := rm.relationships[id]
	if !exists {
		return nil, errors.NotFound("RelationshipManager.Get", "relationship")
	}

	return rel, nil
}

// GetByTarget retrieves a relationship by target path.
// Returns the first relationship matching the target.
func (rm *RelationshipManager) GetByTarget(target string) (*Relationship, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for _, rel := range rm.relationships {
		if rel.Target == target {
			return rel, nil
		}
	}

	return nil, errors.NotFound("RelationshipManager.GetByTarget", "relationship")
}

// All returns all relationships.
func (rm *RelationshipManager) All() []*Relationship {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rels := make([]*Relationship, 0, len(rm.relationships))
	for _, rel := range rm.relationships {
		rels = append(rels, rel)
	}

	return rels
}

// Count returns the number of relationships.
func (rm *RelationshipManager) Count() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return len(rm.relationships)
}

// Delete removes a relationship by ID.
func (rm *RelationshipManager) Delete(id string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.relationships[id]; !exists {
		return errors.NotFound("RelationshipManager.Delete", "relationship")
	}

	delete(rm.relationships, id)
	return nil
}

// AddImage adds an image relationship.
func (rm *RelationshipManager) AddImage(target string) (string, error) {
	return rm.Add(constants.RelTypeImage, target, "Internal")
}

// AddHyperlink adds a hyperlink relationship.
func (rm *RelationshipManager) AddHyperlink(target string) (string, error) {
	return rm.Add(constants.RelTypeHyperlink, target, "External")
}

// AddHeader adds a header relationship.
func (rm *RelationshipManager) AddHeader(target string) (string, error) {
	return rm.Add(constants.RelTypeHeader, target, "Internal")
}

// AddFooter adds a footer relationship.
func (rm *RelationshipManager) AddFooter(target string) (string, error) {
	return rm.Add(constants.RelTypeFooter, target, "Internal")
}

// RegisterExisting adds an existing relationship into the manager without generating a new ID.
func (rm *RelationshipManager) RegisterExisting(id, relType, target, targetMode string) error {
	if id == "" {
		return errors.InvalidArgument("RelationshipManager.RegisterExisting", "id", id, "relationship id cannot be empty")
	}
	if relType == "" {
		return errors.InvalidArgument("RelationshipManager.RegisterExisting", "relType", relType, "relationship type cannot be empty")
	}
	if target == "" {
		return errors.InvalidArgument("RelationshipManager.RegisterExisting", "target", target, "relationship target cannot be empty")
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.relationships[id]; exists {
		return nil
	}

	mode := strings.TrimSpace(targetMode)
	if strings.EqualFold(mode, "internal") || mode == "" {
		mode = ""
	}

	rm.relationships[id] = &Relationship{
		ID:         id,
		Type:       relType,
		Target:     target,
		TargetMode: mode,
	}

	// Update ID generator counter so future relationships avoid collisions.
	numeric := strings.TrimPrefix(strings.ToLower(id), strings.ToLower(constants.IDPrefixRel))
	if n, err := strconv.Atoi(numeric); err == nil {
		rm.idGen.EnsureRelCounterAtLeast(uint64(n))
	}

	return nil
}

// ToXML converts relationships to XML structure.
func (rm *RelationshipManager) ToXML() *xml.Relationships {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rels := &xml.Relationships{
		Xmlns:         constants.NamespacePackageRels,
		Relationships: make([]*xml.Relationship, 0, len(rm.relationships)),
	}

	for _, rel := range rm.relationships {
		xmlRel := &xml.Relationship{
			ID:         rel.ID,
			Type:       rel.Type,
			Target:     rel.Target,
			TargetMode: rel.TargetMode,
		}
		rels.Relationships = append(rels.Relationships, xmlRel)
	}

	return rels
}
