import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { gostopApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeGoStopState } from '../test/stateFactories';
import { GoStopPage } from './GoStopPage';

vi.mock('../api/gameApi', () => ({
  gostopApi: { exec: vi.fn() },
  actionLogApi: { gostop: vi.fn() },
}));

const mockExec = vi.mocked(gostopApi.exec);

const playState = makeGoStopState();
const decisionState = makeGoStopState({
  phase: 1,
  pendingPoints: 7,
  pendingBreakdown: {
    gwang: 3,
    godori: 0,
    tti: 2,
    yeol: 1,
    pi: 1,
    base: 7,
    goCount: 0,
    goMult: 1,
    goScore: 7,
    brightCount: 3,
    ribbonCount: 5,
    animalCount: 5,
    piCount: 10,
  },
});
const roundEndState = makeGoStopState({
  phase: 2,
  roundWinner: 0,
  lastRoundResult: {
    winner: 0,
    breakdown: playState.players[0].breakdown,
    basePoints: 7,
    goScore: 7,
    bakMult: 2,
    total: 14,
    gwangBak: true,
    piBak: false,
    goBak: false,
    goCount: 0,
  },
});
const gameEndState = makeGoStopState({
  phase: 3,
  gameEndFlag: true,
  winner: 0,
  message: 'ゲーム終了！',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playState);
});

describe('GoStopPage', () => {
  it('renders the loading fallback when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<GoStopPage />);
    expect(screen.getByText('読み込み中…')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<GoStopPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders the breakdown chip row during play', async () => {
    renderWithProviders(<GoStopPage />);
    // Human + CPU each get a chip row.
    await waitFor(() => expect(screen.getAllByTestId('gostop-breakdown').length).toBeGreaterThan(0));
  });

  it('plays a hand card with a single match immediately', async () => {
    renderWithProviders(<GoStopPage />);
    const card = await screen.findByTestId('hand-card-0');
    mockExec.mockClear();
    fireEvent.click(card);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('requires a field pick for a two-way match, then plays with fieldIndex', async () => {
    mockExec.mockResolvedValue(makeGoStopState({ captureOptions: { 0: [0, 1] } }));
    renderWithProviders(<GoStopPage />);
    const card = await screen.findByTestId('hand-card-0');
    fireEvent.click(card);
    await waitFor(() => expect(screen.getByTestId('gostop-field-pick')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('field-card-1'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0, fieldIndex: 1 }));
  });

  it('shows the go / stop buttons on the decision phase and dispatches them', async () => {
    mockExec.mockResolvedValue(decisionState);
    renderWithProviders(<GoStopPage />);
    await waitFor(() => expect(screen.getByTestId('gostop-decision')).toBeInTheDocument());
    const goBtn = screen.getByRole('button', { name: 'ゴー' });
    const stopBtn = screen.getByRole('button', { name: 'ストップ' });
    mockExec.mockClear();
    fireEvent.click(goBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('go'));
    fireEvent.click(stopBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stop'));
  });

  it('previews a near-complete yaku with the remaining count on the decision panel', async () => {
    // brightCount 2 -> one card from samgwang (三光).
    mockExec.mockResolvedValue(
      makeGoStopState({
        phase: 1,
        pendingPoints: 3,
        pendingBreakdown: {
          gwang: 0,
          godori: 0,
          tti: 0,
          yeol: 0,
          pi: 0,
          base: 0,
          goCount: 0,
          goMult: 1,
          goScore: 0,
          brightCount: 2,
          ribbonCount: 0,
          animalCount: 0,
          piCount: 0,
        },
      }),
    );
    renderWithProviders(<GoStopPage />);
    await waitFor(() => expect(screen.getByTestId('gostop-yaku-preview')).toBeInTheDocument());
    expect(screen.getByTestId('gostop-yaku-preview-gwang')).toHaveTextContent('三光 あと1枚');
  });

  it('hides the yaku preview when the hand is far from every threshold', async () => {
    mockExec.mockResolvedValue(
      makeGoStopState({
        phase: 1,
        pendingPoints: 0,
        pendingBreakdown: {
          gwang: 0,
          godori: 0,
          tti: 0,
          yeol: 0,
          pi: 0,
          base: 0,
          goCount: 0,
          goMult: 1,
          goScore: 0,
          brightCount: 0,
          ribbonCount: 0,
          animalCount: 0,
          piCount: 0,
        },
      }),
    );
    renderWithProviders(<GoStopPage />);
    await waitFor(() => expect(screen.getByTestId('gostop-decision')).toBeInTheDocument());
    expect(screen.queryByTestId('gostop-yaku-preview')).not.toBeInTheDocument();
  });

  it('shows the next-round button at round end with a bak badge and dispatches nextround', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<GoStopPage />);
    await waitFor(() => expect(screen.getByTestId('gostop-round-result')).toBeInTheDocument());
    expect(screen.getByTestId('gostop-bak-gwang')).toBeInTheDocument();
    const btn = screen.getByRole('button', { name: '次のラウンド' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('renders the game-end result and dispatches a new game', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<GoStopPage />);
    await waitFor(() => expect(screen.getByTestId('gostop-result')).toBeInTheDocument());
    const newGame = screen.getByRole('button', { name: '新しいゲーム' });
    mockExec.mockClear();
    fireEvent.click(newGame);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', expect.objectContaining({ config: expect.any(Object) })),
    );
  });

  it('renders a drawn round-end result without bak badges', async () => {
    mockExec.mockResolvedValue(
      makeGoStopState({
        phase: 2,
        roundWinner: -1,
        lastRoundResult: {
          winner: -1,
          breakdown: null,
          basePoints: 0,
          goScore: 0,
          bakMult: 1,
          total: 0,
          gwangBak: false,
          piBak: false,
          goBak: false,
          goCount: 0,
        },
      }),
    );
    renderWithProviders(<GoStopPage />);
    await waitFor(() => expect(screen.getByTestId('gostop-round-result')).toBeInTheDocument());
    expect(screen.queryByTestId('gostop-bak-badges')).not.toBeInTheDocument();
  });

  it('does not play a hand card when it is the CPU turn', async () => {
    mockExec.mockResolvedValue(makeGoStopState({ isHumanTurn: false, currentTurn: 1 }));
    renderWithProviders(<GoStopPage />);
    const card = await screen.findByTestId('hand-card-0');
    expect(card).toBeDisabled();
    mockExec.mockClear();
    fireEvent.click(card);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores a field click that is not a capture candidate', async () => {
    mockExec.mockResolvedValue(makeGoStopState({ captureOptions: { 0: [0, 1] } }));
    renderWithProviders(<GoStopPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    await waitFor(() => expect(screen.getByTestId('gostop-field-pick')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('field-card-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0, fieldIndex: 0 }));
  });

  it('renders the backend hint tooltip when the hint toggle is enabled (#3519)', async () => {
    // #3519: the toggle was dead because state.hint was always undefined; Output() now populates it.
    localStorage.setItem('hint_enabled_gostop', 'true');
    mockExec.mockResolvedValue(makeGoStopState({ hint: { cardIndex: 0, fieldIndex: -1, go: -1, reason: 'capture' } }));
    renderWithProviders(<GoStopPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
    localStorage.removeItem('hint_enabled_gostop');
  });

  it('hides the hint tooltip when no hint is present even with the toggle enabled', async () => {
    localStorage.setItem('hint_enabled_gostop', 'true');
    mockExec.mockResolvedValue(makeGoStopState({ hint: null }));
    renderWithProviders(<GoStopPage />);
    await screen.findByTestId('hand-card-0');
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();
    localStorage.removeItem('hint_enabled_gostop');
  });
});
