import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fiveHundredApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, FiveHundredResponse } from '../types/card';
import { FiveHundredPage } from './FiveHundredPage';

vi.mock('../api/gameApi', () => ({
  fiveHundredApi: { exec: vi.fn() },
  actionLogApi: { fivehundred: vi.fn() },
}));

const mockExec = vi.mocked(fiveHundredApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function player(
  id: number,
  isHuman: boolean,
  cards: Card[],
  over: Partial<FiveHundredResponse['players'][number]> = {},
) {
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

function makeState(overrides: Partial<FiveHundredResponse> = {}): FiveHundredResponse {
  return {
    players: [
      player(0, true, [card('SPADE', 5), card('HEART', 6), card('DIAMOND', 7)]),
      player(1, false, []),
      player(2, false, []),
      player(3, false, []),
    ],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    leadPlayerIdx: 0,
    trumpSuit: -1,
    contractKind: 0,
    contractTricks: 0,
    contractValue: 0,
    declarerIdx: -1,
    highestBid: null,
    highestBidder: -1,
    jokerLeadSuit: -1,
    kittyCount: 3,
    currentTrick: [],
    teamScores: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { cpuDifficulty: 1, targetScore: 500 },
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('FiveHundredPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<FiveHundredPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', expect.objectContaining({ config: expect.any(Object) })),
    );
  });

  it('labels hand cards with cardAlt and reflects selection via aria-pressed during play', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2, currentPlayerIdx: 0 })); // PLAY, human turn
    renderWithProviders(<FiveHundredPage />);
    const card0 = await screen.findByTestId('hand-card-0');
    expect(card0).toHaveAttribute('aria-label', '♠ 5');
    expect(card0).toHaveAttribute('aria-pressed', 'false');
    expect(card0).not.toBeDisabled();
    fireEvent.click(card0);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toHaveAttribute('aria-pressed', 'true'));
  });

  it('keeps hand cards labeled but disabled in a non-selectable phase (bid)', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 0 })); // BID
    renderWithProviders(<FiveHundredPage />);
    const card0 = await screen.findByTestId('hand-card-0');
    expect(card0).toHaveAttribute('aria-label', '♠ 5');
    expect(card0).toBeDisabled();
  });

  it('renders CPU difficulty options with localized labels', async () => {
    renderWithProviders(<FiveHundredPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // Difficulty options are localized (ja), not the hardcoded Easy/Normal/Hard.
    expect(screen.getByRole('option', { name: 'かんたん' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'ふつう' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'むずかしい' })).toBeInTheDocument();
  });

  it('shows bid controls on the human bid turn', async () => {
    renderWithProviders(<FiveHundredPage />);
    expect(await screen.findByTestId('pass-button')).toBeEnabled();
  });

  it('bids a suit when a suit button is clicked', async () => {
    renderWithProviders(<FiveHundredPage />);
    const spade = await screen.findByTestId('fh-bid-suit-1');
    fireEvent.click(spade);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bidKind: 1, bidTricks: 6, bidSuit: 1 }));
  });

  it('shows a visible label associated with the trick-count selector', async () => {
    renderWithProviders(<FiveHundredPage />);
    const label = await screen.findByTestId('fh-tricks-label');
    expect(label).toHaveTextContent('トリック数を選択');
    expect(label).toHaveAttribute('for', 'fh-bid-tricks');
  });

  it('shows the Avondale score under each bid button (6 = ♠40 / ♥100 / NT120)', async () => {
    renderWithProviders(<FiveHundredPage />);
    expect(await screen.findByTestId('fh-bid-suit-1')).toHaveTextContent('40'); // 6♠
    expect(screen.getByTestId('fh-bid-suit-3')).toHaveTextContent('100'); // 6♥
    expect(screen.getByTestId('fh-bid-nt')).toHaveTextContent('120'); // 6NT
    expect(screen.getByTestId('fh-bid-misere')).toHaveTextContent('250');
    expect(screen.getByTestId('fh-bid-open-misere')).toHaveTextContent('520');
  });

  it('updates suit/NT scores live when the trick count changes (7♠=140, 7NT=220)', async () => {
    renderWithProviders(<FiveHundredPage />);
    await screen.findByTestId('fh-bid-suit-1');
    fireEvent.change(screen.getByLabelText(/トリック数を選択/), { target: { value: '7' } });
    expect(screen.getByTestId('fh-bid-suit-1')).toHaveTextContent('140'); // 7♠
    expect(screen.getByTestId('fh-bid-nt')).toHaveTextContent('220'); // 7NT
  });

  it('passes when pass is clicked', async () => {
    renderWithProviders(<FiveHundredPage />);
    fireEvent.click(await screen.findByTestId('pass-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  it('bids no-trump, misere and open misere', async () => {
    renderWithProviders(<FiveHundredPage />);
    fireEvent.click(await screen.findByTestId('fh-bid-nt'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bidKind: 2, bidTricks: 6 }));
    fireEvent.click(screen.getByTestId('fh-bid-misere'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bidKind: 3 }));
    fireEvent.click(screen.getByTestId('fh-bid-open-misere'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bidKind: 4 }));
  });

  it('nominates a suit when leading the joker in no-trump', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        currentPlayerIdx: 0,
        contractKind: 2,
        trumpSuit: -1,
        currentTrick: [],
        players: [
          player(0, true, [card('JOKER', 0)]),
          player(1, false, []),
          player(2, false, []),
          player(3, false, []),
        ] as FiveHundredResponse['players'],
      }),
    );
    renderWithProviders(<FiveHundredPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    expect(screen.queryByTestId('play-button')).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: '♠' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0, jokerSuit: 1 }));
  });

  it('exchanges three selected cards in the kitty exchange phase', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: 1, declarerIdx: 0, contractKind: 1, contractValue: 40, trumpSuit: 1 }),
    );
    renderWithProviders(<FiveHundredPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('hand-card-1'));
    fireEvent.click(screen.getByTestId('hand-card-2'));
    const exchangeBtn = screen.getByTestId('exchange-button');
    expect(exchangeBtn).toBeEnabled();
    fireEvent.click(exchangeBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('exchange', { discardIndices: [0, 1, 2] }));
  });

  it('plays a selected card in the play phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2, currentPlayerIdx: 0, contractKind: 1, trumpSuit: 1 }));
    renderWithProviders(<FiveHundredPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    const playBtn = screen.getByTestId('play-button');
    expect(playBtn).toBeEnabled();
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0, jokerSuit: undefined }));
  });

  it('advances to the next trick', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 3 }));
    renderWithProviders(<FiveHundredPage />);
    fireEvent.click(await screen.findByTestId('next-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('advances to the next round', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 4 }));
    renderWithProviders(<FiveHundredPage />);
    fireEvent.click(await screen.findByTestId('nextround-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });
});
