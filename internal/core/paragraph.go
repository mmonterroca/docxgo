// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package core

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/internal/manager"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
	"github.com/mmonterroca/docxgo/v2/pkg/errors"
)

// IDGenerator interface for testing purposes
type IDGenerator interface {
	NextParagraphID() string
	NextRunID() string
	NextImageID() string
	GenerateID(prefix string) string
}

// paragraph implements the domain.Paragraph interface.
type paragraph struct {
	id            string
	runs          []domain.Run
	fields        []domain.Field
	images        []domain.Image
	styleName     string
	alignment     domain.Alignment
	indent        domain.Indentation
	spacingBefore int
	spacingAfter  int
	lineSpacing   domain.LineSpacing
	// *Set track whether the caller ever called the corresponding setter, so
	// the serializer can tell "never set" from "explicitly set to a value
	// that happens to equal the default" — most importantly, an explicit 0
	// on a paragraph whose style supplies a non-zero spacing. The domain
	// getters (SpacingBefore, SpacingAfter, LineSpacing) can't express this
	// distinction themselves without changing their public return types, so
	// it's exposed on the concrete type only, the same pattern StyleName()
	// uses for the serializer's style-name access.
	spacingBeforeSet bool
	spacingAfterSet  bool
	lineSpacingSet   bool
	// indent*Set mirror the spacing flags above, but per side rather than per
	// call: SetIndent takes the whole domain.Indentation struct at once, so a
	// single "was SetIndent ever called" flag couldn't tell "Left was set"
	// from "Right was set" -- which the reader needs, since it must mark only
	// the sides actually present in a source <w:ind> element (see
	// SetIndentLeft and friends below).
	indentLeftSet      bool
	indentRightSet     bool
	indentFirstLineSet bool
	indentHangingSet   bool
	numbering          *domain.NumberingReference
	borders            domain.ParagraphBorders
	idGen              IDGenerator
	relManager         *manager.RelationshipManager
	bookmarkID         string // ID for bookmark (if this paragraph needs one for TOC)
	bookmarkName       string // Name for bookmark (e.g., "_Toc123456")
	mediaManager       *manager.MediaManager
	// inHeaderFooter is true for a paragraph that lives in a header or footer
	// part. AddHyperlink refuses to run on such a paragraph -- see its doc
	// comment for why.
	inHeaderFooter bool
}

// NewParagraph creates a new Paragraph.
func NewParagraph(id string, idGen IDGenerator, relManager *manager.RelationshipManager, mediaManager *manager.MediaManager) domain.Paragraph {
	return &paragraph{
		id:            id,
		runs:          make([]domain.Run, 0, constants.DefaultRunCapacity),
		fields:        make([]domain.Field, 0, 4),
		images:        make([]domain.Image, 0, 4),
		alignment:     domain.AlignmentLeft,
		indent:        domain.Indentation{},
		spacingBefore: constants.DefaultParagraphSpacing,
		spacingAfter:  constants.DefaultParagraphSpacing,
		lineSpacing:   domain.LineSpacing{Rule: domain.LineSpacingAuto, Value: constants.DefaultLineSpacing},
		idGen:         idGen,
		relManager:    relManager,
		mediaManager:  mediaManager,
	}
}

// markHeaderFooterParagraph flags a paragraph as living inside a header or
// footer part, so AddHyperlink can refuse to run on it. p is always a
// *paragraph in practice -- NewParagraph is the only constructor -- so a
// failed type assertion is silently ignored rather than panicking.
func markHeaderFooterParagraph(p domain.Paragraph) {
	if concrete, ok := p.(*paragraph); ok {
		concrete.inHeaderFooter = true
	}
}

// markHeaderFooterTable flags a table as living inside a header or footer
// part, so its cells' paragraphs inherit the same AddHyperlink restriction as
// markHeaderFooterParagraph. t is always a *table in practice -- NewTable is
// the only constructor -- so a failed type assertion is silently ignored
// rather than panicking.
func markHeaderFooterTable(t domain.Table) {
	if concrete, ok := t.(*table); ok {
		concrete.inHeaderFooter = true
	}
}

