import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
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

  it('uses the supplied pile position when generating card row keys', () => {
    const keyFor = vi.fn((trickCard: TrickDisplayCard, index: number) => `pile-${trickCard.playerIdx}-${index}`);
    const { rerender } = render(
      <TrickDisplay
        currentTrick={[
          { playerIdx: 1, card },
          { playerIdx: 1, card: { design: 'HEART', value: 2 } },
        ]}
        players={players}
        cardWidth={40}
        label="Pile"
        cardKeyFor={keyFor}
        cardAriaLabelFor={(player, playedCard) => `CPU ${player.id} played ${playedCard.design} ${playedCard.value}`}
      />,
    );

    expect(keyFor).toHaveBeenCalledTimes(2);
    expect(keyFor).toHaveBeenNthCalledWith(1, { playerIdx: 1, card }, 0);
    expect(keyFor).toHaveBeenNthCalledWith(2, { playerIdx: 1, card: { design: 'HEART', value: 2 } }, 1);

    keyFor.mockClear();
    rerender(
      <TrickDisplay
        currentTrick={[
          { playerIdx: 1, card },
          { playerIdx: 2, card: { design: 'DIAMOND', value: 3 } },
          { playerIdx: 1, card: { design: 'HEART', value: 2 } },
        ]}
        players={players}
        cardWidth={40}
        label="Pile"
        cardKeyFor={keyFor}
        cardAriaLabelFor={(player, playedCard) => `CPU ${player.id} played ${playedCard.design} ${playedCard.value}`}
      />,
    );

    expect(keyFor).toHaveBeenCalledTimes(3);
    expect(keyFor).toHaveBeenNthCalledWith(1, { playerIdx: 1, card }, 0);
    expect(keyFor).toHaveBeenNthCalledWith(2, { playerIdx: 2, card: { design: 'DIAMOND', value: 3 } }, 1);
    expect(keyFor).toHaveBeenNthCalledWith(3, { playerIdx: 1, card: { design: 'HEART', value: 2 } }, 2);
    expect(screen.getByRole('img', { name: 'CPU 1 played SPADE 1' }).closest('.relative')).toHaveTextContent('CPU 1');
    expect(screen.getByRole('img', { name: 'CPU 2 played DIAMOND 3' }).closest('.relative')).toHaveTextContent('CPU 2');
    expect(screen.getByRole('img', { name: 'CPU 1 played HEART 2' }).closest('.relative')).toHaveTextContent('CPU 1');
  });

  it('passes server trick details through to the visible detail and card announcement', () => {
    const detailTrick: TrickDisplayCard[] = [{ playerIdx: 0, card, points: 11, isTrump: false }];
    render(
      <TrickDisplay
        currentTrick={detailTrick}
        players={players}
        cardWidth={40}
        label="Current trick"
        cardAriaLabelFor={(_, playedCard) => `${playedCard.value} of Spades`}
        cardDetailFor={(_, trickCard) => `${trickCard.points} points · ${trickCard.isTrump ? 'trump' : 'not trump'}`}
      />,
    );
    expect(screen.getByText('11 points · not trump')).toBeInTheDocument();
    expect(screen.getByAltText('1 of Spades · 11 points · not trump')).toBeInTheDocument();
  });

  it('keeps the legacy parenthesized format when a card has only a badge', () => {
    render(
      <TrickDisplay
        currentTrick={[{ playerIdx: 0, card }]}
        players={players}
        cardWidth={40}
        label="Current trick"
        cardAriaLabelFor={() => '♠ A'}
        cardBadgeFor={() => ({ glyph: '★', title: '切り札: スペード' })}
      />,
    );

    expect(screen.getByAltText('♠ A (切り札: スペード)')).toBeInTheDocument();
  });

  it('appends detail after the legacy badge format', () => {
    render(
      <TrickDisplay
        currentTrick={[{ playerIdx: 0, card, points: 11, isTrump: true }]}
        players={players}
        cardWidth={40}
        label="Current trick"
        cardAriaLabelFor={() => '♠ A'}
        cardBadgeFor={() => ({ glyph: '★', title: '切り札: スペード' })}
        cardDetailFor={(_, trickCard) => `${trickCard.points}点 · 切り札`}
      />,
    );

    expect(screen.getByAltText('♠ A (切り札: スペード) · 11点 · 切り札')).toBeInTheDocument();
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

  it('falls back to the previous trick when the current trick is empty', () => {
    render(
      <TrickDisplay
        currentTrick={[]}
        lastTrick={trick}
        lastTrickWinner={1}
        players={players}
        cardWidth={40}
        label="前のトリック"
      />,
    );
    expect(screen.getByText('前のトリック')).toBeInTheDocument();
    expect(screen.getByTestId('trick-winner-badge')).toHaveTextContent('WIN');
    expect(screen.getAllByTestId('animated-card')).toHaveLength(2);
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

  it('renders an accessible custom card role marker', () => {
    render(
      <TrickDisplay
        currentTrick={trick}
        players={players}
        cardWidth={40}
        label="label"
        cardBadgeFor={(card) => (card.design === 'SPADE' ? { glyph: '↺', title: 'Round suit: 1 is strongest' } : null)}
      />,
    );
    expect(screen.getByTitle('Round suit: 1 is strongest')).toBeInTheDocument();
    expect(screen.getByRole('img', { name: /Round suit: 1 is strongest/ })).toBeInTheDocument();
  });
});

// **場札が席数に縛られないゲームがある** (#5756)。Bhabhi の pile は
// フォローできない人が出るまで精算されないので何十枚にもなり、1 行のままだと
// ページ本体が横スクロールしてしまう。
describe('TrickDisplay wrap mode', () => {
  const longPile: TrickDisplayCard[] = Array.from({ length: 24 }, (_, i) => ({
    playerIdx: i % 3,
    card: { design: 'SPADE', value: (i % 13) + 1 } as Card,
  }));

  it('wraps the row when asked', () => {
    render(<TrickDisplay currentTrick={longPile} players={players} cardWidth={40} label="場" wrap />);
    const row = screen.getByTestId('trick-display-cards');
    expect(row.className).toContain('flex-wrap');
    // 全カードが読める (折り返しても枚数は減らない)。
    expect(screen.getAllByTestId('animated-card')).toHaveLength(24);
  });

  // **既存のゲームの見た目は変えない** (受け入れ条件3)。
  it('keeps a single row by default', () => {
    render(<TrickDisplay currentTrick={trick} players={players} cardWidth={40} label="現在のトリック" />);
    expect(screen.getByTestId('trick-display-cards').className).not.toContain('flex-wrap');
  });
});
