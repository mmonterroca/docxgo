/**
 * Example 04: Template / Mail Merge
 *
 * Creates a template with {{placeholders}}, inspects them, and renders
 * personalized documents.
 * Equivalent to Go example examples/14_mail_merge/main.go
 *
 * Run: npx tsx npm/examples/04_template.ts
 */
import { DocumentBuilder } from '../src';

async function main() {
  const doc = new DocumentBuilder();

  try {
    // Step 1: Create a template document
    console.log('Step 1: Creating invoice template...');
    await doc
      .setOptions({ title: 'Invoice Template', pageSize: 'A4' })
      .addHeading('INVOICE', 1)
      .addParagraph('')
      .addFormattedParagraph([
        { text: 'From: ', bold: true },
        { text: '{{company_name}}' },
      ])
      .addParagraph('{{company_address}}')
      .addParagraph('')
      .addFormattedParagraph([
        { text: 'To: ', bold: true },
        { text: '{{customer_name}}' },
      ])
      .addParagraph('{{customer_email}}')
      .addParagraph('')
      .addFormattedParagraph([
        { text: 'Invoice #: ', bold: true },
        { text: '{{invoice_number}}' },
      ])
      .addFormattedParagraph([
        { text: 'Date: ', bold: true },
        { text: '{{invoice_date}}' },
      ])
      .addParagraph('')
      .addTable([
        ['Description', 'Amount'],
        ['{{item_1}}', '{{price_1}}'],
        ['{{item_2}}', '{{price_2}}'],
        ['Total', '{{total}}'],
      ])
      .addParagraph('')
      .addParagraph('Thank you for your business, {{customer_name}}!')
      .createToFile('/tmp/node_04_template.docx');

    console.log('  ✓ Created /tmp/node_04_template.docx');

    // Step 2: Open and inspect placeholders
    console.log('\nStep 2: Inspecting template...');
    doc.reset();
    await doc.open('/tmp/node_04_template.docx');

    const placeholders = await doc.inspectTemplate();
    console.log(`  Found ${placeholders.count} placeholders (${placeholders.occurrences} occurrences):`);
    for (const name of placeholders.placeholders) {
      console.log(`    - {{${name}}}`);
    }
    await doc.closeDocument();

    // Step 3: Render with data
    console.log('\nStep 3: Rendering invoice...');
    doc.reset();
    await doc.open('/tmp/node_04_template.docx');

    await doc.renderTemplate({
      company_name: 'Acme Corporation',
      company_address: '123 Main Street, Springfield, IL 62701',
      customer_name: 'John Smith',
      customer_email: 'john.smith@example.com',
      invoice_number: 'INV-2025-001',
      invoice_date: 'March 17, 2026',
      item_1: 'Consulting Services',
      price_1: '$2,500.00',
      item_2: 'Software License',
      price_2: '$1,200.00',
      total: '$3,700.00',
    });

    await doc.saveToFile('/tmp/node_04_invoice.docx');
    console.log('  ✓ Created /tmp/node_04_invoice.docx');

    // Verify
    doc.reset();
    await doc.open('/tmp/node_04_invoice.docx');
    const paras = await doc.listParagraphs();
    const hasPlaceholders = paras.paragraphs.some(p => p.text?.includes('{{'));
    console.log(`  Placeholders remaining: ${hasPlaceholders ? 'YES (error!)' : 'none ✓'}`);
    await doc.closeDocument();

    console.log('\n✓ Template workflow complete!');
  } finally {
    doc.dispose();
  }
}

main().catch(console.error);
