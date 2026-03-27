const base =
  'px-3 py-1.5 text-sm font-medium rounded disabled:opacity-70 disabled:cursor-not-allowed disabled:saturate-50 mx-1.5 transition-[transform,box-shadow] duration-150 active:scale-95 hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-400 focus-visible:ring-offset-2 focus-visible:ring-offset-black min-h-[44px]';

/** Primary (blue) button style. */
export const btnPrimary = `${base} text-white bg-blue-600 hover:bg-blue-700`;
/** Warning (yellow) button style. */
export const btnWarning = `${base} text-gray-900 bg-yellow-400 hover:bg-yellow-500`;
/** Success (green) button style. */
export const btnSuccess = `${base} text-white bg-green-600 hover:bg-green-700`;
/** Danger (red) button style. */
export const btnDanger = `${base} text-white bg-red-600 hover:bg-red-700`;
/** Secondary (gray) button style. */
export const btnSecondary = `${base} text-white bg-gray-600 hover:bg-gray-500`;

/** Poker primary (emerald) — call/check action. */
export const btnPokerPrimary = `${base} text-white bg-emerald-600 hover:bg-emerald-700`;
/** Poker accent (sky) — raise/bet action. */
export const btnPokerAccent = `${base} text-white bg-sky-500 hover:bg-sky-600`;
/** Poker all-in (amber) — high-stakes action. */
export const btnPokerAllIn = `${base} text-white bg-amber-500 hover:bg-amber-600`;
/** Poker muted (gray) — fold action. */
export const btnPokerMuted = `${base} text-white bg-gray-500 hover:bg-gray-600`;
/** Outline button — minimal visual weight for reset etc. */
export const btnOutline = `${base} text-gray-300 border border-gray-500 bg-transparent hover:bg-gray-700`;

/** White focus ring style for keyboard navigation. */
export const focusRingWhite = 'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/80';

/** Blue focus ring style for keyboard navigation. */
export const focusRingBlue =
  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-400 focus-visible:ring-offset-2 focus-visible:ring-offset-black';
