// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

// Package ooxmlmerge applies model-owned changes to OOXML fragments while
// retaining source bytes that the domain model does not understand.
package ooxmlmerge

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
)

const xmlNamespace = "http://www.w3.org/XML/1998/namespace"

const (
	transitionalWordNamespace          = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	strictWordNamespace                = "http://purl.oclc.org/ooxml/wordprocessingml/main"
	transitionalRelationshipsNamespace = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
	strictRelationshipsNamespace       = "http://purl.oclc.org/ooxml/officeDocument/relationships"
	transitionalDrawingNamespace       = "http://schemas.openxmlformats.org/drawingml/2006/main"
	strictDrawingNamespace             = "http://purl.oclc.org/ooxml/drawingml/main"
	transitionalPictureNamespace       = "http://schemas.openxmlformats.org/drawingml/2006/picture"
	strictPictureNamespace             = "http://purl.oclc.org/ooxml/drawingml/picture"
	transitionalWordDrawingNamespace   = "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"
	strictWordDrawingNamespace         = "http://purl.oclc.org/ooxml/drawingml/wordprocessingDrawing"
)

func canonicalNamespace(namespace string) string {
	switch namespace {
	case strictWordNamespace:
		return transitionalWordNamespace
	case strictRelationshipsNamespace:
		return transitionalRelationshipsNamespace
	case strictDrawingNamespace:
		return transitionalDrawingNamespace
	case strictPictureNamespace:
		return transitionalPictureNamespace
	case strictWordDrawingNamespace:
		return transitionalWordDrawingNamespace
	default:
		return namespace
	}
}

func equivalentNamespace(a, b string) bool {
	return canonicalNamespace(a) == canonicalNamespace(b)
}

func strictNamespace(namespace string) string {
	switch canonicalNamespace(namespace) {
	case transitionalWordNamespace:
		return strictWordNamespace
	case transitionalRelationshipsNamespace:
		return strictRelationshipsNamespace
	case transitionalDrawingNamespace:
		return strictDrawingNamespace
	case transitionalPictureNamespace:
		return strictPictureNamespace
	case transitionalWordDrawingNamespace:
		return strictWordDrawingNamespace
	default:
		return namespace
	}
}

func ooxmlNamespaceVariant(context map[string]string, preferredNamespace string) (strict bool, found bool) {
	switch preferredNamespace {
	case strictWordNamespace, strictRelationshipsNamespace, strictDrawingNamespace, strictPictureNamespace, strictWordDrawingNamespace:
		return true, true
	case transitionalWordNamespace, transitionalRelationshipsNamespace, transitionalDrawingNamespace, transitionalPictureNamespace, transitionalWordDrawingNamespace:
		return false, true
	}

	strictMain := false
	transitionalMain := false
	for _, namespace := range context {
		switch namespace {
		case strictWordNamespace:
			strictMain = true
		case transitionalWordNamespace:
			transitionalMain = true
		}
	}
	if strictMain {
		return true, true
	}
	if transitionalMain {
		return false, true
	}
	strictFamily := false
	transitionalFamily := false
	for _, namespace := range context {
		switch namespace {
		case strictRelationshipsNamespace, strictDrawingNamespace, strictPictureNamespace, strictWordDrawingNamespace:
			strictFamily = true
		case transitionalRelationshipsNamespace, transitionalDrawingNamespace, transitionalPictureNamespace, transitionalWordDrawingNamespace:
			transitionalFamily = true
		}
	}
	if strictFamily {
		return true, true
	}
	if transitionalFamily {
		return false, true
	}
	return false, false
}

// QName identifies an XML name by namespace and local name. Prefix is kept
// only for rendering; comparisons use Namespace and Local.
type QName struct {
	Namespace string
	Prefix    string
	Local     string
}

// Name creates a QName for a namespace/local pair.
func Name(namespace, local string) QName {
	return QName{Namespace: namespace, Local: local}
}

func (n QName) key() string {
	space := canonicalNamespace(n.Namespace)
	if space == "" {
		space = "prefix:" + n.Prefix
	}
	return space + "\x00" + n.Local
}

func (n QName) rendered() string {
	if n.Prefix == "" {
		return n.Local
	}
	return n.Prefix + ":" + n.Local
}