// AddRun adds a new text run to the paragraph.
func (p *paragraph) AddRun() (domain.Run, error) {
	id := p.idGen.NextRunID()
	run := NewRun(id, p.relManager)
	p.runs = append(p.runs, run)
	return run, nil
}

// AddField adds a field to the paragraph.
//
// Deprecated: Use AddRun() and run.AddField() instead for better control.
func (p *paragraph) AddField(_ domain.FieldType) (domain.Field, error) {
	return nil, errors.Unsupported("Paragraph.AddField", "use AddRun() and run.AddField() instead")
}

// AddHyperlink adds a hyperlink to the paragraph.
//
// The link is written as a real OOXML <w:hyperlink> element: for an external
// URL, via a relationship in this part's .rels (r:id); for an internal link
// (a url starting with "#"), via a bookmark anchor (w:anchor) with no
// relationship at all. Do not call this on a header or footer paragraph:
// docxgo does not yet emit a per-part relationships file
// (word/_rels/headerN.xml.rels / footerN.xml.rels), so an external hyperlink
// relationship minted there would reference a part that doesn't exist.
func (p *paragraph) AddHyperlink(url, displayText string) (domain.Run, error) {
	if url == "" {
		return nil, errors.InvalidArgument("Paragraph.AddHyperlink", "url", url, "URL cannot be empty")
	}

	if p.inHeaderFooter {
		return nil, errors.Unsupported("Paragraph.AddHyperlink",
			"hyperlinks in a header or footer paragraph (no word/_rels/headerN.xml.rels support yet)")
	}

	text := displayText
	if text == "" {
		text = url
	}

	// Build (and validate) the field before touching the paragraph at all.
	// NewHyperlinkField rejects a url containing a double quote by recording
	// a validation error on the field rather than returning nil -- checking
	// it here, before AddRun, means a rejected url leaves the paragraph
	// exactly as it was instead of a stray blue, underlined run behind a
	// returned error.
	field := NewHyperlinkField(url, text)
	if validator, ok := field.(interface{ ValidationError() error }); ok {
		if err := validator.ValidationError(); err != nil {
			return nil, errors.Wrap(err, "Paragraph.AddHyperlink")
		}
	}

	// Create run with hyperlink text
	run, err := p.AddRun()
	if err != nil {
		return nil, errors.Wrap(err, "Paragraph.AddHyperlink")
	}

	err = run.SetText(text)
	if err != nil {
		return nil, errors.Wrap(err, "Paragraph.AddHyperlink")
	}

	// Set hyperlink styling (blue, underlined) -- redundant with the
	// Hyperlink character style the serializer also applies via the field
	// below, but kept so the run renders correctly for any consumer that
	// doesn't resolve character styles.
	_ = run.SetColor(domain.ColorBlue)
	_ = run.SetUnderline(domain.UnderlineSingle)

	// Attach the hyperlink field: this is what makes the serializer emit a
	// real <w:hyperlink> wrapper (see expandRunWithFields in
	// internal/serializer/serializer.go) instead of a plain styled run. It
	// also mints (or, for a "#anchor" url, skips) the relationship -- see
	// run.AddField. The validation check above already ruled out the only
	// failure mode reachable with a non-empty url and an initialized
	// relationship manager, so this should not fail in practice; still
	// wrapped in case a future field or run change adds one.
	if err := run.AddField(field); err != nil {
		return nil, errors.Wrap(err, "Paragraph.AddHyperlink")
	}

	return run, nil
}

// AddImage adds an image to the paragraph from a file path.
func (p *paragraph) AddImage(path string) (domain.Image, error) {
	id := p.idGen.NextImageID()
	img, err := NewImage(id, path)
	if err != nil {
		return nil, errors.Wrap(err, "Paragraph.AddImage")
	}

	if err := p.attachImage(img, filepath.Base(path)); err != nil {
		return nil, err
	}

	return img, nil
}

// AddImageWithSize adds an image with custom dimensions.
func (p *paragraph) AddImageWithSize(path string, size domain.ImageSize) (domain.Image, error) {
	id := p.idGen.NextImageID()
	img, err := NewImageWithSize(id, path, size)
	if err != nil {
		return nil, errors.Wrap(err, "Paragraph.AddImageWithSize")
	}

	if err := p.attachImage(img, filepath.Base(path)); err != nil {
		return nil, err
	}

	return img, nil
}

