/**
 * Selects board card images while excluding role=img containers that group other card images.
 * Standalone role=img elements remain included for procedurally rendered decks such as tarot.
 */
export const BOARD_CARD_SELECTOR = 'img, [role="img"]:not(:has(img, [role="img"]))';
