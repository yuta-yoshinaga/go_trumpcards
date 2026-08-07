import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, crazyeightsApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CrazyEightsResponse } from '../types/card';
import { CrazyEightsPage } from './CrazyEightsPage';

vi.mock('../api/gameApi', () => ({
  crazyeightsApi: { exec: vi.fn() },
  actionLogApi: { crazyeights: vi.fn() },
}));

const mockExec = vi.mocked(crazyeightsApi.exec);

const playPhaseState: CrazyEightsResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], roundScore: 3, cumulativeScore: 10 },
    { id: 2, isHuman: false, cardCount: 5, cards: [], roundScore: 5, cumulativeScore: 20 },
    { id: 3, isHuman: false, cardCount: 5, cards: [], roundScore: 0, cumulativeScore: 5 },
  ],
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 30,
  chosenSuit: 0,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 200 },
};

const chooseSuitState: CrazyEightsResponse = {
  ...playPhaseState,
  phase: 1,
};

const roundEndState: CrazyEightsResponse = {
  ...playPhaseState,
  phase: 2,
};

const gameEndState: CrazyEightsResponse = {
  ...playPhaseState,
  phase: 3,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const gameEndByFlagState: CrazyEightsResponse = {
  ...playPhaseState,
  phase: 0,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const cpuTurnState: CrazyEightsResponse = {
  ...playPhaseState,
  currentPlayerIdx: 1,
};

const chosenSuitState: CrazyEightsResponse = {
  ...playPhaseState,
  chosenSuit: 1,
};

const noDiscardState: CrazyEightsResponse = {
  ...playPhaseState,
  discardTop: null,
};

const unknownSuitState: CrazyEightsResponse = {
  ...playPhaseState,
  chosenSuit: 99,
};

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('CrazyEightsPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<CrazyEightsPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 200,
      }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => {
      expect(screen.getByAltText('\u2660 A')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 J')).toBeInTheDocument();
    });
  });

  it('highlights legal cards and dims illegal ones on the human turn', async () => {
    // Discard top ♥7; hand has ♥J (legal — same suit) and ♠A (illegal).
    renderWithProviders(<CrazyEightsPage />);
    const legal = await screen.findByLabelText('♥ J');
    const illegal = screen.getByLabelText('♠ A');
    expect(legal).toHaveAttribute('data-legal', 'true');
    expect(legal.className).toContain('ring-ds-success');
    expect(illegal).toHaveAttribute('data-legal', 'false');
    expect(illegal.className).toContain('opacity-50');
    expect(illegal).toHaveAttribute('title', 'このカードは出せません（スート・数字・8 のいずれも不一致）');
  });

  it('does not mark card legality on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<CrazyEightsPage />);
    const card = await screen.findByLabelText('♥ J');
    expect(card).not.toHaveAttribute('data-legal');
  });

  it('renders play and draw buttons when human turn', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument();
    });
  });

  it('play button disabled when not 1 card selected', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeDisabled());
  });

  it('play button enabled when 1 card selected', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '出す' })).not.toBeDisabled();
  });

  it('calls play command when play button is clicked', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('calls draw command when draw button is clicked', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '引く' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('does not show play/draw buttons when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '引く' })).not.toBeInTheDocument();
  });

  it('renders choose suit phase with 4 suit buttons', async () => {
    mockExec.mockResolvedValue(chooseSuitState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '♠ スペード' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '♣ クローバー' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '♥ ハート' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '♦ ダイヤ' })).toBeInTheDocument();
    });
  });

  it('calls suit command when suit button is clicked', async () => {
    mockExec.mockResolvedValue(chooseSuitState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ スペード' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '♠ スペード' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('suit', undefined, 1));
  });

  it('calls suit command for clover', async () => {
    mockExec.mockResolvedValue(chooseSuitState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♣ クローバー' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '♣ クローバー' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('suit', undefined, 2));
  });

  it('calls suit command for heart', async () => {
    mockExec.mockResolvedValue(chooseSuitState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♥ ハート' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '♥ ハート' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('suit', undefined, 3));
  });

  it('calls suit command for diamond', async () => {
    mockExec.mockResolvedValue(chooseSuitState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♦ ダイヤ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '♦ ダイヤ' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('suit', undefined, 4));
  });

  it('shows next round button on round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('calls nextround when next round button is clicked', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-3 phase', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows error alert', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('shows CPU player areas', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*5枚/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 2.*5枚/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 3.*5枚/)).toBeInTheDocument();
    });
  });

  it('score table shows all players', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => {
      expect(screen.getByText('スコア')).toBeInTheDocument();
      expect(screen.getByText('あなた')).toBeInTheDocument();
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
      expect(screen.getByText('CPU 2')).toBeInTheDocument();
      expect(screen.getByText('CPU 3')).toBeInTheDocument();
    });
  });

  it('score table headers have scope="col" for accessibility', async () => {
    const { container } = renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByText('あなた')).toBeInTheDocument());
    const ths = container.querySelectorAll('th');
    ths.forEach((th) => {
      expect(th).toHaveAttribute('scope', 'col');
    });
  });

  it('shows discard top card', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => {
      expect(screen.getByText('捨て札')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 7')).toBeInTheDocument();
    });
  });

  it('does not show discard top when null', async () => {
    mockExec.mockResolvedValue(noDiscardState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('捨て札')).not.toBeInTheDocument();
  });

  it('shows chosen suit when chosenSuit > 0', async () => {
    mockExec.mockResolvedValue(chosenSuitState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => {
      expect(screen.getByText(/指定スート/)).toBeInTheDocument();
      expect(screen.getByText(/指定スート: ♠/)).toBeInTheDocument();
    });
  });

  it('does not show chosen suit when chosenSuit is 0', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    expect(screen.queryByText('指定スート')).not.toBeInTheDocument();
  });

  it('renders large suit watermark behind discard pile when chosenSuit > 0', async () => {
    mockExec.mockResolvedValue(chosenSuitState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => {
      const watermark = screen.getByTestId('chosen-suit-watermark');
      expect(watermark).toBeInTheDocument();
      expect(watermark.textContent).toBe('♠');
      // motion-safe: prefix is load-bearing — it pins the keyframe to users who haven't
      // enabled prefers-reduced-motion. Substring on 'animate-suit-watermark' alone would
      // silently pass even if the prefix were dropped.
      expect(watermark.className).toContain('motion-safe:animate-suit-watermark');
      expect(watermark.className).toContain('opacity-15');
    });
  });

  it('does not render watermark when chosenSuit is 0', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    expect(screen.queryByTestId('chosen-suit-watermark')).not.toBeInTheDocument();
  });

  it('shows fallback ? for unknown suit value', async () => {
    mockExec.mockResolvedValue(unknownSuitState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => {
      expect(screen.getByText(/指定スート: \?/)).toBeInTheDocument();
    });
  });

  it('announces the chosen suit in a polite live region when chosenSuit > 0', async () => {
    mockExec.mockResolvedValue(chosenSuitState); // chosenSuit: 1 → spade
    renderWithProviders(<CrazyEightsPage />);
    const region = await screen.findByTestId('ce-suit-live-region');
    expect(region).toHaveAttribute('role', 'status');
    expect(region).toHaveAttribute('aria-live', 'polite');
    expect(region).toHaveTextContent('スートが スペード に変更されました');
  });

  it('keeps the suit live region empty when no suit is chosen', async () => {
    renderWithProviders(<CrazyEightsPage />);
    const region = await screen.findByTestId('ce-suit-live-region');
    expect(region).toHaveTextContent('');
  });

  it('keeps the suit live region empty for an unknown suit value', async () => {
    mockExec.mockResolvedValue(unknownSuitState); // chosenSuit: 99
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByText(/指定スート: \?/)).toBeInTheDocument());
    expect(screen.getByTestId('ce-suit-live-region')).toHaveTextContent('');
  });

  it('associates the illegal-play reason with illegal cards via aria-describedby', async () => {
    // Discard top ♥7; ♠A is illegal on the human turn.
    renderWithProviders(<CrazyEightsPage />);
    const illegal = await screen.findByLabelText('♠ A');
    expect(illegal).toHaveAttribute('aria-describedby', 'ce-illegal-reason');
    const reason = document.getElementById('ce-illegal-reason');
    expect(reason).toHaveTextContent('このカードは出せません（スート・数字・8 のいずれも不一致）');
  });

  it('does not set aria-describedby on legal cards', async () => {
    renderWithProviders(<CrazyEightsPage />);
    const legal = await screen.findByLabelText('♥ J');
    expect(legal).not.toHaveAttribute('aria-describedby');
  });

  it('does not set aria-describedby on cards during a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<CrazyEightsPage />);
    const card = await screen.findByLabelText('♠ A');
    expect(card).not.toHaveAttribute('aria-describedby');
  });

  it('card selection toggle via aria-pressed', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('card buttons have aria-label with card name', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-label', '\u2660 A');

    const cardBtn2 = screen.getByAltText('\u2665 J').closest('button') as HTMLButtonElement;
    expect(cardBtn2).toHaveAttribute('aria-label', '\u2665 J');
  });

  it('reset button calls exec with confirm', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 200,
      }),
    );
  });

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('settings panel changes cpuDifficulty', async () => {
    renderWithProviders(<CrazyEightsPage />);
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
        pointLimit: 200,
      }),
    );
  });

  it('settings panel changes pointLimit', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: '300' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 300,
      }),
    );
  });

  it('round info displayed', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => {
      expect(screen.getByText('ラウンド 1')).toBeInTheDocument();
      expect(screen.getByText('山札: 30枚')).toBeInTheDocument();
    });
  });

  it('does not show message when empty', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('Game end!')).not.toBeInTheDocument();
  });

  it('shows loading state', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    let resolve!: (value: CrazyEightsResponse) => void;
    const slow = new Promise<CrazyEightsResponse>((r) => {
      resolve = r;
    });
    mockExec.mockReturnValueOnce(slow);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(playPhaseState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
  });

  it('sets aria-busy on container', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: 'リセット' }).closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');
  });

  it('no human cards renders empty hand area', async () => {
    const noHuman: CrazyEightsResponse = {
      ...playPhaseState,
      players: playPhaseState.players.map((p) => ({ ...p, isHuman: false })),
    };
    mockExec.mockResolvedValue(noHuman);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByAltText('\u2660 A')).not.toBeInTheDocument();
  });

  it('isHumanTurn false when currentPlayerIdx points to cpu', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.crazyeights).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.crazyeights).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();

    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText(/^棋譜$/)).not.toBeInTheDocument());
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });

  it('does not show action log button when not game end', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('棋譜を見る')).not.toBeInTheDocument();
  });

  it('disables buttons while loading', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    let resolve!: (value: CrazyEightsResponse) => void;
    const slow = new Promise<CrazyEightsResponse>((r) => {
      resolve = r;
    });
    mockExec.mockReturnValueOnce(slow);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(playPhaseState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
  });

  // -- PhaseIndicator coverage --

  it('phase indicator shows your turn when human play turn', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows your turn in choose suit phase', async () => {
    mockExec.mockResolvedValue(chooseSuitState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows waiting when cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('待機中'));
  });

  // -- Keyboard navigation --

  it('pressing number key toggles card in play phase', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('Enter key triggers play in play phase', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.keyDown(document, { key: '1' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);

    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('Escape key clears selection', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('keyboard nav disabled when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('reset hides the action log panel if open', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.crazyeights).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());

    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));

    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  // -- Hand sort toggle (display-only) --

  // Human hand where sorting visibly reorders: original order is [♥5, ♠3],
  // but suit-sorted display is [♠3, ♥5]. The ♠3 stays at original index 1.
  const sortableHandState: CrazyEightsResponse = {
    ...playPhaseState,
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 2,
        cards: [
          { design: 'HEART', value: 5 },
          { design: 'SPADE', value: 3 },
        ],
        roundScore: 0,
        cumulativeScore: 0,
      },
      ...playPhaseState.players.slice(1),
    ],
  };

  /** Returns the human hand's card aria-labels in current DOM (display) order. */
  const handOrder = (container: HTMLElement): string[] =>
    Array.from(container.querySelectorAll<HTMLElement>('[data-tutorial="ce-player-hand"] button[aria-label]')).map(
      (b) => b.getAttribute('aria-label') ?? '',
    );

  it('renders the three sort-mode toggles with original selected by default', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(sortableHandState);
    renderWithProviders(<CrazyEightsPage />);
    const original = await screen.findByTestId('ce-sort-original');
    expect(original).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('ce-sort-rank')).toHaveAttribute('aria-pressed', 'false');
    expect(screen.getByTestId('ce-sort-suit')).toHaveAttribute('aria-pressed', 'false');
  });

  it('suit sort reorders the displayed hand while original order is served', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(sortableHandState);
    const { container } = renderWithProviders(<CrazyEightsPage />);
    await screen.findByTestId('ce-sort-suit');
    // Original (served) order.
    expect(handOrder(container)).toEqual(['♥ 5', '♠ 3']);

    fireEvent.click(screen.getByTestId('ce-sort-suit'));
    // Display re-sorted ♠ before ♥.
    expect(handOrder(container)).toEqual(['♠ 3', '♥ 5']);
    expect(screen.getByTestId('ce-sort-suit')).toHaveAttribute('aria-pressed', 'true');
  });

  it('plays the correct ORIGINAL index after the display is sorted', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(sortableHandState);
    renderWithProviders(<CrazyEightsPage />);
    await screen.findByTestId('ce-sort-suit');

    // Sort by suit: ♠3 now renders first, but its original server index is 1.
    fireEvent.click(screen.getByTestId('ce-sort-suit'));
    fireEvent.click(screen.getByLabelText('♠ 3').closest('button') as HTMLButtonElement);

    mockExec.mockClear();
    mockExec.mockResolvedValue(sortableHandState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  it('keyboard digit selection follows the sorted display order', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(sortableHandState);
    renderWithProviders(<CrazyEightsPage />);
    await screen.findByTestId('ce-sort-suit');

    fireEvent.click(screen.getByTestId('ce-sort-suit'));
    // Digit "1" selects the visually first card (♠3 = original index 1).
    fireEvent.keyDown(document, { key: '1' });
    const spade = screen.getByLabelText('♠ 3').closest('button') as HTMLButtonElement;
    expect(spade).toHaveAttribute('aria-pressed', 'true');

    mockExec.mockClear();
    mockExec.mockResolvedValue(sortableHandState);
    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  it('persists the chosen sort mode to localStorage', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(sortableHandState);
    renderWithProviders(<CrazyEightsPage />);
    await screen.findByTestId('ce-sort-rank');

    fireEvent.click(screen.getByTestId('ce-sort-rank'));
    expect(localStorage.getItem('crazyeights-sort-mode')).toBe('rank');
  });

  it('does not render sort toggles when the human has no cards', async () => {
    localStorage.clear();
    const noCards: CrazyEightsResponse = {
      ...playPhaseState,
      players: [
        { id: 0, isHuman: true, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 0 },
        ...playPhaseState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(noCards);
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByTestId('ce-sort-original')).not.toBeInTheDocument();
  });

  // **Hearts / Spades はサーバー計算の理由付きヒントを返すのに、CrazyEights には
  // ヒントボタンすら無く、全ゲーム共通の簡易ヒューリスティックしか無かった (#4737)。**
  it('fetches a server hint and shows the recommended card with its reason', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByTestId('ce-hint-button')).toBeInTheDocument());

    mockExec.mockResolvedValueOnce({ ...playPhaseState, hint: { cardIndex: 1, reason: 'match_suit' } });
    fireEvent.click(screen.getByTestId('ce-hint-button'));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
    const shown = await screen.findByTestId('ce-server-hint');
    expect(shown).toHaveTextContent('1');
    expect(shown).toHaveTextContent('スートが合う');
  });

  it('shows the recommended suit during the choose-suit phase', async () => {
    renderWithProviders(<CrazyEightsPage />);
    await waitFor(() => expect(screen.getByTestId('ce-hint-button')).toBeInTheDocument());

    mockExec.mockResolvedValueOnce({ ...playPhaseState, hint: { suit: 3, reason: 'choose_longest_suit' } });
    fireEvent.click(screen.getByTestId('ce-hint-button'));

    const shown = await screen.findByTestId('ce-server-hint');
    expect(shown).toHaveTextContent('手札に一番多いスート');
  });

  // 要求する前は出さない。常時表示だとフロント完結のツールチップと二重になる。
  it('shows no server hint before the button is pressed', async () => {
    renderWithProviders(<CrazyEightsPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('ce-server-hint')).not.toBeInTheDocument();
  });
});
