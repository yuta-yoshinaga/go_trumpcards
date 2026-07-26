import type { SettingsItem } from '../components/common/SettingsPanel';

/**
 * Builds the standard "frontend hint" checkbox item for a game's
 * {@link components/common/SettingsPanel.SettingsPanel | SettingsPanel}.
 *
 * Consolidates the identical `{ type: 'checkbox', id: 'frontendHint', … }`
 * object literal that was duplicated across ~147 game pages (issue #4302).
 *
 * @param tc - The `common`-namespace translation function (for the shared label).
 * @param enabled - Current frontend-hint enabled state.
 * @param setEnabled - Toggle handler from `useGameHint`.
 */
export function hintCheckboxItem(
  tc: (key: string, opts?: { ns?: string }) => string,
  enabled: boolean,
  setEnabled: (checked: boolean) => void,
): SettingsItem {
  return {
    type: 'checkbox',
    id: 'frontendHint',
    label: tc('hint.toggle', { ns: 'tutorial' }),
    checked: enabled,
    onToggle: setEnabled,
  };
}
