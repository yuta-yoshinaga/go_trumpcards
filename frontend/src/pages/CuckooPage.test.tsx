import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cuckooApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CuckooPlayer, CuckooResponse } from '../types/card';
import { CuckooPage } from './CuckooPage';

vi.mock('../api/gameApi', () => ({
  cuckooApi: { exec: vi.fn() },
  actionLogApi: { cuckoo: vi.fn() },
}));

const mockExec = vi.mocked(cuckooApi.exec);

function makePlayer(overrides: Partial<CuckooPlayer> = {}): CuckooPlayer {
  return {
    id: 1,
    isHuman: false,
    card: null,
    lives: 3,
    isEliminated: false,
    kingRevealed: false,
    isCurrentTurn: false,
    ...overrides,
  };
}

function makeState(overrides: Partial<CuckooResponse> = {}): CuckooResponse {
  return {
    players: [
      makePlayer({ id: 0, isHuman: true, card: { design: 'SPADE', value: 5 }, isCurrentTurn: true }),
      makePlayer({ id: 1 }),
      makePlayer({ id: 2 }),
      makePlayer({ id: 3 }),
    ],
    phase: 0,
    roundNumber: 1,
    currentPlayerIdx: 0,
    dealerIdx: 3,
    stockCount: 47,
    gameEndFlag: false,
    winnerIdx: -1,
    pendingSwapFrom: -1,
    pendingSwapTo: -1,
    roundLowest: -1,
    roundLosers: [],
    config: { cpuDifficulty: 1, initialLives: 3 },
    message: '',
    ...overrides,
  };
}

/** Players for a refuse-phase state where the human holds a card of `humanValue`. */
function refusePlayers(humanValue: number): CuckooPlayer[] {
  return [
    makePlayer({ id: 0, isHuman: true, card: { design: 'SPADE', value: humanValue }, isCurrentTurn: false }),
    makePlayer({ id: 1 }),
    makePlayer({ id: 2 }),
    makePlayer({ id: 3, isCurrentTurn: true }),
  ];
}

const turnState = makeState();
// Human is targeted for a swap and holds a King (value 13) → may refuse.
const refuseState = makeState({
  phase: 1,
  currentPlayerIdx: 0,
  pendingSwapFrom: 3,
  pendingSwapTo: 0,
  players: refusePlayers(13),
});
// Human is targeted for a swap but holds a non-King (value 5) → cannot refuse.
const refuseNoKingState = makeState({
  phase: 1,
  currentPlayerIdx: 0,
  pendingSwapFrom: 3,
  pendingSwapTo: 0,
  players: refusePlayers(5),
});
const roundEndState = makeState({ phase: 2, currentPlayerIdx: 0, roundLowest: 5, roundLosers: [0] });
const gameEndState = makeState({
  phase: 3,
  gameEndFlag: true,
  winnerIdx: 0,
  players: [
    makePlayer({ id: 0, isHuman: true, lives: 2, card: { design: 'SPADE', value: 5 } }),
    makePlayer({ id: 1, lives: 0, isEliminated: true }),
    makePlayer({ id: 2, lives: 0, isEliminated: true }),
    makePlayer({ id: 3, lives: 0, isEliminated: true }),
  ],
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(turnState);
});

