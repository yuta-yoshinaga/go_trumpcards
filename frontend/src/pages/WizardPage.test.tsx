import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, wizardApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { WizardResponse } from '../types/card';
import { WizardPage } from './WizardPage';

vi.mock('../api/gameApi', () => ({
  wizardApi: { exec: vi.fn() },
  actionLogApi: { wizard: vi.fn() },
}));

const mockExec = vi.mocked(wizardApi.exec);

const playPhaseState: WizardResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      bid: 2,
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 5,
      cards: [],
      bid: 1,
      roundScore: 3,
      cumulativeScore: 10,
      trickCount: 1,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 5,
      cards: [],
      bid: 2,
      roundScore: 5,
      cumulativeScore: 20,
      trickCount: 2,
    },
    {
      id: 3,
      isHuman: false,
      cardCount: 5,
      cards: [],
      bid: 0,
      roundScore: 0,
      cumulativeScore: 5,
      trickCount: 0,
    },
  ],
  phase: 1,
  roundNumber: 3,
  totalRounds: 15,
  handSize: 5,
  trickNumber: 1,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  dealerIdx: 3,
  currentTrick: [],
  trumpCard: { design: 'HEART', value: 7 },
  trumpSuit: 3,
  restrictedBid: -1,
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: 0,
  message: '',
  config: { cpuDifficulty: 1 },
};

const bidPhaseState: WizardResponse = {
  ...playPhaseState,
  phase: 0,
  bidPlayerIdx: 0,
  restrictedBid: -1,
  players: playPhaseState.players.map((p) => ({ ...p, bid: -1 })),
};

const bidPhaseCpuTurnState: WizardResponse = {
  ...bidPhaseState,
  bidPlayerIdx: 1,
};

const trickEndState: WizardResponse = {
  ...playPhaseState,
  phase: 2,
  currentTrick: [
    { playerIdx: 0, card: { design: 'DIAMOND', value: 3 } },
    { playerIdx: 1, card: { design: 'HEART', value: 5 } },
  ],
};

const roundEndState: WizardResponse = {
  ...playPhaseState,
  phase: 3,
};

