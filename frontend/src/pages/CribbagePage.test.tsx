import { act, fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, cribbageApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CribbageResponse } from '../types/card';
import { CribbagePage } from './CribbagePage';

vi.mock('../api/gameApi', () => ({
  cribbageApi: { exec: vi.fn() },
  actionLogApi: { cribbage: vi.fn() },
}));

const mockExec = vi.mocked(cribbageApi.exec);

const discardPhaseState: CribbageResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 6,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
        { design: 'DIAMOND', value: 5 },
        { design: 'CLOVER', value: 8 },
        { design: 'SPADE', value: 3 },
        { design: 'HEART', value: 7 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
    },
    { id: 1, isHuman: false, cardCount: 6, cards: [], roundScore: 0, cumulativeScore: 0 },
  ],
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  dealerIdx: 0,
  crib: [],
  starter: null,
  pegCount: 0,
  pegPlayedCards: [],
  showPhaseStep: 0,
  handScoreDetails: [null, null, null],
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 121 },
};

const peggingPhaseState: CribbageResponse = {
  ...discardPhaseState,
  phase: 2,
  players: [
    {
      ...discardPhaseState.players[0],
      cardCount: 4,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
        { design: 'DIAMOND', value: 5 },
        { design: 'CLOVER', value: 8 },
      ],
    },
    discardPhaseState.players[1],
  ],
  starter: { design: 'SPADE', value: 10 },
  pegCount: 0,
  pegPlayedCards: [],
};

// Cut phase with the human as the non-dealer cutter (dealer=1 → cutter=0).
const cutPhaseState: CribbageResponse = {
  ...discardPhaseState,
  phase: 1,
  dealerIdx: 1,
  currentPlayerIdx: 0,
  starter: null,
  players: [
    { ...discardPhaseState.players[0], cardCount: 4, cards: discardPhaseState.players[0].cards.slice(0, 4) },
    discardPhaseState.players[1],
  ],
};

// Cut phase where the CPU is the cutter (dealer=0 → cutter=1) — no human affordance.
const cutPhaseCpuState: CribbageResponse = {
  ...cutPhaseState,
  dealerIdx: 0,
  currentPlayerIdx: 1,
};

