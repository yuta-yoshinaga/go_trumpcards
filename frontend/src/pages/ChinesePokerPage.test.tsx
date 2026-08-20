import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, ChinesePokerResponse } from '../types/card';
import { ChinesePokerPhase } from '../types/phases';

vi.mock('../api/gameApi', () => ({
  chinesepokerApi: { exec: vi.fn() },
}));
vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

import { chinesepokerApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { ChinesePokerPage } from './ChinesePokerPage';

const mockExec = vi.mocked(chinesepokerApi.exec);

const DESIGNS: CardDesign[] = ['SPADE', 'HEART', 'CLOVER', 'DIAMOND'];
const card = (designIdx: number, value: number): Card => ({ design: DESIGNS[designIdx % 4], value });

const betPhaseState: ChinesePokerResponse = {
  playerCards: [],
  dealerCards: [],
  playerFront: [],
  playerMiddle: [],
  playerBack: [],
  dealerFront: [],
  dealerMiddle: [],
  dealerBack: [],
  phase: ChinesePokerPhase.BET,
  chips: 1000,
  bet: 0,
  result: 0,
  frontResult: 0,
  middleResult: 0,
  backResult: 0,
  payout: 0,
  playerFrontRank: 0,
  playerMiddleRank: 0,
  playerBackRank: 0,
  dealerFrontRank: 0,
  dealerMiddleRank: 0,
  dealerBackRank: 0,
  playerRoyalty: 0,
  dealerRoyalty: 0,
  scoop: false,
  message: '',
};

const setHandsState: ChinesePokerResponse = {
  ...betPhaseState,
  phase: ChinesePokerPhase.SET_HANDS,
  bet: 100,
  chips: 900,
  playerCards: Array.from({ length: 13 }, (_, i) => card(1 + (i % 4), 1 + i)),
};

const endPhaseState: ChinesePokerResponse = {
  ...betPhaseState,
  phase: ChinesePokerPhase.END,
  bet: 100,
  chips: 1100,
  result: 1,
  frontResult: 1,
  middleResult: 1,
  backResult: -1,
  payout: 200,
  playerFront: [card(1, 2), card(2, 3), card(3, 4)],
  playerMiddle: [card(1, 5), card(2, 6), card(3, 7), card(4, 8), card(1, 9)],
  playerBack: [card(2, 10), card(3, 11), card(4, 12), card(1, 13), card(2, 1)],
  dealerFront: [card(3, 2), card(4, 3), card(1, 4)],
  dealerMiddle: [card(2, 5), card(3, 6), card(4, 7), card(1, 8), card(2, 9)],
  dealerBack: [card(3, 10), card(4, 11), card(1, 12), card(2, 13), card(3, 1)],
  playerFrontRank: 1,
  playerMiddleRank: 4,
  playerBackRank: 4,
  dealerFrontRank: 1,
  dealerMiddleRank: 4,
  dealerBackRank: 4,
  message: 'Player wins!',
  messageCode: 'chinesepoker.result.playerWins',
};

describe('ChinesePokerPage', () => {
  it('renders bet phase with bet button', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ChinesePokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
  });

  it('calls bet API on bet button click', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ChinesePokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100));
  });

  it('renders set hands phase with 13 cards labeled by card name', async () => {
    mockExec.mockResolvedValue(setHandsState);
    renderWithProviders(<ChinesePokerPage />);
    await waitFor(() => expect(screen.getByTestId('set-hands-button')).toBeInTheDocument());
    // Hand cards are now named by suit+rank (cardAlt), not the hardcoded "Card N".
    const cardButtons = screen.getAllByRole('button', { name: /^[♠♥♣♦]/ });
    expect(cardButtons.length).toBe(13);
    expect(screen.queryByRole('button', { name: /Card \d+/ })).not.toBeInTheDocument();
  });

  it('includes the assigned row in a hand card aria-label once assigned', async () => {
    mockExec.mockResolvedValue(setHandsState);
    renderWithProviders(<ChinesePokerPage />);
    // First card is ♥ A (design idx 1, value 1).
    const firstCard = await screen.findByRole('button', { name: '♥ A' });
    // First tap assigns it to the front row.
    fireEvent.click(firstCard);
    await waitFor(() => expect(screen.getByRole('button', { name: '♥ A（フロント）' })).toBeInTheDocument());
  });

  it('shows the front/middle/back row preview and updates it on assignment', async () => {
    mockExec.mockResolvedValue(setHandsState);
    renderWithProviders(<ChinesePokerPage />);
    await screen.findByTestId('cp-row-preview');
    const back = () => within(screen.getByTestId('cp-row-back')).queryAllByTestId('animated-card');
    const front = () => within(screen.getByTestId('cp-row-front')).queryAllByTestId('animated-card');
    // Initially everything is unassigned → all 13 cards sit in the back row.
    expect(back()).toHaveLength(13);
    expect(front()).toHaveLength(0);
    const middle = () => within(screen.getByTestId('cp-row-middle')).queryAllByTestId('animated-card');
    // First click assigns the card to the front row.
    fireEvent.click(screen.getByTestId('cp-hand-card-0'));
    expect(front()).toHaveLength(1);
    expect(back()).toHaveLength(12);
    // Second click cycles front → middle.
    fireEvent.click(screen.getByTestId('cp-hand-card-0'));
    expect(front()).toHaveLength(0);
    expect(middle()).toHaveLength(1);
    expect(back()).toHaveLength(12);
  });

  it('set button disabled when not enough cards selected', async () => {
    mockExec.mockResolvedValue(setHandsState);
    renderWithProviders(<ChinesePokerPage />);
    await waitFor(() => expect(screen.getByTestId('set-hands-button')).toBeInTheDocument());
    expect(screen.getByTestId('set-hands-button')).toBeDisabled();
  });

  it('renders end phase with results', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    renderWithProviders(<ChinesePokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument();
  });

  it('calls reset on next game click', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    renderWithProviders(<ChinesePokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    expect(mockExec).toHaveBeenCalledWith('reset');
  });

  it('responds to keyboard shortcut b in bet phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ChinesePokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    fireEvent.keyDown(document, { key: 'b' });
    expect(mockExec).toHaveBeenCalledWith('bet', 100);
  });

  it('submits the edited bet amount via the ChipBetInput', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ChinesePokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    fireEvent.change(screen.getByLabelText('ベット'), { target: { value: '200' } });
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 200));
  });

  it('disables the bet button, blocks the b key, and shows an error for an out-of-range bet', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<ChinesePokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    // 15 is not a multiple of 10 → invalid.
    fireEvent.change(screen.getByLabelText('ベット'), { target: { value: '15' } });
    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();
    expect(screen.getByRole('alert')).toBeInTheDocument();
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'b' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  // Assign the cards at the given indices in order: the first 3 become the front
  // row, the next 5 the middle row, and the remaining 5 stay in the back row.
  const assignFrontMiddle = (frontMid: number[]) => {
    for (const i of frontMid) {
      fireEvent.click(screen.getByTestId(`cp-hand-card-${i}`));
    }
  };

  it('shows a foul warning when the staged arrangement violates back >= middle >= front', async () => {
    // Front = trips 9s (idx 0-2), Middle = pair 6s (idx 3-7), Back = quads 8s.
    // Front (trips) outranks Middle (pair) → foul.
    const foulCards: Card[] = [
      card(0, 9),
      card(1, 9),
      card(2, 9), // front: trips
      card(0, 6),
      card(1, 6),
      card(2, 2),
      card(3, 3),
      card(0, 4), // middle: pair of 6s
      card(0, 8),
      card(1, 8),
      card(2, 8),
      card(3, 8),
      card(0, 5), // back: quads of 8s
    ];
    mockExec.mockResolvedValue({ ...setHandsState, playerCards: foulCards });
    renderWithProviders(<ChinesePokerPage />);
    await screen.findByTestId('cp-row-preview');
    assignFrontMiddle([0, 1, 2, 3, 4, 5, 6, 7]);
    expect(screen.getByTestId('cp-foul-warning')).toBeInTheDocument();
  });

  it('does not show a foul warning for a legal arrangement', async () => {
    // Front = high card (idx 0-2), Middle = pair 6s (idx 3-7), Back = trips 8s.
    const legalCards: Card[] = [
      card(0, 2),
      card(1, 3),
      card(3, 5), // front: high card
      card(0, 6),
      card(1, 6),
      card(2, 10),
      card(3, 11),
      card(0, 12), // middle: pair of 6s
      card(0, 8),
      card(1, 8),
      card(2, 8),
      card(3, 2),
      card(0, 4), // back: trips 8s
    ];
    mockExec.mockResolvedValue({ ...setHandsState, playerCards: legalCards });
    renderWithProviders(<ChinesePokerPage />);
    await screen.findByTestId('cp-row-preview');
    // No assignment yet → incomplete → no warning.
    expect(screen.queryByTestId('cp-foul-warning')).not.toBeInTheDocument();
    assignFrontMiddle([0, 1, 2, 3, 4, 5, 6, 7]);
    expect(screen.queryByTestId('cp-foul-warning')).not.toBeInTheDocument();
  });
});

