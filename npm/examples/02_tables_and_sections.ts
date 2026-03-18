/**
 * Example 02: Tables and Sections
 *
 * Creates a product catalog with tables, page breaks, and landscape sections.
 * Equivalent to Go example examples/02_intermediate/main.go
 *
 * Run: npx tsx npm/examples/02_tables_and_sections.ts
 */
import { DocumentBuilder } from '../src';

async function main() {
  const doc = new DocumentBuilder();

  try {
    await doc
      .setOptions({
        title: 'Product Catalog',
        author: 'ACME Corporation',
        subject: 'Q1 2025 Product Lineup',
        pageSize: 'Letter',
      })

      // Cover page
      .addHeading('ACME Corporation', 1)
      .addParagraph('Product Catalog 2025', { fontSize: 16, color: '#666666' })
      .addPageBreak()

      // Product table
      .addHeading('Electronics', 2)
      .addParagraph('Our electronics lineup features cutting-edge technology.')
      .addTable(
        [
          ['Product', 'Price', 'Stock'],
          ['Laptop Pro 15"', '$1,299', 'In Stock'],
          ['Wireless Mouse', '$49', 'In Stock'],
          ['USB-C Hub', '$79', 'Low Stock'],
          ['4K Monitor', '$599', 'Pre-order'],
        ],
      )

      .addParagraph('')

      // Second table
      .addHeading('Home & Garden', 2)
      .addTable(
        [
          ['Product', 'Price', 'Rating'],
          ['Smart Thermostat', '$199', '★★★★★'],
          ['LED Grow Light', '$89', '★★★★'],
          ['Robot Vacuum', '$449', '★★★★★'],
        ],
      )

      // Landscape section for wide table
      .addSection({ orientation: 'landscape' })
      .addHeading('Full Inventory (Landscape)', 2)
      .addTable(
        [
          ['SKU', 'Product', 'Category', 'Price', 'Stock', 'Warehouse', 'Supplier'],
          ['E-001', 'Laptop Pro 15"', 'Electronics', '$1,299', '142', 'West', 'TechCorp'],
          ['E-002', 'Wireless Mouse', 'Electronics', '$49', '500', 'East', 'PeriphCo'],
          ['H-001', 'Smart Thermostat', 'Home', '$199', '89', 'Central', 'SmartHome'],
          ['H-002', 'Robot Vacuum', 'Home', '$449', '34', 'West', 'CleanBot'],
        ],
      )

      .createToFile('/tmp/node_02_tables.docx');

    console.log('✓ Created /tmp/node_02_tables.docx');
  } finally {
    doc.dispose();
  }
}

main().catch(console.error);
