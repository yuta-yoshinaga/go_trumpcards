import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { FavoriteToggleButton } from './FavoriteToggleButton';

describe('FavoriteToggleButton', () => {
  it('renders ☆ and aria-pressed=false when pressed is false', () => {
    render(
      <FavoriteToggleButton
        path="/blackjack"
        gameLabel="ブラックジャック"
        pressed={false}
        onToggle={() => undefined}
        className={() => 'cls'}
      />,
    );
    const button = screen.getByRole('button', { name: 'ブラックジャック をお気に入りに登録' });
    expect(button).toHaveAttribute('aria-pressed', 'false');
    expect(button.textContent).toBe('☆');
  });

  it('renders ★ and aria-pressed=true when pressed is true', () => {
    render(
      <FavoriteToggleButton
        path="/blackjack"
        gameLabel="ブラックジャック"
        pressed={true}
        onToggle={() => undefined}
        className={() => 'cls'}
      />,
    );
    const button = screen.getByRole('button', { name: 'ブラックジャック をお気に入りに登録' });
    expect(button).toHaveAttribute('aria-pressed', 'true');
    expect(button.textContent).toBe('★');
  });

  it('calls onToggle with the path when clicked', () => {
    const onToggle = vi.fn();
    render(
      <FavoriteToggleButton
        path="/poker"
        gameLabel="ポーカー"
        pressed={false}
        onToggle={onToggle}
        className={() => 'cls'}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'ポーカー をお気に入りに登録' }));
    expect(onToggle).toHaveBeenCalledWith('/poker');
  });

  it('forwards the pressed flag to the className function so callers can vary visuals', () => {
    const className = vi.fn(() => 'cls');
    render(
      <FavoriteToggleButton
        path="/hearts"
        gameLabel="ハーツ"
        pressed={true}
        onToggle={() => undefined}
        className={className}
      />,
    );
    expect(className).toHaveBeenCalledWith(true);
  });

  it('keeps the accessible name as the game favorite action regardless of pressed state', () => {
    const { rerender } = render(
      <FavoriteToggleButton
        path="/spades"
        gameLabel="スペード"
        pressed={false}
        onToggle={() => undefined}
        className={() => 'cls'}
      />,
    );
    expect(screen.getByRole('button')).toHaveAccessibleName('スペード をお気に入りに登録');

    rerender(
      <FavoriteToggleButton
        path="/spades"
        gameLabel="スペード"
        pressed={true}
        onToggle={() => undefined}
        className={() => 'cls'}
      />,
    );
    expect(screen.getByRole('button')).toHaveAccessibleName('スペード をお気に入りに登録');
  });

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
