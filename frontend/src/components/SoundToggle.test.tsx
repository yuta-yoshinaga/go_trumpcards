import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { SoundToggle } from './SoundToggle';

describe('SoundToggle', () => {
  it('renders unmute label when muted', () => {
    render(<SoundToggle muted={true} onToggle={vi.fn()} />);
    expect(screen.getByRole('button', { name: 'サウンドをオンにする' })).toBeInTheDocument();
    expect(screen.getByText('🔇')).toBeInTheDocument();
  });

  it('renders mute label when unmuted', () => {
    render(<SoundToggle muted={false} onToggle={vi.fn()} />);
    expect(screen.getByRole('button', { name: 'サウンドをオフにする' })).toBeInTheDocument();
    expect(screen.getByText('🔊')).toBeInTheDocument();
  });

  it('calls onToggle when clicked', () => {
    const onToggle = vi.fn();
    render(<SoundToggle muted={false} onToggle={onToggle} />);
    fireEvent.click(screen.getByRole('button'));
    expect(onToggle).toHaveBeenCalledTimes(1);
  });
});
