import { spawn, ChildProcess } from 'child_process';
import { createInterface, Interface } from 'readline';
import { resolveBinary } from './binary';
import type { RPCRequest, RPCResponse, RPCError } from './types';

/**
 * Error thrown when a JSON-RPC call returns an error response.
 */
export class DocxgoError extends Error {
  readonly code: string;
  readonly operation?: string;
  readonly data?: Record<string, unknown>;

  constructor(err: RPCError) {
    super(err.message);
    this.name = 'DocxgoError';
    this.code = err.code;
    this.operation = err.operation;
    this.data = err.data;
  }
}

export interface DocxgoRPCOptions {
  /** Explicit path to the docxgo binary. */
  binaryPath?: string;
  /** Timeout in ms for each RPC call (default: 30000). */
  timeout?: number;
}

/**
 * Persistent JSON-RPC client that communicates with the docxgo binary
 * over stdin/stdout using newline-delimited JSON.
 *
 * Ideal for high-frequency usage, batch processing, and Lambda warm starts.
 *
 * @example
 * ```ts
 * const rpc = new DocxgoRPC();
 * const result = await rpc.call('document.create', {
 *   content: [{ type: 'paragraph', runs: [{ text: 'Hello!' }] }],
 *   output: 'buffer',
 * });
 * const buf = Buffer.from(result.data, 'base64');
 * rpc.close();
 * ```
 */
export class DocxgoRPC {
  private proc: ChildProcess;
  private rl: Interface;
  private pending = new Map<number, {
    resolve: (value: unknown) => void;
    reject: (reason: Error) => void;
    timer?: ReturnType<typeof setTimeout>;
  }>();
  private seq = 0;
  private closed = false;
  private defaultTimeout: number;

  constructor(options?: DocxgoRPCOptions) {
    const binPath = resolveBinary(options?.binaryPath);
    this.defaultTimeout = options?.timeout ?? 30_000;

    this.proc = spawn(binPath, ['rpc'], {
      stdio: ['pipe', 'pipe', 'pipe'],
    });

    if (!this.proc.stdout || !this.proc.stdin) {
      throw new Error('Failed to spawn docxgo rpc process');
    }

    this.rl = createInterface({ input: this.proc.stdout });

    this.rl.on('line', (line: string) => {
      if (!line.trim()) return;
      try {
        const resp: RPCResponse = JSON.parse(line);
        const entry = this.pending.get(resp.id);
        if (entry) {
          this.pending.delete(resp.id);
          if (entry.timer) clearTimeout(entry.timer);
          if (resp.error) {
            entry.reject(new DocxgoError(resp.error));
          } else {
            entry.resolve(resp.result);
          }
        }
      } catch {
        // Ignore non-JSON lines (e.g. stderr leak)
      }
    });

    this.proc.on('close', () => {
      this.closed = true;
      // Reject all pending requests
      for (const [, entry] of this.pending) {
        if (entry.timer) clearTimeout(entry.timer);
        entry.reject(new Error('docxgo process exited'));
      }
      this.pending.clear();
    });

    this.proc.on('error', (err: Error) => {
      this.closed = true;
      for (const [, entry] of this.pending) {
        if (entry.timer) clearTimeout(entry.timer);
        entry.reject(err);
      }
      this.pending.clear();
    });
  }

  /**
   * Send a JSON-RPC request and await the response.
   *
   * @param method The RPC method name (e.g. 'document.create').
   * @param params The method parameters.
   * @param timeout Optional per-call timeout in ms.
   * @returns The result value from the response.
   * @throws {DocxgoError} If the RPC returns an error.
   */
  call<T = unknown>(method: string, params?: Record<string, unknown>, timeout?: number): Promise<T> {
    if (this.closed) {
      return Promise.reject(new Error('DocxgoRPC is closed'));
    }

    return new Promise<T>((resolve, reject) => {
      const id = ++this.seq;
      const timeoutMs = timeout ?? this.defaultTimeout;

      const timer = timeoutMs > 0
        ? setTimeout(() => {
            this.pending.delete(id);
            reject(new Error(`RPC call ${method} timed out after ${timeoutMs}ms`));
          }, timeoutMs)
        : undefined;

      this.pending.set(id, {
        resolve: resolve as (value: unknown) => void,
        reject,
        timer,
      });

      const req: RPCRequest = { id, method, params };
      this.proc.stdin!.write(JSON.stringify(req) + '\n');
    });
  }

  /**
   * Gracefully close the RPC process.
   * Closes stdin which causes the Go process to exit.
   */
  close(): void {
    if (!this.closed) {
      this.proc.stdin?.end();
    }
  }

  /**
   * Kill the RPC process immediately.
   */
  kill(): void {
    if (!this.closed) {
      this.proc.kill('SIGTERM');
    }
  }

  /** Whether the RPC process has exited. */
  get isClosed(): boolean {
    return this.closed;
  }
}