// Attribute is one losslessly located start-tag attribute.
type Attribute struct {
	Name       QName
	Value      string
	Start      int
	ValueStart int
	ValueEnd   int
	End        int
}

// Node is an XML element whose offsets point into Raw. Child offsets use the
// same byte slice, which lets a merger retain every gap, comment and unknown
// child without reserializing it.
type Node struct {
	Name         QName
	Attributes   []Attribute
	Children     []*Node
	Raw          []byte
	Start        int
	ContentStart int
	ContentEnd   int
	End          int
	SelfClosing  bool
	namespaces   map[string]string
}

// Bytes returns the exact source bytes for the node.
func (n *Node) Bytes() []byte {
	if n == nil || n.Start < 0 || n.End < n.Start || n.End > len(n.Raw) {
		return nil
	}
	return n.Raw[n.Start:n.End]
}

// Attr returns an attribute by resolved QName.
func (n *Node) Attr(name QName) (string, bool) {
	if n == nil {
		return "", false
	}
	key := name.key()
	for _, attr := range n.Attributes {
		if attr.Name.key() == key {
			return attr.Value, true
		}
	}
	return "", false
}

func cloneNamespaces(source map[string]string) map[string]string {
	result := make(map[string]string, len(source)+1)
	for prefix, namespace := range source {
		result[prefix] = namespace
	}
	result["xml"] = xmlNamespace
	return result
}

func rawQName(name xml.Name, namespaces map[string]string) QName {
	prefix := name.Space
	namespace := namespaces[prefix]
	if prefix == "" {
		namespace = namespaces[""]
	}
	return QName{Namespace: namespace, Prefix: prefix, Local: name.Local}
}

func rawAttributeQName(name xml.Name, namespaces map[string]string) QName {
	if name.Space == "" {
		// A default namespace applies to element names, never to unprefixed
		// attributes.
		return QName{Local: name.Local}
	}
	return QName{Namespace: namespaces[name.Space], Prefix: name.Space, Local: name.Local}
}

func namespaceDeclaration(attr xml.Attr) (string, string, bool) {
	if attr.Name.Space == "xmlns" {
		return attr.Name.Local, attr.Value, true
	}
	if attr.Name.Space == "" && attr.Name.Local == "xmlns" {
		return "", attr.Value, true
	}
	return "", "", false
}

type attributeSpan struct {
	start      int
	valueStart int
	valueEnd   int
	end        int
}

func isXMLSpace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\r' || value == '\n'
}

func scanAttributeSpans(data []byte, start, end, count int) ([]attributeSpan, error) {
	if start < 0 || end < start || end > len(data) {
		return nil, fmt.Errorf("invalid XML start-tag range [%d:%d]", start, end)
	}
	cursor := start
	if cursor >= end || data[cursor] != '<' {
		return nil, fmt.Errorf("XML start tag does not begin with '<'")
	}
	cursor++
	for cursor < end && !isXMLSpace(data[cursor]) && data[cursor] != '>' && data[cursor] != '/' {
		cursor++
	}

	spans := make([]attributeSpan, 0, count)
	for len(spans) < count {
		fullStart := cursor
		for cursor < end && isXMLSpace(data[cursor]) {
			cursor++
		}
		if cursor >= end || data[cursor] == '>' || data[cursor] == '/' {
			return nil, fmt.Errorf("XML start tag contains fewer attributes than the decoder reported")
		}
		for cursor < end && !isXMLSpace(data[cursor]) && data[cursor] != '=' && data[cursor] != '>' && data[cursor] != '/' {
			cursor++
		}
		for cursor < end && isXMLSpace(data[cursor]) {
			cursor++
		}
		if cursor >= end || data[cursor] != '=' {
			return nil, fmt.Errorf("XML attribute is missing '='")
		}
		cursor++
		for cursor < end && isXMLSpace(data[cursor]) {
			cursor++
		}
		if cursor >= end || (data[cursor] != '\'' && data[cursor] != '"') {
			return nil, fmt.Errorf("XML attribute value is not quoted")
		}
		quote := data[cursor]
		cursor++
		valueStart := cursor
		for cursor < end && data[cursor] != quote {
			cursor++
		}
		if cursor >= end {
			return nil, fmt.Errorf("XML attribute value is unterminated")
		}
		valueEnd := cursor
		cursor++
		spans = append(spans, attributeSpan{
			start:      fullStart,
			valueStart: valueStart,
			valueEnd:   valueEnd,
			end:        cursor,
		})
	}
	return spans, nil
}

