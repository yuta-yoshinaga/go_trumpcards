/** Base CSS class for player area containers. */
export const playerAreaBase = 'glass-panel rounded-[10px]';

/** Compact desktop card width for #11702; respects the #929 minimum of 80px. */
export const DESKTOP_COMPACT_CARD_WIDTH = 80;

/** Tailwind classes that constrain and center the card area on large-desktop screens. */
export const lgCardAreaConstraint = 'lg:max-w-6xl lg:mx-auto lg:w-full';

/** Tailwind classes for a 2-column grid on large-desktop: main game area (left) + info sidebar (right). */
export const lgTwoColGrid = 'lg:grid lg:grid-cols-[1fr_minmax(280px,30%)] lg:gap-4';
