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
    expect(screen.getByRole('img')).toHaveAttribute('alt', 'DIAMOND 7');
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

  it('defaults to 80px width when width prop is omitted', () => {
    const card: Card = { design: 'SPADE', value: 1 };
    render(<CardImage card={card} />);
    expect(screen.getByRole('img')).toHaveStyle({ width: '80px' });
  });
});

describe('CardBack', () => {
  it('renders the card back image', () => {
    render(<CardBack />);
    const img = screen.getByRole('img');
    expect(img).toHaveAttribute('src', '/images/z01.png');
    expect(img).toHaveAttribute('alt', 'card back');
  });

  it('applies custom className', () => {
    render(<CardBack className="back-class" />);
    expect(screen.getByRole('img')).toHaveClass('back-class');
  });

  it('calls onClick when clicked', async () => {
    const onClick = vi.fn();
    render(<CardBack onClick={onClick} />);
    screen.getByRole('button', { name: 'card back' }).click();
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it('shows pointer cursor when onClick is provided', () => {
    render(<CardBack onClick={() => undefined} />);
    const img = screen.getByRole('button', { name: 'card back' });
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
    expect(screen.getByRole('img')).toHaveStyle({ width: '80px' });
  });
});
