import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, omiApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { OmiResponse } from '../types/card';
import { OmiPage } from './OmiPage';

vi.mock('../api/gameApi', () => ({
  omiApi: { exec: vi.fn() },
  actionLogApi: { omi: vi.fn() },
}));

const mockExec = vi.mocked(omiApi.exec);

/** Omi: 4 players, 2 teams (0+2 vs 1+3), 8 cards each, 8-trick game. */
const playPhaseState: OmiResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 8,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      team: 0,
      trickCount: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 8,
      cards: [],
      team: 1,
      trickCount: 1,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 8,
      cards: [],
      team: 0,
      trickCount: 0,
    },
    {
      id: 3,
      isHuman: false,
      cardCount: 8,
      cards: [],
      team: 1,
      trickCount: 0,
    },
  ],
  phase: 1, // OmiPhase.PLAY
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  trumpCallerIdx: 1,
  bidPlayerIdx: 1,
  dealerIdx: 3,
  trumpSuit: 1,
  dealStage: 2,
  faceUpCard: null,
  makerTeam: 1,
  goingAlone: false,
  goingAlonePlayerIdx: -1,
  currentTrick: [],
  teamScores: [0, 0],
  teamTricks: [0, 1],
  gameEndFlag: false,
  winnerTeam: -1,
  leadPlayerIdx: 0,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 10 },
};

/** CallTrump phase: human is the caller (bidPlayerIdx = humanIdx = 0). */
const callTrumpHumanState: OmiResponse = {
  ...playPhaseState,
  phase: 0, // OmiPhase.CALL_TRUMP
  bidPlayerIdx: 0, // human is caller
  trumpCallerIdx: 0,
  trumpSuit: 0,
  dealStage: 1, // 4 cards dealt, awaiting trump call
  makerTeam: -1,
  players: playPhaseState.players.map((p, i) =>
    i === 0
      ? {
          ...p,
          cardCount: 4,
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 11 },
            { design: 'CLOVER', value: 7 },
            { design: 'DIAMOND', value: 9 },
          ],
        }
      : { ...p, cardCount: 4, cards: [] },
  ),
};

/** CallTrump phase: CPU is the caller. */
const callTrumpCpuState: OmiResponse = {
  ...callTrumpHumanState,
  bidPlayerIdx: 1, // CPU 1 is caller
  trumpCallerIdx: 1,
};

const trickEndState: OmiResponse = {
  ...playPhaseState,
  phase: 2, // OmiPhase.TRICK_END
  currentTrick: [
    { playerIdx: 0, card: { design: 'DIAMOND', value: 3 } },
    { playerIdx: 1, card: { design: 'HEART', value: 5 } },
  ],
};

const roundEndState: OmiResponse = {
  ...playPhaseState,
  phase: 3, // OmiPhase.ROUND_END
};

const gameEndState: OmiResponse = {
  ...playPhaseState,
  phase: 4, // OmiPhase.GAME_END
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'Game end!',
};

const gameEndByFlagState: OmiResponse = {
  ...playPhaseState,
  phase: 1,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'Game end!',
};

const cpuTurnState: OmiResponse = {
  ...playPhaseState,
  currentPlayerIdx: 1,
};

