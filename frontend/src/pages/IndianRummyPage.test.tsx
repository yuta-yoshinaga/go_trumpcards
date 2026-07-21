import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, indianRummyApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { IndianRummyPlayer, IndianRummyResponse } from '../types/card';
import { IndianRummyPage } from './IndianRummyPage';

vi.mock('../api/gameApi', () => ({
  indianRummyApi: { exec: vi.fn() },
  actionLogApi: { indianrummy: vi.fn() },
}));

const mockExec = vi.mocked(indianRummyApi.exec);

function player(overrides: Partial<IndianRummyPlayer> = {}): IndianRummyPlayer {
  return {
    id: 0,
    isHuman: true,
    cardCount: 13,
    cards: [],
    roundScore: 0,
    cumulativeScore: 0,
    deadwood: 0,
    hasPureSequence: false,
    ...overrides,
  };
}

const drawPhaseState: IndianRummyResponse = {
  players: [
    player({
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
    }),
    player({ id: 1, isHuman: false, roundScore: 3, cumulativeScore: 10 }),
  ],
  phase: 0,
  roundNumber: 1,
  targetRounds: 3,
  currentPlayerIdx: 0,
  dealerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 40,
  wildJoker: { design: 'CLOVER', value: 5 },
  wildRank: 5,
  gameEndFlag: false,
  winnerIdx: -1,
  declarerIdx: -1,
  declarationValid: false,
  message: '',
  config: { playerCount: 4, cpuDifficulty: 1, targetRounds: 3 },
};

const discardPhaseState: IndianRummyResponse = { ...drawPhaseState, phase: 1 };

// 13-card arrangement that IS a valid declaration (two pure runs + two sets),
// verified against the Go domain. The 14th card (♣2) is the finish/discard card.
const declareValidHand: IndianRummyPlayer['cards'] = [
  { design: 'SPADE', value: 3 },
  { design: 'SPADE', value: 4 },
  { design: 'SPADE', value: 5 },
  { design: 'HEART', value: 7 },
  { design: 'HEART', value: 8 },
  { design: 'HEART', value: 9 },
  { design: 'DIAMOND', value: 10 },
  { design: 'CLOVER', value: 10 },
  { design: 'SPADE', value: 10 },
  { design: 'DIAMOND', value: 13 },
  { design: 'CLOVER', value: 13 },
  { design: 'HEART', value: 13 },
  { design: 'SPADE', value: 13 },
  { design: 'CLOVER', value: 2 }, // finish card (index 13)
];

const declareValidState: IndianRummyResponse = {
  ...discardPhaseState,
  wildJoker: null,
  wildRank: 0,
  players: [player({ cards: declareValidHand }), player({ id: 1, isHuman: false })],
};

// 13-card arrangement that is INVALID: only sets, no sequence at all (no pure
// sequence). The 14th card (♥8) is the finish card.
const declareNoPureHand: IndianRummyPlayer['cards'] = [
  { design: 'DIAMOND', value: 2 },
  { design: 'CLOVER', value: 2 },
  { design: 'SPADE', value: 2 },
  { design: 'DIAMOND', value: 6 },
  { design: 'CLOVER', value: 6 },
  { design: 'SPADE', value: 6 },
  { design: 'DIAMOND', value: 10 },
  { design: 'CLOVER', value: 10 },
  { design: 'SPADE', value: 10 },
  { design: 'DIAMOND', value: 13 },
  { design: 'CLOVER', value: 13 },
  { design: 'HEART', value: 13 },
  { design: 'SPADE', value: 13 },
  { design: 'HEART', value: 8 }, // finish card (index 13)
];

const declareNoPureState: IndianRummyResponse = {
  ...discardPhaseState,
  wildJoker: null,
  wildRank: 0,
  players: [player({ cards: declareNoPureHand }), player({ id: 1, isHuman: false })],
};

// 13-card arrangement that is INVALID: one pure run but 10 cards cannot all be
// melded (76 deadwood points). The 14th card (♣9) is the finish card.
const declareUncoveredHand: IndianRummyPlayer['cards'] = [
  { design: 'SPADE', value: 3 },
  { design: 'SPADE', value: 4 },
  { design: 'SPADE', value: 5 },
  { design: 'HEART', value: 7 },
  { design: 'DIAMOND', value: 9 },
  { design: 'CLOVER', value: 11 },
  { design: 'DIAMOND', value: 2 },
  { design: 'CLOVER', value: 13 },
  { design: 'SPADE', value: 8 },
  { design: 'HEART', value: 6 },
  { design: 'DIAMOND', value: 4 },
  { design: 'HEART', value: 12 },
  { design: 'SPADE', value: 1 },
  { design: 'CLOVER', value: 9 }, // finish card (index 13)
];

