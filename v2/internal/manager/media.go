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
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/SlideLang/go-docx/v2/pkg/constants"
	"github.com/SlideLang/go-docx/v2/pkg/errors"
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
