/**
 * Status / state badge style constants.
 *
 * **Use these for any badge that conveys meaningful information** (turn
 * status, forced action, error notice, warning prompt). They use opaque
 * design tokens for both background and foreground so the contrast
 * ratios documented in DESIGN.md are preserved on any game-table
 * background (felt-green, felt-blue, midnight-purple…).
 *
 * Do **not** mix Tailwind opacity suffixes (`/50`, `/80` …) with design
 * tokens for state badges — the resulting effective color depends on
 * whatever sits beneath, which silently breaks the WCAG AA / AAA
 * guarantees DESIGN.md makes.
 *
 * Decorative uses of opacity (glassmorphism, hover tints, animation
 * pulses) remain allowed; this file is specifically for state badges.
 */

const BADGE_BASE = 'rounded-lg py-2 px-3.5 text-xs border';

/** Neutral info badge — surface bg, primary text, subtle border. */
export const badgeInfo = `${BADGE_BASE} bg-ds-surface text-ds-text-primary border-ds-border-subtle`;

/** Success badge — surface bg, success text + border. */
export const badgeSuccess = `${BADGE_BASE} bg-ds-surface text-ds-success border-ds-success`;

/** Warning badge — surface bg, warning text + border. Use for "forced", "involuntary", or "restricted" states. */
export const badgeWarning = `${BADGE_BASE} bg-ds-surface text-ds-warning border-ds-warning`;

/** Error badge — surface bg, error text + border. Use for "fold", "out", "bust", or error notifications. */
export const badgeError = `${BADGE_BASE} bg-ds-surface text-ds-error border-ds-error`;
