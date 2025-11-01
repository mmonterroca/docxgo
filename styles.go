package docx

import (
	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/internal/manager"
)

// NewParagraphStyle creates a custom paragraph style that can be registered with a document style manager.
func NewParagraphStyle(styleID, name string) domain.ParagraphStyle {
	return manager.NewParagraphStyle(styleID, name)
}