// AddImageWithPosition adds an image with custom positioning.
func (p *paragraph) AddImageWithPosition(path string, size domain.ImageSize, pos domain.ImagePosition) (domain.Image, error) {
	id := p.idGen.NextImageID()
	img, err := NewImageWithPosition(id, path, size, pos)
	if err != nil {
		return nil, errors.Wrap(err, "Paragraph.AddImageWithPosition")
	}

	if err := p.attachImage(img, filepath.Base(path)); err != nil {
		return nil, err
	}

	return img, nil
}

// AddImageFromBytes adds an image from raw byte data.
func (p *paragraph) AddImageFromBytes(data []byte, format domain.ImageFormat) (domain.Image, error) {
	id := p.idGen.NextImageID()
	img, err := NewImageFromBytes(id, data, format)
	if err != nil {
		return nil, errors.Wrap(err, "Paragraph.AddImageFromBytes")
	}

	if err := p.attachImage(img, fmt.Sprintf("image%s.%s", id, img.Format())); err != nil {
		return nil, err
	}

	return img, nil
}

// AddImageFromBytesWithSize adds an image from byte data with custom dimensions.
func (p *paragraph) AddImageFromBytesWithSize(data []byte, format domain.ImageFormat, size domain.ImageSize) (domain.Image, error) {
	id := p.idGen.NextImageID()
	img, err := NewImageFromBytesWithSize(id, data, format, size)
	if err != nil {
		return nil, errors.Wrap(err, "Paragraph.AddImageFromBytesWithSize")
	}

	if err := p.attachImage(img, fmt.Sprintf("image%s.%s", id, img.Format())); err != nil {
		return nil, err
	}

	return img, nil
}

// AddImageFromBytesWithPosition adds an image from byte data with custom positioning.
func (p *paragraph) AddImageFromBytesWithPosition(data []byte, format domain.ImageFormat, size domain.ImageSize, pos domain.ImagePosition) (domain.Image, error) {
	id := p.idGen.NextImageID()
	img, err := NewImageFromBytesWithPosition(id, data, format, size, pos)
	if err != nil {
		return nil, errors.Wrap(err, "Paragraph.AddImageFromBytesWithPosition")
	}

	if err := p.attachImage(img, fmt.Sprintf("image%s.%s", id, img.Format())); err != nil {
		return nil, err
	}

	return img, nil
}

// attachImage registers the image with the media and relationship managers and appends it as a drawing run.
func (p *paragraph) attachImage(img domain.Image, sourceName string) error {
	if p.mediaManager == nil {
		return errors.InvalidState("Paragraph.attachImage", "media manager not initialized")
	}

	if sourceName == "" {
		sourceName = img.ID()
	}

	_, mediaPath, err := p.mediaManager.Add(img.Data(), sourceName)
	if err != nil {
		return errors.Wrap(err, "Paragraph.attachImage")
	}

	relativePath := strings.TrimPrefix(mediaPath, "word/")
	if docxImg, ok := img.(*docxImage); ok {
		docxImg.setTarget(relativePath)
	}

	relID, err := p.relManager.AddImage(relativePath)
	if err != nil {
		return errors.Wrap(err, "Paragraph.attachImage")
	}

	if docxImg, ok := img.(*docxImage); ok {
		docxImg.SetRelationshipID(relID)
	}

	run := NewRun(p.idGen.NextRunID(), p.relManager)
	if setter, ok := run.(interface{ setImage(domain.Image) }); ok {
		setter.setImage(img)
	}

	p.runs = append(p.runs, run)
	p.images = append(p.images, img)
	return nil
}

