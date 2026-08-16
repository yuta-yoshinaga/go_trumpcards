import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, ginrummyApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { GinRummyResponse } from '../types/card';
import { GinRummyCpu } from '../types/phases';
import { GinRummyPage } from './GinRummyPage';

vi.mock('../api/gameApi', () => ({
  ginrummyApi: { exec: vi.fn() },
  actionLogApi: { ginrummy: vi.fn() },
}));

const mockExec = vi.mocked(ginrummyApi.exec);

const drawPhaseState: GinRummyResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 10,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
    },
    { id: 1, isHuman: false, cardCount: 10, cards: [], roundScore: 3, cumulativeScore: 10 },
  ],
  layoffTargets: [],
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 30,
  gameEndFlag: false,
  winnerIdx: -1,
  knockerIdx: -1,
  knockerMelds: [],
  knockerDeadwood: [],
  isGin: false,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 100 },
};

const discardPhaseState: GinRummyResponse = {
  ...drawPhaseState,
  phase: 1,
};

const layoffPhaseState: GinRummyResponse = {
  ...drawPhaseState,
  phase: 2,
  knockerMelds: [
    {
      cards: [
        { design: 'SPADE', value: 3 },
        { design: 'HEART', value: 3 },
        { design: 'DIAMOND', value: 3 },
      ],
    },
  ],
  knockerDeadwood: [{ design: 'CLOVER', value: 9 }],
};

const roundEndState: GinRummyResponse = {
  ...drawPhaseState,
  phase: 3,
};

