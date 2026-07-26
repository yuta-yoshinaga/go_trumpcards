/**
 * Auto-dismiss durations (ms) for transient toast / bubble notifications.
 *
 * Single source of truth so every "notice me, then disappear" notification
 * uses one of three consistent timings instead of ad-hoc per-component values
 * (issue #4313). Mirrored in the DESIGN.md Motion section.
 *
 * - `short`  — a glanceable one-liner (CPU action bubble).
 * - `medium` — a short status change worth a beat longer (meta-AI strategy).
 * - `long`   — multi-line content that takes longer to read (betting actions).
 */
export const TOAST_DURATION = {
  short: 2500,
  medium: 4000,
  long: 6000,
} as const;
