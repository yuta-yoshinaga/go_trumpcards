import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, machiavelliApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { MachiavelliMeld, MachiavelliPlayer, MachiavelliResponse } from '../types/card';
import { MachiavelliPage } from './MachiavelliPage';

vi.mock('../api/gameApi', () => ({
  machiavelliApi: { exec: vi.fn() },
  actionLogApi: { machiavelli: vi.fn() },
}));

const mockExec = vi.mocked(machiavelliApi.exec);

function player(overrides: Partial<MachiavelliPlayer> = {}): MachiavelliPlayer {
  return {
    id: 0,
    isHuman: true,
    cardCount: 13,
    cards: [],
    roundScore: 0,
    cumulativeScore: 0,
    deadwood: 0,
    ...overrides,
  };
}

const tableMeld: MachiavelliMeld = {
  cards: [
    { design: 'SPADE', value: 3 },
    { design: 'SPADE', value: 4 },
    { design: 'SPADE', value: 5 },
  ],
  kind: 1,
};

const turnState: MachiavelliResponse = {
  players: [
    player({
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
        { design: 'CLOVER', value: 7 },
      ],
    }),
    player({ id: 1, isHuman: false, roundScore: 3, cumulativeScore: 10 }),
  ],
  table: [tableMeld],
  phase: 0,
  roundNumber: 1,
  targetRounds: 3,
  currentPlayerIdx: 0,
  dealerIdx: 0,
  drawPileCount: 40,
  gameEndFlag: false,
  winnerIdx: -1,
  roundWinnerIdx: -1,
  message: '',
  config: { playerCount: 4, cpuDifficulty: 1, targetRounds: 3 },
};

const roundEndState: MachiavelliResponse = { ...turnState, phase: 1 };
const gameEndState: MachiavelliResponse = {
  ...turnState,
  phase: 2,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};
const gameEndByFlagState: MachiavelliResponse = {
  ...turnState,
  phase: 0,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};
const cpuTurnState: MachiavelliResponse = { ...turnState, currentPlayerIdx: 1 };
const emptyTableState: MachiavelliResponse = { ...turnState, table: [] };
const roundEndCpuCardsState: MachiavelliResponse = {
  ...turnState,
  phase: 1,
  players: [
    turnState.players[0],
    player({
      id: 1,
      isHuman: false,
      cardCount: 2,
      cards: [
        { design: 'DIAMOND', value: 5 },
        { design: 'CLOVER', value: 6 },
      ],
      roundScore: 3,
      cumulativeScore: 10,
      deadwood: 11,
    }),
  ],
};

beforeEach(() => {
  mockExec.mockResolvedValue(turnState);
});

