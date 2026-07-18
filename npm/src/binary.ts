import { existsSync } from 'fs';
import { join } from 'path';

/**
 * Platform/architecture combinations supported by docxgo.
 */
const PLATFORM_MAP: Record<string, string> = {
  'darwin-x64': '@mmonterroca/docxgo-darwin-x64',
  'darwin-arm64': '@mmonterroca/docxgo-darwin-arm64',
  'linux-x64': '@mmonterroca/docxgo-linux-x64',
  'linux-arm64': '@mmonterroca/docxgo-linux-arm64',
  'win32-x64': '@mmonterroca/docxgo-win32-x64',
};

/**
 * Resolves the path to the docxgo binary.
 *
 * Resolution order:
 * 1. DOCXGO_BIN environment variable (explicit override)
 * 2. Platform-specific optional dependency package
 * 3. "docxgo" in PATH (assumes user installed globally)
 *
 * @param customPath Optional explicit path to the binary.
 * @returns The resolved binary path.
 * @throws If no binary can be found.
 */
export function resolveBinary(customPath?: string): string {
  // 1. Explicit override
  if (customPath) {
    return customPath;
  }

  // 2. Environment variable
  const envPath = process.env.DOCXGO_BIN;
  if (envPath) {
    return envPath;
  }

  // 3. Platform-specific package
  const key = `${process.platform}-${process.arch}`;
  const pkg = PLATFORM_MAP[key];
  if (pkg) {
    try {
      // Try to resolve the platform package's binary
      const pkgDir = require.resolve(`${pkg}/package.json`);
      const binPath = join(pkgDir, '..', 'bin', 'docxgo');
      const binPathExe = binPath + (process.platform === 'win32' ? '.exe' : '');
      if (existsSync(binPathExe)) {
        return binPathExe;
      }
      if (existsSync(binPath)) {
        return binPath;
      }
    } catch {
      // Package not installed, fall through
    }
  }

  // 4. Check common local paths
  const localPaths = [
    join(process.cwd(), 'docxgo'),
    join(process.cwd(), 'docxgo.exe'),
    join(__dirname, '..', 'bin', 'docxgo'),
    join(__dirname, '..', 'bin', 'docxgo.exe'),
  ];
  for (const p of localPaths) {
    if (existsSync(p)) {
      return p;
    }
  }

  // 5. Assume it's in PATH
  return 'docxgo';
}
