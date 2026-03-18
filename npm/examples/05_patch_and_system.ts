/**
 * Example 05: Patch Operations and System Methods
 *
 * Demonstrates applyPatch for atomic-style document mutation,
 * system.batch for multiple requests, and metadata round-trip.
 *
 * Run: npx tsx npm/examples/05_patch_and_system.ts
 */
import { DocumentBuilder } from '../src';

async function main() {
  const doc = new DocumentBuilder();

  try {
    // System info
    console.log('System Info:');
    const ver = await doc.version();
    console.log(`  Version: ${ver.version} (protocol ${ver.protocolVersion})`);
    console.log(`  Platform: ${ver.platform}/${ver.arch}`);

    const caps = await doc.capabilities();
    const enabled = Object.entries(caps).filter(([, v]) => v).map(([k]) => k);
    console.log(`  Capabilities: ${enabled.join(', ')}`);

    // Batch: multiple requests in one round-trip
    console.log('\nBatch Request:');
    const batch = await doc.batch([
      { method: 'system.ping' },
      { method: 'system.version' },
    ]);
    console.log(`  ${batch.responses.length} responses received`);

    // Create a base document
    console.log('\nCreating base document...');
    await doc
      .setOptions({ title: 'Base Document', pageSize: 'A4' })
      .addHeading('Original Document', 1)
      .addParagraph('This document will be modified with patch operations.')
      .createToFile('/tmp/node_05_base.docx');
    console.log('  ✓ Created /tmp/node_05_base.docx');

    // Open and apply patch
    console.log('\nApplying patch operations...');
    doc.reset();
    await doc.open('/tmp/node_05_base.docx');

    const result = await doc.applyPatch([
      { op: 'appendParagraph', text: 'Added via patch operation', style: 'Heading2' },
      { op: 'appendParagraph', text: 'Another patched paragraph.' },
      { op: 'setMetadata', title: 'Patched Document', creator: 'Patch Engine' },
      { op: 'setBackgroundColor', color: '#F5F5F5' },
    ]);
    console.log(`  Applied ${result.applied} operations`);

    await doc.saveToFile('/tmp/node_05_patched.docx');
    console.log('  ✓ Saved /tmp/node_05_patched.docx');

    // Verify metadata round-trip
    console.log('\nVerifying metadata...');
    doc.reset();
    await doc.open('/tmp/node_05_patched.docx');
    const info = await doc.inspect();
    console.log(`  Title: "${info.metadata.title}"`);
    console.log(`  Creator: "${info.metadata.creator}"`);
    console.log(`  Paragraphs: ${info.paragraphCount}`);

    const valid = await doc.validate();
    console.log(`  Valid: ${valid.valid} (${valid.errors?.length ?? 0} errors)`);
    await doc.closeDocument();

    console.log('\n✓ Patch and system workflow complete!');
  } finally {
    doc.dispose();
  }
}

main().catch(console.error);