describe('MachiavelliPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MachiavelliPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with default config', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 4,
        cpuDifficulty: 1,
        targetRounds: 3,
      }),
    );
  });

  it('renders the human hand cards', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  it('renders draw and new-meld buttons on the human turn', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'メルドを出す' })).toBeInTheDocument();
    });
  });

  it('new-meld button disabled until 3 cards selected', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'メルドを出す' })).toBeDisabled());
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♥ J').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: 'メルドを出す' })).toBeDisabled();
    fireEvent.click(screen.getByAltText('♣ 7').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: 'メルドを出す' })).not.toBeDisabled();
  });

  it('calls draw when the draw button is clicked', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(turnState);
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('calls newmeld with the selected indices', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♥ J').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♣ 7').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(turnState);
    fireEvent.click(screen.getByRole('button', { name: 'メルドを出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('newmeld', { handIndices: [0, 1, 2] }));
  });

  it('lays off a selected card onto a table meld', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    // Layoff button disabled with no selection.
    expect(screen.getByRole('button', { name: /レイオフ/ })).toBeDisabled();
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: /レイオフ/ })).not.toBeDisabled();
    mockExec.mockClear();
    mockExec.mockResolvedValue(turnState);
    fireEvent.click(screen.getByRole('button', { name: /レイオフ/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('layoff', { meldIdx: 0, handIndex: 0 }));
  });

  it('labels each table meld and its layoff button for assistive tech', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());
    // The meld row is a labelled group describing kind + size + rank range.
    const meldGroup = screen.getByRole('group', { name: /3枚/ });
    expect(meldGroup).toBeInTheDocument();
    // The layoff button's accessible name includes the target meld description.
    const layoff = screen.getByRole('button', { name: /にレイオフ/ });
    expect(layoff.getAttribute('aria-label')).toMatch(/3枚/);
  });

  it('highlights table melds as drop targets when exactly one card is selected', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    const meldGroup = screen.getByRole('group', { name: /3枚/ });
    expect(meldGroup.className).not.toContain('ring-2');
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    await waitFor(() => expect(meldGroup.className).toContain('ring-2'));
  });

  it('shows the table-empty placeholder when there are no melds', async () => {
    mockExec.mockResolvedValue(emptyTableState);
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByText('まだメルドはありません')).toBeInTheDocument());
  });

  it('does not show action buttons when it is not the human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '山札から引く' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'メルドを出す' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /レイオフ/ })).not.toBeInTheDocument();
  });

  it('shows next round button on round end and calls nextround', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(turnState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-2 phase', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());
  });

  it('shows error alert', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('shows CPU player area', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1.*13枚/)).toBeInTheDocument());
  });

  it('score table shows all players', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => {
      expect(screen.getByText('スコア')).toBeInTheDocument();
      expect(screen.getByText('あなた')).toBeInTheDocument();
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
    });
  });

  it('score table headers have scope="col"', async () => {
    const { container } = renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByText('あなた')).toBeInTheDocument());
    for (const th of container.querySelectorAll('th')) {
      expect(th).toHaveAttribute('scope', 'col');
    }
  });

  it('renders the shared table meld cards', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => {
      expect(screen.getByText('場のメルド')).toBeInTheDocument();
      expect(screen.getByAltText('♠ 3')).toBeInTheDocument();
    });
  });

  it('reveals CPU cards and deadwood on round end', async () => {
    mockExec.mockResolvedValue(roundEndCpuCardsState);
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♦ 5')).toBeInTheDocument();
      expect(screen.getByText(/デッドウッド 11/)).toBeInTheDocument();
    });
  });

  it('does not reveal CPU cards during the turn phase', async () => {
    const turnWithCpuCards: MachiavelliResponse = {
      ...turnState,
      players: [
        turnState.players[0],
        player({ id: 1, isHuman: false, cardCount: 2, cards: [{ design: 'DIAMOND', value: 5 }] }),
      ],
    };
    mockExec.mockResolvedValue(turnWithCpuCards);
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByAltText('♦ 5')).not.toBeInTheDocument();
  });

  it('card selection toggles aria-pressed', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('card buttons have aria-label with card name', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.getByAltText('♠ A').closest('button')).toHaveAttribute('aria-label', '♠ A');
  });

  it('reset button calls exec with confirm', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    mockExec.mockClear();
    mockExec.mockResolvedValue(turnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 4,
        cpuDifficulty: 1,
        targetRounds: 3,
      }),
    );
  });

  it('shows and dismisses confirm dialog', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('settings panel changes playerCount', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '5' } });
    mockExec.mockClear();
    mockExec.mockResolvedValue(turnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 5,
        cpuDifficulty: 1,
        targetRounds: 3,
      }),
    );
  });

  it('round info displayed', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => {
      expect(screen.getByText('ラウンド 1/3')).toBeInTheDocument();
      expect(screen.getByText('山札: 40枚')).toBeInTheDocument();
    });
  });

  it('phase indicator shows your turn on the human turn', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows waiting on cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('待機中'));
  });

  it('number key toggles a card on the human turn', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());
    vi.mocked(actionLogApi.machiavelli).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));
    await waitFor(() => expect(actionLogApi.machiavelli).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();
    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText(/^棋譜$/)).not.toBeInTheDocument());
  });

  it('does not show action log button when not game end', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('棋譜を見る')).not.toBeInTheDocument();
  });

  it('renders tutorial button and starts/skips tutorial', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    renderWithProviders(<MachiavelliPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });
});