// Parse parses one XML element. inherited maps prefixes to namespace URIs and
// is used when the fragment relies on declarations from an ancestor outside
// the provided bytes.
func Parse(data []byte, inherited map[string]string) (*Node, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		startOffset := int(dec.InputOffset())
		tok, err := dec.RawToken()
		if err != nil {
			return nil, err
		}
		if start, ok := tok.(xml.StartElement); ok {
			return parseNode(dec, data, start, startOffset, inherited)
		}
	}
}

func parseNode(dec *xml.Decoder, data []byte, start xml.StartElement, startOffset int, inherited map[string]string) (*Node, error) {
	namespaces := cloneNamespaces(inherited)
	for _, attr := range start.Attr {
		if prefix, namespace, ok := namespaceDeclaration(attr); ok {
			namespaces[prefix] = namespace
		}
	}
	node := &Node{
		Name:         rawQName(start.Name, namespaces),
		Raw:          data,
		Start:        startOffset,
		ContentStart: int(dec.InputOffset()),
		namespaces:   namespaces,
	}
	startTag := data[node.Start:node.ContentStart]
	node.SelfClosing = bytes.HasSuffix(bytes.TrimSpace(startTag), []byte("/>"))
	spans, err := scanAttributeSpans(data, node.Start, node.ContentStart, len(start.Attr))
	if err != nil {
		return nil, err
	}
	for index, attr := range start.Attr {
		span := spans[index]
		if prefix, namespace, ok := namespaceDeclaration(attr); ok {
			name := QName{Prefix: "xmlns", Local: prefix}
			if prefix == "" {
				name = QName{Local: "xmlns"}
			}
			node.Attributes = append(node.Attributes, Attribute{
				Name: name, Value: namespace, Start: span.start,
				ValueStart: span.valueStart, ValueEnd: span.valueEnd, End: span.end,
			})
			continue
		}
		node.Attributes = append(node.Attributes, Attribute{
			Name: rawAttributeQName(attr.Name, namespaces), Value: attr.Value, Start: span.start,
			ValueStart: span.valueStart, ValueEnd: span.valueEnd, End: span.end,
		})
	}

	for {
		tokenStart := int(dec.InputOffset())
		tok, err := dec.RawToken()
		if err != nil {
			return nil, err
		}
		switch token := tok.(type) {
		case xml.StartElement:
			child, err := parseNode(dec, data, token, tokenStart, namespaces)
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, child)
		case xml.EndElement:
			endName := rawQName(token.Name, namespaces)
			if endName.key() == node.Name.key() {
				node.ContentEnd = tokenStart
				node.End = int(dec.InputOffset())
				return node, nil
			}
		}
	}
}

// Mode selects how a node is combined.
type Mode int

const (
	// Preserve keeps the source node exactly.
	Preserve Mode = iota
	// Replace uses the current node exactly.
	Replace
	// MergeAttributes applies owned attributes and retains source content.
	MergeAttributes
	// MergeChildren applies owned attributes and recursively merges children.
	MergeChildren
	// Splice applies policy-provided byte edits to the source node.
	Splice
)

// Policy defines model ownership for one element.
type Policy struct {
	Mode            Mode
	OwnedAttributes []QName
	DropAttributes  []QName
	// DropSourceChildren removes source children with these resolved names
	// whenever this policy is evaluated. Other source children remain opaque,
	// and a matching current child is rendered as a modeled replacement.
	DropSourceChildren []QName
	UseCurrentContent  bool
	Children           map[string]*Policy
	ChildRanks         map[string]int
	// ChildKey returns a stable, unique identity for a keyed child. Children
	// for which it returns an empty string are matched by QName and position.
	ChildKey    func(*Node) string
	SpliceEdits func(source, snapshot, current *Node) ([]Edit, error)
}

// PolicyKey returns the map key used by Children and ChildRanks.
func PolicyKey(name QName) string {
	return name.key()
}

func qnameSet(names []QName) map[string]bool {
	result := make(map[string]bool, len(names))
	for _, name := range names {
		result[name.key()] = true
	}
	return result
}

