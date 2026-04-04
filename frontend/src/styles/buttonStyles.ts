const base =
  'px-3 py-1.5 text-sm font-medium rounded disabled:opacity-70 disabled:cursor-not-allowed disabled:saturate-50 mx-1.5 transition-[transform,box-shadow] duration-150 active:scale-95 hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ds-accent focus-visible:ring-offset-2 focus-visible:ring-offset-black min-h-[44px]';

/** Primary (gold accent) button style. */
export const btnPrimary = `${base} text-ds-text-on-accent bg-ds-accent hover:bg-ds-accent-hover`;
/** Warning (orange) button style. */
export const btnWarning = `${base} text-ds-text-on-accent bg-ds-warning hover:bg-ds-warning-hover`;
/** Success (green) button style. */
export const btnSuccess = `${base} text-white bg-ds-success hover:bg-ds-success-hover`;
/** Danger (red) button style. */
export const btnDanger = `${base} text-white bg-ds-error hover:bg-ds-error-hover`;
/** Secondary (surface) button style. */
export const btnSecondary = `${base} text-ds-text-primary bg-ds-surface-elevated hover:bg-ds-surface-elevated-hover`;

/** Poker primary (emerald) — call/check action. */
export const btnPokerPrimary = `${base} text-white bg-emerald-600 hover:bg-emerald-700`;
/** Poker accent (sky) — raise/bet action. */
export const btnPokerAccent = `${base} text-white bg-sky-500 hover:bg-sky-600`;
/** Poker all-in (amber) — high-stakes action. */
export const btnPokerAllIn = `${base} text-white bg-amber-500 hover:bg-amber-600`;
/** Poker muted (gray) — fold action. */
export const btnPokerMuted = `${base} text-white bg-gray-500 hover:bg-gray-600`;
/** Outline button — minimal visual weight for reset etc. */
export const btnOutline = `${base} text-ds-text-muted border border-ds-border-subtle bg-transparent hover:bg-ds-surface-elevated`;

/** Gold accent focus ring style for keyboard navigation. */
export const focusRingAccent =
  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ds-accent focus-visible:ring-offset-2 focus-visible:ring-offset-black';

/** White focus ring style for keyboard navigation. */
export const focusRingWhite = 'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/80';
