import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { GamePageHeading } from './GamePageHeading';

describe('GamePageHeading', () => {
  it('renders an h1 heading with the given title', () => {
    render(<GamePageHeading title="ブラックジャック" />);
    const heading = screen.getByRole('heading', { level: 1 });
    expect(heading).toHaveTextContent('ブラックジャック');
  });

  it('applies sr-only class for visual hiding', () => {
    render(<GamePageHeading title="Poker" />);
    const heading = screen.getByRole('heading', { level: 1 });
    expect(heading).toHaveClass('sr-only');
  });
});
