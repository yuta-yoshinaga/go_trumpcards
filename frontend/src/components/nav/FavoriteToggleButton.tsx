import { useTranslation } from 'react-i18next';

/** Props for {@link FavoriteToggleButton}. */
export interface FavoriteToggleButtonProps {
  /** Game route path (e.g. "/blackjack") used as the favorite key passed to onToggle. */
  path: string;
  /** Translated game name included in this button’s accessible name. */
  gameLabel: string;
  /** Whether the path is currently a favorite. Drives the ★/☆ glyph and aria-pressed. */
  pressed: boolean;
  /** Called when the user activates the button; receives the path. */
  onToggle: (path: string) => void;
  /**
   * Class string for the button. Receives the current pressed state so callers
   * can vary visual treatment (e.g. NavBar fades out unpressed ☆ while
   * DesktopSidebar keeps both ★/☆ in the accent color).
   */
  className: (pressed: boolean) => string;
}

/**
 * Toggle button for marking a game as a favorite. Implements the
 * WAI-ARIA toggle button pattern (`aria-pressed` reflects the current
 * state) so screen readers announce "toggle button, pressed" and voice
 * control / a11y test tooling treats it as a toggle, not a plain button.
 *
 * Presentational: the parent owns favorites state via `useFavoriteGames`
 * and threads `pressed` / `onToggle` here so a single hook instance
 * drives both the button and any sibling list (e.g., the "favorites"
 * section in DesktopSidebar / NavBar).
 */
export function FavoriteToggleButton({ path, gameLabel, pressed, onToggle, className }: FavoriteToggleButtonProps) {
  const { t } = useTranslation('common');
  // Include the game name so each button is identifiable, but leave its
  // favorite state to aria-pressed so the state is not announced twice.
  return (
    <button
      type="button"
      aria-label={t('nav.favoriteToggle', { game: gameLabel })}
      aria-pressed={pressed}
      onClick={() => onToggle(path)}
      className={className(pressed)}
    >
      {pressed ? '★' : '☆'}
    </button>
  );
}
