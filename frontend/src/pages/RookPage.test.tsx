import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { rookApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, RookResponse } from '../types/card';
import { RookPage } from './RookPage';

vi.mock('../api/gameApi', () => ({
  rookApi: { exec: vi.fn() },
  actionLogApi: { rook: vi.fn() },
}));

const mockExec = vi.mocked(rookApi.exec);

const card = (label: string, value: number, color = 'red'): Card =>
  ({ design: 'SPADE', value, label, glyph: label, color, deck: 'rook' }) as unknown as Card;

function player(id: number, isHuman: boolean, cards: Card[], over: Partial<RookResponse['players'][number]> = {}) {
  return {
    id,
    isHuman,
    cardCount: cards.length,
    cards,
    team: id % 2,
    trickCount: 0,
    points: 0,
    bid: 0,
    passed: false,
    isDeclarer: false,
    ...over,
  };
}

function makeState(overrides: Partial<RookResponse> = {}): RookResponse {
  return {
    players: [
      player(0, true, [card('5', 5), card('10', 10, 'green'), card('14', 14, 'black')]),
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
    trumpColor: -1,
    contractBid: 0,
    declarerIdx: -1,
    highestBid: 0,
    highestBidder: -1,
    nestCount: 5,
    nest: [],
    currentTrick: [],
    teamScores: [0, 0],
    teamPoints: [0, 0],
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

describe('RookPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<RookPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', expect.objectContaining({ config: expect.any(Object) })),
    );
  });

  it('shows bid controls on the human bid turn', async () => {
    renderWithProviders(<RookPage />);
    expect(await screen.findByTestId('pass-button')).toBeEnabled();
    expect(screen.getByTestId('bid-button')).toBeEnabled();
  });

  it('shows a visible label associated with the bid selector', async () => {
    renderWithProviders(<RookPage />);
    const label = await screen.findByTestId('rook-bid-label');
    expect(label).toHaveAttribute('for', 'rook-bid');
  });

  it('bids the selected value', async () => {
    renderWithProviders(<RookPage />);
    fireEvent.change(await screen.findByLabelText(/70/), { target: { value: '85' } });
    fireEvent.click(screen.getByTestId('bid-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 85 }));
  });

  it('offers only bids above the current highest bid', async () => {
    mockExec.mockResolvedValue(makeState({ highestBid: 90, highestBidder: 2 }));
    renderWithProviders(<RookPage />);
    const select = (await screen.findByLabelText(/選択/)) as HTMLSelectElement;
    const values = Array.from(select.options).map((o) => o.value);
    expect(values).toEqual(['95', '100', '105', '110', '115', '120']);
  });

  it('shows the bid status with the highest bid and active bidder count on the human bid turn', async () => {
    mockExec.mockResolvedValue(
      makeState({
        highestBid: 80,
        highestBidder: 2,
        players: [
          player(0, true, [card('5', 5)]),
          player(1, false, [], { passed: true }),
          player(2, false, []),
          player(3, false, []),
        ] as RookResponse['players'],
      }),
    );
    renderWithProviders(<RookPage />);
    const status = await screen.findByTestId('rook-bid-status');
    expect(status).toHaveTextContent('現在最高: 80点');
    expect(status).toHaveTextContent('残り入札者: 3人');
    expect(status).toHaveTextContent('パス済み: CPU 1');
  });

  it('shows the bid status as undecided when no bid has been made yet', async () => {
    renderWithProviders(<RookPage />);
    const status = await screen.findByTestId('rook-bid-status');
    expect(status).toHaveTextContent('現在最高: 未決定');
    expect(status).toHaveTextContent('残り入札者: 4人');
  });

  it('hides the bid status outside the bid phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2, currentPlayerIdx: 0, contractBid: 75, trumpColor: 1 }));
    renderWithProviders(<RookPage />);
    await screen.findByTestId('play-button');
    expect(screen.queryByTestId('rook-bid-status')).not.toBeInTheDocument();
  });

  it('passes when pass is clicked', async () => {
    renderWithProviders(<RookPage />);
    fireEvent.click(await screen.findByTestId('pass-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  it('exchanges five cards and a trump color in the nest exchange phase', async () => {
    const cards = [card('5', 5), card('6', 6), card('7', 7), card('8', 8), card('9', 9), card('10', 10)];
    mockExec.mockResolvedValue(
      makeState({
        phase: 1,
        declarerIdx: 0,
        contractBid: 75,
        players: [
          player(0, true, cards, { isDeclarer: true }),
          player(1, false, []),
          player(2, false, []),
          player(3, false, []),
        ] as RookResponse['players'],
      }),
    );
    renderWithProviders(<RookPage />);
    for (let i = 0; i < 5; i++) fireEvent.click(await screen.findByTestId(`hand-card-${i}`));
    const exchangeBtn = screen.getByTestId('exchange-button');
    // Each colour button carries a letter cue in addition to the colour name.
    expect(screen.getByTestId('trump-choice-1')).toHaveTextContent('R');
    expect(screen.getByTestId('trump-choice-3')).toHaveTextContent('G');
    // still disabled until a trump color is chosen
    expect(exchangeBtn).toBeDisabled();
    fireEvent.click(screen.getByTestId('trump-choice-3'));
    expect(exchangeBtn).toBeEnabled();
    fireEvent.click(exchangeBtn);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('exchange', { discardIndices: [0, 1, 2, 3, 4], trumpColor: 3 }),
    );
  });

  it('highlights the backend-recommended discard cards during nest exchange when hints are enabled', async () => {
    localStorage.setItem('hint_enabled_rook', 'true');
    const cards = [card('5', 5), card('6', 6), card('7', 7), card('8', 8), card('9', 9), card('10', 10)];
    const exchangeState = makeState({
      phase: 1,
      declarerIdx: 0,
      contractBid: 75,
      players: [
        player(0, true, cards, { isDeclarer: true }),
        player(1, false, []),
        player(2, false, []),
        player(3, false, []),
      ] as RookResponse['players'],
    });
    mockExec.mockImplementation((cmd) =>
      cmd === 'hint'
        ? Promise.resolve({ ...exchangeState, hint: { discardIndices: [1, 3], reason: 'discard_weakest' } })
        : Promise.resolve(exchangeState),
    );
    renderWithProviders(<RookPage />);

    // The backend hint command is queried to obtain the weak cards.
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
    await waitFor(() => expect(screen.getByTestId('hand-card-1')).toHaveAttribute('data-recommended-discard', 'true'));
    const c1 = screen.getByTestId('hand-card-1');
    const c3 = screen.getByTestId('hand-card-3');
    expect(c1.className).toContain('ring-ds-warning');
    expect(c3).toHaveAttribute('data-recommended-discard', 'true');
    // Non-recommended cards carry no discard highlight.
    expect(screen.getByTestId('hand-card-0')).not.toHaveAttribute('data-recommended-discard');
    // A non-colour cue explains the pale outline.
    expect(screen.getByTestId('rook-discard-hint')).toBeInTheDocument();
  });

  it('shows no discard highlight during nest exchange when hints are disabled', async () => {
    const cards = [card('5', 5), card('6', 6), card('7', 7), card('8', 8), card('9', 9), card('10', 10)];
    mockExec.mockResolvedValue(
      makeState({
        phase: 1,
        declarerIdx: 0,
        contractBid: 75,
        players: [
          player(0, true, cards, { isDeclarer: true }),
          player(1, false, []),
          player(2, false, []),
          player(3, false, []),
        ] as RookResponse['players'],
      }),
    );
    renderWithProviders(<RookPage />);
    await screen.findByTestId('exchange-button');
    expect(screen.getByTestId('hand-card-0')).not.toHaveAttribute('data-recommended-discard');
    expect(screen.queryByTestId('rook-discard-hint')).not.toBeInTheDocument();
    // No hint request is issued while hints are off.
    expect(mockExec).not.toHaveBeenCalledWith('hint');
  });

  it('plays a selected card in the play phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2, currentPlayerIdx: 0, contractBid: 75, trumpColor: 1 }));
    renderWithProviders(<RookPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    const playBtn = screen.getByTestId('play-button');
    expect(playBtn).toBeEnabled();
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('shows the trump color swatch once declared, named for screen readers', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2, currentPlayerIdx: 0, contractBid: 75, trumpColor: 1 }));
    renderWithProviders(<RookPage />);
    const swatch = await screen.findByTestId('trump-swatch');
    // The colour dot itself is named (not colour-only) for SR.
    expect(swatch).toHaveAttribute('role', 'img');
    expect(swatch).toHaveAttribute('aria-label', '赤');
    // The visible label carries the letter cue and is hidden from SR to avoid a repeat.
    expect(screen.getByTestId('trump-name')).toHaveTextContent('R 赤');
  });

  it('advances to the next trick', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 3 }));
    renderWithProviders(<RookPage />);
    fireEvent.click(await screen.findByTestId('next-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('advances to the next round', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 4 }));
    renderWithProviders(<RookPage />);
    fireEvent.click(await screen.findByTestId('nextround-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });
});
