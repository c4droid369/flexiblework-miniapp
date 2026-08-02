import { beforeEach, describe, expect, it, vi } from 'vitest';

// Stub Element Plus before request.ts imports it. request.ts calls
// ElMessage.error() on every error path; we don't care about the toast in
// unit tests, just the response/reject behavior.
vi.mock('element-plus', () => ({
  ElMessage: { error: vi.fn(), success: vi.fn(), warning: vi.fn(), info: vi.fn() },
}));

// The axios mock must run interceptors for real, otherwise the auth header
// and 401 refresh paths can't be exercised. State lives inside vi.hoisted
// because vi.mock factories are hoisted above module imports.
const h = vi.hoisted(() => {
  const reqHandlers: Array<(cfg: any) => any> = [];
  const resOkHandlers: Array<(resp: any) => any> = [];
  const resErrHandlers: Array<(err: any) => any> = [];
  const responseQueue: any[] = [];
  const errorQueue: any[] = [];

  function wrapConfig(config: any) {
    const headers: any = { ...(config?.headers || {}) };
    headers.set = (k: string, v: string) => {
      headers[k] = v;
    };
    return { ...config, headers };
  }

  function call(method: string) {
    return async (url: string, dataOrConfig?: any, maybeConfig?: any) => {
      const config =
        method === 'get' || method === 'delete' ? dataOrConfig || {} : maybeConfig || {};
      let cfg = wrapConfig(config);
      for (const fn of reqHandlers) {
        try {
          cfg = (await fn(cfg)) || cfg;
        } catch (e: any) {
          throw new Error(`reqHandler threw for ${method} ${url}: ${e?.message}`);
        }
      }

      const err = errorQueue.shift();
      if (err) {
        let e: any = err;
        for (const fn of resErrHandlers) {
          try {
            const recovered = await fn(e);
            // Real axios treats a non-throwing error handler that returns a
            // value as recovery: the chain unwinds and that value is the
            // resolved result of the original call.
            return recovered;
          } catch (newErr) {
            e = newErr;
          }
        }
        throw e;
      }
      let r = responseQueue.shift();
      if (r === undefined)
        throw new Error(`no mock response queued for ${method.toUpperCase()} ${url}`);
      for (const fn of resOkHandlers) {
        try {
          r = (await fn(r)) ?? r;
        } catch (e3: any) {
          throw new Error(`resOkHandler threw for ${method} ${url}: ${e3?.message}`);
        }
      }
      return r;
    };
  }

  // The mock instance mirrors real axios: it's callable (instance(config)
  // delegates to instance.get/post based on config.method) AND exposes
  // method shortcuts (.get, .post, ...). request.ts uses both forms.
  const inst: any = (configOrUrl: any, maybeConfig?: any) => {
    // instance(config) → instance.request(config); instance(url, config) → GET.
    if (typeof configOrUrl === 'string') {
      return inst.get(configOrUrl, maybeConfig || {});
    }
    // Treat as a config object — route by method.
    const m = (configOrUrl?.method || 'get').toLowerCase();
    return (inst[m] || inst.get)(configOrUrl?.url || '', configOrUrl);
  };
  inst.defaults = { headers: { common: {} } };
  inst.interceptors = {
    request: {
      use: (fn: any) => {
        reqHandlers.push(fn);
        return reqHandlers.length - 1;
      },
      eject: () => {},
    },
    response: {
      use: (ok: any, err?: any) => {
        resOkHandlers.push(ok);
        if (err) resErrHandlers.push(err);
        return resOkHandlers.length - 1;
      },
      eject: () => {},
    },
  };
  inst.create = () => inst;
  inst.get = vi.fn(call('get'));
  inst.post = vi.fn(call('post'));
  inst.put = vi.fn(call('post'));
  inst.delete = vi.fn(call('delete'));
  inst.request = vi.fn((cfg: any) => {
    const m = (cfg?.method || 'get').toLowerCase();
    return (inst[m] || inst.get)(cfg?.url || '', cfg);
  });

  return {
    inst,
    queueResponse: (r: any) => responseQueue.push(r),
    queueError: (e: any) => errorQueue.push(e),
    reset: () => {
      // Only flush the queues — DO NOT clear handler arrays. request.ts
      // registers its interceptor at module load, and we'd wipe it on
      // every beforeEach otherwise.
      responseQueue.length = 0;
      errorQueue.length = 0;
    },
  };
});

vi.mock('axios', () => ({ default: h.inst }));

import axios from 'axios';
import { clearTokens, setAccessToken, setRefreshToken } from '@/utils/auth';
import { get, post } from '@/utils/request';

const ax = axios as unknown as {
  get: ReturnType<typeof vi.fn>;
  post: ReturnType<typeof vi.fn>;
  delete: ReturnType<typeof vi.fn>;
};

