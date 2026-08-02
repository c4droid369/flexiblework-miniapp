import { beforeEach, describe, expect, it } from 'vitest';
import {
  clearTokens,
  getAccessToken,
  getRefreshToken,
  setAccessToken,
  setRefreshToken,
} from '@/utils/auth';

describe('utils/auth', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('round-trips access token', () => {
    setAccessToken('a-1');
    expect(getAccessToken()).toBe('a-1');
  });

  it('round-trips refresh token', () => {
    setRefreshToken('r-1');
    expect(getRefreshToken()).toBe('r-1');
  });

  it('empty string clears (treated as no token)', () => {
    setAccessToken('a-1');
    setAccessToken('');
    expect(getAccessToken()).toBe('');
  });

  it('clearTokens removes both', () => {
    setAccessToken('a');
    setRefreshToken('r');
    clearTokens();
    expect(getAccessToken()).toBe('');
    expect(getRefreshToken()).toBe('');
  });

  it('returns empty string when nothing stored', () => {
    expect(getAccessToken()).toBe('');
    expect(getRefreshToken()).toBe('');
  });
});
