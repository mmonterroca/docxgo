import { execFileSync } from 'child_process';
import { resolveBinary } from './binary';
import type { RPCRequest, RPCResponse } from './types';
import { DocxgoError } from './rpc';

export interface DocxgoExecOptions {
  /** Explicit path to the docxgo binary. */
  binaryPath?: string;
  /** Timeout in ms (default: 30000). */
  timeout?: number;
}

/**
 * One-shot execution client. Each call spawns a new docxgo process,
 * sends one request, and returns the response.
 *
 * Ideal for: occasional calls, CI pipelines, simple scripts.
 *
 * @example
 * ```ts
 * const exec = new DocxgoExec();
 * const result = exec.call('document.create', {
 *   content: [{ type: 'paragraph', runs: [{ text: 'Hello!' }] }],
 *   output: 'buffer',
 * });
 * ```
 */
export class DocxgoExec {
  private binPath: string;
  private defaultTimeout: number;

  constructor(options?: DocxgoExecOptions) {
    this.binPath = resolveBinary(options?.binaryPath);
    this.defaultTimeout = options?.timeout ?? 30_000;
  }

  /**
   * Execute a single JSON-RPC request synchronously.
   *
   * @param method The RPC method name.
   * @param params The method parameters.
   * @param timeout Optional timeout in ms.
   * @returns The result value.
   * @throws {DocxgoError} If the RPC returns an error.
   */
  call<T = unknown>(method: string, params?: Record<string, unknown>, timeout?: number): T {
    const req: RPCRequest = { id: 1, method, params };
    const requestJSON = JSON.stringify(req);

    let stdout: string;
    try {
      stdout = execFileSync(this.binPath, ['exec', '--request', requestJSON], {
        timeout: timeout ?? this.defaultTimeout,
        maxBuffer: 50 * 1024 * 1024, // 50MB to support large base64 documents
        encoding: 'utf-8',
        stdio: ['pipe', 'pipe', 'pipe'],
      });
    } catch (err: unknown) {
      // execFileSync throws on non-zero exit codes, but the JSON response
      // is still available in stdout. Try to parse it for a proper error.
      const execErr = err as { stdout?: string; message?: string };
      if (execErr.stdout) {
        try {
          const resp: RPCResponse = JSON.parse(execErr.stdout.trim());
          if (resp.error) {
            throw new DocxgoError(resp.error);
          }
        } catch (parseErr) {
          if (parseErr instanceof DocxgoError) throw parseErr;
        }
      }
      throw err;
    }

    const resp: RPCResponse<T> = JSON.parse(stdout.trim());
    if (resp.error) {
      throw new DocxgoError(resp.error);
    }
    return resp.result as T;
  }
}