// RegisterHydratedImage records an image that was rehydrated from an existing document.
// It preserves media metadata so the image can be written back without renaming.
func (p *paragraph) RegisterHydratedImage(img domain.Image, mediaPath, contentType string, data []byte) error {
	if img == nil {
		return errors.InvalidArgument("Paragraph.RegisterHydratedImage", "image", nil, "image cannot be nil")
	}
	if mediaPath == "" {
		return errors.InvalidArgument("Paragraph.RegisterHydratedImage", "mediaPath", mediaPath, "media path cannot be empty")
	}
	if len(data) == 0 {
		return errors.InvalidArgument("Paragraph.RegisterHydratedImage", "data", data, "image data cannot be empty")
	}

	if p.mediaManager != nil {
		if _, err := p.mediaManager.RegisterExisting(img.ID(), mediaPath, contentType, data); err != nil {
			return errors.Wrap(err, "Paragraph.RegisterHydratedImage")
		}
	}

	p.images = append(p.images, img)
	return nil
}

// AttachHydratedImageToRun associates a rehydrated image with the provided run and registers it with the media manager.
func (p *paragraph) AttachHydratedImageToRun(r domain.Run, img domain.Image, mediaPath, contentType string, data []byte) error {
	if r == nil {
		return errors.InvalidArgument("Paragraph.AttachHydratedImageToRun", "run", nil, "run cannot be nil")
	}
	if img == nil {
		return errors.InvalidArgument("Paragraph.AttachHydratedImageToRun", "image", nil, "image cannot be nil")
	}

	coreRun, ok := r.(*run)
	if !ok {
		return errors.InvalidArgument("Paragraph.AttachHydratedImageToRun", "run", r, "unexpected run implementation")
	}

	coreRun.setImage(img)

	relID := ""
	if accessor, ok := img.(interface{ RelationshipID() string }); ok {
		relID = accessor.RelationshipID()
	}

	if relID != "" && coreRun.relManager != nil {
		target := img.Target()
		if err := coreRun.relManager.RegisterExisting(relID, constants.RelTypeImage, target, "Internal"); err != nil {
			return errors.Wrap(err, "Paragraph.AttachHydratedImageToRun")
		}
	}
	return p.RegisterHydratedImage(img, mediaPath, contentType, data)
}

// Images returns all images in this paragraph.
func (p *paragraph) Images() []domain.Image {
	images := make([]domain.Image, len(p.images))
	copy(images, p.images)
	return images
}

// Runs returns all runs in this paragraph.
func (p *paragraph) Runs() []domain.Run {
	// Return a copy to prevent external modification
	runs := make([]domain.Run, len(p.runs))
	copy(runs, p.runs)
	return runs
}

// Fields returns all fields in this paragraph.
func (p *paragraph) Fields() []domain.Field {
	fields := make([]domain.Field, len(p.fields))
	copy(fields, p.fields)
	return fields
}

// Text returns the plain text content of the paragraph.
func (p *paragraph) Text() string {
	var sb strings.Builder
	for _, run := range p.runs {
		sb.WriteString(run.Text())
	}
	return sb.String()
}

// Style returns the style applied to this paragraph.
// Note: Currently only returns the style name, not a full Style object.
// For now, use SetStyle() to apply styles and track the name yourself if needed.
func (p *paragraph) Style() domain.Style {
	// Style retrieval from the style manager is not yet implemented.
	// Return nil for now - users should track the applied style name themselves.
	return nil
}

// SetStyle applies a named style to the paragraph.
func (p *paragraph) SetStyle(styleName string) error {
	if styleName == "" {
		return errors.InvalidArgument("Paragraph.SetStyle", "styleName", styleName, "style name cannot be empty")
	}
	p.styleName = styleName
	return nil
}

// StyleName returns the style name applied to this paragraph.
// This is an internal method used by the serializer.
func (p *paragraph) StyleName() string {
	return p.styleName
}

// SetBookmark sets a bookmark for this paragraph (used for TOC).
// This is an internal method used when generating TOC.
func (p *paragraph) SetBookmark(id, name string) {
	p.bookmarkID = id
	p.bookmarkName = name
}

// BookmarkID returns the bookmark ID for this paragraph.
// This is an internal method used by the serializer.
func (p *paragraph) BookmarkID() string {
	return p.bookmarkID
}

// BookmarkName returns the bookmark name for this paragraph.
// This is an internal method used by the serializer.
func (p *paragraph) BookmarkName() string {
	return p.bookmarkName
}

// Alignment returns the paragraph's horizontal alignment.
func (p *paragraph) Alignment() domain.Alignment {
	return p.alignment
}

