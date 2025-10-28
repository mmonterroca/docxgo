# Example 03: Table of Contents (TOC)

This example shows how to build a Table of Contents using go-docx v2. The document:

- Uses the fluent builder API for cover, chapters, and appendix content
- Inserts a TOC field configured for Heading 1 and Heading 2 entries with hyperlinks
- Seeds the TOC with a friendly placeholder so the initial document looks polished
- Generates chapters and sub-sections using heading styles that drive the TOC
- Adds an appendix with a styled resource list

When you open the generated `.docx` file in Word, press **F9** (or right-click the TOC → **Update Field**) to refresh the table with correct entries and page numbers.

## Running the Example

```bash
cd examples/03_toc
go run .
```

Output: `03_toc_demo.docx`

## Key Concepts

- **Heading Styles**: Use `Heading 1` and `Heading 2` so Word knows which paragraphs belong in the TOC.
- **TOC Field Options**: The example enables hyperlinks and limits levels to 1-2, which is typical for reports.
- **Placeholder Results**: Calling `SetResult` on the TOC field seeds a helpful message until Word recalculates the table.

## Next Steps

- Need page numbers, hyperlinks, and other field types? → See [Example 04: Fields](../04_fields/).
- Want an end-to-end report with headers, footers, and TOC? → See [Example 07: Advanced](../07_advanced/).

## Updating the TOC Manually

1. Open `03_toc_demo.docx` in Microsoft Word (or LibreOffice).
2. Press **Ctrl+A** to select the whole document.
3. Press **F9** (or right-click on the TOC and select **Update Field**).
4. Choose **Update entire table** for accurate page numbers.

You can now export to PDF, print, or continue editing with confidence that your TOC will stay synchronized with your headings.
