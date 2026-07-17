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

  it('marks ally vs foe when partner-team data is supplied', () => {
    const teamedPlayers: TrickDisplayPlayer[] = [
      { id: 0, isHuman: true, team: 0 },
      { id: 1, isHuman: false, team: 1 },
      { id: 2, isHuman: false, team: 0 },
      { id: 3, isHuman: false, team: 1 },
    ];
    const teamedTrick: TrickDisplayCard[] = [
      { playerIdx: 0, card }, // human, team 0 → ally
      { playerIdx: 1, card: { design: 'HEART', value: 10 } }, // CPU 1, team 1 → foe
      { playerIdx: 2, card: { design: 'CLOVER', value: 7 } }, // partner, team 0 → ally
    ];
    const { container } = render(
      <TrickDisplay currentTrick={teamedTrick} players={teamedPlayers} cardWidth={40} label="現在のトリック" />,
    );
    const allies = container.querySelectorAll('[data-team-role="ally"]');
    const foes = container.querySelectorAll('[data-team-role="foe"]');
    expect(allies).toHaveLength(2);
    expect(foes).toHaveLength(1);
  });

  it('omits team coloring when only one team appears in the players list', () => {
    const singleTeamPlayers: TrickDisplayPlayer[] = players.map((p) => ({ ...p, team: 0 }));
    const { container } = render(
      <TrickDisplay currentTrick={trick} players={singleTeamPlayers} cardWidth={40} label="label" />,
    );
    expect(container.querySelector('[data-team-role]')).not.toBeInTheDocument();
  });

  it('highlights the winning card with a badge when winnerIdx is set', () => {
    const { container } = render(
      <TrickDisplay
        currentTrick={trick}
        players={players}
        cardWidth={40}
        label="label"
        winnerIdx={1}
        winnerLabel="勝ち"
      />,
    );
    const badge = screen.getByTestId('trick-winner-badge');
    expect(badge).toHaveTextContent('勝ち');
    // Exactly one card is flagged the winner.
    expect(container.querySelectorAll('[data-trick-winner="true"]')).toHaveLength(1);
  });

  it('shows no winner highlight when winnerIdx is omitted', () => {
    const { container } = render(<TrickDisplay currentTrick={trick} players={players} cardWidth={40} label="label" />);
    expect(screen.queryByTestId('trick-winner-badge')).not.toBeInTheDocument();
    expect(container.querySelector('[data-trick-winner]')).not.toBeInTheDocument();
  });

  it('defaults the winner badge text to WIN', () => {
    render(<TrickDisplay currentTrick={trick} players={players} cardWidth={40} label="label" winnerIdx={0} />);
    expect(screen.getByTestId('trick-winner-badge')).toHaveTextContent('WIN');
  });
});
