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
    kittyIndices: [],
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

  it('highlights kitty-origin cards during the human exchange', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        declarerIdx: 0,
        kittyIndices: [1, 3],
        players: [
          player(0, true, sixCards, { isDeclarer: true }),
          player(1, false, []),
          player(2, false, []),
          player(3, false, []),
        ],
      }),
    );
    renderWithProviders(<BidWhistPage />);
    // Legend appears and only the flagged cards carry the warning ring + badge.
    expect(await screen.findByTestId('kitty-legend')).toBeInTheDocument();
    expect(screen.getByTestId('hand-card-1')).toHaveAttribute('data-kitty', 'true');
    expect(screen.getByTestId('hand-card-3')).toHaveAttribute('data-kitty', 'true');
    expect(screen.getByTestId('hand-card-1').className).toContain('ring-ds-warning');
    expect(screen.getByTestId('kitty-badge-1')).toBeInTheDocument();
    // Non-kitty cards are not flagged.
    expect(screen.getByTestId('hand-card-0')).not.toHaveAttribute('data-kitty');
    expect(screen.queryByTestId('kitty-badge-0')).not.toBeInTheDocument();
  });

  it('does not highlight kitty cards outside the exchange phase', async () => {
    // kittyIndices is only honoured during KITTY_EXCHANGE; in play it is ignored.
    mockExec.mockResolvedValue(makeState({ phase: 3, currentPlayerIdx: 0, kittyIndices: [1, 3] }));
    renderWithProviders(<BidWhistPage />);
    await screen.findByTestId('hand-card-0');
    expect(screen.queryByTestId('kitty-legend')).not.toBeInTheDocument();
    expect(screen.getByTestId('hand-card-1')).not.toHaveAttribute('data-kitty');
  });

  it('hides the kitty progress outside the exchange phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 3, currentPlayerIdx: 0 }));
    renderWithProviders(<BidWhistPage />);
    await screen.findByTestId('hand-card-0');
    expect(screen.queryByTestId('kitty-progress')).not.toBeInTheDocument();
  });

  it('enables every direction when there is no highest bid', async () => {
    renderWithProviders(<BidWhistPage />);
    // Default selected tricks is 1; with no highest bid, all directions are valid.
    expect(await screen.findByTestId('bid-dir-0')).toBeEnabled();
    expect(screen.getByTestId('bid-dir-1')).toBeEnabled();
    expect(screen.getByTestId('bid-dir-2')).toBeEnabled();
    expect(screen.getByTestId('pass-button')).toBeEnabled();
  });

  it('disables directions that do not exceed the current highest bid', async () => {
    // Highest bid = 1 Uptown (order 10). Selected tricks default to 1, so:
    //   Uptown (order 10) is not strictly greater → disabled,
    //   Downtown (11) and No Trump (12) beat it → enabled.
    mockExec.mockResolvedValue(makeState({ highestBid: { tricks: 1, direction: 0 }, highestBidder: 3 }));
    renderWithProviders(<BidWhistPage />);
    const uptown = await screen.findByTestId('bid-dir-0');
    expect(uptown).toBeDisabled();
    expect(uptown).toHaveAttribute('aria-disabled', 'true');
    expect(screen.getByTestId('bid-dir-1')).toBeEnabled();
    expect(screen.getByTestId('bid-dir-2')).toBeEnabled();
    // Pass is always available.
    expect(screen.getByTestId('pass-button')).toBeEnabled();
  });

  it('surfaces a reason tooltip on a disabled direction button', async () => {
    mockExec.mockResolvedValue(makeState({ highestBid: { tricks: 3, direction: 2 }, highestBidder: 1 }));
    renderWithProviders(<BidWhistPage />);
    // Highest bid 3 No Trump (order 32) beats every tricks-1 bid → all disabled with a title.
    const wrap = await screen.findByTestId('bid-dir-wrap-0');
    expect(wrap).toHaveAttribute('title', expect.stringContaining('3'));
    expect(screen.getByTestId('bid-dir-0')).toBeDisabled();
  });

  it('disables lower-order directions but enables higher tricks after selecting them', async () => {
    mockExec.mockResolvedValue(makeState({ highestBid: { tricks: 4, direction: 0 }, highestBidder: 2 }));
    renderWithProviders(<BidWhistPage />);
    // With tricks 1 selected, nothing beats a 4-trick bid.
    expect(await screen.findByTestId('bid-dir-0')).toBeDisabled();
    // Raising the selector to 5 tricks makes all directions valid again (order 50+ > 40).
    fireEvent.change(screen.getByLabelText('トリック数を選択 (1-7):'), { target: { value: '5' } });
    expect(screen.getByTestId('bid-dir-0')).toBeEnabled();
    expect(screen.getByTestId('bid-dir-1')).toBeEnabled();
    expect(screen.getByTestId('bid-dir-2')).toBeEnabled();
  });
});