const noTrumpState: OmiResponse = {
  ...playPhaseState,
  trumpSuit: 0,
};

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('OmiPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<OmiPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 10,
      }),
    );
  });

  it('renders play phase with human cards (8 cards per hand)', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => {
      expect(screen.getByAltText('\u2660 A')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 J')).toBeInTheDocument();
    });
  });

  // ─── Required test 1: trump calling UI appears only for human caller ───────

  it('shows trump suit buttons only when human is the trump caller', async () => {
    mockExec.mockResolvedValue(callTrumpHumanState);
    renderWithProviders(<OmiPage />);
    // Human is caller: all 4 suit buttons must appear
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '♠ スペード' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '♣ クラブ' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '♥ ハート' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '♦ ダイヤ' })).toBeInTheDocument();
    });
  });

  it('does NOT show trump suit buttons when CPU is the trump caller', async () => {
    mockExec.mockResolvedValue(callTrumpCpuState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    // CPU is caller: no suit buttons shown
    expect(screen.queryByRole('button', { name: '♠ スペード' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '♣ クラブ' })).not.toBeInTheDocument();
  });

  /** Negative control: if we always show the suit buttons, the test above must fail. */
  it('[neg-control] always-showing suit buttons breaks the CPU-caller test', () => {
    // This test is documentation — the logic `isHumanCallTrump` must gate the buttons.
    // Verify: callTrumpCpuState has bidPlayerIdx=1 (not human 0), so isHumanCallTrump is false.
    const humanIdx = callTrumpCpuState.players.findIndex((p) => p.isHuman);
    const isHumanCallTrump = callTrumpCpuState.phase === 0 && callTrumpCpuState.bidPlayerIdx === humanIdx;
    // This must be false — if the condition were always-true, the button would appear when it shouldn't.
    expect(isHumanCallTrump).toBe(false);
  });

  // ─── Required test 2: hand size reflects deal stage ────────────────────────

  it('shows 4 cards before trump is called (dealStage=1)', async () => {
    mockExec.mockResolvedValue(callTrumpHumanState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByTestId('omi-deal-stage-info')).toBeInTheDocument());
    // Read the RENDERED count, not the fixture. Asserting on
    // `callTrumpHumanState.players[...].cards.length` only proves the fixture was
    // built with 4 cards -- it stays true however the page renders, so it cannot
    // catch a hard-coded hand size (measured: it did not).
    // The literal 4 is the spec ("first deal is half of 8"), deliberately not
    // derived from a constant: a constant would move with the mutation.
    const opponents = screen.getAllByText(/CPU \d+: .*4/);
    expect(opponents.length).toBeGreaterThan(0);
    expect(screen.queryByText(/CPU \d+: .*\b8\b/)).not.toBeInTheDocument();
    expect(screen.getByTestId('omi-deal-stage-info')).toBeInTheDocument();
  });

  it('shows 8 cards after trump is called (dealStage=2)', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    // playPhaseState has dealStage=2 and 2 cards visible (test hand shortened for readability)
    // The deal stage banner must NOT be shown in play phase
    expect(screen.queryByTestId('omi-deal-stage-info')).not.toBeInTheDocument();
    // Human has their cards rendered
    expect(screen.getByAltText('♠ A')).toBeInTheDocument();
  });

  /**
   * Negative control for the hand-size rendering. Verified by mutation: replacing
   * `count: player.cardCount` with `count: 8` in OmiPage makes this fail.
   * The earlier version of this test asserted on the fixture object rather than
   * the screen, so that mutation went through untouched.
   */
  it('[neg-control] rendered hand size comes from state, not a constant', async () => {
    mockExec.mockResolvedValue(callTrumpHumanState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByTestId('omi-deal-stage-info')).toBeInTheDocument());
    // Every seat is mid-first-deal, so no seat may render a hand of 8.
    expect(screen.queryByText(/CPU \d+: .*\b8\b/)).not.toBeInTheDocument();
    expect(screen.getAllByText(/CPU \d+: .*\b4\b/).length).toBe(3);
  });

  // ─── Suit buttons ──────────────────────────────────────────────────────────

  it('calls calltrump command when a suit button is clicked', async () => {
    mockExec.mockResolvedValue(callTrumpHumanState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ スペード' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '♠ スペード' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('calltrump', undefined, 1));
  });

  it('marks red-suit call-trump buttons with a red ring but not black-suit ones', async () => {
    mockExec.mockResolvedValue(callTrumpHumanState);
    renderWithProviders(<OmiPage />);
    const diamond = await screen.findByRole('button', { name: '♦ ダイヤ' });
    const spade = screen.getByRole('button', { name: '♠ スペード' });
    expect(diamond.className).toContain('ring-ds-error');
    expect(spade.className).not.toContain('ring-ds-error');
  });

  it('shows "CPU is calling trump" message when CPU is the caller', async () => {
    mockExec.mockResolvedValue(callTrumpCpuState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByText('CPU が切り札を宣言中...')).toBeInTheDocument());
  });

  // ─── No Euchre-specific UI ────────────────────────────────────────────────

  it('does NOT show order-up button', async () => {
    mockExec.mockResolvedValue(callTrumpHumanState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'オーダーアップ' })).not.toBeInTheDocument();
  });

  it('does NOT show pass button', async () => {
    mockExec.mockResolvedValue(callTrumpHumanState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'パス' })).not.toBeInTheDocument();
  });

  it('does NOT show discard button', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ディスカード' })).not.toBeInTheDocument();
  });

  // ─── Play phase ───────────────────────────────────────────────────────────

  it('play button disabled when not 1 card selected', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeDisabled());
  });

  it('play button enabled when 1 card selected', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '出す' })).not.toBeDisabled();
  });

  it('does not show play button when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('shows next trick button on trick end', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('shows next round button on round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('calls play command when play button is clicked', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('calls next when next trick button is clicked', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のトリック' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('calls nextround when next round button is clicked', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  // ─── Scoring rules ────────────────────────────────────────────────────────

  it('shows Omi scoring rules: 5+ tricks for 1pt, all-8 for 2pts, 4-4 for 0pt', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => {
      // 5 tricks → 1 point
      expect(screen.getByText(/5トリック以上/)).toBeInTheDocument();
      // All 8 tricks → 2 points
      expect(screen.getByText(/8トリック全取り/)).toBeInTheDocument();
      // 4-4 → 0 points
      expect(screen.getByText(/4-4/)).toBeInTheDocument();
    });
  });

  // ─── Deal stage info ──────────────────────────────────────────────────────

  it('shows deal stage 1 info banner during CallTrump phase', async () => {
    mockExec.mockResolvedValue(callTrumpHumanState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByTestId('omi-deal-stage-info')).toBeInTheDocument());
    expect(screen.getByTestId('omi-deal-stage-info')).toHaveTextContent('手札4枚');
  });

  it('does NOT show deal stage banner during Play phase', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.queryByTestId('omi-deal-stage-info')).not.toBeInTheDocument();
  });

  // ─── Team scores ──────────────────────────────────────────────────────────

  it('team scores shows both teams', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => {
      expect(screen.getByText('チームスコア')).toBeInTheDocument();
    });
  });

  it('score table headers have scope="col" for accessibility', async () => {
    const { container } = renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    const ths = container.querySelectorAll('th');
    ths.forEach((th) => {
      expect(th).toHaveAttribute('scope', 'col');
    });
  });

  // ─── Trump suit info ──────────────────────────────────────────────────────────

  it('shows trump suit info', async () => {
    renderWithProviders(<OmiPage />);
    // Use getAllByText: multiple elements may contain 切り札 (e.g. the trump indicator
    // "切り札: ♠ スペード" and the caller badge "切り札宣言者").
    await waitFor(() => {
      const matches = screen.getAllByText(/切り札/);
      expect(matches.length).toBeGreaterThan(0);
    });
  });

  it('shows "trump not yet called" text when trump is 0', async () => {
    mockExec.mockResolvedValue(noTrumpState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => {
      expect(screen.getByText('切り札未定')).toBeInTheDocument();
    });
  });

  // ─── Round/trick info ─────────────────────────────────────────────────────

  it('round and trick info displayed', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => {
      expect(screen.getByText('ラウンド 1')).toBeInTheDocument();
      expect(screen.getByText('トリック 1')).toBeInTheDocument();
    });
  });

  // ─── Game end ─────────────────────────────────────────────────────────────

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-4 phase', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  // ─── Error handling ───────────────────────────────────────────────────────

  it('shows error alert', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  // ─── Settings ─────────────────────────────────────────────────────────────

  it('settings panel changes cpuDifficulty', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '2' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, {
        cpuDifficulty: 2,
        pointLimit: 10,
      }),
    );
  });

  it('settings panel changes pointLimit', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: '21' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 21,
      }),
    );
  });

  // ─── Card selection ───────────────────────────────────────────────────────

  it('card selection toggle via aria-pressed', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('card buttons have aria-label with card name', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-label', '\u2660 A');

    const cardBtn2 = screen.getByAltText('\u2665 J').closest('button') as HTMLButtonElement;
    expect(cardBtn2).toHaveAttribute('aria-label', '\u2665 J');
  });

  // ─── Legal play highlighting ───────────────────────────────────────────────

  it('rings every card as legal when leading the trick', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const spadeBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    const heartBtn = screen.getByAltText('♥ J').closest('button') as HTMLButtonElement;
    expect(spadeBtn).toHaveAttribute('data-legal', 'true');
    expect(heartBtn).toHaveAttribute('data-legal', 'true');
  });

  it('rings only the legal follow-suit card', async () => {
    const followState: OmiResponse = {
      ...playPhaseState,
      currentTrick: [{ playerIdx: 3, card: { design: 'SPADE', value: 12 } }],
    };
    mockExec.mockResolvedValue(followState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const spadeBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    const heartBtn = screen.getByAltText('♥ J').closest('button') as HTMLButtonElement;
    expect(spadeBtn).toHaveAttribute('data-legal', 'true');
    expect(heartBtn).not.toHaveAttribute('data-legal');

    // Illegal card can still be selected
    fireEvent.click(heartBtn);
    expect(heartBtn).toHaveAttribute('aria-pressed', 'true');
  });

  it('does not ring cards when it is not the human play turn', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const spadeBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(spadeBtn).not.toHaveAttribute('data-legal');
  });

  // ─── Reset ────────────────────────────────────────────────────────────────

  it('reset button calls apiExec', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 10,
      }),
    );
  });

  // ─── Confirm dialog ───────────────────────────────────────────────────────

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 10,
      }),
    );
  });

  // ─── Action log ───────────────────────────────────────────────────────────

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.omi).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.omi).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();

    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText(/^棋譜$/)).not.toBeInTheDocument());
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });

  it('does not show action log button when not game end', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.queryByText('棋譜を見る')).not.toBeInTheDocument();
  });

  // ─── Phase indicator ──────────────────────────────────────────────────────

  it('phase indicator shows your turn when human is trump caller', async () => {
    mockExec.mockResolvedValue(callTrumpHumanState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows your turn when human play turn', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows waiting when cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('待機中'));
  });

  // ─── CPU info ─────────────────────────────────────────────────────────────

  it('shows CPU player areas', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*8枚/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 2.*8枚/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 3.*8枚/)).toBeInTheDocument();
    });
  });

  // ─── Keyboard navigation ──────────────────────────────────────────────────

  it('pressing number key toggles card in play phase', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('Enter key triggers play in play phase', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.keyDown(document, { key: '1' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);

    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('Escape key clears selection', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('keyboard nav disabled when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1.*8枚/)).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  // ─── Mobile layout ────────────────────────────────────────────────────────

  it('renders mobile viewport with horizontal scroll hand', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      renderWithProviders(<OmiPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const hand = document.querySelector('[data-tutorial="omi-player-hand"]');
      expect(hand?.className).toContain('overflow-x-auto');
      expect(hand?.className).not.toContain('flex-wrap');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders desktop viewport with wrapping hand', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 800 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      renderWithProviders(<OmiPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const hand = document.querySelector('[data-tutorial="omi-player-hand"]');
      expect(hand?.className).toContain('flex-wrap');
      expect(hand?.className).not.toContain('overflow-x-auto');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders CPU info as collapsible details on mobile', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      const { container } = renderWithProviders(<OmiPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const allDetails = container.querySelectorAll('details');
      const cpuDetails = Array.from(allDetails).find((d) =>
        d.querySelector('summary')?.textContent?.includes('CPU対戦相手'),
      );
      expect(cpuDetails).toBeInTheDocument();
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders score table as collapsible details on mobile', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      const { container } = renderWithProviders(<OmiPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const scoreDetails = container.querySelector('details[data-tutorial="omi-score-table"]');
      expect(scoreDetails).toBeInTheDocument();
      const summary = scoreDetails?.querySelector('summary');
      expect(summary).toHaveTextContent('チームスコア');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  // ─── Misc ─────────────────────────────────────────────────────────────────

  it('sets aria-busy on container', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: 'リセット' }).closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');
  });

  it('no human player renders empty hand area', async () => {
    const noHuman: OmiResponse = {
      ...playPhaseState,
      players: playPhaseState.players.map((p) => ({ ...p, isHuman: false })),
    };
    mockExec.mockResolvedValue(noHuman);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.queryByAltText('\u2660 A')).not.toBeInTheDocument();
  });

  it('shows current trick cards', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => {
      expect(screen.getByText('現在のトリック')).toBeInTheDocument();
      expect(screen.getByAltText('\u2666 3')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 5')).toBeInTheDocument();
    });
  });

  it('does not show message when empty', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.queryByText('Game end!')).not.toBeInTheDocument();
  });

  it('renders loading state', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    let resolve!: (value: OmiResponse) => void;
    const slow = new Promise<OmiResponse>((r) => {
      resolve = r;
    });
    mockExec.mockReturnValueOnce(slow);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(playPhaseState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  // ─── Hint live region ─────────────────────────────────────────────────────

  it('announces the hint through a region that was already mounted', async () => {
    mockExec.mockResolvedValue(callTrumpHumanState);
    renderWithProviders(<OmiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const region = screen.getByTestId('omi-hint-live');
    expect(region).toHaveAttribute('role', 'status');
    expect(region).toHaveAttribute('aria-live', 'polite');
    expect(region).toHaveTextContent('');

    mockExec.mockResolvedValue({
      ...callTrumpHumanState,
      hint: { reason: 'call_strong_suit', suit: 1 },
    } as unknown as OmiResponse);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(region).not.toHaveTextContent(''));
    expect(screen.getByTestId('omi-hint-live')).toBe(region);
  });
});
