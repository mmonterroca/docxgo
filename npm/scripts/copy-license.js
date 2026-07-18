#!/usr/bin/env node
// copy-license.js — Copy the repo-root LICENSE into npm/ so it ships in the
// published tarball.
//
// The canonical LICENSE lives one directory up (outside what npm packs from
// npm/), so it's copied here on every prepack/publish rather than committed
// as a second, driftable copy.

const fs = require('fs');
const path = require('path');

const src = path.join(__dirname, '..', '..', 'LICENSE');
const dest = path.join(__dirname, '..', 'LICENSE');

if (!fs.existsSync(src)) {
  console.error('Error: repo-root LICENSE not found at', src);
  process.exit(1);
}

fs.copyFileSync(src, dest);
console.log('Copied LICENSE from repo root');
