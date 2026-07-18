#!/usr/bin/env node
// build-esm.js — Convert CJS output to ESM index.mjs
//
// After `tsc` produces CommonJS output in dist/, this script generates
// dist/index.mjs that re-exports everything with ESM syntax so that
// bundlers and ESM consumers can import the package.

const fs = require('fs');
const path = require('path');

const distDir = path.join(__dirname, '..', 'dist');
const cjsIndex = path.join(distDir, 'index.js');

if (!fs.existsSync(cjsIndex)) {
  console.error('Error: dist/index.js not found. Run "tsc" first.');
  process.exit(1);
}

const esmContent = `// Auto-generated ESM wrapper — do not edit
import mod from './index.js';
export const {
  DocxgoRPC,
  DocxgoExec,
  DocxgoError,
  DocumentBuilder,
  resolveBinary,
} = mod;
export default mod;
`;

const outFile = path.join(distDir, 'index.mjs');
fs.writeFileSync(outFile, esmContent, 'utf8');
console.log('Created dist/index.mjs');
