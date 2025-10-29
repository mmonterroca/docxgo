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

package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestDocxError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *DocxError
		contains []string
	}{
		{
			name: "full error",
			err: &DocxError{
				Code:    ErrCodeValidation,
				Op:      "Document.AddParagraph",
				Message: "invalid paragraph",
				Err:     errors.New("underlying error"),
				Context: map[string]interface{}{"field": "value"},
			},
			contains: []string{"operation=Document.AddParagraph", "code=VALIDATION_ERROR", "invalid paragraph", "cause=underlying error"},
		},
		{
			name: "minimal error",
			err: &DocxError{
				Message: "something went wrong",
			},
			contains: []string{"something went wrong"},
		},
		{
			name: "error with op and code only",
			err: &DocxError{
				Code: ErrCodeNotFound,
				Op:   "GetStyle",
			},
			contains: []string{"operation=GetStyle", "code=NOT_FOUND"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("Error() = %q; expected to contain %q", result, substr)
				}
			}
		})
	}
}

func TestDocxError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &DocxError{
		Code: ErrCodeInternal,
		Op:   "TestOp",
		Err:  underlying,
	}

	unwrapped := err.Unwrap()
	if unwrapped != underlying {
		t.Errorf("Unwrap() = %v; want %v", unwrapped, underlying)
	}
}

func TestDocxError_Is(t *testing.T) {
	err1 := &DocxError{Code: ErrCodeValidation}
	err2 := &DocxError{Code: ErrCodeValidation}
	err3 := &DocxError{Code: ErrCodeNotFound}
	err4 := errors.New("other error")

	if !err1.Is(err2) {
		t.Error("Expected err1.Is(err2) to be true")
	}

	if err1.Is(err3) {
		t.Error("Expected err1.Is(err3) to be false")
	}

	if err1.Is(err4) {
		t.Error("Expected err1.Is(err4) to be false")
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name: "with message",
			err: &ValidationError{
				Field:      "fontSize",
				Value:      -1,
				Constraint: "must be positive",
				Message:    "font size cannot be negative",
			},
			expected: "validation error: field=fontSize, value=-1, constraint=must be positive, message=font size cannot be negative",
		},
		{
			name: "without message",
			err: &ValidationError{
				Field:      "color",
				Value:      "invalid",
				Constraint: "must be hex",
			},
			expected: "validation error: field=color, value=invalid, constraint=must be hex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q; want %q", result, tt.expected)
			}
		})
	}
}

func TestBuilderError(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		be := &BuilderError{}
		if be.HasError() {
			t.Error("Expected HasError() to be false")
		}
		if be.Get() != nil {
			t.Error("Expected Get() to be nil")
		}
		if be.Error() != "no error" {
			t.Errorf("Error() = %q; want %q", be.Error(), "no error")
		}
	})

	t.Run("with error", func(t *testing.T) {
		be := &BuilderError{}
		testErr := errors.New("test error")
		be.Set(testErr)

		if !be.HasError() {
			t.Error("Expected HasError() to be true")
		}
		if be.Get() != testErr {
			t.Errorf("Get() = %v; want %v", be.Get(), testErr)
		}
		if be.Error() != "test error" {
			t.Errorf("Error() = %q; want %q", be.Error(), "test error")
		}
	})

	t.Run("set only first error", func(t *testing.T) {
		be := &BuilderError{}
		err1 := errors.New("first error")
		err2 := errors.New("second error")

		be.Set(err1)
		be.Set(err2)

		if be.Get() != err1 {
			t.Errorf("Get() = %v; want %v", be.Get(), err1)
		}
	})

	t.Run("unwrap", func(t *testing.T) {
		testErr := errors.New("test error")
		be := &BuilderError{err: testErr}

		unwrapped := be.Unwrap()
		if unwrapped != testErr {
			t.Errorf("Unwrap() = %v; want %v", unwrapped, testErr)
		}
	})
}

func TestErrorf(t *testing.T) {
	err := Errorf(ErrCodeValidation, "TestOp", "value %d is invalid", 42)

	docxErr, ok := err.(*DocxError)
	if !ok {
		t.Fatal("Expected *DocxError")
	}

	if docxErr.Code != ErrCodeValidation {
		t.Errorf("Code = %s; want %s", docxErr.Code, ErrCodeValidation)
	}
	if docxErr.Op != "TestOp" {
		t.Errorf("Op = %s; want %s", docxErr.Op, "TestOp")
	}
	if !strings.Contains(docxErr.Message, "42") {
		t.Errorf("Message = %s; expected to contain '42'", docxErr.Message)
	}
}

func TestWrap(t *testing.T) {
	t.Run("wrap error", func(t *testing.T) {
		underlying := errors.New("underlying")
		err := Wrap(underlying, "TestOp")

		docxErr, ok := err.(*DocxError)
		if !ok {
			t.Fatal("Expected *DocxError")
		}

		if docxErr.Op != "TestOp" {
			t.Errorf("Op = %s; want %s", docxErr.Op, "TestOp")
		}
		if docxErr.Err != underlying {
			t.Errorf("Err = %v; want %v", docxErr.Err, underlying)
		}
	})

	t.Run("wrap nil", func(t *testing.T) {
		err := Wrap(nil, "TestOp")
		if err != nil {
			t.Errorf("Wrap(nil) = %v; want nil", err)
		}
	})
}

