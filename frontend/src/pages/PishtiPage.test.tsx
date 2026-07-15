import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { pishtiApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, PishtiPlayer, PishtiResponse } from '../types/card';
import { PishtiPage } from './PishtiPage';

vi.mock('../api/gameApi', () => ({
  pishtiApi: { exec: vi.fn() },
  actionLogApi: { pishti: vi.fn() },
}));

const mockExec = vi.mocked(pishtiApi.exec);

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makePlayer(overrides: Partial<PishtiPlayer> = {}): PishtiPlayer {
  return {
    id: 1,
    isHuman: false,
    cardCount: 4,
    cards: [],
    capturedCount: 0,
    pistiBonus: 0,
    finalScore: 0,
    ...overrides,
  };
}

function makeState(overrides: Partial<PishtiResponse> = {}): PishtiResponse {
  return {
    players: [
      makePlayer({
        id: 0,
        isHuman: true,
        cards: [card('SPADE', 5), card('HEART', 11), card('DIAMOND', 1), card('CLOVER', 9)],
      }),
      makePlayer({ id: 1 }),
      makePlayer({ id: 2 }),
      makePlayer({ id: 3 }),
    ],
    currentTurn: 0,
    pile: [card('CLOVER', 7)],
    pileTop: card('CLOVER', 7),
    pileCount: 1,
    lastCaptureIdx: -1,
    gameEndFlag: false,
    phase: 'play',
    remainingDeck: 36,
    winners: [],
    finalScores: [],
    config: { playerCnt: 4, cpuDifficulty: 1 },
    message: '',
    ...overrides,
  };
}

const playState = makeState();
const emptyPileState = makeState({ pile: [], pileTop: null, pileCount: 0 });
const gameEndState = makeState({
  phase: 'gameEnd',
  gameEndFlag: true,
  currentTurn: -1,
  winners: [0],
  finalScores: [11, 7, 4, 8],
  players: [
    makePlayer({ id: 0, isHuman: true, cardCount: 0, cards: [], capturedCount: 28, pistiBonus: 10, finalScore: 11 }),
    makePlayer({ id: 1, cardCount: 0, finalScore: 7 }),
    makePlayer({ id: 2, cardCount: 0, finalScore: 4 }),
    makePlayer({ id: 3, cardCount: 0, finalScore: 8 }),
  ],
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playState);
});