// #5615: ヒントは「ファウルの危険がある」としか言わず、**どの札をどこへ**は
// プレイヤーの試行錯誤だった。サーバー (CUI と同じ計算) が前列に勧める札を
// 名指しできるようになったので、その札を盤面で示す。
describe('ChinesePokerPage suggested front row', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('marks the cards the hint names for the front row', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: {
        targetAction: 'setHands',
        reason: 'frontendHint.chinesepokerSplit',
        confidence: 'moderate',
        targetIndices: [10, 11, 12],
      },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    mockExec.mockResolvedValue(setHandsState);
    renderWithProviders(<ChinesePokerPage />);

    await waitFor(() => expect(screen.getByTestId('cp-hand-card-10')).toHaveAttribute('data-hint-front', 'true'));
    expect(screen.getByTestId('cp-hand-card-12')).toHaveAttribute('data-hint-front', 'true');
    // 名指しされていない札は印を持たない。
    expect(screen.getByTestId('cp-hand-card-0')).not.toHaveAttribute('data-hint-front');
  });

  it('marks nothing while the hint is switched off', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: {
        targetAction: 'setHands',
        reason: 'frontendHint.chinesepokerSplit',
        confidence: 'moderate',
        targetIndices: [10, 11, 12],
      },
      hintEnabled: false,
      setHintEnabled: vi.fn(),
    });
    mockExec.mockResolvedValue(setHandsState);
    renderWithProviders(<ChinesePokerPage />);

    await waitFor(() => expect(screen.getByTestId('cp-hand-card-10')).toBeInTheDocument());
    expect(screen.getByTestId('cp-hand-card-10')).not.toHaveAttribute('data-hint-front');
  });
});
