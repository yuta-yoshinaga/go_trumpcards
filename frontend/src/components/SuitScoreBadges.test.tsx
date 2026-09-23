import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { fiftyOneSuitScores } from '../utils/fiftyOneSuitScores';
import { SuitScoreBadges } from './SuitScoreBadges';

const cards: Card[] = [
  { design: 'SPADE', value: 1 },
  { design: 'SPADE', value: 10 },
  { design: 'HEART', value: 5 },
  { design: 'DIAMOND', value: 13 },
  { design: 'JOKER', value: 0 },
];

describe('SuitScoreBadges', () => {
  it('renders the default test IDs and suit totals', () => {
    render(<SuitScoreBadges cards={cards} ariaLabel="Suit scores" />);

    expect(screen.getByTestId('suit-score-badges')).toBeInTheDocument();
    const totals = fiftyOneSuitScores(cards);
    for (const design of ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as const) {
      expect(screen.getByTestId(`suit-badge-${design}`)).toHaveTextContent(String(totals[design]));
    }
  });

  it('renders explicitly provided test IDs', () => {
    render(
      <SuitScoreBadges
        cards={cards}
        ariaLabel="CPU suit scores"
        listTestId="custom-list"
        badgeTestId={(design) => `custom-badge-${design}`}
      />,
    );

    expect(screen.getByTestId('custom-list')).toBeInTheDocument();
    expect(screen.getByTestId('custom-badge-SPADE')).toBeInTheDocument();
    expect(screen.queryByTestId('suit-score-badges')).not.toBeInTheDocument();
  });

  it('highlights only the highest positive suit', () => {
    render(<SuitScoreBadges cards={cards} ariaLabel="Suit scores" />);

    expect(screen.getByTestId('suit-badge-SPADE')).toHaveClass('bg-ds-accent');
    expect(screen.getByTestId('suit-badge-CLOVER')).not.toHaveClass('bg-ds-accent');
    expect(screen.getByTestId('suit-badge-HEART')).not.toHaveClass('bg-ds-accent');
    expect(screen.getByTestId('suit-badge-DIAMOND')).not.toHaveClass('bg-ds-accent');
  });

  it('uses fixed-order tie breaking and highlights no suit when all totals are zero', () => {
    const { rerender } = render(
      <SuitScoreBadges
        cards={[
          { design: 'SPADE', value: 5 },
          { design: 'CLOVER', value: 5 },
        ]}
        ariaLabel="Suit scores"
      />,
    );
    expect(screen.getByTestId('suit-badge-SPADE')).toHaveClass('bg-ds-accent');
    expect(screen.getByTestId('suit-badge-CLOVER')).not.toHaveClass('bg-ds-accent');

    rerender(<SuitScoreBadges cards={[]} ariaLabel="Suit scores" />);
    for (const design of ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as const) {
      expect(screen.getByTestId(`suit-badge-${design}`)).not.toHaveClass('bg-ds-accent');
    }
  });
});
