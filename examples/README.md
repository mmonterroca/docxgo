# go-docx Examples

This directory contains practical, runnable examples demonstrating the capabilities of the go-docx library.

## Running the Examples

Each example is a standalone Go program. To run any example:

```bash
# Navigate to the examples directory
cd examples

# Run an example
go run 01_hello_world.go

# The generated .docx file will be created in the current directory
```

## Available Examples

### 01_hello_world.go
The simplest possible document - creates a basic "Hello, World!" document.

**Run:**
```bash
go run 01_hello_world.go
```

### 02_formatted_text.go
Shows various text formatting options including bold, italic, colors, and sizes.

**Run:**
```bash
go run 02_formatted_text.go
```

### 03_table_of_contents.go
Professional document with automatic Table of Contents, headings, and page numbers.

**Run:**
```bash
go run 03_table_of_contents.go
```

**After opening in Word:**
- Press `Ctrl+A` then `F9` to update all fields
- Or right-click TOC → "Update Field" → "Update entire table"

## Testing All Examples

Run all examples at once:

```bash
for f in *.go; do go run "$f"; done
```

## Requirements

- Go 1.20 or higher
- Microsoft Word or LibreOffice to view generated files

## More Information

- [Complete API Documentation](../docs/API_DOCUMENTATION.md)
- [Project README](../README.md)
- [Demo Test](../demo_test.go)