const showPhaseState: CribbageResponse = {
  ...discardPhaseState,
  phase: 3,
  dealerIdx: 1,
  starter: { design: 'SPADE', value: 10 },
  handScoreDetails: [{ fifteens: 2, pairs: 0, runs: 3, flush: 0, nobs: 0, total: 5 }, null, null],
  players: [
    {
      ...discardPhaseState.players[0],
      cardCount: 4,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
        { design: 'DIAMOND', value: 5 },
        { design: 'CLOVER', value: 8 },
      ],
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 4,
      cards: [
        { design: 'DIAMOND', value: 2 },
        { design: 'CLOVER', value: 3 },
        { design: 'HEART', value: 4 },
        { design: 'SPADE', value: 6 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
    },
  ],
};

const roundEndState: CribbageResponse = {
  ...discardPhaseState,
  phase: 4,
};

const gameEndState: CribbageResponse = {
  ...discardPhaseState,
  phase: 5,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const gameEndByFlagState: CribbageResponse = {
  ...discardPhaseState,
  phase: 0,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const cpuTurnState: CribbageResponse = {
  ...discardPhaseState,
  currentPlayerIdx: 1,
};

beforeEach(() => {
  mockExec.mockResolvedValue(discardPhaseState);
});

describe('CribbagePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<CribbagePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 121,
      }),
    );
  });

  it('renders discard phase with human cards', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => {
      expect(screen.getByAltText('\u2660 A')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 J')).toBeInTheDocument();
    });
  });

  it('renders discard button when human discard turn', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'クリブに捨てる' })).toBeInTheDocument();
    });
  });

  it('discard button disabled when not 2 cards selected', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'クリブに捨てる' })).toBeDisabled());
  });

  it('discard button enabled when 2 cards selected', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('\u2665 J').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: 'クリブに捨てる' })).not.toBeDisabled();
  });

  it('calls discard command when discard button is clicked', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('\u2665 J').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(peggingPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'クリブに捨てる' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', undefined, [0, 1]));
  });

  it('renders cut deck and cut button when human is the cutter', async () => {
    mockExec.mockResolvedValue(cutPhaseState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => {
      expect(screen.getByTestId('cb-cut-deck')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'デッキをカットする' })).toBeInTheDocument();
    });
  });

  it('calls cut command when the cut button is clicked', async () => {
    mockExec.mockResolvedValue(cutPhaseState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByTestId('cb-cut-button')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(peggingPhaseState);
    fireEvent.click(screen.getByTestId('cb-cut-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('cut'));
  });

  it('calls cut command when the deck is clicked', async () => {
    mockExec.mockResolvedValue(cutPhaseState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByTestId('cb-cut-deck')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(peggingPhaseState);
    fireEvent.click(screen.getByTestId('cb-cut-deck'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('cut'));
  });

  it('does not show cut affordance when the CPU is the cutter', async () => {
    mockExec.mockResolvedValue(cutPhaseCpuState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByTestId('cb-cut-button')).not.toBeInTheDocument();
    expect(screen.queryByTestId('cb-cut-deck')).not.toBeInTheDocument();
  });

  it('shows the His Heels note when the starter is a Jack', async () => {
    mockExec.mockResolvedValue({ ...peggingPhaseState, starter: { design: 'HEART', value: 11 } });
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByTestId('cb-his-heels')).toBeInTheDocument());
  });

  it('renders peg and go buttons when human pegging turn', async () => {
    mockExec.mockResolvedValue(peggingPhaseState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'カードを出す' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Go' })).toBeInTheDocument();
    });
  });

  it('peg button disabled when not 1 card selected', async () => {
    mockExec.mockResolvedValue(peggingPhaseState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'カードを出す' })).toBeDisabled());
  });

  it('calls peg command when peg button is clicked', async () => {
    mockExec.mockResolvedValue(peggingPhaseState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(peggingPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'カードを出す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('peg', 0));
  });

  it('auto-dispatches go after a brief delay when the human has no playable card and shows the notice', async () => {
    vi.useFakeTimers();
    try {
      const goState: CribbageResponse = {
        ...peggingPhaseState,
        pegCount: 28,
        players: [
          {
            ...peggingPhaseState.players[0],
            cardCount: 1,
            cards: [{ design: 'HEART', value: 10 }], // 28+10=38 > 31, no playable card.
          },
          peggingPhaseState.players[1],
        ],
      };
      mockExec.mockResolvedValue(goState);
      renderWithProviders(<CribbagePage />);
      // The auto-Go notice appears as soon as the no-playable state is reflected.
      await vi.waitFor(() => expect(screen.getByTestId('cb-auto-go-notice')).toBeInTheDocument());

      mockExec.mockClear();
      mockExec.mockResolvedValue(goState);
      // Advance past the 1s delay; the timeout should fire 'go' without any user click.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(1000);
      });
      expect(mockExec).toHaveBeenCalledWith('go');
    } finally {
      vi.useRealTimers();
    }
  });

  it('does not auto-dispatch go when the human has at least one playable card', async () => {
    vi.useFakeTimers();
    try {
      mockExec.mockResolvedValue(peggingPhaseState);
      renderWithProviders(<CribbagePage />);
      await vi.waitFor(() => expect(screen.getByRole('button', { name: 'Go' })).toBeInTheDocument());
      // No notice while the player can still peg.
      expect(screen.queryByTestId('cb-auto-go-notice')).not.toBeInTheDocument();
      mockExec.mockClear();
      await act(async () => {
        await vi.advanceTimersByTimeAsync(2000);
      });
      expect(mockExec).not.toHaveBeenCalledWith('go');
    } finally {
      vi.useRealTimers();
    }
  });

  it('calls go command when go button is clicked', async () => {
    // Go is only enabled when no playable cards (pegCount high, only high cards left)
    const goState: CribbageResponse = {
      ...peggingPhaseState,
      pegCount: 28,
      players: [
        {
          ...peggingPhaseState.players[0],
          cardCount: 1,
          cards: [{ design: 'HEART', value: 10 }], // 10 value, 28+10=38>31 → can't play
        },
        peggingPhaseState.players[1],
      ],
    };
    mockExec.mockResolvedValue(goState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'Go' })).toBeEnabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(goState);
    fireEvent.click(screen.getByRole('button', { name: 'Go' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('go'));
  });

  it('shows show next button in show phase', async () => {
    mockExec.mockResolvedValue(showPhaseState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次を表示' })).toBeInTheDocument());
  });

  it('gives the score-breakdown table a caption and a labeled first column header', async () => {
    mockExec.mockResolvedValue(showPhaseState);
    renderWithProviders(<CribbagePage />);
    // The breakdown table is captioned...
    const table = await screen.findByRole('table', { name: 'ショー得点の内訳' });
    expect(table).toBeInTheDocument();
    // ...and its previously-empty first column header now has an accessible name.
    const headers = within(table).getAllByRole('columnheader');
    expect(headers[0]).toHaveTextContent('対象');
  });

  it('calls shownext when show next button is clicked', async () => {
    mockExec.mockResolvedValue(showPhaseState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次を表示' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(roundEndState);
    fireEvent.click(screen.getByRole('button', { name: '次を表示' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('shownext'));
  });

  it('shows next round button on round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('calls nextround when next round button is clicked', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('does not show discard buttons when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'クリブに捨てる' })).not.toBeInTheDocument();
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-5 phase', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows error alert', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('shows CPU player area', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*6枚/)).toBeInTheDocument();
    });
  });

  it('score table shows all players', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => {
      expect(screen.getByText('スコア')).toBeInTheDocument();
      expect(screen.getAllByText('あなた').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('CPU 1').length).toBeGreaterThanOrEqual(1);
    });
  });

  it('card selection toggle via aria-pressed', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('card buttons have aria-label with card name', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-label', '\u2660 A');
  });

  it('reset button calls exec with confirm', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 121,
      }),
    );
  });

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('settings panel changes cpuDifficulty', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '2' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 2,
        pointLimit: 121,
      }),
    );
  });

  it('settings panel changes pointLimit', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: '61' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 61,
      }),
    );
  });

  it('round info displayed', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => {
      expect(screen.getByText('ラウンド 1')).toBeInTheDocument();
    });
  });

  it('shows starter card when present', async () => {
    mockExec.mockResolvedValue(peggingPhaseState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => {
      expect(screen.getByText('スターター')).toBeInTheDocument();
      expect(screen.getByAltText('\u2660 10')).toBeInTheDocument();
    });
  });

  it('does not show starter when null', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('スターター')).not.toBeInTheDocument();
  });

  it('shows hand score details in show phase', async () => {
    mockExec.mockResolvedValue(showPhaseState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => {
      expect(screen.getByText('あなたのハンド')).toBeInTheDocument();
    });
  });

  it('shows CPU cards during show phase', async () => {
    mockExec.mockResolvedValue(showPhaseState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => {
      expect(screen.getByAltText('\u2666 2')).toBeInTheDocument();
    });
  });

  it('does not show CPU cards during discard phase', async () => {
    const discardWithCpuCards: CribbageResponse = {
      ...discardPhaseState,
      players: [
        discardPhaseState.players[0],
        {
          id: 1,
          isHuman: false,
          cardCount: 3,
          cards: [{ design: 'DIAMOND', value: 13 }],
          roundScore: 0,
          cumulativeScore: 0,
        },
      ],
    };
    mockExec.mockResolvedValue(discardWithCpuCards);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByAltText('\u2666 K')).not.toBeInTheDocument();
  });

  it('shows peg board with a localized label and screen-reader score summaries', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => {
      expect(screen.getByTestId('peg-board')).toBeInTheDocument();
    });
    // aria-label is localized (no longer the literal "peg-board").
    expect(screen.getByLabelText('ペグボード（得点）')).toBeInTheDocument();
    expect(screen.queryByLabelText('peg-board')).not.toBeInTheDocument();
    // Each player has an sr-only score/goal summary.
    expect(screen.getByTestId('peg-score-summary-0')).toHaveTextContent('/ 121 点');
  });

  it('renders the rear-peg trail after the player score jumps', async () => {
    const startState: CribbageResponse = {
      ...discardPhaseState,
      players: [{ ...discardPhaseState.players[0], cumulativeScore: 10 }, { ...discardPhaseState.players[1] }],
    };
    mockExec.mockResolvedValue(startState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByTestId('front-peg-0')).toBeInTheDocument());
    // First render captures 10 into prevScoresRef → no rear peg yet.
    expect(screen.queryByTestId('rear-peg-0')).not.toBeInTheDocument();

    // Trigger a re-render with a higher score.
    mockExec.mockResolvedValueOnce({
      ...startState,
      players: [{ ...startState.players[0], cumulativeScore: 22 }, { ...startState.players[1] }],
    });
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.getByTestId('rear-peg-0')).toBeInTheDocument());
  });

  it('uses 15-point ticks on a 61-point board', async () => {
    mockExec.mockResolvedValue({
      ...discardPhaseState,
      config: { cpuDifficulty: 1, pointLimit: 61 },
    });
    renderWithProviders(<CribbagePage />);
    await screen.findByTestId('peg-board');
    // 2 tracks × 4 ticks each (15, 30, 45, 60 — strictly < pointLimit 61) = 8 tick elements.
    expect(screen.getAllByTestId('peg-tick')).toHaveLength(8);
  });

  it('shows loading state', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    let resolve!: (value: CribbageResponse) => void;
    const slow = new Promise<CribbageResponse>((r) => {
      resolve = r;
    });
    mockExec.mockReturnValueOnce(slow);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(discardPhaseState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
  });

  it('sets aria-busy on container', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: 'リセット' }).closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.cribbage).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.cribbage).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();

    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText(/^棋譜$/)).not.toBeInTheDocument());
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });

  it('does not show action log button when not game end', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('棋譜を見る')).not.toBeInTheDocument();
  });

  // -- PhaseIndicator coverage --

  it('phase indicator shows your turn when human discard turn', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows your turn in pegging phase', async () => {
    mockExec.mockResolvedValue(peggingPhaseState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows waiting when cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('待機中'));
  });

  // -- Keyboard navigation --

  it('pressing number key toggles card in discard phase', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('Enter key triggers discard in discard phase', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.keyDown(document, { key: '1' });
    fireEvent.keyDown(document, { key: '2' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(peggingPhaseState);

    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', undefined, [0, 1]));
  });

  it('Escape key clears selection', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('keyboard nav disabled when not human turn', async () => {
    const cpuDiscardState: CribbageResponse = { ...discardPhaseState, currentPlayerIdx: 1 };
    mockExec.mockResolvedValue(cpuDiscardState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getAllByText('CPU 1').length).toBeGreaterThanOrEqual(1));

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('reset hides the action log panel if open', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.cribbage).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());

    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));

    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('marks a peg card aria-disabled when its value would push the count over 31', async () => {
    const nearLimitState: CribbageResponse = { ...peggingPhaseState, pegCount: 28 };
    mockExec.mockResolvedValue(nearLimitState);
    renderWithProviders(<CribbagePage />);
    // Hand is [♠A, ♥J→10, ♦5, ♣8]. With pegCount=28: A=1→29 (OK); J=10→38, 5→33, 8→36 all exceed 31.
    // Verify the Jack of Hearts is the restricted one we test against.
    const restrictedJack = await screen.findByTestId('cb-card-restricted-HEART-11');
    // Use aria-disabled (not native `disabled`) so the card stays focusable for the tooltip.
    expect(restrictedJack).toHaveAttribute('aria-disabled', 'true');
    expect(restrictedJack).not.toBeDisabled();
  });

  it('highlights the Go button when the human has no playable card', async () => {
    const allRestrictedState: CribbageResponse = {
      ...peggingPhaseState,
      pegCount: 30,
      players: [
        {
          ...peggingPhaseState.players[0],
          cardCount: 2,
          cards: [
            { design: 'SPADE', value: 5 },
            { design: 'HEART', value: 7 },
          ],
        },
        peggingPhaseState.players[1],
      ],
    };
    mockExec.mockResolvedValue(allRestrictedState);
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByTestId('cb-go-button')).toBeInTheDocument());
    expect(screen.getByTestId('cb-go-button')).toHaveClass('animate-pulse');
  });

  it('shows the hover preview when hovering an eligible peg card', async () => {
    mockExec.mockResolvedValue({ ...peggingPhaseState, pegCount: 4 });
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByLabelText('♥ J')).toBeInTheDocument());
    fireEvent.pointerEnter(screen.getByLabelText('♥ J'));
    const preview = await screen.findByTestId('cb-peg-hover-preview');
    expect(preview.textContent).toContain('4');
    expect(preview.textContent).toContain('14');
  });

  // The preview text already updated on hover before this change, so asserting
  // that it changes proves nothing. What was missing is the region that makes
  // the update reach a screen reader at all.
  it('announces the peg preview through a permanently mounted live region', async () => {
    mockExec.mockResolvedValue({ ...peggingPhaseState, pegCount: 4 });
    renderWithProviders(<CribbagePage />);
    await waitFor(() => expect(screen.getByLabelText('♥ J')).toBeInTheDocument());

    const region = screen.getByTestId('cb-peg-preview-live');
    expect(region).toHaveAttribute('aria-live', 'polite');
    expect(region).toHaveAttribute('role', 'status');
    // Mounted and empty before the hover: a region appearing together with its
    // text is not a change and may go unread (#5955).
    expect(region).toHaveTextContent('');

    fireEvent.pointerEnter(screen.getByLabelText('♥ J'));
    await waitFor(() => expect(region).toHaveTextContent('14'));
  });
});
