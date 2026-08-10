import { describe, expect, it, vi } from 'vitest';
import { assertFloor } from './floor.mjs';

/**
 * `assertFloor` reports by exiting the process, so each case swaps `process.exit` for a throw
 * and asserts on whether it fired. Both directions are covered deliberately: a floor helper
 * that rejected everything would satisfy every "must fail" test on its own.
 */
function run(fn) {
  const exit = vi.spyOn(process, 'exit').mockImplementation((code) => {
    throw new Error(`exit:${code}`);
  });
  const err = vi.spyOn(console, 'error').mockImplementation(() => {});
  try {
    const value = fn();
    return { exited: false, value, messages: err.mock.calls.map((c) => c.join(' ')) };
  } catch (e) {
    if (!String(e.message).startsWith('exit:')) throw e;
    return { exited: true, code: e.message.slice(5), messages: err.mock.calls.map((c) => c.join(' ')) };
  } finally {
    exit.mockRestore();
    err.mockRestore();
  }
}

describe('assertFloor', () => {
  it('passes a count above the floor and returns it', () => {
    const r = run(() => assertFloor('demo', 264, 200, 'games'));
    expect(r.exited).toBe(false);
    expect(r.value).toBe(264);
  });

  it('passes a count exactly at the floor', () => {
    const r = run(() => assertFloor('demo', 200, 200, 'games'));
    expect(r.exited).toBe(false);
    expect(r.value).toBe(200);
  });

  it('exits 1 when the count is below the floor', () => {
    const r = run(() => assertFloor('demo', 12, 200, 'games'));
    expect(r.exited).toBe(true);
    expect(r.code).toBe('1');
    expect(r.messages.join('\n')).toContain('only 12 games found (expected at least 200)');
  });

  it('exits 1 when the walk found nothing at all', () => {
    const r = run(() => assertFloor('demo', 0, 1, 'files'));
    expect(r.exited).toBe(true);
  });

  it.each([0, -5, 1.5, Number.NaN])('rejects a floor of %s as asserting nothing', (min) => {
    const r = run(() => assertFloor('demo', 999, min, 'files'));
    expect(r.exited).toBe(true);
    expect(r.messages.join('\n')).toContain('floor must be a positive integer');
  });

  it('rejects a non-finite count rather than comparing it', () => {
    const r = run(() => assertFloor('demo', Number.NaN, 10, 'files'));
    expect(r.exited).toBe(true);
  });

  it('names the guard and what was counted, so the failure is actionable', () => {
    const r = run(() => assertFloor('trademark-terms', 3, 4000, 'files scanned'));
    expect(r.messages.join('\n')).toContain('trademark-terms');
    expect(r.messages.join('\n')).toContain('files scanned');
  });
});
