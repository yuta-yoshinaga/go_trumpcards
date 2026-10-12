import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { RotatedTrumpStock } from './RotatedTrumpStock';

describe('RotatedTrumpStock', () => {
  it('keeps the rotated card in its reserved area beside the label', () => {
    render(
      <RotatedTrumpStock
        label={<span data-testid="label">Trump</span>}
        trumpCard={<div data-testid="card-face" />}
        stockRemaining={5}
        cardWidth={100}
        deckTestId="stock-deck"
      />,
    );

    const label = screen.getByTestId('label');
    const trump = screen.getByTestId('rotated-trump-card');
    const cardFace = screen.getByTestId('card-face');
    expect(label.parentElement).toBe(trump.parentElement?.parentElement);
    expect(trump.style.transform).toBe('translate(-50%, -50%) rotate(90deg)');
    expect(trump.parentElement?.getAttribute('style')).toContain('width: 98px');
    expect(cardFace).toBeInTheDocument();
    expect(screen.getByTestId('stock-deck')).toBeInTheDocument();
  });

  it('omits the deck stack when no cards remain', () => {
    render(
      <RotatedTrumpStock
        label={<span>Trump</span>}
        trumpCard={<div>Card</div>}
        stockRemaining={0}
        cardWidth={100}
        deckTestId="stock-deck"
      />,
    );
    expect(screen.getByTestId('rotated-trump-card')).toBeInTheDocument();
    expect(screen.queryByTestId('stock-deck')).not.toBeInTheDocument();
  });
});