describe('CuckooPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<CuckooPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<CuckooPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows round, dealer and stock', async () => {
    renderWithProviders(<CuckooPage />);
    await waitFor(() => expect(screen.getByText(/ラウンド 1/)).toBeInTheDocument());
    expect(screen.getByText(/山札: 47枚/)).toBeInTheDocument();
  });

  it('renders the players list with lives, exposing the remaining count to screen readers', async () => {
    renderWithProviders(<CuckooPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤー/)).toBeInTheDocument());
    // Each life row's aria-label includes the remaining count (e.g. "ライフ 3"), not just "ライフ".
    const lives = screen.getAllByLabelText('ライフ 3');
    expect(lives.length).toBe(4);
    expect(screen.queryByLabelText('ライフ')).not.toBeInTheDocument();
  });

  it('labels an eliminated player with the out label instead of a life count', async () => {
    mockExec.mockResolvedValue(gameEndState); // seats 1-3 eliminated (lives 0)
    renderWithProviders(<CuckooPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤー/)).toBeInTheDocument());
    expect(screen.getAllByLabelText('脱落').length).toBe(3);
  });

  it('announces round losers via a polite live region', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CuckooPage />);
    const losers = await screen.findByTestId('cuckoo-losers');
    expect(losers).toHaveAttribute('role', 'status');
    expect(losers).toHaveAttribute('aria-live', 'polite');
  });

  it('dispatches keep on the human turn', async () => {
    renderWithProviders(<CuckooPage />);
    const btn = await screen.findByRole('button', { name: 'キープ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('keep'));
  });

  it('dispatches swap on the human turn', async () => {
    renderWithProviders(<CuckooPage />);
    const btn = await screen.findByRole('button', { name: '交換' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('swap'));
  });

  it('shows the stock-swap label when the human is the dealer', async () => {
    mockExec.mockResolvedValue(makeState({ dealerIdx: 0 }));
    renderWithProviders(<CuckooPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札と交換' })).toBeInTheDocument());
  });

  it('dispatches refuse via the King-holder label when the human is the swap target', async () => {
    mockExec.mockResolvedValue(refuseState);
    renderWithProviders(<CuckooPage />);
    const btn = await screen.findByRole('button', { name: /キングを公開して拒否/ });
    expect(btn).toBeEnabled();
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('refuse'));
  });

  it('disables the refuse button with a reason when the human holds no King', async () => {
    mockExec.mockResolvedValue(refuseNoKingState);
    renderWithProviders(<CuckooPage />);
    const btn = await screen.findByTestId('cuckoo-refuse-button');
    expect(btn).toBeDisabled();
    expect(btn).toHaveAttribute('title', '拒否できるのはキングを持っているときだけです。');
    // The accept button remains available.
    expect(screen.getByRole('button', { name: '受け入れる' })).toBeEnabled();
  });

  it('shows a King-aware refuse notice for each branch', async () => {
    mockExec.mockResolvedValue(refuseState);
    const { unmount } = renderWithProviders(<CuckooPage />);
    expect(await screen.findByTestId('cuckoo-refuse-notice')).toHaveTextContent('公開して拒否できます');
    unmount();
    mockExec.mockResolvedValue(refuseNoKingState);
    renderWithProviders(<CuckooPage />);
    expect(await screen.findByTestId('cuckoo-refuse-notice')).toHaveTextContent(
      'キングを持っていないため拒否できません',
    );
  });

  it('dispatches accept when the human is the swap target', async () => {
    mockExec.mockResolvedValue(refuseState);
    renderWithProviders(<CuckooPage />);
    const btn = await screen.findByRole('button', { name: '受け入れる' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('accept'));
  });

  it('supports k/s keyboard shortcuts on the human turn', async () => {
    renderWithProviders(<CuckooPage />);
    await screen.findByRole('button', { name: 'キープ' });
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'k' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('keep'));
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 's' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('swap'));
  });

  it('supports r/a keyboard shortcuts only when targeted for a swap', async () => {
    mockExec.mockResolvedValue(refuseState);
    renderWithProviders(<CuckooPage />);
    await screen.findByRole('button', { name: '受け入れる' });
    mockExec.mockClear();
    // keep/swap keys are inactive in the refuse phase.
    fireEvent.keyDown(document, { key: 'k' });
    fireEvent.keyDown(document, { key: 'a' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('accept'));
    expect(mockExec).not.toHaveBeenCalledWith('keep');
  });

  it('shows keyboard hints on the action buttons', async () => {
    renderWithProviders(<CuckooPage />);
    const keep = await screen.findByRole('button', { name: 'キープ' });
    expect(keep.querySelector('kbd')).toHaveTextContent('K');
  });

  it('shows the round-loser reveal and dispatches nextround', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CuckooPage />);
    await waitFor(() => expect(screen.getByText(/今ラウンドの敗者/)).toBeInTheDocument());
    const btn = screen.getByRole('button', { name: '次のラウンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('shows the win message when the human wins', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CuckooPage />);
    await waitFor(() => expect(screen.getByText('あなたの勝利です！')).toBeInTheDocument());
  });

  it('changes CPU difficulty via the settings panel and resets', async () => {
    renderWithProviders(<CuckooPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤー/)).toBeInTheDocument());
    mockExec.mockClear();
    const select = screen.getByLabelText('CPU難易度');
    fireEvent.change(select, { target: { value: '2' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 2 } }));
  });

  it('toggles the CLI terminal', async () => {
    renderWithProviders(<CuckooPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤー/)).toBeInTheDocument());
    const toggle = screen.getByRole('button', { name: /CLI/i });
    fireEvent.click(toggle);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
  });
});
