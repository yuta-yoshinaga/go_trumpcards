/**
 * URL query encoding / decoding for the `/discover/result` route.
 *
 * Wire format example: `m=1,0&s=2,1&so=1,0&t=0,1` — two comma-separated
 * integers per axis. Each integer is the option index **within that
 * sub-question's options array** (not the master option list), so the
 * legal range differs across Q1 and Q2 for the same axis. A skipped
 * answer is the literal string `-`.
 *
 * The codec is the contract between `/discover` (writer) and
 * `/discover/result` (reader). Changing key names, ordering, or the
 * skip token is a wire-format break.
 */

import { AXES, AXIS_KEYS, type AxisKey } from '../constants/discoverAxes';

/**
 * The user's survey answers as either an option index inside the
 * sub-question's options array, or `null` (skipped).
 */
export interface UserMoodInput {
  readonly mood: readonly [number | null, number | null];
  readonly skill: readonly [number | null, number | null];
  readonly social: readonly [number | null, number | null];
  readonly theme: readonly [number | null, number | null];
}

/** URL query key for each axis. */
const QUERY_KEYS: Readonly<Record<AxisKey, string>> = {
  mood: 'm',
  skill: 's',
  social: 'so',
  theme: 't',
} as const;

const SKIP_TOKEN = '-';

/** Serialize a mood to the `m=...&s=...&so=...&t=...` form. */
export function encodeMood(mood: UserMoodInput): string {
  const parts: string[] = [];
  for (const axis of AXIS_KEYS) {
    const values = mood[axis].map((v) => (v === null ? SKIP_TOKEN : String(v))).join(',');
    parts.push(`${QUERY_KEYS[axis]}=${values}`);
  }
  return parts.join('&');
}

/** Parse a query-string fragment (no leading `?`) into a UserMoodInput, or null if invalid. */
export function parseMood(query: string): UserMoodInput | null {
  if (!query) return null;
  const trimmed = query.startsWith('?') ? query.slice(1) : query;
  const params = new URLSearchParams(trimmed);
  return parseSearchParams(params);
}

/**
 * Parse a `URLSearchParams` (e.g. from `useSearchParams()`) into a UserMoodInput.
 *
 * Returns `null` if any axis is missing, has the wrong arity, or contains
 * values outside the legal per-question range. Callers should
 * `navigate('/discover', { replace: true })` when this returns null.
 */
export function parseSearchParams(params: URLSearchParams): UserMoodInput | null {
  // Built up to a full Record over AXIS_KEYS; the loop returns null on the
  // first missing/invalid axis, so by exit every key is set.
  const out: Record<AxisKey, [number | null, number | null]> = {
    mood: [null, null],
    skill: [null, null],
    social: [null, null],
    theme: [null, null],
  };
  for (const axis of AXIS_KEYS) {
    const raw = params.get(QUERY_KEYS[axis]);
    if (raw === null) return null;
    const tokens = raw.split(',');
    if (tokens.length !== 2) return null;
    const parsed: [number | null, number | null] = [null, null];
    for (const qIdx of [0, 1] as const) {
      const token = tokens[qIdx];
      if (token === SKIP_TOKEN) continue;
      if (!/^\d+$/.test(token)) return null;
      const n = Number.parseInt(token, 10);
      const optCount = AXES[axis].questions[qIdx].options.length;
      if (!Number.isInteger(n) || n < 0 || n >= optCount) return null;
      parsed[qIdx] = n;
    }
    out[axis] = parsed;
  }
  return out;
}

/**
 * `true` when at least one survey question (any axis, any index) was
 * actually answered. The Result page uses this to decide whether to
 * show the editor's-pick hero (some signal) or the warm fallback hero
 * (zero signal). A partial-skip user — for example, 6/8 answered with
 * one axis fully skipped — still has signal and should see real
 * recommendations; the skipped axis is treated as neutral (0.5) by
 * `axisScore` rather than rejected at the page level.
 */
export function hasAnyAnswer(mood: UserMoodInput): boolean {
  for (const axis of AXIS_KEYS) {
    for (const a of mood[axis]) {
      if (a !== null) return true;
    }
  }
  return false;
}