// SetAlignment sets the paragraph's horizontal alignment.
func (p *paragraph) SetAlignment(align domain.Alignment) error {
	if align < domain.AlignmentLeft || align > domain.AlignmentDistribute {
		return errors.InvalidArgument("Paragraph.SetAlignment", "align", align, "invalid alignment value")
	}
	p.alignment = align
	return nil
}

// Indent returns the paragraph's indentation settings.
func (p *paragraph) Indent() domain.Indentation {
	return p.indent
}

// SetIndent sets the paragraph's indentation.
func (p *paragraph) SetIndent(indent domain.Indentation) error {
	// Validate indentation values
	if indent.Left < constants.MinIndent || indent.Left > constants.MaxIndent {
		return errors.InvalidArgument("Paragraph.SetIndent", "indent.Left", indent.Left,
			"left indent must be between -31680 and 31680 twips (-22 to 22 inches)")
	}
	if indent.Right < constants.MinIndent || indent.Right > constants.MaxIndent {
		return errors.InvalidArgument("Paragraph.SetIndent", "indent.Right", indent.Right,
			"right indent must be between -31680 and 31680 twips (-22 to 22 inches)")
	}
	if indent.FirstLine < 0 || indent.FirstLine > constants.MaxIndent {
		return errors.InvalidArgument("Paragraph.SetIndent", "indent.FirstLine", indent.FirstLine,
			"first line indent must be between 0 and 31680 twips (0 to 22 inches)")
	}
	if indent.Hanging < 0 || indent.Hanging > constants.MaxIndent {
		return errors.InvalidArgument("Paragraph.SetIndent", "indent.Hanging", indent.Hanging,
			"hanging indent must be between 0 and 31680 twips (0 to 22 inches)")
	}
	if indent.FirstLine > 0 && indent.Hanging > 0 {
		return errors.InvalidArgument("Paragraph.SetIndent", "indent", indent,
			"cannot have both first line indent and hanging indent")
	}

	// SetIndent replaces the whole struct in one call, so it clears all four
	// set-flags along with it: a zero-valued side in the struct is the
	// caller's explicit "no indent on this side" for the whole paragraph,
	// not a per-side override, and any indent*Set left over from an earlier
	// SetIndentLeft/Right/FirstLine/Hanging call would otherwise make the
	// serializer emit that stale side's old value (e.g. a leftover
	// w:left="0" from a prior SetIndentLeft) even though this call replaced
	// it. Callers that need to override one side while leaving the other
	// three genuinely untouched must use SetIndentLeft/Right/FirstLine/
	// Hanging instead, each of which marks only its own side.
	p.indent = indent
	p.indentLeftSet = false
	p.indentRightSet = false
	p.indentFirstLineSet = false
	p.indentHangingSet = false
	return nil
}

// SetIndentLeft sets only the left indentation, leaving the other three
// sides and their set-flags untouched. Unlike SetIndent, an explicit 0 here
// is distinguishable from a side that was never set, so the serializer can
// emit it even when it overrides a style's own non-zero value.
func (p *paragraph) SetIndentLeft(twips int) error {
	if twips < constants.MinIndent || twips > constants.MaxIndent {
		return errors.InvalidArgument("Paragraph.SetIndentLeft", "twips", twips,
			"left indent must be between -31680 and 31680 twips (-22 to 22 inches)")
	}
	p.indent.Left = twips
	p.indentLeftSet = true
	return nil
}

// IndentLeftSet reports whether SetIndentLeft was ever called.
func (p *paragraph) IndentLeftSet() bool {
	return p.indentLeftSet
}

// SetIndentRight is SetIndentLeft's counterpart for the right side.
func (p *paragraph) SetIndentRight(twips int) error {
	if twips < constants.MinIndent || twips > constants.MaxIndent {
		return errors.InvalidArgument("Paragraph.SetIndentRight", "twips", twips,
			"right indent must be between -31680 and 31680 twips (-22 to 22 inches)")
	}
	p.indent.Right = twips
	p.indentRightSet = true
	return nil
}

