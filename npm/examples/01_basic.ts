/**
 * Example 01: Basic Document
 *
 * Creates a simple document with title, paragraphs, and formatted text.
 * Equivalent to Go example examples/01_basic/main.go
 *
 * Run: npx tsx npm/examples/01_basic.ts
 */
import { DocumentBuilder } from '../src';

async function main() {
  const doc = new DocumentBuilder();

  try {
    // Create a document with title, headings, and paragraphs
    await doc
      .setOptions({ title: 'Basic Document', author: 'docxgo Node.js', pageSize: 'A4' })
      .addHeading('Welcome to docxgo for Node.js', 1)
      .addParagraph('Creating Word documents from Node.js is simple with the DocumentBuilder API.')
      .addParagraph('')
      .addHeading('Formatted Text', 2)
      .addParagraph('This text is bold.', { bold: true })
      .addParagraph('This text is italic.', { italic: true })
      .addParagraph('This text is colored.', { color: '#2E74B5' })
      .addParagraph('This text is large.', { fontSize: 18 })
      .addParagraph('')
      .addHeading('Multiple Runs in One Paragraph', 2)
      .addFormattedParagraph([
        { text: 'You can mix ' },
        { text: 'bold', bold: true },
        { text: ', ' },
        { text: 'italic', italic: true },
        { text: ', and ' },
        { text: 'colored', color: '#FF0000' },
        { text: ' text in a single paragraph.' },
      ])
      .createToFile('/tmp/node_01_basic.docx');

    console.log('✓ Created /tmp/node_01_basic.docx');
  } finally {
    doc.dispose();
  }
}

main().catch(console.error);
