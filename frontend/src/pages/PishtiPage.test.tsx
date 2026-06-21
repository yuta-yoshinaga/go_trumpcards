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

  it('plays a hand card on the human turn', async () => {
    renderWithProviders(<PishtiPage />);
    const cardBtn = await screen.findByTestId('hand-card-1');
    mockExec.mockClear();
    fireEvent.click(cardBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { handIndex: 1 }));
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