func attrMap(node *Node) map[string]Attribute {
	result := make(map[string]Attribute, len(node.Attributes))
	for _, attr := range node.Attributes {
		result[attr.Name.key()] = attr
	}
	return result
}

func writeAttribute(out *bytes.Buffer, attr Attribute) error {
	out.WriteByte(' ')
	out.WriteString(attr.Name.rendered())
	out.WriteString(`="`)
	if err := xml.EscapeText(out, []byte(attr.Value)); err != nil {
		return err
	}
	out.WriteByte('"')
	return nil
}

func isNamespaceAttribute(attr Attribute) bool {
	return attr.Name.Prefix == "xmlns" || (attr.Name.Prefix == "" && attr.Name.Local == "xmlns")
}

type renderBindings struct {
	context      map[string]string
	prefixes     map[string]string
	declarations map[string]string
	strictOOXML  bool
	hasOOXML     bool
}

func newRenderBindings(context map[string]string, preferredNamespace string) *renderBindings {
	strictOOXML, hasOOXML := ooxmlNamespaceVariant(context, preferredNamespace)
	return &renderBindings{
		context:      cloneNamespaces(context),
		prefixes:     make(map[string]string),
		declarations: make(map[string]string),
		strictOOXML:  strictOOXML,
		hasOOXML:     hasOOXML,
	}
}

func (b *renderBindings) namespace(namespace string) string {
	if !b.hasOOXML {
		return namespace
	}
	if b.strictOOXML {
		return strictNamespace(namespace)
	}
	return canonicalNamespace(namespace)
}

func (b *renderBindings) prefix(name QName) string {
	if name.Namespace == "" {
		return name.Prefix
	}
	canonical := canonicalNamespace(name.Namespace)
	if prefix, ok := b.prefixes[canonical]; ok {
		return prefix
	}
	targetNamespace := b.namespace(name.Namespace)
	if name.Prefix != "" && b.context[name.Prefix] == targetNamespace {
		b.prefixes[canonical] = name.Prefix
		return name.Prefix
	}

	var candidates []string
	for prefix, namespace := range b.context {
		if prefix != "" && namespace == targetNamespace {
			candidates = append(candidates, prefix)
		}
	}
	if len(candidates) > 0 {
		sort.Strings(candidates)
		b.prefixes[canonical] = candidates[0]
		return candidates[0]
	}

	base := name.Prefix
	if base == "" || base == "xml" || base == "xmlns" {
		base = "ns"
	}
	declarationNamespace := targetNamespace
	if defaultNamespace := b.context[""]; defaultNamespace == targetNamespace {
		// Elements can use a default namespace while qualified attributes
		// still require a prefix. Keep the source document's Strict or
		// Transitional namespace family when introducing that prefix.
		declarationNamespace = defaultNamespace
	}
	prefix := base
	for suffix := 1; ; suffix++ {
		if namespace, used := b.context[prefix]; !used || equivalentNamespace(namespace, declarationNamespace) {
			break
		}
		prefix = fmt.Sprintf("%s%d", base, suffix)
	}
	b.context[prefix] = declarationNamespace
	b.prefixes[canonical] = prefix
	b.declarations[prefix] = declarationNamespace
	return prefix
}

func collectBindings(node *Node, bindings *renderBindings) {
	if node == nil {
		return
	}
	bindings.prefix(node.Name)
	for _, attr := range node.Attributes {
		if !isNamespaceAttribute(attr) {
			bindings.prefix(attr.Name)
		}
	}
	for _, child := range node.Children {
		collectBindings(child, bindings)
	}
}

func renderedQName(name QName, bindings *renderBindings) string {
	prefix := bindings.prefix(name)
	if prefix == "" {
		return name.Local
	}
	return prefix + ":" + name.Local
}

