import { describe, expect, it, vi } from 'vitest';
import { hintCheckboxItem } from './settingsItems';

describe('hintCheckboxItem', () => {
  it('builds the frontend-hint checkbox item with the shared label and wiring', () => {
    const tc = vi.fn((key: string, opts?: { ns?: string }) => `${opts?.ns ?? ''}:${key}`);
    const setEnabled = vi.fn();
    const item = hintCheckboxItem(tc, true, setEnabled);

    expect(item).toEqual({
      type: 'checkbox',
      id: 'frontendHint',
      label: 'tutorial:hint.toggle',
      checked: true,
      onToggle: setEnabled,
    });
    expect(tc).toHaveBeenCalledWith('hint.toggle', { ns: 'tutorial' });
  });

  it('reflects the disabled state via checked', () => {
    const tc = (key: string) => key;
    expect(hintCheckboxItem(tc, false, () => {}).checked).toBe(false);
  });
});
