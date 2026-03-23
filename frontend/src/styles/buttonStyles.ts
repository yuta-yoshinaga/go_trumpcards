const base =
  'px-3 py-1.5 text-sm font-medium rounded disabled:opacity-70 disabled:cursor-not-allowed disabled:saturate-50 mx-1.5 transition-transform duration-150 active:scale-95 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-400 focus-visible:ring-offset-2 focus-visible:ring-offset-black';

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

/** White focus ring style for keyboard navigation. */
export const focusRingWhite = 'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/80';

/** Blue focus ring style for keyboard navigation. */
export const focusRingBlue =
  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-400 focus-visible:ring-offset-2 focus-visible:ring-offset-black';