// IndentRightSet reports whether SetIndentRight was ever called.
func (p *paragraph) IndentRightSet() bool {
	return p.indentRightSet
}

// SetIndentFirstLine is SetIndentLeft's counterpart for the first-line
// indent. Setting it does not clear an existing hanging indent -- the
// mutual-exclusivity check lives in SetIndent, which sees the whole struct
// at once; a caller using the per-side setters is responsible for not
// setting both.
func (p *paragraph) SetIndentFirstLine(twips int) error {
	if twips < 0 || twips > constants.MaxIndent {
		return errors.InvalidArgument("Paragraph.SetIndentFirstLine", "twips", twips,
			"first line indent must be between 0 and 31680 twips (0 to 22 inches)")
	}
	p.indent.FirstLine = twips
	p.indentFirstLineSet = true
	return nil
}

// IndentFirstLineSet reports whether SetIndentFirstLine was ever called.
func (p *paragraph) IndentFirstLineSet() bool {
	return p.indentFirstLineSet
}

// SetIndentHanging is SetIndentLeft's counterpart for the hanging indent.
func (p *paragraph) SetIndentHanging(twips int) error {
	if twips < 0 || twips > constants.MaxIndent {
		return errors.InvalidArgument("Paragraph.SetIndentHanging", "twips", twips,
			"hanging indent must be between 0 and 31680 twips (0 to 22 inches)")
	}
	p.indent.Hanging = twips
	p.indentHangingSet = true
	return nil
}

// IndentHangingSet reports whether SetIndentHanging was ever called.
func (p *paragraph) IndentHangingSet() bool {
	return p.indentHangingSet
}

// SpacingBefore returns spacing before the paragraph (in twips).
func (p *paragraph) SpacingBefore() int {
	return p.spacingBefore
}

// SetSpacingBefore sets spacing before the paragraph.
func (p *paragraph) SetSpacingBefore(twips int) error {
	if twips < constants.MinSpacing || twips > constants.MaxSpacing {
		return errors.InvalidArgument("Paragraph.SetSpacingBefore", "twips", twips,
			"spacing must be between 0 and 31680 twips (0 to 22 inches)")
	}
	p.spacingBefore = twips
	p.spacingBeforeSet = true
	return nil
}

// SpacingBeforeSet reports whether SetSpacingBefore was ever called, so the
// serializer can distinguish an explicit 0 from a value that was never set.
func (p *paragraph) SpacingBeforeSet() bool {
	return p.spacingBeforeSet
}

// SpacingAfter returns spacing after the paragraph (in twips).
func (p *paragraph) SpacingAfter() int {
	return p.spacingAfter
}

// SetSpacingAfter sets spacing after the paragraph.
func (p *paragraph) SetSpacingAfter(twips int) error {
	if twips < constants.MinSpacing || twips > constants.MaxSpacing {
		return errors.InvalidArgument("Paragraph.SetSpacingAfter", "twips", twips,
			"spacing must be between 0 and 31680 twips (0 to 22 inches)")
	}
	p.spacingAfter = twips
	p.spacingAfterSet = true
	return nil
}

// SpacingAfterSet reports whether SetSpacingAfter was ever called, so the
// serializer can distinguish an explicit 0 from a value that was never set.
func (p *paragraph) SpacingAfterSet() bool {
	return p.spacingAfterSet
}

// LineSpacing returns the line spacing setting.
func (p *paragraph) LineSpacing() domain.LineSpacing {
	return p.lineSpacing
}

// SetLineSpacing sets the line spacing.
func (p *paragraph) SetLineSpacing(spacing domain.LineSpacing) error {
	if spacing.Rule < domain.LineSpacingAuto || spacing.Rule > domain.LineSpacingAtLeast {
		return errors.InvalidArgument("Paragraph.SetLineSpacing", "spacing.Rule", spacing.Rule,
			"invalid line spacing rule")
	}
	if spacing.Value < constants.MinLineSpacing || spacing.Value > constants.MaxLineSpacing {
		return errors.InvalidArgument("Paragraph.SetLineSpacing", "spacing.Value", spacing.Value,
			"line spacing value must be between 0 and 31680 twips")
	}
	p.lineSpacing = spacing
	p.lineSpacingSet = true
	return nil
}

