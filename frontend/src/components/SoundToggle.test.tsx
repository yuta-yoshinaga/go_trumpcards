import { fireEvent, render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { describe, expect, it, vi } from 'vitest';
import { SoundToggle } from './SoundToggle';

const mockToggleMute = vi.fn();

vi.mock('../providers/SoundProvider', () => ({
  useSound: vi.fn(() => ({
    muted: false,
    toggleMute: mockToggleMute,
    playSound: vi.fn(),
    claimExecSound: vi.fn(),
    consumeExecClaim: vi.fn(() => false),
  })),
}));

import { useSound } from '../providers/SoundProvider';

describe('SoundToggle', () => {
  it('renders mute label when unmuted', () => {
    vi.mocked(useSound).mockReturnValue({
      muted: false,
      toggleMute: mockToggleMute,
      playSound: vi.fn(),
      claimExecSound: vi.fn(),
      consumeExecClaim: vi.fn(() => false),
    });
    render(<SoundToggle />);
    const btn = screen.getByRole('button', { name: i18n.t('sound.mute') });
    expect(btn).toBeInTheDocument();
    expect(btn.querySelector('svg')).toBeInTheDocument();
  });

  it('renders unmute label when muted', () => {
    vi.mocked(useSound).mockReturnValue({
      muted: true,
      toggleMute: mockToggleMute,
      playSound: vi.fn(),
      claimExecSound: vi.fn(),
      consumeExecClaim: vi.fn(() => false),
    });
    render(<SoundToggle />);
    const btn = screen.getByRole('button', { name: i18n.t('sound.unmute') });
    expect(btn).toBeInTheDocument();
    expect(btn.querySelector('svg')).toBeInTheDocument();
  });

  it('calls toggleMute when clicked', () => {
    mockToggleMute.mockClear();
    vi.mocked(useSound).mockReturnValue({
      muted: false,
      toggleMute: mockToggleMute,
      playSound: vi.fn(),
      claimExecSound: vi.fn(),
      consumeExecClaim: vi.fn(() => false),
    });
    render(<SoundToggle />);
    fireEvent.click(screen.getByRole('button'));
    expect(mockToggleMute).toHaveBeenCalledTimes(1);
  });
});