const gameEndState: WizardResponse = {
  ...playPhaseState,
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const gameEndByFlagState: WizardResponse = {
  ...playPhaseState,
  phase: 1,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const cpuTurnState: WizardResponse = {
  ...playPhaseState,
  currentPlayerIdx: 1,
};

/** Round end where the human made their bid exactly and CPUs over/undershoot. */
const bidAccuracyRoundEndState: WizardResponse = {
  ...playPhaseState,
  phase: 3,
  players: [
    { ...playPhaseState.players[0], bid: 3, trickCount: 3 },
    { ...playPhaseState.players[1], bid: 2, trickCount: 4 },
    { ...playPhaseState.players[2], bid: 2, trickCount: 1 },
    { ...playPhaseState.players[3], bid: 1, trickCount: 2 },
  ],
};

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

/** Human following a HEART lead while holding HEART, SPADE, and a Wizard card. */
const followSuitState: WizardResponse = {
  ...playPhaseState,
  currentPlayerIdx: 0,
  leadPlayerIdx: 1,
  currentTrick: [{ playerIdx: 1, card: { design: 'HEART', value: 5 } }],
  players: [
    {
      ...playPhaseState.players[0],
      cardCount: 3,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
        { design: 'JOKER', value: 1, label: 'Wizard', glyph: '✦', deck: 'wizard' },
      ],
    },
    ...playPhaseState.players.slice(1),
  ],
};

describe('WizardPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<WizardPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
      }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  it('shows the bid-progress chip in the header during play', async () => {
    renderWithProviders(<WizardPage />);
    const chip = await screen.findByTestId('bid-progress-chip');
    expect(chip).toHaveTextContent('宣言: 2 / 獲得: 0');
    expect(chip.className).toContain('border-ds-border-subtle');
  });

  it('colors the progress chip green when tricks match the bid', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      players: [{ ...playPhaseState.players[0], trickCount: 2 }, ...playPhaseState.players.slice(1)],
    });
    renderWithProviders(<WizardPage />);
    const chip = await screen.findByTestId('bid-progress-chip');
    expect(chip).toHaveTextContent('宣言: 2 / 獲得: 2');
    expect(chip.className).toContain('border-ds-success');
  });

  it('colors the progress chip yellow when tricks exceed the bid', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      players: [{ ...playPhaseState.players[0], trickCount: 3 }, ...playPhaseState.players.slice(1)],
    });
    renderWithProviders(<WizardPage />);
    const chip = await screen.findByTestId('bid-progress-chip');
    expect(chip.className).toContain('border-ds-warning');
  });

  it('colors the progress chip red when the bid is no longer reachable', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      players: [{ ...playPhaseState.players[0], cardCount: 1 }, ...playPhaseState.players.slice(1)],
    });
    renderWithProviders(<WizardPage />);
    const chip = await screen.findByTestId('bid-progress-chip');
    expect(chip.className).toContain('border-ds-error');
  });

  it('does not turn red while the human card in the unresolved trick can still win', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      currentTrick: [{ playerIdx: 0, card: { design: 'SPADE', value: 9 } }],
      players: [{ ...playPhaseState.players[0], cardCount: 1 }, ...playPhaseState.players.slice(1)],
    });
    renderWithProviders(<WizardPage />);
    const chip = await screen.findByTestId('bid-progress-chip');
    expect(chip.className).toContain('border-ds-border-subtle');
    expect(chip.className).not.toContain('border-ds-error');
  });

  it('hides the progress chip during the bid phase and at game end', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    let view = renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    expect(screen.queryByTestId('bid-progress-chip')).not.toBeInTheDocument();
    view.unmount();

    mockExec.mockResolvedValue(gameEndState);
    view = renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    expect(screen.queryByTestId('bid-progress-chip')).not.toBeInTheDocument();
    view.unmount();

    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    expect(screen.queryByTestId('bid-progress-chip')).not.toBeInTheDocument();
  });

  it('renders bid phase as a button group of bid choices', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => {
      // handSize = 5 → buttons 0..5
      for (let i = 0; i <= 5; i++) {
        expect(screen.getByRole('button', { name: `ビッド ${i}` })).toBeInTheDocument();
      }
      expect(screen.queryByLabelText('bid-input')).not.toBeInTheDocument();
    });
  });

  it('shows the bid-total chip during the bid phase (under when nothing bid yet)', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<WizardPage />);
    const chip = await screen.findByTestId('bid-total-chip');
    expect(chip).toHaveTextContent('総ビッド 0 / 手札 5');
    expect(chip).toHaveTextContent('(アンダー)');
  });

  it('colors the bid-total chip for over and exact tables', async () => {
    mockExec.mockResolvedValue({
      ...bidPhaseState,
      players: bidPhaseState.players.map((p, i) => ({ ...p, bid: i < 3 ? [3, 2, 2, -1][i] : -1 })),
    });
    const { unmount } = renderWithProviders(<WizardPage />);
    expect(await screen.findByTestId('bid-total-chip')).toHaveTextContent('(オーバー)');
    unmount();

    mockExec.mockResolvedValue({
      ...bidPhaseState,
      players: bidPhaseState.players.map((p, i) => ({ ...p, bid: [2, 2, 1, -1][i] })),
    });
    renderWithProviders(<WizardPage />);
    expect(await screen.findByTestId('bid-total-chip')).toHaveTextContent('(ぴったり)');
  });

  it('shows bid phase instruction when human bid turn', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => {
      expect(screen.getByText('ビッド宣言 (0-5)')).toBeInTheDocument();
    });
  });

  it('does not show bid instruction when cpu bid turn', async () => {
    mockExec.mockResolvedValue(bidPhaseCpuTurnState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText(/ビッド宣言/)).not.toBeInTheDocument();
  });

  it('calls bid command with the clicked button value', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ビッド 3' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ビッド 3' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 3));
  });

  it('play button disabled when not 1 card selected', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeDisabled());
  });

  it('play button enabled when 1 card selected', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '出す' })).not.toBeDisabled();
  });

  it('does not show play button when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('shows next trick button on trick end', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('shows next round button on round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('shows bid-accuracy summary at round end with made/over/under outcomes', async () => {
    mockExec.mockResolvedValue(bidAccuracyRoundEndState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByTestId('wiz-bid-accuracy')).toBeInTheDocument());
    // Human bid 3, took 3 -> exact hit.
    expect(screen.getByText('的中')).toBeInTheDocument();
    // CPU 1 bid 2, took 4 -> +2 overshoot delta.
    expect(screen.getByText('+2 超過')).toBeInTheDocument();
    // CPU 2 bid 2, took 1 -> -1 undershoot delta.
    expect(screen.getByText('-1 不足')).toBeInTheDocument();
  });

  it('does not show the bid-accuracy summary during play', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    expect(screen.queryByTestId('wiz-bid-accuracy')).not.toBeInTheDocument();
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-4 phase', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows error alert', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('settings panel changes cpuDifficulty', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '2' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 2,
      }),
    );
  });

  it('card selection toggle via aria-pressed', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('reset button calls exec', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
      }),
    );
  });

  it('score table shows all players', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() => {
      expect(screen.getByText('スコア')).toBeInTheDocument();
      expect(screen.getByText('あなた')).toBeInTheDocument();
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
      expect(screen.getByText('CPU 2')).toBeInTheDocument();
      expect(screen.getByText('CPU 3')).toBeInTheDocument();
    });
  });

  it('score table headers have scope="col" for accessibility', async () => {
    const { container } = renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByText('あなた')).toBeInTheDocument());
    const ths = container.querySelectorAll('th');
    ths.forEach((th) => {
      expect(th).toHaveAttribute('scope', 'col');
    });
  });

  it('shows current trick cards', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => {
      expect(screen.getByText('現在のトリック')).toBeInTheDocument();
      expect(screen.getByAltText('♦ 3')).toBeInTheDocument();
      expect(screen.getByAltText('♥ 5')).toBeInTheDocument();
    });
  });

  it('does not show current trick when empty', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('現在のトリック')).not.toBeInTheDocument();
  });

  it('shows CPU player areas', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*5枚/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 2.*5枚/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 3.*5枚/)).toBeInTheDocument();
    });
  });

  it('calls play command when play button is clicked', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('calls next when next trick button is clicked', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のトリック' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('calls nextround when next round button is clicked', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('round and trick info displayed', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() => {
      expect(screen.getByText('ラウンド 3/15')).toBeInTheDocument();
      expect(screen.getByText('トリック 1')).toBeInTheDocument();
    });
  });

  it('trump and dealer info displayed', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() => {
      expect(screen.getByText('切り札: ハート')).toBeInTheDocument();
      expect(screen.getByText('ディーラー: CPU 3')).toBeInTheDocument();
    });
  });

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.wizard).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.wizard).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();

    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText(/^棋譜$/)).not.toBeInTheDocument());
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('score table shows dash for bid < 0', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    const rows = screen.getAllByRole('row');
    expect(rows.length).toBe(5);
  });

  it('shows bid value for player with bid >= 0', async () => {
    renderWithProviders(<WizardPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*ビッド 1/)).toBeInTheDocument();
    });
  });

  it('shows unbid text for player with bid < 0', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<WizardPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*未ビッド/)).toBeInTheDocument();
    });
  });

  it('renders CPU info as collapsible details on mobile', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      const { container } = renderWithProviders(<WizardPage />);
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
      const { container } = renderWithProviders(<WizardPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const scoreDetails = container.querySelector('details[data-tutorial="wiz-score-table"]');
      expect(scoreDetails).toBeInTheDocument();
      const summary = scoreDetails?.querySelector('summary');
      expect(summary).toHaveTextContent('スコア');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  describe('legal-play highlighting', () => {
    it('rings every card when leading (empty trick)', async () => {
      mockExec.mockResolvedValue(playPhaseState);
      renderWithProviders(<WizardPage />);
      const spade = await screen.findByRole('button', { name: '♠ A' });
      const heart = screen.getByRole('button', { name: '♥ J' });
      expect(spade.className).toContain('ring-ds-success');
      expect(heart.className).toContain('ring-ds-success');
      expect(spade).not.toHaveAttribute('title');
    });

    it('rings only led-suit and Wizard/Jester cards when following, dimming the rest', async () => {
      mockExec.mockResolvedValue(followSuitState);
      renderWithProviders(<WizardPage />);
      const heart = await screen.findByRole('button', { name: '♥ J' });
      const wizard = screen.getByRole('button', { name: 'Wizard ✦' });
      const spade = screen.getByRole('button', { name: '♠ A' });
      // Must follow HEART: the held HEART and the always-legal Wizard are ringed.
      expect(heart.className).toContain('ring-ds-success');
      expect(wizard.className).toContain('ring-ds-success');
      // The off-suit SPADE is illegal while a HEART is held: dimmed with a reason tooltip.
      expect(spade.className).toContain('opacity-50');
      expect(spade.className).not.toContain('ring-ds-success');
      expect(spade).toHaveAttribute('title');
      expect(spade).toHaveAttribute('aria-describedby', 'wiz-illegal-reason');
    });

    it('does not highlight the hand off the human play turn', async () => {
      mockExec.mockResolvedValue({ ...playPhaseState, currentPlayerIdx: 1 });
      renderWithProviders(<WizardPage />);
      const spade = await screen.findByRole('button', { name: '♠ A' });
      expect(spade.className).not.toContain('ring-ds-success');
      expect(spade.className).not.toContain('opacity-50');
    });
  });
});