func renderNode(out *bytes.Buffer, node *Node, bindings *renderBindings, root bool) error {
	if node == nil {
		return nil
	}
	out.WriteByte('<')
	out.WriteString(renderedQName(node.Name, bindings))
	for _, attr := range node.Attributes {
		if isNamespaceAttribute(attr) {
			continue
		}
		if canonicalNamespace(node.Name.Namespace) == transitionalDrawingNamespace &&
			node.Name.Local == "graphicData" && attr.Name.Namespace == "" && attr.Name.Local == "uri" {
			attr.Value = bindings.namespace(attr.Value)
		}
		attr.Name.Prefix = bindings.prefix(attr.Name)
		if err := writeAttribute(out, attr); err != nil {
			return err
		}
	}
	if root {
		prefixes := make([]string, 0, len(bindings.declarations))
		for prefix := range bindings.declarations {
			prefixes = append(prefixes, prefix)
		}
		sort.Strings(prefixes)
		for _, prefix := range prefixes {
			if err := writeAttribute(out, Attribute{
				Name:  QName{Prefix: "xmlns", Local: prefix},
				Value: bindings.declarations[prefix],
			}); err != nil {
				return err
			}
		}
	}
	if node.SelfClosing {
		out.WriteString("/>")
		return nil
	}
	out.WriteByte('>')
	cursor := node.ContentStart
	for _, child := range node.Children {
		out.Write(node.Raw[cursor:child.Start])
		if err := renderNode(out, child, bindings, false); err != nil {
			return err
		}
		cursor = child.End
	}
	out.Write(node.Raw[cursor:node.ContentEnd])
	out.WriteString("</")
	out.WriteString(renderedQName(node.Name, bindings))
	out.WriteByte('>')
	return nil
}

// RenderInContext renders a generated node with prefixes that are valid in
// the destination context. Existing compatible prefixes are reused; missing
// namespaces are declared on the rendered root with collision-free prefixes.
func RenderInContext(node *Node, context map[string]string) ([]byte, error) {
	return RenderInContextWithMainNamespace(node, context, "")
}