const gameEndState: GinRummyResponse = {
  ...drawPhaseState,
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const gameEndByFlagState: GinRummyResponse = {
  ...drawPhaseState,
  phase: 0,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const cpuTurnState: GinRummyResponse = {
  ...drawPhaseState,
  currentPlayerIdx: 1,
};

const noDiscardState: GinRummyResponse = {
  ...drawPhaseState,
  discardTop: null,
};

const layoffCpuCardsState: GinRummyResponse = {
  ...drawPhaseState,
  phase: 2,
  knockerMelds: [],
  players: [
    drawPhaseState.players[0],
    {
      id: 1,
      isHuman: false,
      cardCount: 3,
      cards: [
        { design: 'DIAMOND', value: 5 },
        { design: 'CLOVER', value: 6 },
        { design: 'HEART', value: 8 },
      ],
      roundScore: 3,
      cumulativeScore: 10,
    },
  ],
};

const roundEndCpuCardsState: GinRummyResponse = {
  ...drawPhaseState,
  phase: 3,
  players: [
    drawPhaseState.players[0],
    {
      id: 1,
      isHuman: false,
      cardCount: 3,
      cards: [
        { design: 'DIAMOND', value: 5 },
        { design: 'CLOVER', value: 6 },
        { design: 'HEART', value: 8 },
      ],
      roundScore: 3,
      cumulativeScore: 10,
    },
  ],
};

// Round-end fixtures for the score breakdown. The opponent's (CPU, id 1) hand
// is scored via the same meld search the domain uses.
const knockRoundEndState: GinRummyResponse = {
  ...drawPhaseState,
  phase: 3,
  knockerIdx: 0,
  isGin: false,
  knockerDeadwood: [{ design: 'CLOVER', value: 9 }], // knocker deadwood 9
  players: [
    drawPhaseState.players[0],
    {
      id: 1,
      isHuman: false,
      cardCount: 3,
      // No meld: 5 + 6 + 8 = 19 deadwood > 9 → plain knock, knocker scores 10.
      cards: [
        { design: 'DIAMOND', value: 5 },
        { design: 'CLOVER', value: 6 },
        { design: 'HEART', value: 8 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
    },
  ],
};

const ginRoundEndState: GinRummyResponse = {
  ...knockRoundEndState,
  isGin: true,
  knockerDeadwood: [], // gin → knocker deadwood 0, scores opponent 19 + 25 bonus = 44
};

const undercutRoundEndState: GinRummyResponse = {
  ...drawPhaseState,
  phase: 3,
  knockerIdx: 0,
  isGin: false,
  knockerDeadwood: [{ design: 'CLOVER', value: 9 }], // knocker deadwood 9
  players: [
    drawPhaseState.players[0],
    {
      id: 1,
      isHuman: false,
      cardCount: 4,
      // ♦5-6-7 run (melded) + ♥2 → deadwood 2 ≤ 9 → undercut, opponent scores 7 + 25 = 32.
      cards: [
        { design: 'DIAMOND', value: 5 },
        { design: 'DIAMOND', value: 6 },
        { design: 'DIAMOND', value: 7 },
        { design: 'HEART', value: 2 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
    },
  ],
};

beforeEach(() => {
  mockExec.mockResolvedValue(drawPhaseState);
});

describe('GinRummyPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<GinRummyPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 1,
        pointLimit: 100,
      }),
    );
  });

  it('shows live deadwood indicator during discard phase', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByTestId('ginrummy-deadwood-indicator')).toBeInTheDocument());
  });

  it('does not show deadwood indicator outside discard phase', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.queryByTestId('ginrummy-deadwood-indicator')).not.toBeInTheDocument();
  });

  it('color-codes the hand into meld and deadwood cards during discard phase', async () => {
    // ♠5-6-7 form a run (melded); ♥K is deadwood.
    const meldHandState: GinRummyResponse = {
      ...discardPhaseState,
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 4,
          cards: [
            { design: 'SPADE', value: 5 },
            { design: 'SPADE', value: 6 },
            { design: 'SPADE', value: 7 },
            { design: 'HEART', value: 13 },
          ],
          roundScore: 0,
          cumulativeScore: 0,
        },
        discardPhaseState.players[1],
      ],
    };
    mockExec.mockResolvedValue(meldHandState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByTestId('gr-hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('gr-hand-card-0')).toHaveAttribute('data-meld', 'meld');
    expect(screen.getByTestId('gr-hand-card-1')).toHaveAttribute('data-meld', 'meld');
    expect(screen.getByTestId('gr-hand-card-2')).toHaveAttribute('data-meld', 'meld');
    expect(screen.getByTestId('gr-hand-card-3')).toHaveAttribute('data-meld', 'deadwood');
    expect(screen.getByTestId('ginrummy-meld-legend')).toBeInTheDocument();
  });

  it('does not color-code the hand outside discard phase', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByTestId('gr-hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('gr-hand-card-0')).not.toHaveAttribute('data-meld');
    expect(screen.queryByTestId('ginrummy-meld-legend')).not.toBeInTheDocument();
  });

  it('pulses the knock button when deadwood ≤10 during discard phase', async () => {
    const lowDeadwoodHand: GinRummyResponse = {
      ...discardPhaseState,
      players: [
        {
          ...discardPhaseState.players[0],
          cards: [
            { design: 'SPADE', value: 5 },
            { design: 'SPADE', value: 6 },
            { design: 'SPADE', value: 7 },
            { design: 'HEART', value: 3 },
          ],
        },
        ...discardPhaseState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(lowDeadwoodHand);
    renderWithProviders(<GinRummyPage />);
    const knockBtn = await screen.findByTestId('ginrummy-knock-button');
    expect(knockBtn.className).toContain('animate-pulse');
  });

  it('considers post-discard deadwood, not the full 11-card hand', async () => {
    // 11-card hand whose full deadwood is 11 (K + A = 11) but drops to
    // 0 once the King is discarded (♠5-6-7 run + 7♥-7♣-7♦ set + ace).
    // The knock button must pulse because a single discard makes it ≤ 10.
    const eleven: GinRummyResponse = {
      ...discardPhaseState,
      players: [
        {
          ...discardPhaseState.players[0],
          cardCount: 11,
          cards: [
            { design: 'SPADE', value: 5 },
            { design: 'SPADE', value: 6 },
            { design: 'SPADE', value: 7 },
            { design: 'HEART', value: 7 },
            { design: 'CLOVER', value: 7 },
            { design: 'DIAMOND', value: 7 },
            { design: 'HEART', value: 8 },
            { design: 'HEART', value: 9 },
            { design: 'HEART', value: 10 },
            { design: 'CLOVER', value: 1 },
            { design: 'SPADE', value: 13 },
          ],
        },
        ...discardPhaseState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(eleven);
    renderWithProviders(<GinRummyPage />);
    const knockBtn = await screen.findByTestId('ginrummy-knock-button');
    expect(knockBtn.className).toContain('animate-pulse');
  });

  it('renders draw phase with human cards', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => {
      expect(screen.getByAltText('\u2660 A')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 J')).toBeInTheDocument();
    });
  });

  it('renders draw stock and draw discard buttons when human draw turn', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeInTheDocument();
    });
  });

  it('draw discard button disabled when no discard top', async () => {
    mockExec.mockResolvedValue(noDiscardState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeDisabled());
  });

  it('calls drawstock command when draw stock button is clicked', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('calls drawdiscard command when draw discard button is clicked', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '捨て札から引く' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('renders discard and knock buttons when human discard turn', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'ノック' })).toBeInTheDocument();
    });
  });

  it('discard button disabled when not 1 card selected', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨てる' })).toBeDisabled());
  });

  it('discard button enabled when 1 card selected', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '捨てる' })).not.toBeDisabled();
  });

  it('calls discard command when discard button is clicked', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '捨てる' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 0));
  });

  it('calls knock command when knock button is clicked', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(layoffPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ノック' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('knock', 0));
  });

  it('knock button disabled when not 1 card selected', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ノック' })).toBeDisabled());
  });

  it('renders layoff and skip buttons when human layoff turn', async () => {
    mockExec.mockResolvedValue(layoffPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'レイオフ' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument();
    });
  });

  it('layoff button disabled when no card selected', async () => {
    mockExec.mockResolvedValue(layoffPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'レイオフ' })).toBeDisabled());
  });

  it('shows a type badge on each knocker meld during layoff', async () => {
    mockExec.mockResolvedValue(layoffPhaseState);
    renderWithProviders(<GinRummyPage />);
    // The fixture meld is a set of three 3s → rank badge "3".
    await waitFor(() => expect(screen.getByTestId('gr-meld-badge-0')).toHaveTextContent('3'));
  });

  it('calls layoff command when layoff button is clicked', async () => {
    mockExec.mockResolvedValue(layoffPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(roundEndState);
    fireEvent.click(screen.getByRole('button', { name: 'レイオフ' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('layoff', undefined, undefined, [0]));
  });

  it('calls layoff with empty array when skip button is clicked', async () => {
    mockExec.mockResolvedValue(layoffPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(roundEndState);
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('layoff', undefined, undefined, []));
  });

  it('does not show draw buttons when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '山札から引く' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '捨て札から引く' })).not.toBeInTheDocument();
  });

  it('shows next round button on round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('calls nextround when next round button is clicked', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  // -- Round-end score breakdown --

  it('shows a plain-knock score breakdown at round end', async () => {
    mockExec.mockResolvedValue(knockRoundEndState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByTestId('ginrummy-score-breakdown')).toBeInTheDocument());
    expect(screen.getByTestId('ginrummy-breakdown-outcome')).toHaveTextContent('ノック');
    // opponent deadwood 19 − knocker 9 = 10, no bonus.
    expect(screen.getByTestId('ginrummy-breakdown-formula')).toHaveTextContent('デッドウッド差 10 = 10');
    expect(screen.getByText('あなた が 10 点獲得')).toBeInTheDocument();
  });

  it('shows a gin score breakdown at round end', async () => {
    mockExec.mockResolvedValue(ginRoundEndState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByTestId('ginrummy-score-breakdown')).toBeInTheDocument());
    expect(screen.getByTestId('ginrummy-breakdown-outcome')).toHaveTextContent('ジン');
    // opponent deadwood 19 + gin bonus 25 = 44.
    expect(screen.getByTestId('ginrummy-breakdown-formula')).toHaveTextContent(
      '相手デッドウッド 19 + ジンボーナス 25 = 44',
    );
    expect(screen.getByText('あなた が 44 点獲得')).toBeInTheDocument();
  });

  it('shows an undercut score breakdown crediting the defender at round end', async () => {
    mockExec.mockResolvedValue(undercutRoundEndState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByTestId('ginrummy-score-breakdown')).toBeInTheDocument());
    expect(screen.getByTestId('ginrummy-breakdown-outcome')).toHaveTextContent('アンダーカット');
    // knocker 9 − opponent 2 = 7 difference + undercut bonus 25 = 32, credited to CPU.
    expect(screen.getByTestId('ginrummy-breakdown-formula')).toHaveTextContent(
      'デッドウッド差 7 + アンダーカットボーナス 25 = 32',
    );
    expect(screen.getByText('CPU 1 が 32 点獲得')).toBeInTheDocument();
  });

  it('hides the score breakdown before round end', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByTestId('ginrummy-score-breakdown')).not.toBeInTheDocument();
  });

  it('hides the score breakdown on a drawn round (no knocker)', async () => {
    // roundEndState keeps knockerIdx -1 (stock exhausted) → no scoring to show.
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.queryByTestId('ginrummy-score-breakdown')).not.toBeInTheDocument();
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-4 phase', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows error alert', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('shows CPU player area', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*10枚/)).toBeInTheDocument();
    });
  });

  it('score table shows all players', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => {
      expect(screen.getByText('スコア')).toBeInTheDocument();
      expect(screen.getByText('あなた')).toBeInTheDocument();
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
    });
  });

  it('score table headers have scope="col" for accessibility', async () => {
    const { container } = renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByText('あなた')).toBeInTheDocument());
    const ths = container.querySelectorAll('th');
    ths.forEach((th) => {
      expect(th).toHaveAttribute('scope', 'col');
    });
  });

  it('shows discard top card', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => {
      expect(screen.getByText('捨て札')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 7')).toBeInTheDocument();
    });
  });

  it('does not show discard top when null', async () => {
    mockExec.mockResolvedValue(noDiscardState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('捨て札')).not.toBeInTheDocument();
  });

  it('shows knocker melds when present', async () => {
    mockExec.mockResolvedValue(layoffPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => {
      expect(screen.getByText('ノッカーのメルド')).toBeInTheDocument();
    });
  });

  it('does not show knocker melds when empty', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('ノッカーのメルド')).not.toBeInTheDocument();
  });

  it('card selection toggle via aria-pressed', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('card buttons have aria-label with card name', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-label', '\u2660 A');

    const cardBtn2 = screen.getByAltText('\u2665 J').closest('button') as HTMLButtonElement;
    expect(cardBtn2).toHaveAttribute('aria-label', '\u2665 J');
  });

  it('reset button calls exec with confirm', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 1,
        pointLimit: 100,
      }),
    );
  });

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('settings panel changes cpuDifficulty', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '2' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 2,
        pointLimit: 100,
      }),
    );
  });

  it('settings panel changes pointLimit', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: '150' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 1,
        pointLimit: 150,
      }),
    );
  });

  it('round info displayed', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => {
      expect(screen.getByText('ラウンド 1')).toBeInTheDocument();
      expect(screen.getByText('山札: 30枚')).toBeInTheDocument();
    });
  });

  it('does not show message when empty', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('Game end!')).not.toBeInTheDocument();
  });

  it('shows loading state', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    let resolve!: (value: GinRummyResponse) => void;
    const slow = new Promise<GinRummyResponse>((r) => {
      resolve = r;
    });
    mockExec.mockReturnValueOnce(slow);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(drawPhaseState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
  });

  it('sets aria-busy on container', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: 'リセット' }).closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');
  });

  it('no human cards renders empty hand area', async () => {
    const noHuman: GinRummyResponse = {
      ...drawPhaseState,
      players: drawPhaseState.players.map((p) => ({ ...p, isHuman: false })),
    };
    mockExec.mockResolvedValue(noHuman);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByAltText('\u2660 A')).not.toBeInTheDocument();
  });

  it('isHumanTurn false when currentPlayerIdx points to cpu', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '山札から引く' })).not.toBeInTheDocument();
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.ginrummy).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.ginrummy).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();

    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText(/^棋譜$/)).not.toBeInTheDocument());
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });

  it('does not show action log button when not game end', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('棋譜を見る')).not.toBeInTheDocument();
  });

  it('disables buttons while loading', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    let resolve!: (value: GinRummyResponse) => void;
    const slow = new Promise<GinRummyResponse>((r) => {
      resolve = r;
    });
    mockExec.mockReturnValueOnce(slow);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(drawPhaseState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
  });

  // -- PhaseIndicator coverage --

  it('phase indicator shows your turn when human draw turn', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows your turn in discard phase', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows your turn in layoff phase', async () => {
    mockExec.mockResolvedValue(layoffPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows waiting when cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('待機中'));
  });

  // -- Keyboard navigation --

  it('pressing number key toggles card in discard phase', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('Enter key triggers discard in discard phase', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.keyDown(document, { key: '1' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);

    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 0));
  });

  it('Enter key triggers layoff in layoff phase', async () => {
    mockExec.mockResolvedValue(layoffPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.keyDown(document, { key: '1' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(roundEndState);

    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('layoff', undefined, undefined, [0]));
  });

  it('Escape key clears selection', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('keyboard nav disabled when not human turn', async () => {
    const cpuDiscardState: GinRummyResponse = { ...discardPhaseState, currentPlayerIdx: 1 };
    mockExec.mockResolvedValue(cpuDiscardState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('keyboard nav disabled in draw phase', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('reset hides the action log panel if open', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.ginrummy).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());

    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));

    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
  });

  it('shows CPU cards during layoff phase', async () => {
    mockExec.mockResolvedValue(layoffCpuCardsState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => {
      expect(screen.getByAltText('\u2666 5')).toBeInTheDocument();
      expect(screen.getByAltText('\u2663 6')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 8')).toBeInTheDocument();
    });
  });

  it('shows CPU cards during round end phase', async () => {
    mockExec.mockResolvedValue(roundEndCpuCardsState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => {
      expect(screen.getByAltText('\u2666 5')).toBeInTheDocument();
    });
  });

  it('does not show CPU cards during draw phase', async () => {
    const drawWithCpuCards: GinRummyResponse = {
      ...drawPhaseState,
      players: [
        drawPhaseState.players[0],
        {
          id: 1,
          isHuman: false,
          cardCount: 3,
          cards: [{ design: 'DIAMOND', value: 5 }],
          roundScore: 3,
          cumulativeScore: 10,
        },
      ],
    };
    mockExec.mockResolvedValue(drawWithCpuCards);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByAltText('\u2666 5')).not.toBeInTheDocument();
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(drawPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  // **レイオフフェーズの主題そのもの (#4823)。**ディスカードでは
  // メルド/デッドウッドを見せているのに、レイオフには補助が無かった。
  it('marks which hand cards can be laid off during the layoff phase', async () => {
    mockExec.mockResolvedValue({
      ...layoffPhaseState,
      // 手札 0 枚目だけがノッカーのメルドに足せる。
      layoffTargets: [[0], [], []],
    });
    renderWithProviders(<GinRummyPage />);

    await waitFor(() => expect(document.querySelectorAll('[data-layoff="yes"]')).toHaveLength(1));
    expect(document.querySelectorAll('[data-layoff="no"]').length).toBeGreaterThan(0);
  });

  it('does not mark layoffable cards outside the layoff phase', async () => {
    mockExec.mockResolvedValue({ ...discardPhaseState, layoffTargets: [[0], [], []] });
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(document.querySelectorAll('[data-layoff]')).toHaveLength(0);
  });
});

// #5500: 難易度で変わるのは CPU のノック判断と捨て札の拾い方なのに、選択肢には
// Easy/Normal/Hard のラベルしか出ておらず、実際の判断基準はプレイヤーから見えない。
describe('GinRummyPage difficulty policies', () => {
  it('summarises what each difficulty actually does', async () => {
    mockExec.mockResolvedValue(drawPhaseState);
    renderWithProviders(<GinRummyPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    fireEvent.click(screen.getByText('設定'));

    const select = screen.getByRole('combobox', { name: 'CPU難易度' });
    const labels = Array.from(select.querySelectorAll('option')).map((o) => o.textContent ?? '');
    expect(labels).toHaveLength(3);

    // **数値は GinRummyCpu から補間される。** ラベルに直接書くと domain と乖離する。
    expect(labels[1]).toContain(String(GinRummyCpu.KNOCK_DEADWOOD_NORMAL));
    expect(labels[2]).toContain(String(GinRummyCpu.KNOCK_DEADWOOD_HARD));
    // Easy は法定上限でノックし、拾いは確率。
    expect(labels[0]).toContain(String(GinRummyCpu.KNOCK_THRESHOLD));
    expect(labels[0]).toContain(String(GinRummyCpu.EASY_PICK_ONE_IN));
    // 3つとも素のラベルのままではないこと。
    for (const label of labels) {
      expect(label).toMatch(/ノック/);
    }
  });
});
