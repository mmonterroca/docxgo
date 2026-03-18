/**
 * Example 03: Read, Modify, Save
 *
 * Demonstrates the open → inspect → modify → save workflow.
 * Equivalent to Go example examples/12_read_and_modify/main.go
 *
 * Run: npx tsx npm/examples/03_read_modify.ts
 */
import { DocumentBuilder } from '../src';

async function main() {
  const doc = new DocumentBuilder();

  try {
    // Step 1: Create a document to work with
    console.log('Step 1: Creating initial document...');
    await doc
      .setOptions({ title: 'Original Report', author: 'Alice' })
      .addHeading('Quarterly Report', 1)
      .addParagraph('This is the original content from Q1.')
      .addTable([
        ['Metric', 'Value'],
        ['Revenue', '$1.2M'],
        ['Growth', '15%'],
      ])
      .createToFile('/tmp/node_03_original.docx');
    console.log('  ✓ Created /tmp/node_03_original.docx');

    // Step 2: Open and inspect
    console.log('\nStep 2: Opening and inspecting...');
    doc.reset();
    await doc.open('/tmp/node_03_original.docx');

    const info = await doc.inspect();
    console.log(`  Title: ${info.metadata?.title ?? '(none)'}`);
    console.log(`  Paragraphs: ${info.paragraphCount}`);
    console.log(`  Tables: ${info.tableCount}`);

    const paras = await doc.listParagraphs();
    console.log('  Content:');
    for (const p of paras.paragraphs) {
      if (p.text) console.log(`    [${p.index}] ${p.text}`);
    }

    // Step 3: Modify — add new content
    console.log('\nStep 3: Adding new content...');
    doc
      .addHeading('Q2 Update', 2)
      .addParagraph('Business continued to grow in Q2.')
      .addTable([
        ['Metric', 'Value'],
        ['Revenue', '$1.5M'],
        ['Growth', '25%'],
      ]);

    // Step 4: Save modified version
    await doc.saveToFile('/tmp/node_03_modified.docx');
    console.log('  ✓ Saved /tmp/node_03_modified.docx');

    // Step 5: Verify
    console.log('\nStep 4: Verifying modified document...');
    doc.reset();
    await doc.open('/tmp/node_03_modified.docx');
    const info2 = await doc.inspect();
    console.log(`  Paragraphs: ${info2.paragraphCount} (was ${info.paragraphCount})`);
    console.log(`  Tables: ${info2.tableCount} (was ${info.tableCount})`);
    await doc.closeDocument();

    console.log('\n✓ Read-modify-save workflow complete!');
  } finally {
    doc.dispose();
  }
}

main().catch(console.error);
