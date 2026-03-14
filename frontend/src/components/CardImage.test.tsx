import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { CardBack, CardImage } from './CardImage';

describe('CardImage', () => {
  it.each<[CardDesign, string]>([
    ['SPADE', 's'],
    ['CLOVER', 'c'],
    ['HEART', 'h'],
    ['DIAMOND', 'd'],
    ['JOKER', 'x'],
  ])('renders correct image src for %s', (design, prefix) => {
    const card: Card = { design, value: 1 };
    render(<CardImage card={card} />);
    expect(screen.getByRole('img')).toHaveAttribute('src', `/images/${prefix}01.png`);
  });

  it('zero-pads single-digit card values', () => {
    const card: Card = { design: 'HEART', value: 5 };
    render(<CardImage card={card} />);
    expect(screen.getByRole('img')).toHaveAttribute('src', '/images/h05.png');
  });

  it('does not pad two-digit card values', () => {
    const card: Card = { design: 'SPADE', value: 13 };
    render(<CardImage card={card} />);
    expect(screen.getByRole('img')).toHaveAttribute('src', '/images/s13.png');
  });

  it('renders alt text with design and value', () => {
    const card: Card = { design: 'DIAMOND', value: 7 };
    render(<CardImage card={card} />);
    expect(screen.getByRole('img')).toHaveAttribute('alt', '♦ 7');
  });

  it('applies custom className', () => {
    const card: Card = { design: 'SPADE', value: 1 };
    render(<CardImage card={card} className="my-class" />);
    expect(screen.getByRole('img')).toHaveClass('my-class');
  });

  it('applies custom style', () => {
    const card: Card = { design: 'SPADE', value: 1 };
    render(<CardImage card={card} style={{ opacity: 0.5 }} />);
    const img = screen.getByRole('img');
    expect(img).toHaveStyle({ opacity: '0.5' });
  });

  it('applies width prop as image width', () => {
    const card: Card = { design: 'SPADE', value: 1 };
    render(<CardImage card={card} width={60} />);
    expect(screen.getByRole('img')).toHaveStyle({ width: '60px' });
  });

  it('falls back to x prefix for unknown design', () => {
    const card = { design: 'UNKNOWN' as CardDesign, value: 0 };
    render(<CardImage card={card} />);
    expect(screen.getByRole('img')).toHaveAttribute('src', '/images/x00.png');
  });

  it('defaults to 80px width when width prop is omitted', () => {
    const card: Card = { design: 'SPADE', value: 1 };
    render(<CardImage card={card} />);
    expect(screen.getByRole('img')).toHaveStyle({ width: '80px', maxWidth: '100%' });
  });

  it('sets draggable attribute when draggable prop is true', () => {
    const card: Card = { design: 'SPADE', value: 1 };
    render(<CardImage card={card} draggable />);
    expect(screen.getByRole('img')).toHaveAttribute('draggable', 'true');
  });

  it('does not set draggable when prop is omitted', () => {
    const card: Card = { design: 'SPADE', value: 1 };
    render(<CardImage card={card} />);
    expect(screen.getByRole('img')).not.toHaveAttribute('draggable', 'true');
  });

  it('fires onDragStart event handler', () => {
    const card: Card = { design: 'SPADE', value: 1 };
    const onDragStart = vi.fn();
    render(<CardImage card={card} draggable onDragStart={onDragStart} />);
    const img = screen.getByRole('img');
    const event = new Event('dragstart', { bubbles: true });
    img.dispatchEvent(event);
    expect(onDragStart).toHaveBeenCalledTimes(1);
  });

  it('fires onDragOver event handler', () => {
    const card: Card = { design: 'SPADE', value: 1 };
    const onDragOver = vi.fn();
    render(<CardImage card={card} onDragOver={onDragOver} />);
    const img = screen.getByRole('img');
    const event = new Event('dragover', { bubbles: true });
    img.dispatchEvent(event);
    expect(onDragOver).toHaveBeenCalledTimes(1);
  });

  it('fires onDrop event handler', () => {
    const card: Card = { design: 'SPADE', value: 1 };
    const onDrop = vi.fn();
    render(<CardImage card={card} onDrop={onDrop} />);
    const img = screen.getByRole('img');
    const event = new Event('drop', { bubbles: true });
    img.dispatchEvent(event);
    expect(onDrop).toHaveBeenCalledTimes(1);
  });
});

describe('CardBack', () => {
  it('renders the card back image', () => {
    render(<CardBack />);
    const img = screen.getByRole('img');
    expect(img).toHaveAttribute('src', '/images/z01.png');
    expect(img).toHaveAttribute('alt', 'カード裏面');
  });

  it('applies custom className', () => {
    render(<CardBack className="back-class" />);
    expect(screen.getByRole('img')).toHaveClass('back-class');
  });

  it('calls onClick when clicked', async () => {
    const onClick = vi.fn();
    render(<CardBack onClick={onClick} />);
    screen.getByRole('button', { name: 'カード裏面' }).click();
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it('renders button with custom aria-label when ariaLabel is provided', () => {
    const onClick = vi.fn();
    render(<CardBack onClick={onClick} ariaLabel="カード 1 枚目を引く" />);
    expect(screen.getByRole('button', { name: 'カード 1 枚目を引く' })).toBeInTheDocument();
  });

  it('sets empty alt on img when onClick and ariaLabel are both provided', () => {
    render(<CardBack onClick={() => undefined} ariaLabel="カード 1 枚目を引く" />);
    const btn = screen.getByRole('button', { name: 'カード 1 枚目を引く' });
    const img = btn.querySelector('img') as HTMLImageElement;
    expect(img).toHaveAttribute('alt', '');
  });

  it('uses i18n fallback aria-label and empty alt when onClick is provided without ariaLabel', () => {
    render(<CardBack onClick={() => undefined} />);
    const btn = screen.getByRole('button', { name: 'カード裏面' });
    expect(btn).toHaveAttribute('aria-label', 'カード裏面');
    expect(btn.querySelector('img')).toHaveAttribute('alt', '');
  });

  it('shows pointer cursor when onClick is provided', () => {
    render(<CardBack onClick={() => undefined} />);
    const img = screen.getByRole('button', { name: 'カード裏面' });
    expect(img).toHaveStyle({ cursor: 'pointer' });
  });

  it('has no pointer cursor when onClick is absent', () => {
    render(<CardBack />);
    const img = screen.getByRole('img');
    expect(img.style.cursor).toBe('');
  });

  it('applies width prop as image width', () => {
    render(<CardBack width={40} />);
    expect(screen.getByRole('img')).toHaveStyle({ width: '40px' });
  });

  it('defaults to 80px width when width prop is omitted', () => {
    render(<CardBack />);
    expect(screen.getByRole('img')).toHaveStyle({ width: '80px', maxWidth: '100%' });
  });
});
