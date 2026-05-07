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

/**
 * Neutral info badge — surface bg, primary text, subtle border.
 *
 * Intentionally uses `text-ds-text-primary` (10.1:1 AAA on surface) rather
 * than `text-ds-info`: the info token (#5B8FB9) hits only ~4.5:1 on the
 * surface background, right at the AA boundary, so we lean on the
 * border-coloured `border-ds-info` would have given for semantic signal
 * is replaced by `border-ds-border-subtle` — info badges read as quiet
 * notifications, not warnings.
 */
export const badgeInfo = `${BADGE_BASE} bg-ds-surface text-ds-text-primary border-ds-border-subtle`;

/**
 * Success badge — surface bg, success text (5.8:1 AA on surface) + matching
 * border. Use for "round won", "auto-go", or other positive confirmations.
 */
export const badgeSuccess = `${BADGE_BASE} bg-ds-surface text-ds-success border-ds-success`;

/**
 * Warning badge — surface bg, warning text (6.3:1 AA on surface) + matching
 * border. Use for "forced", "involuntary", or "restricted" states.
 */
export const badgeWarning = `${BADGE_BASE} bg-ds-surface text-ds-warning border-ds-warning`;

/**
 * Error badge — surface bg, primary text (10.1:1 AAA on surface) + error
 * border. Use for "fold", "out", "bust", or error notifications.
 *
 * Foreground is intentionally `text-ds-text-primary` rather than
 * `text-ds-error`: the error token (#B83A3A) only hits ~2.7:1 on the
 * surface background — well below WCAG AA — so the semantic signal
 * comes entirely from the coloured border while the message stays
 * fully readable.
 */
export const badgeError = `${BADGE_BASE} bg-ds-surface text-ds-text-primary border-ds-error`;
