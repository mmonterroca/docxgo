/**
 * Example 06: Document Language
 *
 * Demonstrates document.setLanguage (via applyPatch) on a freshly created
 * document, and the round-trip guard that applies once a document has been
 * saved and reopened.
 *
 * Run: npx tsx npm/examples/06_language.ts
 */
import { DocumentBuilder, DocxgoError } from '../src';

async function main() {
  const doc = new DocumentBuilder();

  try {
    // Step 1: Create a document and set its proofing language.
    // create() tracks the new document's ID internally, so applyPatch()
    // can target it directly — no open()/openFromBase64() round-trip needed.
    console.log('Step 1: Creating document and setting language...');
    await doc
      .setOptions({ title: 'Documento en Español' })
      .addHeading('Bienvenido', 1)
      .addParagraph('Este documento usa el idioma español de México.')
      .create();

    const patchResult = await doc.applyPatch([{ op: 'setLanguage', val: 'es-MX' }]);
    console.log(`  ✓ Language set (applied: ${patchResult.applied})`);

    const info = await doc.inspect();
    console.log(`  Language: ${JSON.stringify(info.language)}`);

    await doc.saveToFile('/tmp/node_06_language.docx');
    console.log('  ✓ Saved /tmp/node_06_language.docx');
    await doc.closeDocument();

    // Step 2: The round-trip guard — setLanguage rejects documents opened
    // via open()/openFromBase64()/openFromBuffer(), since their
    // styles.xml/settings.xml are preserved verbatim from the source file.
    console.log('\nStep 2: Attempting setLanguage on a reopened document...');
    doc.reset();
    await doc.open('/tmp/node_06_language.docx');

    const reopenedInfo = await doc.inspect();
    console.log(`  Language survives the round-trip: ${JSON.stringify(reopenedInfo.language)}`);

    try {
      await doc.applyPatch([{ op: 'setLanguage', val: 'en-US' }]);
      console.log('  ✗ Unexpected: setLanguage should have failed on a reopened document');
    } catch (err) {
      const message = err instanceof DocxgoError ? err.message : String(err);
      console.log(`  ✓ Rejected as expected: ${message}`);
    }
    await doc.closeDocument();

    console.log('\n✓ Language workflow complete!');
  } finally {
    doc.dispose();
  }
}

main().catch(console.error);
