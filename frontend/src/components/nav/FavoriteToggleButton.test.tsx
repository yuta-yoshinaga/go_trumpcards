import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { FavoriteToggleButton } from './FavoriteToggleButton';

describe('FavoriteToggleButton', () => {
  it('renders ☆ and aria-pressed=false when pressed is false', () => {
    render(
      <FavoriteToggleButton path="/blackjack" pressed={false} onToggle={() => undefined} className={() => 'cls'} />,
    );
    const button = screen.getByRole('button');
    expect(button).toHaveAttribute('aria-pressed', 'false');
    expect(button.textContent).toBe('☆');
  });

  it('renders ★ and aria-pressed=true when pressed is true', () => {
    render(
      <FavoriteToggleButton path="/blackjack" pressed={true} onToggle={() => undefined} className={() => 'cls'} />,
    );
    const button = screen.getByRole('button');
    expect(button).toHaveAttribute('aria-pressed', 'true');
    expect(button.textContent).toBe('★');
  });

  it('calls onToggle with the path when clicked', () => {
    const onToggle = vi.fn();
    render(<FavoriteToggleButton path="/poker" pressed={false} onToggle={onToggle} className={() => 'cls'} />);
    fireEvent.click(screen.getByRole('button'));
    expect(onToggle).toHaveBeenCalledWith('/poker');
  });

  it('forwards the pressed flag to the className function so callers can vary visuals', () => {
    const className = vi.fn(() => 'cls');
    render(<FavoriteToggleButton path="/hearts" pressed={true} onToggle={() => undefined} className={className} />);
    expect(className).toHaveBeenCalledWith(true);
  });

  it('uses different aria-labels for add vs. remove', () => {
    const { rerender } = render(
      <FavoriteToggleButton path="/spades" pressed={false} onToggle={() => undefined} className={() => 'cls'} />,
    );
    expect(screen.getByRole('button')).toHaveAccessibleName('お気に入りに追加');

    rerender(<FavoriteToggleButton path="/spades" pressed={true} onToggle={() => undefined} className={() => 'cls'} />);
    expect(screen.getByRole('button')).toHaveAccessibleName('お気に入りから削除');
  });
});