// LineSpacingSet reports whether SetLineSpacing was ever called, so the
// serializer can distinguish an explicit auto/240 (the default values) from
// line spacing that was never set at all.
func (p *paragraph) LineSpacingSet() bool {
	return p.lineSpacingSet
}

// Numbering returns the numbering reference applied to the paragraph, if any.
func (p *paragraph) Numbering() (domain.NumberingReference, bool) {
	if p == nil || p.numbering == nil {
		return domain.NumberingReference{}, false
	}
	return *p.numbering, true
}

// SetNumbering applies a numbering reference (list id + level) to the paragraph.
func (p *paragraph) SetNumbering(ref domain.NumberingReference) error {
	if p == nil {
		return errors.InvalidState("Paragraph.SetNumbering", "paragraph is nil")
	}
	if ref.Level < domain.NumberingLevelMin || ref.Level > domain.NumberingLevelMax {
		return errors.InvalidArgument("Paragraph.SetNumbering", "ref.Level", ref.Level,
			"numbering level must be between 0 and 8")
	}
	if ref.ID < 0 {
		return errors.InvalidArgument("Paragraph.SetNumbering", "ref.ID", ref.ID,
			"numbering id cannot be negative")
	}
	copyRef := ref
	p.numbering = &copyRef
	return nil
}

// ClearNumbering removes any numbering reference from the paragraph.
func (p *paragraph) ClearNumbering() {
	if p == nil {
		return
	}
	p.numbering = nil
}

// Borders returns the paragraph borders.
func (p *paragraph) Borders() domain.ParagraphBorders {
	return p.borders
}

// SetBorders sets all paragraph borders at once.
func (p *paragraph) SetBorders(borders domain.ParagraphBorders) error {
	if p == nil {
		return errors.InvalidState("Paragraph.SetBorders", "paragraph is nil")
	}
	p.borders = borders
	return nil
}

// SetBorderTop sets the top border.
func (p *paragraph) SetBorderTop(border domain.BorderStyle) error {
	if p == nil {
		return errors.InvalidState("Paragraph.SetBorderTop", "paragraph is nil")
	}
	p.borders.Top = border
	return nil
}

// SetBorderBottom sets the bottom border.
func (p *paragraph) SetBorderBottom(border domain.BorderStyle) error {
	if p == nil {
		return errors.InvalidState("Paragraph.SetBorderBottom", "paragraph is nil")
	}
	p.borders.Bottom = border
	return nil
}

// SetBorderLeft sets the left border.
func (p *paragraph) SetBorderLeft(border domain.BorderStyle) error {
	if p == nil {
		return errors.InvalidState("Paragraph.SetBorderLeft", "paragraph is nil")
	}
	p.borders.Left = border
	return nil
}

// SetBorderRight sets the right border.
func (p *paragraph) SetBorderRight(border domain.BorderStyle) error {
	if p == nil {
		return errors.InvalidState("Paragraph.SetBorderRight", "paragraph is nil")
	}
	p.borders.Right = border
	return nil
}

// ClearRuns removes all runs from the paragraph.
func (p *paragraph) ClearRuns() {
	p.runs = p.runs[:0]
}

// RemoveRun removes the run at the given index.
func (p *paragraph) RemoveRun(index int) error {
	if index < 0 || index >= len(p.runs) {
		return errors.NewValidationError("Paragraph.RemoveRun", "index", index, "run index out of range")
	}
	p.runs = append(p.runs[:index], p.runs[index+1:]...)
	return nil
}

// InsertRunAt inserts a new empty run at the given index and returns it.
// Index must be in [0, len(runs)].
func (p *paragraph) InsertRunAt(index int) (domain.Run, error) {
	if index < 0 || index > len(p.runs) {
		return nil, errors.NewValidationError("Paragraph.InsertRunAt", "index", index, "run index out of range")
	}
	id := p.idGen.NextRunID()
	r := NewRun(id, p.relManager)
	if index == len(p.runs) {
		p.runs = append(p.runs, r)
	} else {
		p.runs = append(p.runs[:index+1], p.runs[index:]...)
		p.runs[index] = r
	}
	return r, nil
}
