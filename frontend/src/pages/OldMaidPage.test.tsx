import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, oldmaidApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { OldMaidResponse } from '../types/card';
import { OldMaidPage } from './OldMaidPage';

vi.mock('../api/gameApi', () => ({
  oldmaidApi: { exec: vi.fn() },
  actionLogApi: { oldmaid: vi.fn() },
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
  drawHistory: [],
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
  // Game auto-starts with default settings on mount. Multiple elements may
  // contain "あなた" (player area label + timeline chips), so wait until at
  // least one renders rather than asserting uniqueness here.
  await waitFor(() => expect(screen.getAllByText('あなた').length).toBeGreaterThan(0));
}

describe('OldMaidPage', () => {
  it('auto-starts game on mount and calls reset with defaults', async () => {
    renderWithProviders(<OldMaidPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, 0, false, undefined, false, false, false),
    );
  });

  it('shows keyboard shortcut hints and lowercase aria-keyshortcuts on the human turn', async () => {
    await startGame();
    expect(screen.getByTestId('oldmaid-key-hints')).toHaveTextContent('D: 引く / S: シャッフル');
    // Lowercase: an unmodified key (uppercase would imply Shift+D per WAI-ARIA).
    expect(screen.getByRole('button', { name: 'ランダムに引く' })).toHaveAttribute('aria-keyshortcuts', 'd');
    expect(screen.getByRole('button', { name: 'シャッフル' })).toHaveAttribute('aria-keyshortcuts', 's');
  });

  it('hides the key hints and drops aria-keyshortcuts on a CPU turn (bindings inactive)', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<OldMaidPage />);
    await waitFor(() => expect(screen.getAllByText('あなた').length).toBeGreaterThan(0));
    expect(screen.queryByTestId('oldmaid-key-hints')).not.toBeInTheDocument();
    // The shuffle button stays clickable on a CPU turn, but the `s` key binding is
    // inactive, so it must not advertise a shortcut that does nothing.
    expect(screen.getByRole('button', { name: 'シャッフル' })).not.toHaveAttribute('aria-keyshortcuts');
  });

  it('renders skeleton while API call is pending on mount', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<OldMaidPage />);
    expect(screen.queryByText('あなた')).not.toBeInTheDocument();
  });

  it('設定 button opens settings dialog', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    expect(screen.getByText('Old Maid 設定')).toBeInTheDocument();
  });

  it('settings dialog has mode 0 radio selected by default', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    const radios = screen.getAllByRole('radio');
    expect(radios[0]).toBeChecked();
    expect(radios[1]).not.toBeChecked();
  });

  it('settings dialog can select mode 1', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    const radios = screen.getAllByRole('radio');
    fireEvent.click(radios[1]);
    expect(radios[1]).toBeChecked();
  });

  it('settings dialog CPU心理戦 checkbox is unchecked by default', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    expect(screen.getByLabelText('CPU心理戦（奇数カードを端に配置）')).not.toBeChecked();
  });

  it('settings dialog CPU心理戦 checkbox can be toggled', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    const checkbox = screen.getByLabelText('CPU心理戦（奇数カードを端に配置）');
    fireEvent.click(checkbox);
    expect(checkbox).toBeChecked();
  });

  it('settings dialog CPU記憶AI checkbox is unchecked by default', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    expect(screen.getByLabelText('CPU記憶AI（引いた位置を記憶して戦略的に選択）')).not.toBeChecked();
  });

  it('settings dialog CPU記憶AI checkbox can be toggled', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    const checkbox = screen.getByLabelText('CPU記憶AI（引いた位置を記憶して戦略的に選択）');
    fireEvent.click(checkbox);
    expect(checkbox).toBeChecked();
  });

  it('apply calls reset with mode=1 and cpuPlacementStrategy=true after changes', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    const radios = screen.getAllByRole('radio');
    fireEvent.click(radios[1]);
    fireEvent.click(screen.getByLabelText('CPU心理戦（奇数カードを端に配置）'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: '適用' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, 1, true, undefined, false, false, false),
    );
  });

  it('apply calls reset with cpuMemoryAI=true when checkbox enabled', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    fireEvent.click(screen.getByLabelText('CPU記憶AI（引いた位置を記憶して戦略的に選択）'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: '適用' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, 0, false, undefined, true, false, false),
    );
  });

  it('apply calls reset with cpuHesitationEnabled=true when checkbox enabled', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    fireEvent.click(screen.getByLabelText('CPU迷い時間ディレイ（カード内容により反応速度が変化）'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: '適用' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, 0, false, undefined, false, true, false),
    );
  });

  it('apply calls reset with cpuMetaAI=true when checkbox enabled', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    fireEvent.click(screen.getByLabelText('メタAI（CPUがプレイスタイルを学習）'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: '適用' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, 0, false, undefined, false, false, true),
    );
  });

  it('settings dialog closes on cancel without resetting and reverts changes', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    // Change mode to 1 in the dialog
    const radios = screen.getAllByRole('radio');
    fireEvent.click(radios[1]);
    expect(radios[1]).toBeChecked();
    // Cancel
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByText('Old Maid 設定')).not.toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
    // Reopen dialog: mode should be reverted to 0 (original)
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    const radiosAfter = screen.getAllByRole('radio');
    expect(radiosAfter[0]).toBeChecked();
    expect(radiosAfter[1]).not.toBeChecked();
  });

  it('settings dialog closes on apply', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    expect(screen.getByText('Old Maid 設定')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '適用' }));
    expect(screen.queryByText('Old Maid 設定')).not.toBeInTheDocument();
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
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, 0, false, undefined, false, false, false),
    );
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

    // auto-start returns humanTurnState, draw returns stateWithCpuActions
    mockExec.mockResolvedValueOnce(humanTurnState).mockResolvedValueOnce(stateWithCpuActions);
    renderWithProviders(<OldMaidPage />);
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
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  }, 10000);

  it('clears error message on successful API call after failure', async () => {
    await startGame();
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());

    mockExec.mockReset();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
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
    await waitFor(() => expect(screen.getByText('ランダムに引く')).not.toBeDisabled());

    fireEvent.click(screen.getByText('ランダムに引く'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));

    // Should show the game end message without crashing
    await waitFor(() => expect(screen.getByText(humanActionNoCpuState.message)).toBeInTheDocument(), {
      timeout: 4000,
    });
  }, 10000);

  it('settings dialog groups radio buttons in fieldset with legend', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    const legend = screen.getByText('モード選択');
    const fieldset = legend.closest('fieldset');
    expect(fieldset).toBeInTheDocument();
    const radios = within(fieldset as HTMLElement).getAllByRole('radio');
    expect(radios).toHaveLength(2);
  });

  it('settings dialog groups checkboxes in fieldset with legend', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    const legend = screen.getByText('CPU設定');
    const fieldset = legend.closest('fieldset');
    expect(fieldset).toBeInTheDocument();
    const checkboxes = within(fieldset as HTMLElement).getAllByRole('checkbox');
    expect(checkboxes).toHaveLength(4);
  });

  it('sets aria-busy on game screen while loading', async () => {
    await startGame();

    const container = screen.getByRole('button', { name: 'リセット' }).closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');

    let resolve!: (value: OldMaidResponse) => void;
    const slowPromise = new Promise<OldMaidResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ランダムに引く' }));

    expect(container).toHaveAttribute('aria-busy', 'true');

    resolve(humanTurnState);
    await waitFor(() => {
      expect(container).toHaveAttribute('aria-busy', 'false');
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
    // Reorder now wraps each card in a tap-target button with the draggable attribute
    const humanCards = screen.getAllByAltText(/♠ A|♥ 2|ジョーカー/);
    for (const card of humanCards) {
      const draggableAncestor = card.closest('[draggable]');
      expect(draggableAncestor).toHaveAttribute('draggable', 'true');
    }
  });

  it('human cards are not draggable when game has ended', async () => {
    mockExec.mockResolvedValue(gameEndState);
    await startGame();
    const humanCards = screen.getAllByAltText(/♠ A|♥ 2|ジョーカー/);
    for (const card of humanCards) {
      const draggableAncestor = card.closest('[draggable="true"]');
      expect(draggableAncestor).toBeNull();
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
      const container = screen.getByRole('button', { name: 'リセット' }).closest('[aria-busy]') as HTMLElement;
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
    const container = screen.getByRole('button', { name: 'リセット' }).closest('[aria-busy]') as HTMLElement;
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

  // ── Draw History Timeline ──────────────────────────────────────────────

  it('shows draw history timeline when drawHistory is non-empty', async () => {
    const stateWithHistory: OldMaidResponse = {
      ...humanTurnState,
      drawHistory: [
        { drawPlayerIdx: 0, drawFromIdx: 1, discardedPairs: 0, drawerFinished: false, targetFinished: true },
        { drawPlayerIdx: 2, drawFromIdx: 0, discardedPairs: 1, drawerFinished: true, targetFinished: false },
      ],
    };
    mockExec.mockResolvedValue(stateWithHistory);
    await startGame();
    expect(screen.getByTestId('draw-history-timeline')).toBeInTheDocument();
    expect(screen.getByText('引き履歴')).toBeInTheDocument();
    const rows = screen.getAllByTestId('draw-history-entry');
    expect(rows).toHaveLength(2);
    // First entry: no discard burst, target finished
    expect(rows[0].querySelector('[data-testid="discard-burst"]')).toBeNull();
    // Second entry: pair discarded
    expect(rows[1].querySelector('[data-testid="discard-burst"]')?.textContent).toBe('💥');
  });

  it('does not show draw history timeline when drawHistory is empty', async () => {
    await startGame();
    expect(screen.queryByTestId('draw-history-timeline')).not.toBeInTheDocument();
  });

  it('shows draw history timeline when drawHistory is undefined (backward compat)', async () => {
    const stateNoHistory: OldMaidResponse = {
      ...humanTurnState,
      drawHistory: undefined as unknown as OldMaidResponse['drawHistory'],
    };
    mockExec.mockResolvedValue(stateNoHistory);
    await startGame();
    expect(screen.queryByTestId('draw-history-timeline')).not.toBeInTheDocument();
  });

  // ── Suspect Pin ──────────────────────────────────────────────────────

  it('shows suspect pin button for CPU players', async () => {
    await startGame();
    const pinButtons = screen.getAllByRole('button', { name: /容疑者にマーク/ });
    expect(pinButtons.length).toBeGreaterThan(0);
  });

  it('toggles suspect pin on click', async () => {
    await startGame();
    const pinButton = screen.getAllByRole('button', { name: /容疑者にマーク/ })[0];
    fireEvent.click(pinButton);
    expect(screen.getByText('容疑者')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /マークを外す/ })).toBeInTheDocument();
  });

  it('unpins suspect on second click', async () => {
    await startGame();
    const pinButton = screen.getAllByRole('button', { name: /容疑者にマーク/ })[0];
    fireEvent.click(pinButton);
    expect(screen.getByText('容疑者')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /マークを外す/ }));
    expect(screen.queryByText('容疑者')).not.toBeInTheDocument();
  });

  it('clears suspect pins on reset', async () => {
    await startGame();
    const pinButton = screen.getAllByRole('button', { name: /容疑者にマーク/ })[0];
    fireEvent.click(pinButton);
    expect(screen.getByText('容疑者')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.queryByText('容疑者')).not.toBeInTheDocument());
  });

  it('clears suspect pins when applying settings', async () => {
    await startGame();
    const pinButton = screen.getAllByRole('button', { name: /容疑者にマーク/ })[0];
    fireEvent.click(pinButton);
    expect(screen.getByText('容疑者')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '設定' }));
    fireEvent.click(screen.getByRole('button', { name: '適用' }));
    await waitFor(() => expect(screen.queryByText('容疑者')).not.toBeInTheDocument());
  });

  it('does not show suspect pin button for finished CPU players', async () => {
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
    // Only CPU 2 (not finished) should have pin button
    const pinButtons = screen.getAllByRole('button', { name: /容疑者にマーク/ });
    expect(pinButtons).toHaveLength(1);
  });

  it('does not show suspect pin button when game has ended', async () => {
    mockExec.mockResolvedValue(gameEndState);
    await startGame();
    expect(screen.queryByRole('button', { name: /容疑者にマーク/ })).not.toBeInTheDocument();
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
    await waitFor(() => expect(screen.getByText('ランダムに引く')).not.toBeDisabled());

    fireEvent.click(screen.getByText('ランダムに引く'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));

    // After CPU replay, the CPU action log should appear
    await waitFor(() => expect(screen.getByText(/\[CPUの行動\]/)).toBeInTheDocument(), {
      timeout: 4000,
    });
  }, 10000);

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue({
      ...humanTurnState,
      gameEndFlag: true,
    });

    renderWithProviders(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.oldmaid).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.oldmaid).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();

    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });

  // --- ConfirmDialog tests ---

  it('shows confirm dialog when reset button is clicked', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    await startGame();
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, 0, false, undefined, false, false, false),
    );
  });

  // --- Keyboard navigation tests ---

  it('pressing d triggers draw random on human turn', async () => {
    await startGame();
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.keyDown(document, { key: 'd' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('pressing s triggers shuffle on human turn', async () => {
    await startGame();
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.keyDown(document, { key: 's' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('shuffle'));
  });

  it('keyboard shortcuts are disabled on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    await startGame();
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'd' });
    fireEvent.keyDown(document, { key: 's' });
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('keyboard shortcuts are disabled at game end', async () => {
    mockExec.mockResolvedValue(gameEndState);
    await startGame();
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'd' });
    fireEvent.keyDown(document, { key: 's' });
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('renders tutorial button', async () => {
    await startGame();
    expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument();
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    await startGame();
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    await startGame();
    expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();
  });

  // -- PhaseIndicator coverage --

  it('phase indicator shows your turn when human turn', async () => {
    await startGame();
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows waiting when cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    await startGame();
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('待機中'));
  });

  it('phase indicator shows end phase', async () => {
    mockExec.mockResolvedValue(gameEndState);
    await startGame();
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('終了'));
  });

  // -- Drawn-card landing highlight (issue #2984) --

  it('ring-highlights the position where the human just drew a card', async () => {
    // Human (id 0) drew HEART-2, which now sits at index 1 of their hand.
    mockExec.mockResolvedValue({
      ...humanTurnState,
      hasDrawn: true,
      lastDrawPlayerIdx: 0,
      lastDrawFromIdx: 1,
      lastDrawCard: { design: 'HEART', value: 2 },
    });
    await startGame();
    const container = await screen.findByTestId('human-card-container');
    const highlighted = within(container)
      .getAllByRole('button')
      .filter((b) => b.dataset.drawnHighlight === 'true');
    expect(highlighted).toHaveLength(1);
    // Second card (index 1) is the drawn HEART-2.
    expect(within(container).getAllByRole('button')[1]).toBe(highlighted[0]);
    // Animation is gated behind motion-safe so reduced-motion users get no pulse.
    expect(highlighted[0].className).toContain('ring-ds-accent');
    expect(highlighted[0].className).toContain('motion-safe:animate-pulse');
  });

  it('does not highlight when a CPU drew from another player', async () => {
    // CPU 1 drew from CPU 2 (no human involvement): nothing to highlight.
    mockExec.mockResolvedValue(cpuTurnState);
    await startGame();
    const container = await screen.findByTestId('human-card-container');
    const highlighted = within(container)
      .getAllByRole('button')
      .filter((b) => b.dataset.drawnHighlight === 'true');
    expect(highlighted).toHaveLength(0);
  });

  it('does not highlight when the human draw was discarded as a pair', async () => {
    // Human drew a card not present in their hand (it paired off and was discarded).
    mockExec.mockResolvedValue({
      ...humanTurnState,
      hasDrawn: true,
      lastDrawPlayerIdx: 0,
      lastDrawFromIdx: 1,
      lastDrawCard: { design: 'CLOVER', value: 9 },
      lastDiscardedPairs: 1,
    });
    await startGame();
    const container = await screen.findByTestId('human-card-container');
    const highlighted = within(container)
      .getAllByRole('button')
      .filter((b) => b.dataset.drawnHighlight === 'true');
    expect(highlighted).toHaveLength(0);
  });
});
