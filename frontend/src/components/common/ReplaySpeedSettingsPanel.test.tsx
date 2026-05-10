import { fireEvent, render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { afterEach, describe, expect, it } from 'vitest';
import { REPLAY_SPEED_STORAGE_KEY } from '../../hooks/useReplaySpeed';
import { ReplaySpeedSettingsPanel } from './ReplaySpeedSettingsPanel';

const labelText = () => i18n.t('settings.replaySpeed.label', { ns: 'common' });

describe('ReplaySpeedSettingsPanel', () => {
  afterEach(() => {
    localStorage.clear();
  });

  it('renders the speed select with the persisted value', () => {
    localStorage.setItem(REPLAY_SPEED_STORAGE_KEY, 'fast');
    render(<ReplaySpeedSettingsPanel />);
    const select = screen.getByLabelText(labelText()) as HTMLSelectElement;
    expect(select.value).toBe('fast');
  });

  it('persists the chosen speed back to localStorage', () => {
    render(<ReplaySpeedSettingsPanel />);
    const select = screen.getByLabelText(labelText()) as HTMLSelectElement;
    fireEvent.change(select, { target: { value: 'instant' } });
    expect(select.value).toBe('instant');
    expect(localStorage.getItem(REPLAY_SPEED_STORAGE_KEY)).toBe('instant');
  });

  it('offers all three speed options', () => {
    render(<ReplaySpeedSettingsPanel />);
    const select = screen.getByLabelText(labelText()) as HTMLSelectElement;
    const values = Array.from(select.options).map((o) => o.value);
    expect(values).toEqual(['normal', 'fast', 'instant']);
  });
});
