import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bidWhistApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BidWhistResponse, Card } from '../types/card';
import { BidWhistPage } from './BidWhistPage';

vi.mock('../api/gameApi', () => ({
  bidWhistApi: { exec: vi.fn() },
  actionLogApi: { bidwhist: vi.fn() },
}));

const mockExec = vi.mocked(bidWhistApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function player(id: number, isHuman: boolean, cards: Card[], over: Partial<BidWhistResponse['players'][number]> = {}) {
  return {
    id,
    isHuman,
    cardCount: cards.length,
    cards,
    team: id % 2,
    trickCount: 0,
    passed: false,
    isDeclarer: false,
    ...over,
  };
}

const sixCards = [
  card('SPADE', 5),
  card('HEART', 6),
  card('DIAMOND', 7),
  card('CLOVER', 8),
  card('SPADE', 9),
  card('HEART', 10),
];

function makeState(overrides: Partial<BidWhistResponse> = {}): BidWhistResponse {
  return {
    players: [player(0, true, sixCards), player(1, false, []), player(2, false, []), player(3, false, [])],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    leadPlayerIdx: 0,
    trumpSuit: -1,
    contractTricks: 0,
    contractDirection: 0,
    declarerIdx: -1,
    highestBid: null,
    highestBidder: -1,
    kittyCount: 6,
    currentTrick: [],
    teamScores: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { cpuDifficulty: 1, targetScore: 7 },
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('BidWhistPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<BidWhistPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', expect.objectContaining({ config: expect.any(Object) })),
    );
  });

  it('renders CPU difficulty options with localized labels', async () => {
    renderWithProviders(<BidWhistPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // Difficulty options are localized (ja), not the hardcoded Easy/Normal/Hard.
    expect(screen.getByRole('option', { name: 'かんたん' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'ふつう' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'むずかしい' })).toBeInTheDocument();
  });

  it('shows bid controls on the human bid turn', async () => {
    renderWithProviders(<BidWhistPage />);
    expect(await screen.findByTestId('pass-button')).toBeEnabled();
  });

  it('shows a visible label for the trick-count selector', async () => {
    renderWithProviders(<BidWhistPage />);
    const label = await screen.findByTestId('bw-tricks-label');
    expect(label).toHaveTextContent('トリック数を選択');
  });

  it('bids a direction when a direction button is clicked', async () => {
    renderWithProviders(<BidWhistPage />);
    const uptown = await screen.findByRole('button', { name: 'アップタウン' });
    fireEvent.click(uptown);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bidTricks: 1, bidDirection: 0 }));
  });

  it('passes when pass is clicked', async () => {
    renderWithProviders(<BidWhistPage />);
    fireEvent.click(await screen.findByTestId('pass-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  it('declares a trump suit during the trump declaration phase', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 1,
        declarerIdx: 0,
        players: [
          player(0, true, sixCards, { isDeclarer: true }),
          player(1, false, []),
          player(2, false, []),
          player(3, false, []),
        ],
      }),
    );
    renderWithProviders(<BidWhistPage />);
    const spade = await screen.findByTestId('trump-1');
    fireEvent.click(spade);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', { trumpSuit: 1 }));
  });

  it('plays a selected card during the play phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 3, currentPlayerIdx: 0 }));
    renderWithProviders(<BidWhistPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    fireEvent.click(await screen.findByTestId('play-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('shows the kitty selection progress and fills the bar at six cards', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        declarerIdx: 0,
        players: [
          player(0, true, sixCards, { isDeclarer: true }),
          player(1, false, []),
          player(2, false, []),
          player(3, false, []),
        ],
      }),
    );
    renderWithProviders(<BidWhistPage />);
    const progress = await screen.findByTestId('kitty-progress');
    expect(progress).toHaveTextContent('0 / 6');
    // Bar is not yet full → info colour, no success colour.
    expect(progress.querySelector('.bg-ds-success')).toBeNull();

    // Select all six cards → counter reaches 6/6 and the bar turns success.
    for (let i = 0; i < 6; i++) {
      fireEvent.click(screen.getByTestId(`hand-card-${i}`));
    }
    expect(progress).toHaveTextContent('6 / 6');
    expect(progress.querySelector('.bg-ds-success')).not.toBeNull();
  });

  it('hides the kitty progress outside the exchange phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 3, currentPlayerIdx: 0 }));
    renderWithProviders(<BidWhistPage />);
    await screen.findByTestId('hand-card-0');
    expect(screen.queryByTestId('kitty-progress')).not.toBeInTheDocument();
  });
});
