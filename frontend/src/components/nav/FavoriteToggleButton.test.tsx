import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { FavoriteToggleButton } from './FavoriteToggleButton';

describe('FavoriteToggleButton', () => {
  it('names the game and exposes pressed state separately', () => {
    const { rerender } = render(
      <FavoriteToggleButton
        path="/hearts"
        gameLabel="ハーツ"
        pressed={false}
        onToggle={vi.fn()}
        className={() => ''}
      />,
    );

    const button = screen.getByRole('button', { name: 'ハーツ をお気に入りに登録' });
    expect(button).toHaveAttribute('aria-pressed', 'false');

    rerender(
      <FavoriteToggleButton path="/hearts" gameLabel="ハーツ" pressed onToggle={vi.fn()} className={() => ''} />,
    );
    expect(screen.getByRole('button', { name: 'ハーツ をお気に入りに登録' })).toHaveAttribute('aria-pressed', 'true');
  });
});
