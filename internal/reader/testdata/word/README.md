# testdata/word

Fixtures in this directory are `.docx` files not produced by docxgo's own
writer, used to test the reader against input shaped the way real
applications (chiefly Microsoft Word) actually emit it -- see
`internal/reader/reader_test.go`'s `buildRawZipPackage` fixtures for the
same reasoning applied to hand-authored raw XML instead of a binary file.

## issue-102-input.docx

Attached by the reporter to
[issue #102](https://github.com/mmonterroca/docxgo/issues/102) ("Formatting
issues in pre-authored documents"), authored in Microsoft Word. Minimal:
a title paragraph in all caps (`w:caps`), a table using the built-in
`TableGrid` style for its borders (`w:tblStyle`, not an explicit
`w:tblBorders`), and a mid-document section break with no explicit
`w:sectPr/w:type` (defaults to `nextPage` per the OOXML schema). Round-tripping
it through `OpenDocument` + `WriteTo` dropped all three -- see
`internal/reader/reader_test.go` and `docx_read_test.go` for the regression
tests built against it.

Shared voluntarily by the reporter as a minimal repro attached to a public
GitHub issue; contains no content beyond placeholder text ("Title: should be
all caps", "Copyright", "Author", "Organization", table cell placeholders).

## issue-116-input.docx

Attached by the reporter to
[issue #116](https://github.com/mmonterroca/docxgo/issues/116). It contains
markup that docxgo does not model directly, including `mc:AlternateContent`,
a body-level content control, floating-table properties and explicit border
values. The regression test verifies that an unedited open-and-save preserves
the original `word/document.xml` byte-for-byte.
