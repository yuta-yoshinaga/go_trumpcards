import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { oldmaidApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { OldMaidResponse } from '../types/card';
import { OldMaidPage } from './OldMaidPage';

vi.mock('../api/gameApi', () => ({
  oldmaidApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(oldmaidApi.exec);

const humanTurnState: OldMaidResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      isFinished: false,
      cardCount: 3,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 2 },
        { design: 'JOKER', value: 0 },
      ],
    },
    { id: 1, isHuman: false, isFinished: false, cardCount: 4, cards: [] },
    { id: 2, isHuman: false, isFinished: false, cardCount: 2, cards: [] },
  ],
  currentTurn: 0,
  nextDrawTargetIdx: 1,
  gameEndFlag: false,
  hasDrawn: false,
  lastDrawPlayerIdx: 0,
  lastDrawFromIdx: 0,
  lastDrawCard: null,
  lastDiscardedPairs: 0,
  cpuActions: [],
  humanAction: null,
  cpuHighlightedCardIdx: -1,
  removedCard: null,
  mode: 0,
  message: '',
};

const cpuTurnState: OldMaidResponse = {
  ...humanTurnState,
  currentTurn: 1,
  nextDrawTargetIdx: 0,
  hasDrawn: true,
  lastDrawPlayerIdx: 1,
  lastDrawFromIdx: 2,
  lastDrawCard: { design: 'HEART', value: 5 },
  lastDiscardedPairs: 1,
};

const gameEndState: OldMaidResponse = {
  ...humanTurnState,
  gameEndFlag: true,
  message: 'あなたが負けました',
};

beforeEach(() => {
  mockExec.mockResolvedValue(humanTurnState);
});

async function startGame() {
  renderWithProviders(<OldMaidPage />);
  fireEvent.click(screen.getByRole('button', { name: 'ゲーム開始' }));
  await waitFor(() => expect(screen.getByText('あなた')).toBeInTheDocument());
}

