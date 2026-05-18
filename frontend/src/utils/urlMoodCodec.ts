/**
 * URL query encoding / decoding for the `/discover/result` route.
 *
 * Wire format example: `m=2,3&s=0,1&so=1,1&t=0,0` — two comma-separated
 * integers per axis. A skipped answer is the literal string `-`.
 *
 * The codec is the contract between `/discover` (writer) and
 * `/discover/result` (reader). Changing key names, ordering, or the
 * skip token is a wire-format break.
 */

import { AXES, AXIS_KEYS, type AxisKey } from '../constants/discoverAxes';

/**
 * The user's survey answers as either an answer index (0..option_count-1)
 * or `null` (skipped).
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
 * values outside the legal range. Callers should `navigate('/discover',
 * { replace: true })` when this returns null.
 */
export function parseSearchParams(params: URLSearchParams): UserMoodInput | null {
  const out: Partial<Record<AxisKey, [number | null, number | null]>> = {};
  for (const axis of AXIS_KEYS) {
    const raw = params.get(QUERY_KEYS[axis]);
    if (raw === null) return null;
    const tokens = raw.split(',');
    if (tokens.length !== 2) return null;
    const parsed: (number | null)[] = [];
    for (const token of tokens) {
      if (token === SKIP_TOKEN) {
        parsed.push(null);
        continue;
      }
      if (!/^\d+$/.test(token)) return null;
      const n = Number.parseInt(token, 10);
      if (!Number.isInteger(n) || n < 0 || n >= AXES[axis].options.length) return null;
      parsed.push(n);
    }
    out[axis] = parsed as [number | null, number | null];
  }
  // Every axis is set because AXIS_KEYS was fully iterated above; assert via cast.
  return {
    mood: out.mood ?? [null, null],
    skill: out.skill ?? [null, null],
    social: out.social ?? [null, null],
    theme: out.theme ?? [null, null],
  };
}

/** Convenience: validate a `UserMoodInput` shape (used by tests and reducer). */
export function isFullyAnswered(mood: UserMoodInput): boolean {
  for (const axis of AXIS_KEYS) {
    if (mood[axis][0] === null && mood[axis][1] === null) return false;
  }
  return true;
}
