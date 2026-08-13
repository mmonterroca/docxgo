// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

// Package manager provides internal management services for docxgo v2.
package manager

import (
	"fmt"
	"sync/atomic"

	"github.com/mmonterroca/docxgo/v2/pkg/constants"
)

// IDGenerator generates unique IDs for document elements.
// It is thread-safe and can be used concurrently.
type IDGenerator struct {
	paragraphCounter atomic.Uint64
	runCounter       atomic.Uint64
	tableCounter     atomic.Uint64
	rowCounter       atomic.Uint64
	cellCounter      atomic.Uint64
	imageCounter     atomic.Uint64
	shapeCounter     atomic.Uint64
	relCounter       atomic.Uint64
	commentCounter   atomic.Uint64
	footnoteCounter  atomic.Uint64
	endnoteCounter   atomic.Uint64
}

// NewIDGenerator creates a new ID generator.
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

// NextParagraphID generates the next paragraph ID.
func (g *IDGenerator) NextParagraphID() string {
	id := g.paragraphCounter.Add(1)
	return fmt.Sprintf("%s%d", constants.IDPrefixParagraph, id)
}

// NextRunID generates the next run ID.
func (g *IDGenerator) NextRunID() string {
	id := g.runCounter.Add(1)
	return fmt.Sprintf("%s%d", constants.IDPrefixRun, id)
}

// NextTableID generates the next table ID.
func (g *IDGenerator) NextTableID() string {
	id := g.tableCounter.Add(1)
	return fmt.Sprintf("%s%d", constants.IDPrefixTable, id)
}

// NextRowID generates the next row ID.
func (g *IDGenerator) NextRowID() string {
	id := g.rowCounter.Add(1)
	return fmt.Sprintf("%s%d", constants.IDPrefixRow, id)
}

// NextCellID generates the next cell ID.
func (g *IDGenerator) NextCellID() string {
	id := g.cellCounter.Add(1)
	return fmt.Sprintf("%s%d", constants.IDPrefixCell, id)
}

// NextImageID generates the next image ID.
func (g *IDGenerator) NextImageID() string {
	id := g.imageCounter.Add(1)
	return fmt.Sprintf("%s%d", constants.IDPrefixImage, id)
}

// NextShapeID generates the next shape ID.
func (g *IDGenerator) NextShapeID() string {
	id := g.shapeCounter.Add(1)
	return fmt.Sprintf("%s%d", constants.IDPrefixShape, id)
}

// NextRelID generates the next relationship ID.
func (g *IDGenerator) NextRelID() string {
	id := g.relCounter.Add(1)
	return fmt.Sprintf("%s%d", constants.IDPrefixRel, id)
}

// NextCommentID generates the next comment ID.
func (g *IDGenerator) NextCommentID() string {
	id := g.commentCounter.Add(1)
	return fmt.Sprintf("%s%d", constants.IDPrefixComment, id)
}

// NextFootnoteID generates the next footnote ID.
func (g *IDGenerator) NextFootnoteID() string {
	id := g.footnoteCounter.Add(1)
	return fmt.Sprintf("%s%d", constants.IDPrefixFootnote, id)
}

// NextEndnoteID generates the next endnote ID.
func (g *IDGenerator) NextEndnoteID() string {
	id := g.endnoteCounter.Add(1)
	return fmt.Sprintf("%s%d", constants.IDPrefixEndnote, id)
}

// GenerateID generates an ID with a custom prefix.
// This is a generic method for any element type.
func (g *IDGenerator) GenerateID(prefix string) string {
	// Use table counter for generic IDs
	id := g.tableCounter.Add(1)
	return fmt.Sprintf("%s%d", prefix, id)
}

// Reset resets all counters to zero.
// This should only be used when starting a new document.
func (g *IDGenerator) Reset() {
	g.paragraphCounter.Store(0)
	g.runCounter.Store(0)
	g.tableCounter.Store(0)
	g.rowCounter.Store(0)
	g.cellCounter.Store(0)
	g.imageCounter.Store(0)
	g.shapeCounter.Store(0)
	g.relCounter.Store(0)
	g.commentCounter.Store(0)
	g.footnoteCounter.Store(0)
	g.endnoteCounter.Store(0)
}

// EnsureRelCounterAtLeast ensures the relationship counter is at least the provided value.
func (g *IDGenerator) EnsureRelCounterAtLeast(value uint64) {
	for {
		current := g.relCounter.Load()
		if current >= value {
			return
		}
		if g.relCounter.CompareAndSwap(current, value) {
			return
		}
	}
}