func TestWrapWithCode(t *testing.T) {
	t.Run("wrap with code", func(t *testing.T) {
		underlying := errors.New("underlying")
		err := WrapWithCode(underlying, ErrCodeIO, "TestOp")

		docxErr, ok := err.(*DocxError)
		if !ok {
			t.Fatal("Expected *DocxError")
		}

		if docxErr.Code != ErrCodeIO {
			t.Errorf("Code = %s; want %s", docxErr.Code, ErrCodeIO)
		}
		if docxErr.Op != "TestOp" {
			t.Errorf("Op = %s; want %s", docxErr.Op, "TestOp")
		}
		if docxErr.Err != underlying {
			t.Errorf("Err = %v; want %v", docxErr.Err, underlying)
		}
	})

	t.Run("wrap nil with code", func(t *testing.T) {
		err := WrapWithCode(nil, ErrCodeIO, "TestOp")
		if err != nil {
			t.Errorf("WrapWithCode(nil) = %v; want nil", err)
		}
	})
}

func TestWrapWithContext(t *testing.T) {
	t.Run("wrap with context", func(t *testing.T) {
		underlying := errors.New("underlying")
		ctx := map[string]interface{}{"key": "value"}
		err := WrapWithContext(underlying, "TestOp", ctx)

		docxErr, ok := err.(*DocxError)
		if !ok {
			t.Fatal("Expected *DocxError")
		}

		if docxErr.Op != "TestOp" {
			t.Errorf("Op = %s; want %s", docxErr.Op, "TestOp")
		}
		if docxErr.Context["key"] != "value" {
			t.Errorf("Context[key] = %v; want 'value'", docxErr.Context["key"])
		}
	})

	t.Run("wrap nil with context", func(t *testing.T) {
		ctx := map[string]interface{}{"key": "value"}
		err := WrapWithContext(nil, "TestOp", ctx)
		if err != nil {
			t.Errorf("WrapWithContext(nil) = %v; want nil", err)
		}
	})
}

func TestNotFound(t *testing.T) {
	err := NotFound("GetStyle", "Heading1")

	docxErr, ok := err.(*DocxError)
	if !ok {
		t.Fatal("Expected *DocxError")
	}

	if docxErr.Code != ErrCodeNotFound {
		t.Errorf("Code = %s; want %s", docxErr.Code, ErrCodeNotFound)
	}
	if !strings.Contains(docxErr.Message, "not found") {
		t.Errorf("Message should contain 'not found'")
	}
}

func TestInvalidState(t *testing.T) {
	err := InvalidState("Build", "document is empty")

	docxErr, ok := err.(*DocxError)
	if !ok {
		t.Fatal("Expected *DocxError")
	}

	if docxErr.Code != ErrCodeInvalidState {
		t.Errorf("Code = %s; want %s", docxErr.Code, ErrCodeInvalidState)
	}
	if docxErr.Message != "document is empty" {
		t.Errorf("Message = %s; want 'document is empty'", docxErr.Message)
	}
}

func TestValidation(t *testing.T) {
	err := Validation("fontSize", -1, "must be positive", "negative values not allowed")

	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatal("Expected *ValidationError")
	}

	if valErr.Field != "fontSize" {
		t.Errorf("Field = %s; want 'fontSize'", valErr.Field)
	}
	if valErr.Value != -1 {
		t.Errorf("Value = %v; want -1", valErr.Value)
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("AddParagraph", "text", nil, "text is required")

	docxErr, ok := err.(*DocxError)
	if !ok {
		t.Fatal("Expected *DocxError")
	}

	if docxErr.Code != ErrCodeValidation {
		t.Errorf("Code = %s; want %s", docxErr.Code, ErrCodeValidation)
	}

	valErr, ok := docxErr.Err.(*ValidationError)
	if !ok {
		t.Fatal("Expected underlying *ValidationError")
	}

	if valErr.Field != "text" {
		t.Errorf("Field = %s; want 'text'", valErr.Field)
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("GetStyle", "styleName", "Heading1", "style does not exist")

	docxErr, ok := err.(*DocxError)
	if !ok {
		t.Fatal("Expected *DocxError")
	}

	if docxErr.Code != ErrCodeNotFound {
		t.Errorf("Code = %s; want %s", docxErr.Code, ErrCodeNotFound)
	}
	if !strings.Contains(docxErr.Message, "Heading1") {
		t.Errorf("Message should contain 'Heading1'")
	}
}

func TestInvalidArgument(t *testing.T) {
	err := InvalidArgument("SetColor", "hex", "GGGGGG", "invalid hex color")

	docxErr, ok := err.(*DocxError)
	if !ok {
		t.Fatal("Expected *DocxError")
	}

	if docxErr.Code != ErrCodeValidation {
		t.Errorf("Code = %s; want %s", docxErr.Code, ErrCodeValidation)
	}

	valErr, ok := docxErr.Err.(*ValidationError)
	if !ok {
		t.Fatal("Expected underlying *ValidationError")
	}

	if valErr.Field != "hex" {
		t.Errorf("Field = %s; want 'hex'", valErr.Field)
	}
}

func TestUnsupported(t *testing.T) {
	err := Unsupported("AddChart", "3D charts")

	docxErr, ok := err.(*DocxError)
	if !ok {
		t.Fatal("Expected *DocxError")
	}

	if docxErr.Code != ErrCodeUnsupported {
		t.Errorf("Code = %s; want %s", docxErr.Code, ErrCodeUnsupported)
	}
	if !strings.Contains(docxErr.Message, "not supported") {
		t.Errorf("Message should contain 'not supported'")
	}
}
