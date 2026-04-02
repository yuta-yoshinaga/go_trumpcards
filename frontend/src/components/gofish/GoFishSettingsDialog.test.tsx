import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../test/renderWithProviders';
import { GoFishSettingsDialog } from './GoFishSettingsDialog';

describe('GoFishSettingsDialog', () => {
  it('renders CPU difficulty selector', () => {
    renderWithProviders(<GoFishSettingsDialog cpuDifficulty={1} onCpuDifficultyChange={vi.fn()} />);
    expect(screen.getByRole('combobox')).toBeInTheDocument();
  });

  it('shows current difficulty selected', () => {
    renderWithProviders(<GoFishSettingsDialog cpuDifficulty={2} onCpuDifficultyChange={vi.fn()} />);
    const select = screen.getByRole('combobox');
    expect((select as HTMLSelectElement).value).toBe('2');
  });

  it('calls onCpuDifficultyChange when selection changes', () => {
    const onChange = vi.fn();
    renderWithProviders(<GoFishSettingsDialog cpuDifficulty={1} onCpuDifficultyChange={onChange} />);
    fireEvent.change(screen.getByRole('combobox'), { target: { value: '0' } });
    expect(onChange).toHaveBeenCalledWith('0');
  });

  it('renders all difficulty options', () => {
    renderWithProviders(<GoFishSettingsDialog cpuDifficulty={1} onCpuDifficultyChange={vi.fn()} />);
    const options = screen.getAllByRole('option');
    expect(options).toHaveLength(3);
  });
});
