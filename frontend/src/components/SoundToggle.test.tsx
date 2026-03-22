import { fireEvent, render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { describe, expect, it, vi } from 'vitest';
import { SoundToggle } from './SoundToggle';

describe('SoundToggle', () => {
  it('renders unmute label when muted', () => {
    render(<SoundToggle muted={true} onToggle={vi.fn()} />);
    const btn = screen.getByRole('button', { name: i18n.t('sound.unmute') });
    expect(btn).toBeInTheDocument();
    expect(btn.querySelector('svg')).toBeInTheDocument();
  });

  it('renders mute label when unmuted', () => {
    render(<SoundToggle muted={false} onToggle={vi.fn()} />);
    const btn = screen.getByRole('button', { name: i18n.t('sound.mute') });
    expect(btn).toBeInTheDocument();
    expect(btn.querySelector('svg')).toBeInTheDocument();
  });

  it('calls onToggle when clicked', () => {
    const onToggle = vi.fn();
    render(<SoundToggle muted={false} onToggle={onToggle} />);
    fireEvent.click(screen.getByRole('button'));
    expect(onToggle).toHaveBeenCalledTimes(1);
  });
});