// RenderInContextWithMainNamespace is RenderInContext with an explicit OOXML
// family anchor. mainNamespace should be the resolved namespace of the
// destination document element; it takes priority over unused declarations
// from the other conformance family that may also exist in context.
func RenderInContextWithMainNamespace(node *Node, context map[string]string, mainNamespace string) ([]byte, error) {
	if node == nil {
		return nil, nil
	}
	bindings := newRenderBindings(context, mainNamespace)
	collectBindings(node, bindings)
	var out bytes.Buffer
	if err := renderNode(&out, node, bindings, true); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func mergedOpen(source, current *Node, policy *Policy, selfClosing bool) ([]byte, error) {
	owned := qnameSet(policy.OwnedAttributes)
	dropped := qnameSet(policy.DropAttributes)
	currentAttrs := attrMap(current)
	written := make(map[string]bool, len(currentAttrs))
	bindings := newRenderBindings(source.namespaces, source.Name.Namespace)
	edits := make([]Edit, 0, len(source.Attributes))
	for _, attr := range source.Attributes {
		key := attr.Name.key()
		if dropped[key] {
			edits = append(edits, Edit{Start: attr.Start - source.Start, End: attr.End - source.Start})
			continue
		}
		if !owned[key] {
			continue
		}
		replacement, ok := currentAttrs[key]
		if !ok {
			edits = append(edits, Edit{Start: attr.Start - source.Start, End: attr.End - source.Start})
			continue
		}
		var escaped bytes.Buffer
		if err := xml.EscapeText(&escaped, []byte(replacement.Value)); err != nil {
			return nil, err
		}
		edits = append(edits, Edit{
			Start:       attr.ValueStart - source.Start,
			End:         attr.ValueEnd - source.Start,
			Replacement: escaped.Bytes(),
		})
		written[key] = true
	}

	open, err := ApplyEdits(source.Raw[source.Start:source.ContentStart], edits)
	if err != nil {
		return nil, err
	}
	var additions bytes.Buffer
	for _, attr := range current.Attributes {
		key := attr.Name.key()
		if !owned[key] || dropped[key] || written[key] {
			continue
		}
		attr.Name.Prefix = bindings.prefix(attr.Name)
		if err := writeAttribute(&additions, attr); err != nil {
			return nil, err
		}
		written[key] = true
	}
	prefixes := make([]string, 0, len(bindings.declarations))
	for prefix := range bindings.declarations {
		prefixes = append(prefixes, prefix)
	}
	sort.Strings(prefixes)
	for _, prefix := range prefixes {
		if err := writeAttribute(&additions, Attribute{
			Name:  QName{Prefix: "xmlns", Local: prefix},
			Value: bindings.declarations[prefix],
		}); err != nil {
			return nil, err
		}
	}

	markerStart := len(open) - 1
	if source.SelfClosing {
		markerStart--
	}
	if markerStart < 0 || markerStart > len(open) {
		return nil, fmt.Errorf("invalid preserved XML start tag")
	}
	var out bytes.Buffer
	out.Write(open[:markerStart])
	out.Write(additions.Bytes())
	if selfClosing {
		out.WriteString("/>")
	} else {
		out.WriteByte('>')
	}
	return out.Bytes(), nil
}

type childEntry struct {
	node   *Node
	key    string
	rank   int
	order  int
	unique bool
}

const unrankedChild = 1 << 30

func policyForChild(policy *Policy, child *Node) *Policy {
	if policy == nil || child == nil {
		return nil
	}
	return policy.Children[child.Name.key()]
}

func childBaseKey(policy *Policy, child *Node) (string, bool) {
	if policy != nil && policy.ChildKey != nil {
		if key := policy.ChildKey(child); key != "" {
			return key, true
		}
	}
	return child.Name.key(), false
}

func childEntries(node *Node, policy *Policy) []childEntry {
	if node == nil {
		return nil
	}
	occurrences := make(map[string]int)
	entries := make([]childEntry, 0, len(node.Children))
	for index, child := range node.Children {
		base, unique := childBaseKey(policy, child)
		occurrence := occurrences[base]
		occurrences[base]++
		key := base
		if !unique {
			key = fmt.Sprintf("%s\x00%d", base, occurrence)
		}
		rank := unrankedChild
		if policy != nil {
			if value, ok := policy.ChildRanks[child.Name.key()]; ok {
				rank = value
			}
		}
		entries = append(entries, childEntry{
			node:   child,
			key:    key,
			rank:   rank,
			order:  index,
			unique: unique,
		})
	}
	return entries
}

func entryMap(entries []childEntry) map[string]childEntry {
	result := make(map[string]childEntry, len(entries))
	for _, entry := range entries {
		result[entry.key] = entry
	}
	return result
}

func sameNode(a, b *Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	return bytes.Equal(a.Bytes(), b.Bytes())
}

// Merge applies policy to source using snapshot as the original model
// projection and current as the requested model projection.
func Merge(source, snapshot, current *Node, policy *Policy) ([]byte, error) {
	if source != nil && policy != nil && policy.Mode == Preserve {
		return append([]byte(nil), source.Bytes()...), nil
	}
	if current == nil {
		return nil, nil
	}
	if source == nil || policy == nil {
		return append([]byte(nil), current.Bytes()...), nil
	}
	if policy.Mode == Replace {
		return RenderInContextWithMainNamespace(current, source.namespaces, source.Name.Namespace)
	}
	if sameNode(snapshot, current) {
		return append([]byte(nil), source.Bytes()...), nil
	}
	if policy.Mode == Splice {
		if policy.SpliceEdits == nil {
			return nil, fmt.Errorf("splice policy has no edit function")
		}
		edits, err := policy.SpliceEdits(source, snapshot, current)
		if err != nil {
			return nil, err
		}
		return ApplyEdits(source.Bytes(), edits)
	}

	if policy.Mode == MergeAttributes {
		useCurrentContent := policy.UseCurrentContent
		selfClosing := source.SelfClosing
		if useCurrentContent {
			selfClosing = current.SelfClosing
		}
		open, err := mergedOpen(source, current, policy, selfClosing)
		if err != nil {
			return nil, err
		}
		if selfClosing {
			return open, nil
		}
		var out bytes.Buffer
		out.Write(open)
		if useCurrentContent {
			cursor := current.ContentStart
			for _, child := range current.Children {
				out.Write(current.Raw[cursor:child.Start])
				rendered, err := RenderInContextWithMainNamespace(child, source.namespaces, source.Name.Namespace)
				if err != nil {
					return nil, err
				}
				out.Write(rendered)
				cursor = child.End
			}
			out.Write(current.Raw[cursor:current.ContentEnd])
		} else {
			out.Write(source.Raw[source.ContentStart:source.ContentEnd])
		}
		if !source.SelfClosing {
			out.Write(source.Raw[source.ContentEnd:source.End])
		} else {
			out.WriteString("</")
			out.WriteString(source.Name.rendered())
			out.WriteByte('>')
		}
		return out.Bytes(), nil
	}

	return mergeChildren(source, snapshot, current, policy)
}

func mergeChildren(source, snapshot, current *Node, policy *Policy) ([]byte, error) {
	sourceEntries := childEntries(source, policy)
	snapshotMap := entryMap(childEntries(snapshot, policy))
	currentEntries := childEntries(current, policy)
	currentMap := entryMap(currentEntries)
	sourceMap := entryMap(sourceEntries)
	droppedSourceChildren := qnameSet(policy.DropSourceChildren)
	activeSourceMap := make(map[string]childEntry, len(sourceMap))
	for key, entry := range sourceMap {
		if !droppedSourceChildren[entry.node.Name.key()] {
			activeSourceMap[key] = entry
		}
	}

	var inserts []childEntry
	plannedUnique := make(map[string]bool)
	for _, entry := range currentEntries {
		if policyForChild(policy, entry.node) == nil {
			continue
		}
		if entry.unique && plannedUnique[entry.key] {
			continue
		}
		if entry.unique {
			plannedUnique[entry.key] = true
		}
		if _, exists := activeSourceMap[entry.key]; exists {
			continue
		}
		snapshotEntry, hadSnapshot := snapshotMap[entry.key]
		_, replacesDroppedSource := sourceMap[entry.key]
		if hadSnapshot && sameNode(snapshotEntry.node, entry.node) && !replacesDroppedSource {
			continue
		}
		inserts = append(inserts, entry)
	}
	sort.SliceStable(inserts, func(i, j int) bool {
		if inserts[i].rank != inserts[j].rank {
			return inserts[i].rank < inserts[j].rank
		}
		return inserts[i].order < inserts[j].order
	})

	open, close := containerTags(source)
	if len(policy.OwnedAttributes) > 0 || len(policy.DropAttributes) > 0 {
		var err error
		open, err = mergedOpen(source, current, policy, false)
		if err != nil {
			return nil, err
		}
	}
	var out bytes.Buffer
	out.Write(open)
	cursor := source.ContentStart
	nextInsert := 0
	seenUnique := make(map[string]bool)
	// Opaque children have no OOXML rank. Treat the entire opaque run before
	// a ranked child as part of that following child's anchor; otherwise a
	// pending insertion would stop at the first opaque node and could leapfrog
	// a lower-ranked modeled child that appears after it.
	insertionRanks := make([]int, len(sourceEntries))
	nextRank := unrankedChild
	for index := len(sourceEntries) - 1; index >= 0; index-- {
		if sourceEntries[index].rank != unrankedChild {
			nextRank = sourceEntries[index].rank
		}
		insertionRanks[index] = nextRank
	}
	for index, entry := range sourceEntries {
		for nextInsert < len(inserts) && inserts[nextInsert].rank < insertionRanks[index] {
			rendered, err := RenderInContextWithMainNamespace(inserts[nextInsert].node, source.namespaces, source.Name.Namespace)
			if err != nil {
				return nil, err
			}
			out.Write(rendered)
			nextInsert++
		}
		out.Write(source.Raw[cursor:entry.node.Start])
		if droppedSourceChildren[entry.node.Name.key()] {
			cursor = entry.node.End
			continue
		}
		childPolicy := policyForChild(policy, entry.node)
		currentEntry, hasCurrent := currentMap[entry.key]
		snapshotEntry, hasSnapshot := snapshotMap[entry.key]
		keyChanged := hasCurrent != hasSnapshot ||
			(hasCurrent && hasSnapshot && !sameNode(snapshotEntry.node, currentEntry.node))
		if childPolicy != nil && entry.unique && keyChanged && seenUnique[entry.key] {
			cursor = entry.node.End
			continue
		}
		if childPolicy != nil && entry.unique {
			seenUnique[entry.key] = true
		}
		if childPolicy == nil {
			out.Write(entry.node.Bytes())
			cursor = entry.node.End
			continue
		}
		if !hasCurrent {
			if hasSnapshot {
				cursor = entry.node.End
				continue
			}
			out.Write(entry.node.Bytes())
			cursor = entry.node.End
			continue
		}
		if hasSnapshot && sameNode(snapshotEntry.node, currentEntry.node) {
			out.Write(entry.node.Bytes())
			cursor = entry.node.End
			continue
		}
		merged, err := Merge(entry.node, snapshotEntry.node, currentEntry.node, childPolicy)
		if err != nil {
			return nil, err
		}
		out.Write(merged)
		cursor = entry.node.End
	}
	for nextInsert < len(inserts) {
		rendered, err := RenderInContextWithMainNamespace(inserts[nextInsert].node, source.namespaces, source.Name.Namespace)
		if err != nil {
			return nil, err
		}
		out.Write(rendered)
		nextInsert++
	}
	out.Write(source.Raw[cursor:source.ContentEnd])
	out.Write(close)
	return out.Bytes(), nil
}

func containerTags(node *Node) ([]byte, []byte) {
	if node == nil {
		return nil, nil
	}
	if !node.SelfClosing {
		return node.Raw[node.Start:node.ContentStart], node.Raw[node.ContentEnd:node.End]
	}
	open := append([]byte(nil), node.Raw[node.Start:node.ContentStart]...)
	open = bytes.TrimSpace(open)
	if bytes.HasSuffix(open, []byte("/>")) {
		open = append(open[:len(open)-2], '>')
	}
	close := []byte("</" + node.Name.rendered() + ">")
	return open, close
}

// ApplyEdits applies non-overlapping byte edits from the end of the source.
type Edit struct {
	Start       int
	End         int
	Replacement []byte
}

func ApplyEdits(source []byte, edits []Edit) ([]byte, error) {
	type indexedEdit struct {
		Edit
		index int
	}
	sorted := make([]indexedEdit, len(edits))
	for index, edit := range edits {
		sorted[index] = indexedEdit{Edit: edit, index: index}
	}
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Start != sorted[j].Start {
			return sorted[i].Start > sorted[j].Start
		}
		iReplaces := sorted[i].End > sorted[i].Start
		jReplaces := sorted[j].End > sorted[j].Start
		if iReplaces != jReplaces {
			// Replace the original range before inserting at its leading
			// boundary, otherwise the replacement would consume the inserted
			// bytes after offsets shift.
			return iReplaces
		}
		// Applying equal-offset insertions in reverse caller order preserves
		// their original order in the final byte stream.
		return sorted[i].index > sorted[j].index
	})
	result := append([]byte(nil), source...)
	lastStart := len(source)
	for _, indexed := range sorted {
		edit := indexed.Edit
		if edit.Start < 0 || edit.End < edit.Start || edit.End > len(source) || edit.End > lastStart {
			return nil, fmt.Errorf("overlapping or invalid XML edit [%d:%d]", edit.Start, edit.End)
		}
		next := make([]byte, 0, len(result)-(edit.End-edit.Start)+len(edit.Replacement))
		next = append(next, result[:edit.Start]...)
		next = append(next, edit.Replacement...)
		next = append(next, result[edit.End:]...)
		result = next
		lastStart = edit.Start
	}
	return result, nil
}

