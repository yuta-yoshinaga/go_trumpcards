import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, panApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { PanPlayer, PanResponse } from '../types/card';
import { PanPage } from './PanPage';

vi.mock('../api/gameApi', () => ({
  panApi: { exec: vi.fn() },
  actionLogApi: { pan: vi.fn() },
}));

const mockExec = vi.mocked(panApi.exec);

function player(overrides: Partial<PanPlayer> = {}): PanPlayer {
  return {
    id: 0,
    isHuman: true,
    cardCount: 11,
    cards: [],
    laidMelds: [],
    meldedCount: 0,
    chips: 50,
    handPoints: 0,
    roundScore: 0,
    cumulativeScore: 0,
    ...overrides,
  };
}

const humanHand = [
  { design: 'SPADE', value: 1 },
  { design: 'CLOVER', value: 5 },
  { design: 'DIAMOND', value: 6 },
] as const;

const drawPhaseState: PanResponse = {
  players: [
    player({ cards: [...humanHand] }),
    player({ id: 1, isHuman: false, cardCount: 10, roundScore: 3, cumulativeScore: 10, chips: 45 }),
  ],
  phase: 0,
  roundNumber: 1,
  targetRounds: 3,
  currentPlayerIdx: 0,
  dealerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 250,
  deckSize: 320,
  winMeldCount: 11,
  gameEndFlag: false,
  winnerIdx: -1,
  panDeclarerIdx: -1,
  message: '',
  config: { playerCount: 4, cpuDifficulty: 1, targetRounds: 3 },
};

const playPhaseState: PanResponse = {
  ...drawPhaseState,
  phase: 1,
  players: [
    player({ cards: [...humanHand] }),
    player({
      id: 1,
      isHuman: false,
      cardCount: 7,
      chips: 45,
      meldedCount: 3,
      laidMelds: [
        {
          cards: [
            { design: 'DIAMOND', value: 4 },
            { design: 'HEART', value: 4 },
            { design: 'CLOVER', value: 4 },
          ],
        },
      ],
    }),
  ],
};

// バジェ (3/5/7 のセット) を場に出した状態。4 のセットは valle ではない。
const valleMeldState: PanResponse = {
  ...drawPhaseState,
  phase: 1,
  players: [
    player({ cards: [...humanHand] }),
    player({
      id: 1,
      isHuman: false,
      cardCount: 7,
      laidMelds: [
        {
          cards: [
            { design: 'DIAMOND', value: 5 },
            { design: 'HEART', value: 5 },
            { design: 'CLOVER', value: 5 },
          ],
        },
        {
          cards: [
            { design: 'DIAMOND', value: 4 },
            { design: 'HEART', value: 4 },
            { design: 'CLOVER', value: 4 },
          ],
        },
      ],
    }),
  ],
};

const roundEndState: PanResponse = { ...drawPhaseState, phase: 2 };
const gameEndState: PanResponse = {
  ...drawPhaseState,
  phase: 3,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};
const cpuTurnState: PanResponse = { ...drawPhaseState, currentPlayerIdx: 1 };
const noDiscardState: PanResponse = { ...drawPhaseState, discardTop: null };
const roundEndCpuCardsState: PanResponse = {
  ...drawPhaseState,
  phase: 2,
  players: [
    drawPhaseState.players[0],
    player({
      id: 1,
      isHuman: false,
      cardCount: 3,
      cards: [
        { design: 'DIAMOND', value: 5 },
        { design: 'CLOVER', value: 6 },
        { design: 'HEART', value: 7 },
      ],
      chips: 45,
      roundScore: 3,
      cumulativeScore: 10,
    }),
  ],
};

beforeEach(() => {
  mockExec.mockResolvedValue(drawPhaseState);
});