describe('OldMaidPage', () => {
  it('shows setup screen on initial render without calling exec', () => {
    renderWithProviders(<OldMaidPage />);
    expect(screen.getByText('Old Maid 設定')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('renders null game area while API call is pending after setup', async () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<OldMaidPage />);
    fireEvent.click(screen.getByRole('button', { name: 'ゲーム開始' }));
    expect(screen.queryByText('あなた')).not.toBeInTheDocument();
  });

  it('mode 0 radio is selected by default', () => {
    renderWithProviders(<OldMaidPage />);
    const radios = screen.getAllByRole('radio');
    expect(radios[0]).toBeChecked();
    expect(radios[1]).not.toBeChecked();
  });

  it('can select mode 1', () => {
    renderWithProviders(<OldMaidPage />);
    const radios = screen.getAllByRole('radio');
    fireEvent.click(radios[1]);
    expect(radios[1]).toBeChecked();
  });

  it('CPU心理戦 checkbox is unchecked by default', () => {
    renderWithProviders(<OldMaidPage />);
    expect(screen.getByRole('checkbox')).not.toBeChecked();
  });

  it('CPU心理戦 checkbox can be toggled', () => {
    renderWithProviders(<OldMaidPage />);
    const checkbox = screen.getByRole('checkbox');
    fireEvent.click(checkbox);
    expect(checkbox).toBeChecked();
  });

  it('ゲーム開始 calls reset with mode=1 and cpuPlacementStrategy=true after changes', async () => {
    renderWithProviders(<OldMaidPage />);
    const radios = screen.getAllByRole('radio');
    fireEvent.click(radios[1]);
    fireEvent.click(screen.getByRole('checkbox'));
    fireEvent.click(screen.getByRole('button', { name: 'ゲーム開始' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, 1, true));
  });

  it('hides setup screen after ゲーム開始', async () => {
    await startGame();
    expect(screen.queryByText('Old Maid 設定')).not.toBeInTheDocument();
  });

  it('設定 button returns to setup screen', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    expect(screen.getByText('Old Maid 設定')).toBeInTheDocument();
  });

  it('shows ジジ抜き mode badge when mode is 1', async () => {
    const jijiState: OldMaidResponse = { ...humanTurnState, mode: 1 };
    mockExec.mockResolvedValue(jijiState);
    await startGame();
    expect(screen.getByText('ジジ抜き')).toBeInTheDocument();
  });

  it('highlighted card has translateY style', async () => {
    const highlightedState: OldMaidResponse = {
      ...humanTurnState,
      cpuHighlightedCardIdx: 0,
      nextDrawTargetIdx: 1,
      currentTurn: 0,
    };
    mockExec.mockResolvedValue(highlightedState);
    await startGame();
    const btn = screen.getByRole('button', { name: 'カード 1 枚目を引く' });
    const img = btn.querySelector('img') as HTMLImageElement;
    expect(img).toHaveStyle({ transform: 'translateY(-8px)' });
  });

  it('non-highlighted cards do not have translateY style', async () => {
    await startGame();
    const cardBacks = screen.getAllByAltText('カード裏面');
    for (const cardBack of cardBacks) {
      expect(cardBack).not.toHaveStyle({ transform: 'translateY(-8px)' });
    }
  });

  it('shows removed card info at game end in JijiNuki mode', async () => {
    const jijiEndState: OldMaidResponse = {
      ...humanTurnState,
      gameEndFlag: true,
      removedCard: { design: 'SPADE', value: 3 },
      mode: 1,
      message: 'ゲーム終了',
    };
    mockExec.mockResolvedValue(jijiEndState);
    await startGame();
    expect(screen.getByText('除外カード: SPADE 3')).toBeInTheDocument();
  });

  it('does not show removed card when gameEndFlag is false', async () => {
    const notEndedState: OldMaidResponse = {
      ...humanTurnState,
      gameEndFlag: false,
      removedCard: { design: 'SPADE', value: 3 },
    };
    mockExec.mockResolvedValue(notEndedState);
    await startGame();
    expect(screen.queryByText(/除外カード:/)).not.toBeInTheDocument();
  });

  it('renders human player area labeled あなた', async () => {
    await startGame();
    expect(screen.getByText('あなた')).toBeInTheDocument();
  });

  it('renders CPU player areas with correct labels', async () => {
    await startGame();
    expect(screen.getByText('CPU 1')).toBeInTheDocument();
    expect(screen.getByText('CPU 2')).toBeInTheDocument();
  });

  it('shows human player cards', async () => {
    await startGame();
    expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    expect(screen.getByAltText('♥ 2')).toBeInTheDocument();
    expect(screen.getByAltText('ジョーカー')).toBeInTheDocument();
  });

  it('shows card counts for CPU players', async () => {
    await startGame();
    expect(screen.getByText('4枚')).toBeInTheDocument();
    expect(screen.getByText('2枚')).toBeInTheDocument();
  });

  it('shows target player badge (← 引く相手) on human turn', async () => {
    await startGame();
    expect(screen.getByText('← 引く相手')).toBeInTheDocument();
  });

  it('shows instruction to click target player on human turn', async () => {
    await startGame();
    expect(screen.getByText(/あなたの番！/)).toBeInTheDocument();
  });

  it('random draw button is enabled on human turn', async () => {
    await startGame();
    expect(screen.getByText('ランダムに引く')).not.toBeDisabled();
  });

  it('random draw button is disabled on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    await startGame();
    expect(screen.getByText('ランダムに引く')).toBeDisabled();
  });

  it('random draw button is disabled when game has ended', async () => {
    mockExec.mockResolvedValue(gameEndState);
    await startGame();
    expect(screen.getByText('ランダムに引く')).toBeDisabled();
  });

  it('calls reset when reset button is clicked', async () => {
    await startGame();
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByText('リセット'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, 0, false));
  });

  it('calls draw when random draw button is clicked', async () => {
    await startGame();
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent.click(screen.getByText('ランダムに引く'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('shows status message when hasDrawn is true', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    await startGame();
    expect(screen.getByText(/CPU 1がCPU 2から1枚引きました/)).toBeInTheDocument();
  });

  it('shows discarded pairs in status message', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    await startGame();
    expect(screen.getByText(/1組捨てました/)).toBeInTheDocument();
  });

  it('shows finished badge for completed players', async () => {
    const stateWithFinished: OldMaidResponse = {
      ...humanTurnState,
      players: [
        { ...humanTurnState.players[0] },
        { id: 1, isHuman: false, isFinished: true, cardCount: 0, cards: [] },
        { ...humanTurnState.players[2] },
      ],
    };
    mockExec.mockResolvedValue(stateWithFinished);
    await startGame();
    expect(screen.getByText('上がり')).toBeInTheDocument();
  });

  it('shows game result message when game ends', async () => {
    mockExec.mockResolvedValue(gameEndState);
    await startGame();
    expect(screen.getByText('あなたが負けました')).toBeInTheDocument();
  });

  it('shows CPU actions log when cpuActions is non-empty', async () => {
    const stateWithCpuActions: OldMaidResponse = {
      ...humanTurnState,
      cpuActions: [{ drawPlayerIdx: 1, drawFromIdx: 2, drawnCard: null, discardedPairs: 0 }],
    };
    mockExec.mockResolvedValue(stateWithCpuActions);
    await startGame();
    await waitFor(() => {
      expect(screen.getByText(/\[CPUの行動\]/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 1がCPU 2から1枚引きました/)).toBeInTheDocument();
    });
  });

  it('does not show drawn card in CPU actions log', async () => {
    const stateWithCpuActions: OldMaidResponse = {
      ...humanTurnState,
      cpuActions: [{ drawPlayerIdx: 1, drawFromIdx: 2, drawnCard: null, discardedPairs: 0 }],
    };
    mockExec.mockResolvedValue(stateWithCpuActions);
    await startGame();
    expect(screen.getByText(/\[CPUの行動\]/)).toBeInTheDocument();
    expect(screen.queryByText(/SPADE 3/)).not.toBeInTheDocument();
  });

  it('renders aria-label on clickable card backs for screen reader accessibility', async () => {
    await startGame();
    // CPU 1 is target with 4 cards; only first 4 shown (capped at 10)
    const drawButtons = screen.getAllByRole('button', { name: /枚目を引く/ });
    expect(drawButtons[0]).toHaveAttribute('aria-label', 'カード 1 枚目を引く');
    expect(drawButtons[1]).toHaveAttribute('aria-label', 'カード 2 枚目を引く');
    expect(drawButtons[2]).toHaveAttribute('aria-label', 'カード 3 枚目を引く');
    expect(drawButtons[3]).toHaveAttribute('aria-label', 'カード 4 枚目を引く');
  });

  it('calls draw with drawIdx when a target player card back is clicked', async () => {
    await startGame();
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    // CPU 1 is target player: click its first card back
    const btn = screen.getByRole('button', { name: 'カード 1 枚目を引く' });
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw', 0));
  });

  it('shows stacked discarded cards when lastDiscardedCards has a pair', async () => {
    const stateWithDiscarded: OldMaidResponse = {
      ...humanTurnState,
      hasDrawn: true,
      lastDrawPlayerIdx: 0,
      lastDrawFromIdx: 1,
      lastDrawCard: { design: 'SPADE', value: 5 },
      lastDiscardedPairs: 1,
      lastDiscardedCards: [
        { design: 'SPADE', value: 5 },
        { design: 'CLOVER', value: 5 },
      ],
    };
    mockExec.mockResolvedValue(stateWithDiscarded);
    await startGame();
    expect(screen.getByAltText('♠ 5')).toBeInTheDocument();
    expect(screen.getByAltText('♣ 5')).toBeInTheDocument();
  });

  it('shows human draw status before CPU replay and eventually shows CPU log', async () => {
    const stateWithCpuActions: OldMaidResponse = {
      ...humanTurnState,
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          cardCount: 4,
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 2 },
            { design: 'JOKER', value: 0 },
            { design: 'DIAMOND', value: 7 },
          ],
        },
        { id: 1, isHuman: false, isFinished: false, cardCount: 3, cards: [] },
        { id: 2, isHuman: false, isFinished: false, cardCount: 1, cards: [] },
      ],
      currentTurn: 0,
      nextDrawTargetIdx: 1,
      hasDrawn: true,
      lastDrawPlayerIdx: 1,
      lastDrawFromIdx: 2,
      lastDrawCard: null,
      lastDiscardedPairs: 0,
      cpuActions: [{ drawPlayerIdx: 1, drawFromIdx: 2, drawnCard: null, discardedPairs: 0, discardedCards: [] }],
      humanAction: {
        drawPlayerIdx: 0,
        drawFromIdx: 1,
        drawnCard: { design: 'HEART', value: 3 },
        discardedPairs: 0,
        discardedCards: [],
      },
    };

    // reset returns humanTurnState, draw returns stateWithCpuActions
    mockExec.mockResolvedValueOnce(humanTurnState).mockResolvedValueOnce(stateWithCpuActions);
    renderWithProviders(<OldMaidPage />);
    fireEvent.click(screen.getByRole('button', { name: 'ゲーム開始' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'ランダムに引く' })).not.toBeDisabled());

    // Trigger draw
    fireEvent.click(screen.getByText('ランダムに引く'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));

    // After all replay delays, final state CPU log should be visible
    // Each delay is 800ms; humanAction + 1 cpuAction = 2 delays = 1600ms max
    await waitFor(() => expect(screen.getByText(/\[CPUの行動\]/)).toBeInTheDocument(), {
      timeout: 4000,
    });
  }, 10000);

  it('disables buttons while loading', async () => {
    await startGame();
    let resolve!: (value: OldMaidResponse) => void;
    const slowPromise = new Promise<OldMaidResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ランダムに引く' }));

    expect(screen.getByRole('button', { name: 'ランダムに引く' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(humanTurnState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ランダムに引く' })).not.toBeDisabled());
  });

  it('enables random draw button after CPU replay animation completes (Bug 1 regression)', async () => {
    // Final server state has currentTurn=0 (human's turn)
    const finalState: OldMaidResponse = {
      ...humanTurnState,
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          cardCount: 4,
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 2 },
            { design: 'JOKER', value: 0 },
            { design: 'DIAMOND', value: 7 },
          ],
        },
        { id: 1, isHuman: false, isFinished: false, cardCount: 3, cards: [] },
        { id: 2, isHuman: false, isFinished: false, cardCount: 1, cards: [] },
      ],
      currentTurn: 0,
      nextDrawTargetIdx: 1,
      hasDrawn: true,
      lastDrawPlayerIdx: 1,
      lastDrawFromIdx: 2,
      lastDrawCard: null,
      lastDiscardedPairs: 0,
      cpuActions: [{ drawPlayerIdx: 1, drawFromIdx: 2, drawnCard: null, discardedPairs: 0, discardedCards: [] }],
      humanAction: {
        drawPlayerIdx: 0,
        drawFromIdx: 1,
        drawnCard: { design: 'HEART', value: 3 },
        discardedPairs: 0,
        discardedCards: [],
      },
    };

    mockExec.mockResolvedValueOnce(humanTurnState).mockResolvedValueOnce(finalState);
    renderWithProviders(<OldMaidPage />);
    fireEvent.click(screen.getByRole('button', { name: 'ゲーム開始' }));
    await waitFor(() => expect(screen.getByText('ランダムに引く')).not.toBeDisabled());

    fireEvent.click(screen.getByText('ランダムに引く'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));

    // After all replay delays, the final state (currentTurn=0) must be applied
    // so the button is re-enabled for the human's next turn
    await waitFor(() => expect(screen.getByText('ランダムに引く')).not.toBeDisabled(), {
      timeout: 4000,
    });
  }, 10000);

  it('shows error message when API call fails', async () => {
    await startGame();
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  }, 10000);

  it('clears error message on successful API call after failure', async () => {
    await startGame();
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());

    mockExec.mockReset();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() => expect(screen.queryByText(NETWORK_ERROR_MESSAGE())).not.toBeInTheDocument());
  }, 10000);

  it('shows JOKER label in status when lastDrawCard is JOKER', async () => {
    const jokerDrawState: OldMaidResponse = {
      ...cpuTurnState,
      lastDrawCard: { design: 'JOKER', value: 0 },
    };
    mockExec.mockResolvedValue(jokerDrawState);
    await startGame();
    expect(screen.getByText(/CPU 1がCPU 2から1枚引きました \(JOKER\)/)).toBeInTheDocument();
  });

  it('shows status without card info when hasDrawn is true but lastDrawCard is null', async () => {
    const noCardDrawState: OldMaidResponse = {
      ...cpuTurnState,
      lastDrawCard: null,
      lastDiscardedPairs: 0,
    };
    mockExec.mockResolvedValue(noCardDrawState);
    await startGame();
    expect(screen.getByText(/CPU 1がCPU 2から1枚引きました$/)).toBeInTheDocument();
    expect(screen.queryByText(/組捨てました/)).not.toBeInTheDocument();
  });

  it('shows overflow indicator when CPU player has more than 10 cards', async () => {
    const overflowState: OldMaidResponse = {
      ...humanTurnState,
      nextDrawTargetIdx: 1,
      players: [
        { ...humanTurnState.players[0] },
        { id: 1, isHuman: false, isFinished: false, cardCount: 13, cards: [] }, // target player: selectable path
        { id: 2, isHuman: false, isFinished: false, cardCount: 11, cards: [] }, // non-target: non-selectable path
      ],
    };
    mockExec.mockResolvedValue(overflowState);
    await startGame();
    expect(screen.getByText('+3')).toBeInTheDocument(); // 13-10=3
    expect(screen.getByText('+1')).toBeInTheDocument(); // 11-10=1
  });

  it('shows single odd discarded card without pairing', async () => {
    const stateWithOddDiscards: OldMaidResponse = {
      ...humanTurnState,
      lastDiscardedCards: [{ design: 'SPADE', value: 3 }],
    };
    mockExec.mockResolvedValue(stateWithOddDiscards);
    await startGame();
    expect(screen.getByAltText('♠ 3')).toBeInTheDocument();
  });

  it('shows discarded pair count in CPU log when discardedPairs > 0', async () => {
    const stateWithDiscardedPairs: OldMaidResponse = {
      ...humanTurnState,
      cpuActions: [{ drawPlayerIdx: 1, drawFromIdx: 2, drawnCard: null, discardedPairs: 2, discardedCards: [] }],
    };
    mockExec.mockResolvedValue(stateWithDiscardedPairs);
    await startGame();
    await waitFor(() => {
      expect(screen.getByText(/CPU 1がCPU 2から1枚引きました。2組捨てました/)).toBeInTheDocument();
    });
  });

  it('replays two CPU actions covering non-last-action branches', async () => {
    const twoCpuActionsState: OldMaidResponse = {
      ...humanTurnState,
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          cardCount: 3,
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 2 },
            { design: 'JOKER', value: 0 },
          ],
        },
        { id: 1, isHuman: false, isFinished: false, cardCount: 3, cards: [] },
        { id: 2, isHuman: false, isFinished: false, cardCount: 1, cards: [] },
      ],
      humanAction: null,
      cpuActions: [
        { drawPlayerIdx: 1, drawFromIdx: 2, drawnCard: null, discardedPairs: 0, discardedCards: [] },
        { drawPlayerIdx: 2, drawFromIdx: 0, drawnCard: null, discardedPairs: 0, discardedCards: [] },
      ],
      currentTurn: 0,
      hasDrawn: true,
      lastDrawPlayerIdx: 1,
      lastDrawFromIdx: 2,
      lastDrawCard: null,
      lastDiscardedPairs: 0,
    };

    mockExec.mockResolvedValueOnce(humanTurnState).mockResolvedValueOnce(twoCpuActionsState);
    renderWithProviders(<OldMaidPage />);
    fireEvent.click(screen.getByRole('button', { name: 'ゲーム開始' }));
    await waitFor(() => expect(screen.getByText('ランダムに引く')).not.toBeDisabled());

    fireEvent.click(screen.getByText('ランダムに引く'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));

    // After CPU replay with 2 actions, the CPU action log should appear
    await waitFor(() => expect(screen.getByText(/\[CPUの行動\]/)).toBeInTheDocument(), {
      timeout: 6000,
    });
  }, 15000);

  it('shows replay state when humanAction.discardedCards is undefined', async () => {
    const stateWithUndefinedDiscards: OldMaidResponse = {
      ...humanTurnState,
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          cardCount: 3,
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 2 },
            { design: 'JOKER', value: 0 },
          ],
        },
        { id: 1, isHuman: false, isFinished: false, cardCount: 3, cards: [] },
        { id: 2, isHuman: false, isFinished: false, cardCount: 1, cards: [] },
      ],
      currentTurn: 0,
      nextDrawTargetIdx: 1,
      hasDrawn: true,
      lastDrawPlayerIdx: 1,
      lastDrawFromIdx: 2,
      lastDrawCard: null,
      lastDiscardedPairs: 0,
      cpuActions: [{ drawPlayerIdx: 1, drawFromIdx: 2, drawnCard: null, discardedPairs: 0 }],
      humanAction: {
        drawPlayerIdx: 0,
        drawFromIdx: 1,
        drawnCard: { design: 'HEART', value: 3 },
        discardedPairs: 0,
      },
    };

    mockExec.mockResolvedValueOnce(humanTurnState).mockResolvedValueOnce(stateWithUndefinedDiscards);
    renderWithProviders(<OldMaidPage />);
    fireEvent.click(screen.getByRole('button', { name: 'ゲーム開始' }));
    await waitFor(() => expect(screen.getByText('ランダムに引く')).not.toBeDisabled());

    fireEvent.click(screen.getByText('ランダムに引く'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));

    await waitFor(() => expect(screen.getByText(/\[CPUの行動\]/)).toBeInTheDocument(), {
      timeout: 4000,
    });
  }, 10000);

  it('handles humanAction with empty cpuActions without crashing', async () => {
    const humanActionNoCpuState: OldMaidResponse = {
      ...humanTurnState,
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          cardCount: 2,
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'JOKER', value: 0 },
          ],
        },
        { id: 1, isHuman: false, isFinished: true, cardCount: 0, cards: [] },
        { id: 2, isHuman: false, isFinished: true, cardCount: 0, cards: [] },
      ],
      currentTurn: 0,
      nextDrawTargetIdx: 0,
      hasDrawn: true,
      lastDrawPlayerIdx: 0,
      lastDrawFromIdx: 1,
      lastDrawCard: { design: 'HEART', value: 3 },
      lastDiscardedPairs: 1,
      cpuActions: [],
      humanAction: {
        drawPlayerIdx: 0,
        drawFromIdx: 1,
        drawnCard: { design: 'HEART', value: 3 },
        discardedPairs: 1,
        discardedCards: [
          { design: 'HEART', value: 3 },
          { design: 'DIAMOND', value: 3 },
        ],
      },
      gameEndFlag: true,
      message: 'あなたが負けました',
    };

    mockExec.mockResolvedValueOnce(humanTurnState).mockResolvedValueOnce(humanActionNoCpuState);
    renderWithProviders(<OldMaidPage />);
    fireEvent.click(screen.getByRole('button', { name: 'ゲーム開始' }));
    await waitFor(() => expect(screen.getByText('ランダムに引く')).not.toBeDisabled());

    fireEvent.click(screen.getByText('ランダムに引く'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));

    // Should show the game end message without crashing
    await waitFor(() => expect(screen.getByText(humanActionNoCpuState.message)).toBeInTheDocument(), {
      timeout: 4000,
    });
  }, 10000);

  it('setup screen has aria-busy attribute', () => {
    renderWithProviders(<OldMaidPage />);
    const setupContainer = screen.getByText('Old Maid 設定').closest('[aria-busy]') as HTMLElement;
    expect(setupContainer).toHaveAttribute('aria-busy', 'false');
  });

  it('sets aria-busy and sr-only loading text on game screen while loading', async () => {
    await startGame();

    const container = screen.getByRole('button', { name: 'リセット' }).closest('[aria-live]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');
    expect(screen.queryByText('処理中...')).not.toBeInTheDocument();

    let resolve!: (value: OldMaidResponse) => void;
    const slowPromise = new Promise<OldMaidResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ランダムに引く' }));

    expect(container).toHaveAttribute('aria-busy', 'true');
    expect(screen.getByText('処理中...')).toBeInTheDocument();

    resolve(humanTurnState);
    await waitFor(() => {
      expect(container).toHaveAttribute('aria-busy', 'false');
      expect(screen.queryByText('処理中...')).not.toBeInTheDocument();
    });
  });

  it('shuffle button is rendered and enabled when game is active', async () => {
    await startGame();
    const btn = screen.getByRole('button', { name: 'シャッフル' });
    expect(btn).not.toBeDisabled();
  });

  it('shuffle button is disabled when game has ended', async () => {
    mockExec.mockResolvedValue(gameEndState);
    await startGame();
    const btn = screen.getByRole('button', { name: 'シャッフル' });
    expect(btn).toBeDisabled();
  });

  it('shuffle button is disabled while loading', async () => {
    await startGame();
    let resolve!: (value: OldMaidResponse) => void;
    const slowPromise = new Promise<OldMaidResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'シャッフル' }));
    expect(screen.getByRole('button', { name: 'シャッフル' })).toBeDisabled();
    resolve(humanTurnState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'シャッフル' })).not.toBeDisabled());
  });

  it('shuffle button calls shuffle API and updates state', async () => {
    await startGame();
    mockExec.mockClear();
    const shuffledState: OldMaidResponse = {
      ...humanTurnState,
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          cardCount: 3,
          cards: [
            { design: 'JOKER', value: 0 },
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 2 },
          ],
        },
        { id: 1, isHuman: false, isFinished: false, cardCount: 4, cards: [] },
        { id: 2, isHuman: false, isFinished: false, cardCount: 2, cards: [] },
      ],
    };
    mockExec.mockResolvedValue(shuffledState);
    fireEvent.click(screen.getByRole('button', { name: 'シャッフル' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('shuffle'));
  });

  it('shuffle shows error on API failure', async () => {
    await startGame();
    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'シャッフル' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('human cards are draggable when game is active', async () => {
    await startGame();
    const humanCards = screen.getAllByAltText(/♠ A|♥ 2|ジョーカー/);
    for (const card of humanCards) {
      expect(card).toHaveAttribute('draggable', 'true');
    }
  });

  it('human cards are not draggable when game has ended', async () => {
    mockExec.mockResolvedValue(gameEndState);
    await startGame();
    const humanCards = screen.getAllByAltText(/♠ A|♥ 2|ジョーカー/);
    for (const card of humanCards) {
      expect(card).not.toHaveAttribute('draggable', 'true');
    }
  });

  it('drag and drop reorders cards via API', async () => {
    await startGame();
    mockExec.mockClear();
    const reorderedState: OldMaidResponse = {
      ...humanTurnState,
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          cardCount: 3,
          cards: [
            { design: 'HEART', value: 2 },
            { design: 'SPADE', value: 1 },
            { design: 'JOKER', value: 0 },
          ],
        },
        { id: 1, isHuman: false, isFinished: false, cardCount: 4, cards: [] },
        { id: 2, isHuman: false, isFinished: false, cardCount: 2, cards: [] },
      ],
    };
    mockExec.mockResolvedValue(reorderedState);

    const cards = screen.getAllByAltText(/♠ A|♥ 2|ジョーカー/);
    // Drag card at index 0 to index 1
    fireEvent.dragStart(cards[0], { dataTransfer: { setData: vi.fn() } });
    fireEvent.dragOver(cards[1]);
    fireEvent.drop(cards[1], {
      dataTransfer: { getData: () => '0' },
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reorder', undefined, undefined, undefined, [1, 0, 2]));
  });

  it('drag and drop same position is no-op', async () => {
    await startGame();
    mockExec.mockClear();

    const cards = screen.getAllByAltText(/♠ A|♥ 2|ジョーカー/);
    // Drag card at index 0 to same index 0
    fireEvent.dragStart(cards[0], { dataTransfer: { setData: vi.fn() } });
    fireEvent.dragOver(cards[0]);
    fireEvent.drop(cards[0], {
      dataTransfer: { getData: () => '0' },
    });

    // Should NOT call API since from === to
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('drag and drop with empty data is no-op', async () => {
    await startGame();
    mockExec.mockClear();

    const cards = screen.getAllByAltText(/♠ A|♥ 2|ジョーカー/);
    fireEvent.drop(cards[1], {
      dataTransfer: { getData: () => '' },
    });

    expect(mockExec).not.toHaveBeenCalled();
  });

  it('reorder shows error on API failure', async () => {
    await startGame();
    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));

    const cards = screen.getAllByAltText(/♠ A|♥ 2|ジョーカー/);
    fireEvent.drop(cards[1], {
      dataTransfer: { getData: () => '0' },
    });

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  // ── Card reveal suspense & Joker shake ─────────────────────────────────

  it('shows card-back during reveal delay and revealed card after timer', async () => {
    const drawState: OldMaidResponse = {
      ...humanTurnState,
      hasDrawn: true,
      lastDrawPlayerIdx: 0,
      lastDrawFromIdx: 1,
      lastDrawCard: { design: 'HEART', value: 5 },
      lastDiscardedPairs: 0,
    };
    mockExec.mockResolvedValue(drawState);
    await startGame();

    // Card-back is shown during the 600ms reveal delay
    const revealArea = screen.getByTestId('card-reveal-area');
    expect(revealArea.querySelector('img[alt="カード裏面"]')).toBeInTheDocument();

    // After 600ms, the actual card is revealed with flip animation
    await waitFor(() => {
      expect(screen.getByTestId('card-reveal-area').querySelector('img[alt="♥ 5"]')).toBeInTheDocument();
    });
  });

  it('applies animate-shake class when Joker is drawn', async () => {
    const jokerDrawState: OldMaidResponse = {
      ...humanTurnState,
      hasDrawn: true,
      lastDrawPlayerIdx: 0,
      lastDrawFromIdx: 1,
      lastDrawCard: { design: 'JOKER', value: 0 },
      lastDiscardedPairs: 0,
    };
    mockExec.mockResolvedValue(jokerDrawState);
    await startGame();

    // After the 600ms reveal timer, shake is triggered
    await waitFor(() => {
      const container = screen.getByRole('button', { name: 'リセット' }).closest('[aria-live]') as HTMLElement;
      expect(container.className).toContain('animate-shake');
    });
  });

  it('does not apply animate-shake for non-Joker cards', async () => {
    const normalDrawState: OldMaidResponse = {
      ...humanTurnState,
      hasDrawn: true,
      lastDrawPlayerIdx: 0,
      lastDrawFromIdx: 1,
      lastDrawCard: { design: 'HEART', value: 5 },
      lastDiscardedPairs: 0,
    };
    mockExec.mockResolvedValue(normalDrawState);
    await startGame();

    // Wait for reveal to complete (600ms)
    await waitFor(() => {
      expect(screen.getByTestId('card-reveal-area').querySelector('img[alt="♥ 5"]')).toBeInTheDocument();
    });
    const container = screen.getByRole('button', { name: 'リセット' }).closest('[aria-live]') as HTMLElement;
    expect(container.className).not.toContain('animate-shake');
  });

  it('does not show card reveal area when lastDrawCard is null', async () => {
    await startGame();
    expect(screen.queryByTestId('card-reveal-area')).not.toBeInTheDocument();
  });

  it('does not show card reveal area when game has ended', async () => {
    const endWithDraw: OldMaidResponse = {
      ...gameEndState,
      hasDrawn: true,
      lastDrawCard: { design: 'HEART', value: 5 },
    };
    mockExec.mockResolvedValue(endWithDraw);
    await startGame();
    expect(screen.queryByTestId('card-reveal-area')).not.toBeInTheDocument();
  });

  it('goes directly to CPU replay when humanAction is null', async () => {
    const directCpuState: OldMaidResponse = {
      ...humanTurnState,
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          cardCount: 3,
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 2 },
            { design: 'JOKER', value: 0 },
          ],
        },
        { id: 1, isHuman: false, isFinished: false, cardCount: 3, cards: [] },
        { id: 2, isHuman: false, isFinished: false, cardCount: 1, cards: [] },
      ],
      humanAction: null,
      cpuActions: [{ drawPlayerIdx: 1, drawFromIdx: 2, drawnCard: null, discardedPairs: 0, discardedCards: [] }],
      currentTurn: 0,
      hasDrawn: true,
      lastDrawPlayerIdx: 1,
      lastDrawFromIdx: 2,
      lastDrawCard: null,
      lastDiscardedPairs: 0,
    };

    mockExec.mockResolvedValueOnce(humanTurnState).mockResolvedValueOnce(directCpuState);
    renderWithProviders(<OldMaidPage />);
    fireEvent.click(screen.getByRole('button', { name: 'ゲーム開始' }));
    await waitFor(() => expect(screen.getByText('ランダムに引く')).not.toBeDisabled());

    fireEvent.click(screen.getByText('ランダムに引く'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));

    // After CPU replay, the CPU action log should appear
    await waitFor(() => expect(screen.getByText(/\[CPUの行動\]/)).toBeInTheDocument(), {
      timeout: 4000,
    });
  }, 10000);
});
