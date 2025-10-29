/*
MIT License

Copyright (c) 2025 Misael Monterroca
Copyright (c) 2020-2023 fumiama (original go-docx)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package manager

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/mmonterroca/docxgo/pkg/constants"
	"github.com/mmonterroca/docxgo/pkg/errors"
)

// MediaFile represents a media file in the document.
type MediaFile struct {
	ID          string // Unique ID
	Name        string // File name (e.g., "image1.png")
	Path        string // Path within .docx (e.g., "word/media/image1.png")
	ContentType string // MIME type
	Data        []byte // File data
}

// MediaManager manages media files (images, etc.) in a document.
// It is thread-safe.
type MediaManager struct {
	mu      sync.RWMutex
	files   map[string]*MediaFile // key is ID
	idGen   *IDGenerator
	counter int // Counter for generating unique file names
}

// NewMediaManager creates a new media manager.
func NewMediaManager(idGen *IDGenerator) *MediaManager {
	return &MediaManager{
		files: make(map[string]*MediaFile, constants.DefaultMediaCapacity),
		idGen: idGen,
	}
}

// Add adds a media file and returns its ID and path.
func (mm *MediaManager) Add(data []byte, filename string) (id, path string, err error) {
	if len(data) == 0 {
		return "", "", errors.InvalidArgument("MediaManager.Add", "data", data, "media data cannot be empty")
	}
	if filename == "" {
		return "", "", errors.InvalidArgument("MediaManager.Add", "filename", filename, "filename cannot be empty")
	}

	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Generate unique ID
	id = mm.idGen.NextImageID()

	// Detect content type from extension
	ext := strings.ToLower(filepath.Ext(filename))
	contentType := mm.detectContentType(ext)

	// Generate unique filename
	mm.counter++
	uniqueName := fmt.Sprintf("image%d%s", mm.counter, ext)
	path = constants.PathMediaPrefix + uniqueName

	file := &MediaFile{
		ID:          id,
		Name:        uniqueName,
		Path:        path,
		ContentType: contentType,
		Data:        data,
	}

	mm.files[id] = file
	return id, path, nil
}

// RegisterExisting registers a media file that already exists in the DOCX package.
// It preserves the original path, name, and content type so round-tripping does not
// duplicate or rename embedded assets when hydrating documents.
func (mm *MediaManager) RegisterExisting(id, path, contentType string, data []byte) (string, error) {
	if len(data) == 0 {
		return "", errors.InvalidArgument("MediaManager.RegisterExisting", "data", data, "media data cannot be empty")
	}
	if path == "" {
		return "", errors.InvalidArgument("MediaManager.RegisterExisting", "path", path, "media path cannot be empty")
	}

	m := strings.ReplaceAll(path, "\\", "/")
	m = strings.TrimSpace(m)

	mm.mu.Lock()
	defer mm.mu.Unlock()

	if id == "" {
		id = mm.idGen.NextImageID()
	}

	filename := filepath.Base(m)
	if filename == "" {
		filename = id
	}

	if !strings.HasPrefix(m, constants.PathMediaPrefix) {
		m = constants.PathMediaPrefix + strings.TrimPrefix(m, "/")
	}

	if contentType == "" {
		ext := strings.ToLower(filepath.Ext(filename))
		contentType = mm.detectContentType(ext)
	}

	// Update counter so future generated names do not collide with hydrated ones.
	lowerName := strings.ToLower(filename)
	if strings.HasPrefix(lowerName, "image") {
		digits := strings.TrimPrefix(lowerName, "image")
		digits = strings.TrimSuffix(digits, strings.ToLower(filepath.Ext(filename)))
		if n, err := strconv.Atoi(digits); err == nil && n > mm.counter {
			mm.counter = n
		}
	}

	copyData := make([]byte, len(data))
	copy(copyData, data)

	mm.files[id] = &MediaFile{
		ID:          id,
		Name:        filename,
		Path:        m,
		ContentType: contentType,
		Data:        copyData,
	}

	return id, nil
}

// Get retrieves a media file by ID.
func (mm *MediaManager) Get(id string) (*MediaFile, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	file, exists := mm.files[id]
	if !exists {
		return nil, errors.NotFound("MediaManager.Get", "media file")
	}

	return file, nil
}

// GetByPath retrieves a media file by path.
func (mm *MediaManager) GetByPath(path string) (*MediaFile, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	for _, file := range mm.files {
		if file.Path == path {
			return file, nil
		}
	}

	return nil, errors.NotFound("MediaManager.GetByPath", "media file")
}

// All returns all media files.
func (mm *MediaManager) All() []*MediaFile {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	files := make([]*MediaFile, 0, len(mm.files))
	for _, file := range mm.files {
		files = append(files, file)
	}

	return files
}

// Count returns the number of media files.
func (mm *MediaManager) Count() int {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	return len(mm.files)
}

// Delete removes a media file by ID.
func (mm *MediaManager) Delete(id string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if _, exists := mm.files[id]; !exists {
		return errors.NotFound("MediaManager.Delete", "media file")
	}

	delete(mm.files, id)
	return nil
}

// detectContentType returns the MIME type for a file extension.
func (mm *MediaManager) detectContentType(ext string) string {
	switch ext {
	case ".png":
		return constants.ContentTypePNG
	case ".jpg", ".jpeg":
		return constants.ContentTypeJPEG
	case ".gif":
		return constants.ContentTypeGIF
	case ".bmp":
		return constants.ContentTypeBMP
	case ".tiff", ".tif":
		return constants.ContentTypeTIFF
	case ".wmf":
		return constants.ContentTypeWMF
	case ".emf":
		return constants.ContentTypeEMF
	default:
		return "application/octet-stream"
	}
}