describe('PanPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<PanPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with default config', async () => {
    renderWithProviders(<PanPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 4,
        cpuDifficulty: 1,
        targetRounds: 3,
      }),
    );
  });

  it('renders draw phase with human cards', async () => {
    renderWithProviders(<PanPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♣ 5')).toBeInTheDocument();
    });
  });

  it('renders draw stock and draw discard buttons on human draw turn', async () => {
    renderWithProviders(<PanPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeInTheDocument();
    });
  });

  it('draw discard button disabled when no discard top', async () => {
    mockExec.mockResolvedValue(noDiscardState);
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeDisabled());
  });

  it('calls drawstock when draw stock button clicked', async () => {
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('calls drawdiscard when draw discard button clicked', async () => {
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '捨て札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('renders meld and discard buttons on human play turn', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<PanPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'メルド' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument();
    });
  });

  it('meld disabled with fewer than 3 cards, discard disabled without exactly 1', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'メルド' })).toBeDisabled());
    expect(screen.getByRole('button', { name: '捨てる' })).toBeDisabled();
  });

  it('meld enabled when 3 cards selected and calls meld', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♣ 5').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♦ 6').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: 'メルド' })).not.toBeDisabled();
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'メルド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('meld', { cardIndices: [0, 1, 2] }));
  });

  it('discard enabled when 1 card selected and calls discard', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '捨てる' })).not.toBeDisabled();
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '捨てる' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndex: 0 }));
  });

  it('shows layoff button when 1 card selected in play phase and calls layoff', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    // No layoff target before selecting a card.
    expect(screen.queryByTestId('pan-layoff-1-0')).not.toBeInTheDocument();
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    const layoffBtn = await screen.findByTestId('pan-layoff-1-0');
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(layoffBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('layoff', { meldOwner: 1, meldIdx: 0, cardIndex: 0 }));
  });

  it('does not show draw buttons when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '山札から引く' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '捨て札から引く' })).not.toBeInTheDocument();
  });

  it('shows next round button on round end and calls nextround', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PanPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows error alert', async () => {
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('shows CPU player area with chips', async () => {
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1.*10枚/)).toBeInTheDocument());
  });

  it('score table shows all players and chips column', async () => {
    renderWithProviders(<PanPage />);
    await waitFor(() => {
      expect(screen.getByText('スコア')).toBeInTheDocument();
      expect(screen.getByText('あなた')).toBeInTheDocument();
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
      expect(screen.getByText('チップ')).toBeInTheDocument();
    });
  });

  it('score table headers have scope="col"', async () => {
    const { container } = renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByText('あなた')).toBeInTheDocument());
    for (const th of container.querySelectorAll('th')) {
      expect(th).toHaveAttribute('scope', 'col');
    }
  });

  it('shows discard top card', async () => {
    renderWithProviders(<PanPage />);
    await waitFor(() => {
      expect(screen.getByText('捨て札')).toBeInTheDocument();
    });
  });

  it('does not show discard top label when null', async () => {
    mockExec.mockResolvedValue(noDiscardState);
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('捨て札')).not.toBeInTheDocument();
  });

  it('reveals CPU cards on round end', async () => {
    mockExec.mockResolvedValue(roundEndCpuCardsState);
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByAltText('♦ 5')).toBeInTheDocument());
  });

  it('does not reveal CPU cards during draw phase', async () => {
    const drawWithCpuCards: PanResponse = {
      ...drawPhaseState,
      players: [
        drawPhaseState.players[0],
        player({ id: 1, isHuman: false, cardCount: 3, cards: [{ design: 'DIAMOND', value: 5 }] }),
      ],
    };
    mockExec.mockResolvedValue(drawWithCpuCards);
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByAltText('♦ 5')).not.toBeInTheDocument();
  });

  it('card selection toggles aria-pressed', async () => {
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('card buttons have aria-label with card name', async () => {
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.getByAltText('♠ A').closest('button')).toHaveAttribute('aria-label', '♠ A');
  });

  it('reset button calls exec with confirm', async () => {
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
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
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('settings panel changes playerCount', async () => {
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '5' } });
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
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
    renderWithProviders(<PanPage />);
    await waitFor(() => {
      expect(screen.getByText('ラウンド 1/3')).toBeInTheDocument();
      expect(screen.getByText('山札: 250枚')).toBeInTheDocument();
    });
  });

  it('shows human meld progress toward the win condition', async () => {
    const progressState: PanResponse = {
      ...drawPhaseState,
      players: [player({ cards: [...humanHand], meldedCount: 4 }), drawPhaseState.players[1]],
    };
    mockExec.mockResolvedValue(progressState);
    renderWithProviders(<PanPage />);
    const bar = await screen.findByTestId('pan-meld-progress');
    expect(bar).toHaveTextContent('メルド 4/11');
    const progressbar = screen.getByRole('progressbar');
    expect(progressbar).toHaveAttribute('aria-valuenow', '4');
    expect(progressbar).toHaveAttribute('aria-valuemax', '11');
    // Not yet in the final stretch, so no "remaining" callout.
    expect(screen.queryByText('あと 2 枚')).not.toBeInTheDocument();
  });

  it('highlights remaining melds when close to going out', async () => {
    const closeState: PanResponse = {
      ...drawPhaseState,
      players: [player({ cards: [...humanHand], meldedCount: 9 }), drawPhaseState.players[1]],
    };
    mockExec.mockResolvedValue(closeState);
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByTestId('pan-meld-progress')).toHaveTextContent('メルド 9/11'));
    expect(screen.getByText('あと 2 枚')).toBeInTheDocument();
  });

  it('phase indicator shows your turn on human draw turn', async () => {
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows waiting on cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('待機中'));
  });

  it('number key toggles a card in play phase', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());
    vi.mocked(actionLogApi.pan).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));
    await waitFor(() => expect(actionLogApi.pan).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();
  });

  it('renders tutorial button and starts/skips tutorial', async () => {
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });
});