describe('PishtiPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<PishtiPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<PishtiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the remaining stock', async () => {
    renderWithProviders(<PishtiPage />);
    await waitFor(() => expect(screen.getByText(/山札: 36枚/)).toBeInTheDocument());
  });

  it('renders the players list with captured counts', async () => {
    renderWithProviders(<PishtiPage />);
    await waitFor(() => expect(screen.getByText('プレイヤー')).toBeInTheDocument());
    expect(screen.getAllByText(/捕獲 0枚/).length).toBe(4);
  });

  it('renders the human hand cards', async () => {
    renderWithProviders(<PishtiPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('hand-card-3')).toBeInTheDocument();
  });

  it('names each hand card in its aria-label and announces the turn', async () => {
    renderWithProviders(<PishtiPage />);
    // hand[0] is ♠5 → "♠ 5 を出す".
    expect(await screen.findByRole('button', { name: '♠ 5 を出す' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♦ A を出す' })).toBeInTheDocument();
    // The turn notice is a live region so turn arrival is announced.
    expect(screen.getByTestId('pishti-turn-notice')).toHaveAttribute('role', 'status');
  });

  it('plays a hand card on the human turn', async () => {
    renderWithProviders(<PishtiPage />);
    const cardBtn = await screen.findByTestId('hand-card-1');
    mockExec.mockClear();
    fireEvent.click(cardBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { handIndex: 1 }));
  });

  it('celebrates a Pişti when the bonus total rises (+10)', async () => {
    renderWithProviders(<PishtiPage />);
    const cardBtn = await screen.findByTestId('hand-card-1');
    expect(screen.queryByTestId('pishti-celebration')).not.toBeInTheDocument();
    // The play resolves a state where the human just scored a +10 Pişti.
    mockExec.mockResolvedValueOnce(
      makeState({ players: [makePlayer({ id: 0, isHuman: true, pistiBonus: 10 }), ...playState.players.slice(1)] }),
    );
    fireEvent.click(cardBtn);
    const badge = await screen.findByTestId('pishti-celebration');
    expect(badge).toHaveTextContent('+10');
  });

  it('marks a Jack Pişti as +20', async () => {
    renderWithProviders(<PishtiPage />);
    const cardBtn = await screen.findByTestId('hand-card-1');
    expect(screen.queryByTestId('pishti-celebration')).not.toBeInTheDocument();
    mockExec.mockResolvedValueOnce(
      makeState({ players: [makePlayer({ id: 0, isHuman: true, pistiBonus: 20 }), ...playState.players.slice(1)] }),
    );
    fireEvent.click(cardBtn);
    const badge = await screen.findByTestId('pishti-celebration');
    expect(badge).toHaveTextContent('+20');
  });

  it('does not fake a +20 Jack when two players each score +10 in one response', async () => {
    renderWithProviders(<PishtiPage />);
    const cardBtn = await screen.findByTestId('hand-card-1');
    // Human +10 and a CPU +10 land in the same response → aggregate 20, but neither is a Jack.
    mockExec.mockResolvedValueOnce(
      makeState({
        players: [
          makePlayer({ id: 0, isHuman: true, pistiBonus: 10 }),
          makePlayer({ id: 1, pistiBonus: 10 }),
          makePlayer({ id: 2 }),
          makePlayer({ id: 3 }),
        ],
      }),
    );
    fireEvent.click(cardBtn);
    const badge = await screen.findByTestId('pishti-celebration');
    expect(badge).toHaveTextContent('+10');
    expect(badge).not.toHaveTextContent('+20');
  });

  it('clears the celebration badge when a new game resets the bonuses', async () => {
    renderWithProviders(<PishtiPage />);
    const cardBtn = await screen.findByTestId('hand-card-1');
    mockExec.mockResolvedValueOnce(
      makeState({ players: [makePlayer({ id: 0, isHuman: true, pistiBonus: 10 }), ...playState.players.slice(1)] }),
    );
    fireEvent.click(cardBtn);
    await screen.findByTestId('pishti-celebration');
    // A reset (via settings change) resolves a fresh game with bonuses back to 0.
    fireEvent.change(screen.getByLabelText('CPU難易度'), { target: { value: '2' } });
    await waitFor(() => expect(screen.queryByTestId('pishti-celebration')).not.toBeInTheDocument());
  });

  it('shows the empty-pile label when the pile is empty', async () => {
    mockExec.mockResolvedValue(emptyPileState);
    renderWithProviders(<PishtiPage />);
    await waitFor(() => expect(screen.getByText('場札なし')).toBeInTheDocument());
  });

  it('does not dispatch play when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentTurn: 2 }));
    renderWithProviders(<PishtiPage />);
    const cardBtn = await screen.findByTestId('hand-card-0');
    mockExec.mockClear();
    fireEvent.click(cardBtn);
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('shows the win message and a next-game button when the human wins', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PishtiPage />);
    await waitFor(() => expect(screen.getByText('あなたの勝利です！')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '次のゲームへ' })).toBeInTheDocument();
  });

  it('dispatches next when next-game is clicked', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PishtiPage />);
    const btn = await screen.findByRole('button', { name: '次のゲームへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('shows final scores on game end', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PishtiPage />);
    await waitFor(() => expect(screen.getByText(/11点/)).toBeInTheDocument());
  });

  it('changes CPU difficulty via the settings panel and resets', async () => {
    renderWithProviders(<PishtiPage />);
    await waitFor(() => expect(screen.getByText('プレイヤー')).toBeInTheDocument());
    mockExec.mockClear();
    const select = screen.getByLabelText('CPU難易度');
    fireEvent.change(select, { target: { value: '2' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 2, playerCnt: 4 } }));
  });

  it('changes player count via the settings panel and resets', async () => {
    renderWithProviders(<PishtiPage />);
    await waitFor(() => expect(screen.getByText('プレイヤー')).toBeInTheDocument());
    mockExec.mockClear();
    const select = screen.getByLabelText('プレイヤー数');
    fireEvent.change(select, { target: { value: '2' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, playerCnt: 2 } }));
  });

  it('toggles the CLI terminal', async () => {
    renderWithProviders(<PishtiPage />);
    await waitFor(() => expect(screen.getByText('プレイヤー')).toBeInTheDocument());
    const toggle = screen.getByRole('button', { name: /CLI/i });
    fireEvent.click(toggle);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
  });
});
