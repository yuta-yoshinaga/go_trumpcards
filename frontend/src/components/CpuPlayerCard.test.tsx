import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { CpuPlayerCard } from './CpuPlayerCard';

const cards: Card[] = [
  { design: 'SPADE', value: 1 },
  { design: 'HEART', value: 13 },
];

function makePlayer(overrides: Partial<Parameters<typeof CpuPlayerCard>[0]['player']> = {}) {
  return {
    id: 1,
    playStyleName: 'TAG',
    chips: 500,
    currentBet: 0,
    folded: false,
    allIn: false,
    handName: '',
    cards,
    ...overrides,
  };
}

describe('CpuPlayerCard', () => {
  it('renders player info', () => {
    render(<CpuPlayerCard player={makePlayer()} showCards={false} faceDownCount={2} showHandName={false} />);
    expect(screen.getByText(/CPU 1/)).toBeInTheDocument();
    expect(screen.getByText(/TAG/)).toBeInTheDocument();
    expect(screen.getByText(/チップ: 500/)).toBeInTheDocument();
  });

  it('shows face-down cards when showCards is false', () => {
    render(<CpuPlayerCard player={makePlayer()} showCards={false} faceDownCount={5} showHandName={false} />);
    const backs = screen.getAllByAltText('カード裏面');
    expect(backs).toHaveLength(5);
  });

  it('shows face-up cards when showCards is true', () => {
    render(<CpuPlayerCard player={makePlayer()} showCards={true} faceDownCount={2} showHandName={false} />);
    expect(screen.queryByAltText('カード裏面')).not.toBeInTheDocument();
    expect(screen.getByAltText('♠ A')).toBeInTheDocument();
  });

  it('shows face-down cards when showCards is true but player is folded', () => {
    render(
      <CpuPlayerCard player={makePlayer({ folded: true })} showCards={true} faceDownCount={2} showHandName={false} />,
    );
    expect(screen.getAllByAltText('カード裏面')).toHaveLength(2);
  });

  it('shows hand name badge when showHandName is true', () => {
    render(
      <CpuPlayerCard
        player={makePlayer({ handName: 'フルハウス' })}
        showCards={true}
        faceDownCount={2}
        showHandName={true}
      />,
    );
    expect(screen.getByText('フルハウス')).toBeInTheDocument();
  });

  it('hides hand name badge when showHandName is false', () => {
    render(
      <CpuPlayerCard
        player={makePlayer({ handName: 'フルハウス' })}
        showCards={true}
        faceDownCount={2}
        showHandName={false}
      />,
    );
    expect(screen.queryByText('フルハウス')).not.toBeInTheDocument();
  });

  it('does not show hand name when folded', () => {
    render(
      <CpuPlayerCard
        player={makePlayer({ handName: 'フルハウス', folded: true })}
        showCards={true}
        faceDownCount={2}
        showHandName={true}
      />,
    );
    expect(screen.queryByText('フルハウス')).not.toBeInTheDocument();
  });

  it('does not render hand name badge when handName is empty', () => {
    const { container } = render(
      <CpuPlayerCard player={makePlayer({ handName: '' })} showCards={true} faceDownCount={2} showHandName={true} />,
    );
    // No badge element with handNameBadgeStyle background
    const badges = container.querySelectorAll('[style*="background"]');
    expect(badges).toHaveLength(0);
  });

  it('shows bet amount when currentBet > 0', () => {
    render(
      <CpuPlayerCard
        player={makePlayer({ currentBet: 30 })}
        showCards={false}
        faceDownCount={2}
        showHandName={false}
      />,
    );
    expect(screen.getByText(/ベット: 30/)).toBeInTheDocument();
  });

  it('does not show bet when currentBet is 0', () => {
    render(<CpuPlayerCard player={makePlayer()} showCards={false} faceDownCount={2} showHandName={false} />);
    expect(screen.queryByText(/ベット:/)).not.toBeInTheDocument();
  });

  it('shows folded label', () => {
    render(
      <CpuPlayerCard player={makePlayer({ folded: true })} showCards={false} faceDownCount={2} showHandName={false} />,
    );
    expect(screen.getByText('[フォールド]')).toBeInTheDocument();
  });

  it('shows all-in label', () => {
    render(
      <CpuPlayerCard player={makePlayer({ allIn: true })} showCards={false} faceDownCount={2} showHandName={false} />,
    );
    expect(screen.getByText('[オールイン]')).toBeInTheDocument();
  });

  it('renders extraInfo when provided', () => {
    render(
      <CpuPlayerCard
        player={makePlayer()}
        showCards={false}
        faceDownCount={2}
        showHandName={false}
        extraInfo={<span>交換: 2枚</span>}
      />,
    );
    expect(screen.getByText('交換: 2枚')).toBeInTheDocument();
  });

  it('does not render extraInfo when not provided', () => {
    render(<CpuPlayerCard player={makePlayer()} showCards={false} faceDownCount={2} showHandName={false} />);
    expect(screen.queryByText(/交換/)).not.toBeInTheDocument();
  });

  it('shows face-down when showCards true but cards empty', () => {
    render(
      <CpuPlayerCard player={makePlayer({ cards: [] })} showCards={true} faceDownCount={3} showHandName={false} />,
    );
    expect(screen.getAllByAltText('カード裏面')).toHaveLength(3);
  });
});
