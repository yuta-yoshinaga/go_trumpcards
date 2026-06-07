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

const BADGE_SIZE = 'rounded-lg py-2 px-3.5 text-xs';

/**
 * Opaque color tokens (1px border + surface bg + token foreground + token
 * border) for an **info** state badge that needs custom sizing.
 *
 * Combine with your own `rounded-*` / padding / `text-*` classes for compact
 * inline pills, where the full {@link badgeInfo} (which forces `rounded-lg
 * py-2 px-3.5`) would be too large. The contrast guarantees are identical.
 *
 * Intentionally uses `text-ds-text-primary` (10.1:1 AAA on surface) rather
 * than `text-ds-info`: the info token (#5B8FB9) hits only ~4.5:1 on the
 * surface background, right at the AA boundary, so the semantic signal comes
 * from the subtle border while the text stays fully readable.
 */
export const badgeInfoColors = 'border bg-ds-surface text-ds-text-primary border-ds-border-subtle';

/**
 * Opaque color tokens for a **success** state badge that needs custom sizing.
 * Success text is 5.8:1 (AA) on the surface background. See
 * {@link badgeInfoColors} for usage.
 */
export const badgeSuccessColors = 'border bg-ds-surface text-ds-success border-ds-success';

/**
 * Opaque color tokens for a **warning** state badge that needs custom sizing.
 * Warning text is 6.3:1 (AA) on the surface background. See
 * {@link badgeInfoColors} for usage.
 */
export const badgeWarningColors = 'border bg-ds-surface text-ds-warning border-ds-warning';

/**
 * Opaque color tokens for an **error** state badge that needs custom sizing.
 *
 * Foreground is intentionally `text-ds-text-primary` (10.1:1 AAA on surface)
 * rather than `text-ds-error`: the error token (#B83A3A) only hits ~2.7:1 on
 * the surface background — well below WCAG AA — so the semantic signal comes
 * entirely from the coloured border while the message stays fully readable.
 * See {@link badgeInfoColors} for usage.
 */
export const badgeErrorColors = 'border bg-ds-surface text-ds-text-primary border-ds-error';

/**
 * Neutral info badge — surface bg, primary text, subtle border. Use for quiet
 * notifications. For compact inline pills use {@link badgeInfoColors}.
 */
export const badgeInfo = `${BADGE_SIZE} ${badgeInfoColors}`;

/**
 * Success badge — surface bg, success text + matching border. Use for "round
 * won", "auto-go", or other positive confirmations.
 */
export const badgeSuccess = `${BADGE_SIZE} ${badgeSuccessColors}`;

/**
 * Warning badge — surface bg, warning text + matching border. Use for
 * "forced", "involuntary", or "restricted" states.
 */
export const badgeWarning = `${BADGE_SIZE} ${badgeWarningColors}`;

/**
 * Error badge — surface bg, primary text + error border. Use for "fold",
 * "out", "bust", or error notifications.
 */
export const badgeError = `${BADGE_SIZE} ${badgeErrorColors}`;
