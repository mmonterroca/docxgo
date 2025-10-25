/*
   Copyright (c) 2025 SlideLang

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package manager

import (
	"sync"

	"github.com/SlideLang/go-docx/v2/pkg/constants"
	"github.com/SlideLang/go-docx/v2/pkg/errors"
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

	// Default to Internal if not specified
	if targetMode == "" {
		targetMode = "Internal"
	}

	rel := &Relationship{
		ID:         id,
		Type:       relType,
		Target:     target,
		TargetMode: targetMode,
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
