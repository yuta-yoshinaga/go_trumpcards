import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, conquianApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { ConquianResponse } from '../types/card';
import { ConquianPage } from './ConquianPage';

vi.mock('../api/gameApi', () => ({
  conquianApi: { exec: vi.fn() },
  actionLogApi: { conquian: vi.fn() },
}));

const mockExec = vi.mocked(conquianApi.exec);

const drawPhaseState: ConquianResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 10,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      melds: [],
      wins: 0,
    },
    { id: 1, isHuman: false, cardCount: 10, cards: [], melds: [], wins: 1 },
  ],
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 28,
  gameEndFlag: false,
  winnerIdx: -1,
  roundWinnerIdx: -1,
  tookDiscard: false,
  message: '',
  messageCode: '',
  config: { cpuDifficulty: 1, targetWins: 3 },
};

const meldPhaseState: ConquianResponse = {
  ...drawPhaseState,
  phase: 1,
};

const meldPhaseTookDiscard: ConquianResponse = {
  ...meldPhaseState,
  tookDiscard: true,
};

// Meld phase with a 3-card hand so all three can be selected to form a new meld.
const meldPhaseThreeCards: ConquianResponse = {
  ...meldPhaseState,
  players: [
    {
      ...drawPhaseState.players[0],
      cardCount: 3,
      cards: [
        { design: 'SPADE', value: 5 },
        { design: 'HEART', value: 5 },
        { design: 'CLOVER', value: 5 },
      ],
    },
    { id: 1, isHuman: false, cardCount: 10, cards: [], melds: [], wins: 1 },
  ],
};

const roundEndState: ConquianResponse = {
  ...drawPhaseState,
  phase: 2,
};