describe('PanPage meld candidates', () => {
  // Hand with a set of three 5s (a legal meld candidate) plus a ♠4 that can be
  // laid off onto CPU 1's existing set of 4s (see playPhaseState).
  const meldableHand = [
    { design: 'SPADE', value: 5 },
    { design: 'HEART', value: 5 },
    { design: 'DIAMOND', value: 5 },
    { design: 'SPADE', value: 4 },
  ] as const;
  const meldablePlayState: PanResponse = {
    ...playPhaseState,
    players: [player({ cards: [...meldableHand] }), playPhaseState.players[1]],
  };

  // Near-miss: 6-J-Q of one suit LOOKS like a run, but 6 and J are not adjacent
  // in Pan's reduced deck (the 7 sits between them), so it is NOT a legal meld.
  const nearMissHand = [
    { design: 'SPADE', value: 6 },
    { design: 'SPADE', value: 11 },
    { design: 'SPADE', value: 12 },
  ] as const;
  const nearMissState: PanResponse = {
    ...playPhaseState,
    players: [player({ cards: [...nearMissHand] }), playPhaseState.players[1]],
  };

  afterEach(() => {
    localStorage.removeItem('hint_enabled_pan');
  });

  it('surfaces a set candidate when the hint toggle is on', async () => {
    localStorage.setItem('hint_enabled_pan', 'true');
    mockExec.mockResolvedValue(meldablePlayState);
    renderWithProviders(<PanPage />);
    expect(await screen.findByTestId('pan-meld-candidates')).toBeInTheDocument();
    expect(screen.getByTestId('pan-candidate-0-1-2')).toBeInTheDocument();
    expect(screen.getByText('セット')).toBeInTheDocument();
  });

  it('selecting a candidate selects those cards and enables meld', async () => {
    localStorage.setItem('hint_enabled_pan', 'true');
    mockExec.mockResolvedValue(meldablePlayState);
    renderWithProviders(<PanPage />);
    const candidate = await screen.findByTestId('pan-candidate-0-1-2');
    expect(screen.getByRole('button', { name: 'メルド' })).toBeDisabled();
    fireEvent.click(candidate);
    expect(screen.getByRole('button', { name: 'メルド' })).not.toBeDisabled();
    mockExec.mockClear();
    mockExec.mockResolvedValue(meldablePlayState);
    fireEvent.click(screen.getByRole('button', { name: 'メルド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('meld', { cardIndices: [0, 1, 2] }));
  });

  it('marks a layoff-able hand card', async () => {
    localStorage.setItem('hint_enabled_pan', 'true');
    mockExec.mockResolvedValue(meldablePlayState);
    renderWithProviders(<PanPage />);
    await screen.findByTestId('pan-meld-candidates');
    const layoffCard = screen.getByAltText('♠ 4').closest('button') as HTMLButtonElement;
    expect(layoffCard).toHaveAttribute('data-layoff-target', 'true');
  });

  it('does NOT surface a candidate for a near-but-invalid hand', async () => {
    localStorage.setItem('hint_enabled_pan', 'true');
    mockExec.mockResolvedValue(nearMissState);
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'メルド' })).toBeInTheDocument());
    expect(screen.queryByTestId('pan-meld-candidates')).not.toBeInTheDocument();
  });

  it('hides candidates when the hint toggle is off', async () => {
    localStorage.removeItem('hint_enabled_pan');
    mockExec.mockResolvedValue(meldablePlayState);
    renderWithProviders(<PanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'メルド' })).toBeInTheDocument());
    expect(screen.queryByTestId('pan-meld-candidates')).not.toBeInTheDocument();
  });

  // **チップ列が動いた理由を盤面から読めるようにする。**バジェ (3/5/7 のセット) は
  // 全員にチップを配るのに、どのメルドがそれなのか表示が無かった (#4853)。
  it('badges only the valle melds on the table', async () => {
    mockExec.mockResolvedValue(valleMeldState);
    renderWithProviders(<PanPage />);

    await waitFor(() => expect(screen.getByTestId('pan-valle-1-0')).toBeInTheDocument());
    expect(screen.queryByTestId('pan-valle-1-1')).not.toBeInTheDocument();
  });
});