// AppendChild inserts raw child XML at the end of a container. A self-closing
// source element is expanded without changing its attributes.
func AppendChild(node *Node, child []byte) ([]byte, error) {
	if node == nil {
		return nil, fmt.Errorf("cannot append a child to a nil XML node")
	}
	open, close := containerTags(node)
	var out bytes.Buffer
	out.Write(open)
	if !node.SelfClosing {
		out.Write(node.Raw[node.ContentStart:node.ContentEnd])
	}
	out.Write(child)
	out.Write(close)
	return out.Bytes(), nil
}

// First finds the first descendant, including root, with the requested QName.
func First(root *Node, name QName) *Node {
	if root == nil {
		return nil
	}
	if root.Name.key() == name.key() {
		return root
	}
	for _, child := range root.Children {
		if found := First(child, name); found != nil {
			return found
		}
	}
	return nil
}

// ParseAll parses consecutive top-level elements from one byte slice.
func ParseAll(data []byte, inherited map[string]string) ([]*Node, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var nodes []*Node
	for {
		startOffset := int(dec.InputOffset())
		tok, err := dec.RawToken()
		if err != nil {
			if err == io.EOF {
				return nodes, nil
			}
			return nil, err
		}
		if start, ok := tok.(xml.StartElement); ok {
			node, err := parseNode(dec, data, start, startOffset, inherited)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, node)
		}
	}
}