describe('utils/request — envelope unwrapping', () => {
  beforeEach(() => {
    clearTokens();
    h.reset();
    vi.clearAllMocks();
  });

  it('returns data on code=0', async () => {
    h.queueResponse({ data: { code: 0, message: 'ok', data: { id: 7, name: 'alice' } } });
    const r = await get<{ id: number; name: string }>('/users/7');
    expect(r).toEqual({ id: 7, name: 'alice' });
  });

  it('rejects when envelope code != 0', async () => {
    h.queueResponse({ data: { code: 1001, message: 'bad request', data: null } });
    await expect(get('/x')).rejects.toThrow('bad request');
  });
});

describe('utils/request — auth header', () => {
  beforeEach(() => {
    clearTokens();
    h.reset();
    vi.clearAllMocks();
  });

  it('attaches Bearer when access token is set', async () => {
    setAccessToken('a-1');
    let observed: string | undefined;
    h.queueResponse({ data: { code: 0, message: 'ok', data: null } });
    // Probe via a request interceptor that captures the final cfg.
    // (We add this AFTER the real interceptor runs, but before the next call.)
    await get('/users1');

    // Add the probe and re-run a fresh call so it sees the configured chain.
    let probeObserved: string | undefined;
    const probe = (cfg: any) => {
      probeObserved = cfg?.headers?.Authorization ?? cfg?.headers?.authorization;
      return cfg;
    };
    // Push directly into the hoisted reqHandlers queue.
    h.inst.interceptors.request.use(probe);
    h.queueResponse({ data: { code: 0, message: 'ok', data: null } });
    await get('/users2');
    void observed;
    expect(probeObserved).toBe('Bearer a-1');
  });

  it('omits Authorization when no token', async () => {
    let probeObserved: string | undefined = 'sentinel';
    const probe = (cfg: any) => {
      probeObserved = cfg?.headers?.Authorization ?? cfg?.headers?.authorization;
      return cfg;
    };
    h.inst.interceptors.request.use(probe);
    h.queueResponse({ data: { code: 0, message: 'ok', data: null } });
    await get('/users');
    expect(probeObserved).toBeUndefined();
  });
});

describe('utils/request — 401 refresh', () => {
  beforeEach(() => {
    clearTokens();
    h.reset();
    vi.clearAllMocks();
  });

  it('redirects to /login when no refresh token', async () => {
    setAccessToken('stale');
    h.queueError({
      response: { status: 401 },
      config: { headers: { set: (_k: string, _v: string) => {} } },
      message: '401',
    });

    const origLocation = (window as any).location;
    Object.defineProperty(window, 'location', {
      writable: true,
      configurable: true,
      value: { ...origLocation, href: '' },
    });

    await expect(get('/x')).rejects.toBeTruthy();
    expect((window as any).location.href).toBe('/login');

    Object.defineProperty(window, 'location', { writable: true, value: origLocation });
  });

  it('calls /auth/refresh on 401 with the current refresh token', async () => {
    setAccessToken('stale');
    setRefreshToken('r-1');

    // Original GET → 401. config.headers.set must exist because the refresh
    // handler calls original.headers.set('Authorization', ...).
    h.queueError({
      response: { status: 401 },
      config: { headers: { set: (_k: string, _v: string) => {} } },
      message: '401',
    });
    // /auth/refresh → 200
    h.queueResponse({
      data: {
        code: 0,
        message: 'ok',
        data: {
          access_token: 'a-new',
          refresh_token: 'r-new',
          expires_in: 3600,
          token_type: 'Bearer',
          permissions: [],
        },
      },
    });
    // Retried GET → 200 — must succeed or the refresh handler will loop on
    // another 401 and queue a pending refresh that never resolves.
    h.queueResponse({ data: { code: 0, message: 'ok', data: 'OK' } });

    let r: string | undefined;
    try {
      r = await get<string>('/x');
    } catch (e: any) {
      throw new Error(
        `get() rejected with: ${e?.message}; ` +
          `get calls=${ax.get.mock.calls.length}; ` +
          `post calls=${ax.post.mock.calls.length}`,
      );
    }
    expect(r).toBe('OK');
    expect(ax.post).toHaveBeenCalledWith(
      expect.stringContaining('/auth/refresh'),
      expect.objectContaining({ refresh_token: 'r-1' }),
    );
  }, 5000);
});

describe('utils/request — typed helpers', () => {
  beforeEach(() => {
    h.reset();
    vi.clearAllMocks();
  });

  it('post passes body and unwraps envelope', async () => {
    h.queueResponse({ data: { code: 0, message: 'ok', data: { token: 'xyz' } } });
    const r = await post<{ token: string }>('/login', { user: 'a' });
    expect(r.token).toBe('xyz');
    expect(ax.post).toHaveBeenCalled();
  });
});