const gameEndState: ConquianResponse = {
  ...drawPhaseState,
  phase: 3,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const cpuTurnState: ConquianResponse = {
  ...drawPhaseState,
  currentPlayerIdx: 1,
};

const noDiscardState: ConquianResponse = {
  ...drawPhaseState,
  discardTop: null,
};

const meldsDisplayState: ConquianResponse = {
  ...roundEndState,
  players: [
    {
      ...drawPhaseState.players[0],
      melds: [
        {
          cards: [
            { design: 'SPADE', value: 5 },
            { design: 'HEART', value: 5 },
            { design: 'CLOVER', value: 5 },
          ],
        },
      ],
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 3,
      cards: [{ design: 'DIAMOND', value: 5 }],
      melds: [],
      wins: 1,
    },
  ],
};

// Human has laid a 3-card set and a 6-card run (9 of 11 melded, 2 remaining ->
// the "close to winning" highlight should appear).
const meldProgressState: ConquianResponse = {
  ...meldPhaseState,
  players: [
    {
      ...drawPhaseState.players[0],
      melds: [
        {
          cards: [
            { design: 'SPADE', value: 5 },
            { design: 'HEART', value: 5 },
            { design: 'CLOVER', value: 5 },
          ],
        },
        {
          cards: [
            { design: 'DIAMOND', value: 3 },
            { design: 'DIAMOND', value: 4 },
            { design: 'DIAMOND', value: 5 },
            { design: 'DIAMOND', value: 6 },
            { design: 'DIAMOND', value: 7 },
            { design: 'DIAMOND', value: 8 },
          ],
        },
      ],
    },
    { id: 1, isHuman: false, cardCount: 10, cards: [], melds: [], wins: 1 },
  ],
};

beforeEach(() => {
  localStorage.clear();
  mockExec.mockResolvedValue(drawPhaseState);
});

describe('ConquianPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<ConquianPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<ConquianPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 1,
        targetWins: 3,
      }),
    );
  });

  it('renders draw phase with human cards and draw buttons', async () => {
    renderWithProviders(<ConquianPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeInTheDocument();
    });
  });

  it('draw discard button disabled when no discard top', async () => {
    mockExec.mockResolvedValue(noDiscardState);
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeDisabled());
  });

  it('calls drawstock when draw stock button clicked', async () => {
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('calls drawdiscard when draw discard button clicked', async () => {
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseTookDiscard);
    fireEvent.click(screen.getByRole('button', { name: '捨て札から引く' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('keyboard: "s" draws from the stock and "d" takes the discard', async () => {
    const { unmount } = renderWithProviders(<ConquianPage />);
    await screen.findByRole('button', { name: '山札から引く' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseState);
    fireEvent.keyDown(document.body, { key: 's' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
    unmount();

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    renderWithProviders(<ConquianPage />);
    await screen.findByRole('button', { name: '捨て札から引く' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseTookDiscard);
    fireEvent.keyDown(document.body, { key: 'd' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('keyboard: "d" does nothing when there is no discard top', async () => {
    mockExec.mockResolvedValue(noDiscardState);
    renderWithProviders(<ConquianPage />);
    await screen.findByRole('button', { name: '捨て札から引く' });
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'd' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('drawdiscard');
  });

  it('keyboard: draw keys are disabled on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, expect.anything()));
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 's' });
    fireEvent.keyDown(document.body, { key: 'd' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('drawstock');
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('drawdiscard');
  });

  it('advertises the draw keyboard shortcuts on the buttons', async () => {
    renderWithProviders(<ConquianPage />);
    const stockBtn = await screen.findByRole('button', { name: '山札から引く' });
    expect(stockBtn).toHaveAttribute('aria-keyshortcuts', 's');
    expect(stockBtn.querySelector('kbd')?.textContent).toBe('S');
    expect(screen.getByRole('button', { name: '捨て札から引く' })).toHaveAttribute('aria-keyshortcuts', 'd');
  });

  it('renders meld, layoff and discard buttons during meld phase', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<ConquianPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'メルドする' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'レイオフ' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument();
    });
  });

  it('meld and layoff buttons are both disabled with no selection', async () => {
    mockExec.mockResolvedValue(meldPhaseThreeCards);
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByTestId('conquian-meld-button')).toBeInTheDocument());
    expect(screen.getByTestId('conquian-meld-button')).toBeDisabled();
    expect(screen.getByTestId('conquian-layoff-button')).toBeDisabled();
  });

  it('layoff enabled and meld disabled when exactly 1 card selected; dispatches meld', async () => {
    mockExec.mockResolvedValue(meldPhaseThreeCards);
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 5')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ 5').closest('button') as HTMLButtonElement);
    expect(screen.getByTestId('conquian-layoff-button')).not.toBeDisabled();
    expect(screen.getByTestId('conquian-meld-button')).toBeDisabled();

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByTestId('conquian-layoff-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('meld', undefined, undefined, [[0]]));
  });

  it('both meld and layoff disabled when exactly 2 cards selected', async () => {
    mockExec.mockResolvedValue(meldPhaseThreeCards);
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 5')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ 5').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♥ 5').closest('button') as HTMLButtonElement);
    expect(screen.getByTestId('conquian-meld-button')).toBeDisabled();
    expect(screen.getByTestId('conquian-layoff-button')).toBeDisabled();
  });

  it('meld enabled and layoff disabled when 3 cards selected; dispatches meld', async () => {
    mockExec.mockResolvedValue(meldPhaseThreeCards);
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 5')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ 5').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♥ 5').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♣ 5').closest('button') as HTMLButtonElement);
    expect(screen.getByTestId('conquian-meld-button')).not.toBeDisabled();
    expect(screen.getByTestId('conquian-layoff-button')).toBeDisabled();

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByTestId('conquian-meld-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('meld', undefined, undefined, [[0, 1, 2]]));
  });

  it('discard button disabled when not exactly 1 card selected', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨てる' })).toBeDisabled());
  });

  it('calls discard command when discard button clicked', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '捨てる' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 0));
  });

  it('shows forced-use hint when discard was taken', async () => {
    mockExec.mockResolvedValue(meldPhaseTookDiscard);
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByTestId('conquian-forced-use')).toBeInTheDocument());
  });

  it('does not show forced-use hint when discard was not taken', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'メルドする' })).toBeInTheDocument());
    expect(screen.queryByTestId('conquian-forced-use')).not.toBeInTheDocument();
  });

  it('renders table melds and reveals CPU cards at round end', async () => {
    mockExec.mockResolvedValue(meldsDisplayState);
    renderWithProviders(<ConquianPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♦ 5')).toBeInTheDocument();
    });
    expect(screen.getAllByText(/テーブルメルド/).length).toBeGreaterThan(0);
  });

  it('shows next round button on round end and calls nextround', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('does not show draw buttons when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByText('勝利数')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '山札から引く' })).not.toBeInTheDocument();
  });

  it('wins table shows all players', async () => {
    renderWithProviders(<ConquianPage />);
    await waitFor(() => {
      expect(screen.getByText('勝利数')).toBeInTheDocument();
      expect(screen.getByText('あなた')).toBeInTheDocument();
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
    });
  });

  it('shows discard top card', async () => {
    renderWithProviders(<ConquianPage />);
    await waitFor(() => {
      expect(screen.getByText('捨て札')).toBeInTheDocument();
      expect(screen.getByAltText('♥ 7')).toBeInTheDocument();
    });
  });

  it('round info displayed', async () => {
    renderWithProviders(<ConquianPage />);
    await waitFor(() => {
      expect(screen.getByText('ラウンド 1')).toBeInTheDocument();
      expect(screen.getByText('山札: 28枚')).toBeInTheDocument();
    });
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<ConquianPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows error alert', async () => {
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('reset button calls exec with confirm', async () => {
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 1,
        targetWins: 3,
      }),
    );
  });

  it('settings panel changes targetWins', async () => {
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByText('勝利数')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: '5' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 1,
        targetWins: 5,
      }),
    );
  });

  it('card selection toggle via aria-pressed', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');
  });

  it('shows meld progress toward 11 for the human and CPU', async () => {
    renderWithProviders(<ConquianPage />);
    const humanProgress = await screen.findByTestId('conquian-meld-progress');
    // Fresh game: nobody has melded yet.
    expect(humanProgress).toHaveTextContent('メルド 0/11');
    expect(screen.getByTestId('conquian-meld-progress-cpu-1')).toHaveTextContent('メルド 0/11');
  });

  it('counts melded cards and highlights the final stretch', async () => {
    mockExec.mockResolvedValue(meldProgressState);
    renderWithProviders(<ConquianPage />);
    const humanProgress = await screen.findByTestId('conquian-meld-progress');
    // 3 + 6 = 9 melded of 11.
    expect(humanProgress).toHaveTextContent('メルド 9/11');
    expect(humanProgress).toHaveTextContent('あと 2 枚');
    const bar = humanProgress.querySelector('[role="progressbar"]');
    expect(bar).toHaveAttribute('aria-valuenow', '9');
    expect(bar).toHaveAttribute('aria-valuemax', '11');
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<ConquianPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.conquian).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.conquian).toHaveBeenCalledTimes(1));
  });
});