const declareUncoveredState: IndianRummyResponse = {
  ...discardPhaseState,
  wildJoker: null,
  wildRank: 0,
  players: [player({ cards: declareUncoveredHand }), player({ id: 1, isHuman: false })],
};
const roundEndState: IndianRummyResponse = { ...drawPhaseState, phase: 2 };
const gameEndState: IndianRummyResponse = {
  ...drawPhaseState,
  phase: 3,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};
const gameEndByFlagState: IndianRummyResponse = {
  ...drawPhaseState,
  phase: 0,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};
const cpuTurnState: IndianRummyResponse = { ...drawPhaseState, currentPlayerIdx: 1 };
const noDiscardState: IndianRummyResponse = { ...drawPhaseState, discardTop: null };
const roundEndCpuCardsState: IndianRummyResponse = {
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
        { design: 'HEART', value: 8 },
      ],
      roundScore: 3,
      cumulativeScore: 10,
      deadwood: 19,
      hasPureSequence: true,
    }),
  ],
};

beforeEach(() => {
  mockExec.mockResolvedValue(drawPhaseState);
});

describe('IndianRummyPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<IndianRummyPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with default config', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 4,
        cpuDifficulty: 1,
        targetRounds: 3,
      }),
    );
  });

  it('renders draw phase with human cards', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  it('marks wild-rank cards in the human hand with a WILD badge', async () => {
    const wildInHandState: IndianRummyResponse = {
      ...drawPhaseState,
      players: [
        player({
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'DIAMOND', value: 5 }, // matches wildJoker rank 5 -> wild
            { design: 'JOKER', value: 0 }, // printed joker -> wild
          ],
        }),
        drawPhaseState.players[1],
      ],
    };
    mockExec.mockResolvedValue(wildInHandState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    // Two of the three hand cards are wild (the rank-5 card and the printed joker).
    expect(screen.getAllByTestId('ir-wild-badge')).toHaveLength(2);
    // The non-wild ace carries no wild annotation in its label.
    expect(screen.getByRole('button', { name: '♠ A' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /♦ 5 .*ワイルド/ })).toBeInTheDocument();
  });

  it('marks wild-rank cards in revealed CPU hands with a WILD badge', async () => {
    mockExec.mockResolvedValue(roundEndCpuCardsState);
    renderWithProviders(<IndianRummyPage />);
    // CPU hand has a rank-5 card (DIAMOND 5) matching the wild joker.
    await waitFor(() => expect(screen.getAllByTestId('ir-wild-badge').length).toBeGreaterThanOrEqual(1));
  });

  it('renders draw stock and draw discard buttons on human draw turn', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeInTheDocument();
    });
  });

  it('draw discard button disabled when no discard top', async () => {
    mockExec.mockResolvedValue(noDiscardState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeDisabled());
  });

  it('calls drawstock when draw stock button clicked', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('calls drawdiscard when draw discard button clicked', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '捨て札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('renders discard and declare buttons on human discard turn', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '宣言' })).toBeInTheDocument();
    });
  });

  it('discard and declare disabled when not exactly 1 card selected', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨てる' })).toBeDisabled());
    expect(screen.getByRole('button', { name: '宣言' })).toBeDisabled();
  });

  it('discard button enabled when 1 card selected', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '捨てる' })).not.toBeDisabled();
  });

  it('calls discard command when discard button clicked', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '捨てる' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 0));
  });

  it('calls declare command when declare button clicked', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(roundEndState);
    fireEvent.click(screen.getByRole('button', { name: '宣言' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare', 0));
  });

  // -- Declaration preview (client-side) --
  it('does not show the declare preview until a finish card is selected', async () => {
    mockExec.mockResolvedValue(declareValidState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '宣言' })).toBeInTheDocument());
    expect(screen.queryByTestId('indianrummy-declare-preview')).not.toBeInTheDocument();
  });

  it('shows a valid declare preview when the remaining 13 cards form a valid declaration', async () => {
    mockExec.mockResolvedValue(declareValidState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByAltText('♣ 2')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♣ 2').closest('button') as HTMLButtonElement);
    const preview = await screen.findByTestId('indianrummy-declare-preview');
    expect(preview).toBeInTheDocument();
    expect(screen.getByTestId('indianrummy-declare-preview-valid')).toBeInTheDocument();
    expect(screen.queryByTestId('indianrummy-declare-preview-invalid')).not.toBeInTheDocument();
    // Declare button is never blocked by the preview.
    expect(screen.getByRole('button', { name: '宣言' })).not.toBeDisabled();
  });

  it('warns about a missing pure sequence and the penalty for an invalid declaration', async () => {
    mockExec.mockResolvedValue(declareNoPureState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByAltText('♥ 8')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♥ 8').closest('button') as HTMLButtonElement);
    await screen.findByTestId('indianrummy-declare-preview-invalid');
    expect(screen.getByText('純シーケンス未成立')).toBeInTheDocument();
    expect(screen.getByText('このまま宣言すると +80 点')).toBeInTheDocument();
    // Player can still force the declaration through.
    expect(screen.getByRole('button', { name: '宣言' })).not.toBeDisabled();
  });

  it('warns about unmelded cards when a pure sequence exists but cards are uncovered', async () => {
    mockExec.mockResolvedValue(declareUncoveredState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByAltText('♣ 9')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♣ 9').closest('button') as HTMLButtonElement);
    await screen.findByTestId('indianrummy-declare-preview-invalid');
    expect(screen.queryByText('純シーケンス未成立')).not.toBeInTheDocument();
    expect(screen.getByText(/未メルド .*枚（76 点）/)).toBeInTheDocument();
  });

  it('does not show draw buttons when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '山札から引く' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '捨て札から引く' })).not.toBeInTheDocument();
  });

  it('shows next round button on round end and calls nextround', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-3 phase', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());
  });

  it('shows error alert', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('shows CPU player area', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1.*13枚/)).toBeInTheDocument());
  });

  it('score table shows all players', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => {
      expect(screen.getByText('スコア')).toBeInTheDocument();
      expect(screen.getByText('あなた')).toBeInTheDocument();
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
    });
  });

  it('score table headers have scope="col"', async () => {
    const { container } = renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByText('あなた')).toBeInTheDocument());
    for (const th of container.querySelectorAll('th')) {
      expect(th).toHaveAttribute('scope', 'col');
    }
  });

  it('shows wild joker indicator', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => {
      expect(screen.getByTestId('indianrummy-wild-joker')).toBeInTheDocument();
      expect(screen.getByText('ワイルドジョーカー')).toBeInTheDocument();
    });
  });

  it('shows discard top card', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => {
      expect(screen.getByText('捨て札')).toBeInTheDocument();
      expect(screen.getByAltText('♥ 7')).toBeInTheDocument();
    });
  });

  it('does not show discard top when null', async () => {
    mockExec.mockResolvedValue(noDiscardState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('捨て札')).not.toBeInTheDocument();
  });

  it('reveals CPU cards, deadwood, and pure-sequence badge on round end', async () => {
    mockExec.mockResolvedValue(roundEndCpuCardsState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♦ 5')).toBeInTheDocument();
      expect(screen.getByText(/デッドウッド 19/)).toBeInTheDocument();
      expect(screen.getByText('純シーケンス有')).toBeInTheDocument();
    });
  });

  it('does not reveal CPU cards during draw phase', async () => {
    const drawWithCpuCards: IndianRummyResponse = {
      ...drawPhaseState,
      players: [
        drawPhaseState.players[0],
        player({ id: 1, isHuman: false, cardCount: 3, cards: [{ design: 'DIAMOND', value: 5 }] }),
      ],
    };
    mockExec.mockResolvedValue(drawWithCpuCards);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByAltText('♦ 5')).not.toBeInTheDocument();
  });

  it('card selection toggles aria-pressed', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('card buttons have aria-label with card name', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.getByAltText('♠ A').closest('button')).toHaveAttribute('aria-label', '♠ A');
  });

  it('reset button calls exec with confirm', async () => {
    renderWithProviders(<IndianRummyPage />);
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
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('settings panel changes playerCount', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '3' } });
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 3,
        cpuDifficulty: 1,
        targetRounds: 3,
      }),
    );
  });

  it('settings panel changes targetRounds', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[2], { target: { value: '5' } });
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 4,
        cpuDifficulty: 1,
        targetRounds: 5,
      }),
    );
  });

  it('round info displayed', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => {
      expect(screen.getByText('ラウンド 1/3')).toBeInTheDocument();
      expect(screen.getByText('山札: 40枚')).toBeInTheDocument();
    });
  });

  it('shows loading state and disables reset while loading', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    let resolve!: (value: IndianRummyResponse) => void;
    const slow = new Promise<IndianRummyResponse>((r) => {
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
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    const container = screen.getByRole('button', { name: 'リセット' }).closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');
  });

  // -- PhaseIndicator --
  it('phase indicator shows your turn on human draw turn', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows waiting on cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('待機中'));
  });

  // -- Keyboard navigation --
  it('number key toggles a card in discard phase', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('Enter key triggers discard in discard phase', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    fireEvent.keyDown(document, { key: '1' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 0));
  });

  it('Escape key clears selection', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('keyboard nav disabled in draw phase', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('keyboard nav disabled when not human turn', async () => {
    mockExec.mockResolvedValue({ ...discardPhaseState, currentPlayerIdx: 1 });
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());
    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  // -- Action log --
  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());
    vi.mocked(actionLogApi.indianrummy).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));
    await waitFor(() => expect(actionLogApi.indianrummy).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();
    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText(/^棋譜$/)).not.toBeInTheDocument());
  });

  it('does not show action log button when not game end', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('棋譜を見る')).not.toBeInTheDocument();
  });

  // -- Tutorial --
  it('renders tutorial button and starts/skips tutorial', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    renderWithProviders(<IndianRummyPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });
});
