import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { TrickDisplay, type TrickDisplayCard, type TrickDisplayPlayer } from './TrickDisplay';

const card: Card = { design: 'SPADE', value: 1 };

const players: TrickDisplayPlayer[] = [
  { id: 0, isHuman: true },
  { id: 1, isHuman: false },
  { id: 2, isHuman: false },
];

const trick: TrickDisplayCard[] = [
  { playerIdx: 0, card },
  { playerIdx: 1, card: { design: 'HEART', value: 10 } },
];

describe('TrickDisplay', () => {
  it('renders nothing when the trick is empty', () => {
    const { container } = render(
      <TrickDisplay currentTrick={[]} players={players} cardWidth={40} label="Current trick" />,
    );
    expect(container).toBeEmptyDOMElement();
  });

  it('renders one card per trick entry with the label', () => {
    render(<TrickDisplay currentTrick={trick} players={players} cardWidth={40} label="現在のトリック" />);
    expect(screen.getByText('現在のトリック')).toBeInTheDocument();
    expect(screen.getAllByTestId('animated-card')).toHaveLength(2);
  });

  it('resolves player display names from the players array', () => {
    render(<TrickDisplay currentTrick={trick} players={players} cardWidth={40} label="label" />);
    expect(screen.getByText('あなた')).toBeInTheDocument();
    expect(screen.getByText('CPU 1')).toBeInTheDocument();
  });

  it('applies data-tutorial attribute when provided', () => {
    const { container } = render(
      <TrickDisplay
        currentTrick={trick}
        players={players}
        cardWidth={40}
        label="label"
        dataTutorial="ht-trick-display"
      />,
    );
    expect(container.querySelector('[data-tutorial="ht-trick-display"]')).toBeInTheDocument();
  });

  it('falls back to CPU label when player index is out of range', () => {
    render(<TrickDisplay currentTrick={[{ playerIdx: 7, card }]} players={players} cardWidth={40} label="label" />);
    expect(screen.getByText('CPU 7')).toBeInTheDocument();
  });

  it('renders one AnimatedCard per trick entry', () => {
    render(<TrickDisplay currentTrick={trick} players={players} cardWidth={40} label="label" />);
    expect(screen.getAllByTestId('animated-card')).toHaveLength(2);
  });
});
