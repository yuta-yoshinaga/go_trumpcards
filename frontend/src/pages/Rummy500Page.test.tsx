import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { rummy500Api } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Rummy500Response } from '../types/card';
import { Rummy500Page } from './Rummy500Page';

vi.mock('../api/gameApi', () => ({
  rummy500Api: { exec: vi.fn() },
  actionLogApi: { rummy500: vi.fn() },
}));

const mockExec = vi.mocked(rummy500Api.exec);

const drawPhaseState: Rummy500Response = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 2,
      cards: [
        { design: 'SPADE', value: 7 },
        { design: 'HEART', value: 7 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
      laidMelds: [],
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 13,
      cards: [],
      roundScore: 0,
      cumulativeScore: 0,
      laidMelds: [],
    },
  ],
  layoffTargets: [],
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardPile: [
    { design: 'HEART', value: 3 },
    { design: 'CLOVER', value: 5 },
  ],
  drawPileCount: 25,
  gameEndFlag: false,
  winnerIdx: -1,
  roundEnderIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 500 },
};

const playPhaseState: Rummy500Response = {
  ...drawPhaseState,
  phase: 1,
  players: [
    {
      ...drawPhaseState.players[0],
      cardCount: 3,
      cards: [
        { design: 'SPADE', value: 7 },
        { design: 'HEART', value: 7 },
        { design: 'CLOVER', value: 7 },
      ],
    },
    drawPhaseState.players[1],
  ],
};

const playPhaseInvalidHandState: Rummy500Response = {
  ...drawPhaseState,
  phase: 1,
  players: [
    {
      ...drawPhaseState.players[0],
      cardCount: 3,
      cards: [
        { design: 'SPADE', value: 3 },
        { design: 'HEART', value: 8 },
        { design: 'CLOVER', value: 11 },
      ],
    },
    drawPhaseState.players[1],
  ],
};

const roundEndState: Rummy500Response = {
  ...drawPhaseState,
  phase: 2,
};

const gameEndState: Rummy500Response = {
  ...drawPhaseState,
  phase: 3,
  gameEndFlag: true,
  winnerIdx: 0,
};

describe('Rummy500Page', () => {
  beforeEach(() => {
    mockExec.mockReset();
    mockExec.mockResolvedValue(drawPhaseState);
  });

  it('calls reset on mount with default config', async () => {
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 500 });
    });
  });

  it('shows draw stock button in Draw phase', async () => {
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /山札から引く/ })).toBeInTheDocument();
    });
  });

  it('clicking discard card in Draw phase calls drawdiscard', async () => {
    renderWithProviders(<Rummy500Page />);
    const card = await screen.findByTestId('disc-card-0');
    fireEvent.click(card);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith('drawdiscard', undefined, undefined, undefined, 0);
    });
  });

  it('previews the whole take-range when hovering a mid-pile discard card', async () => {
    renderWithProviders(<Rummy500Page />);
    // drawPhaseState.discardPile has 2 cards: [♥3 (idx 0), ♣5 (idx 1, top)].
    const bottom = await screen.findByTestId('disc-card-0');
    const top = screen.getByTestId('disc-card-1');
    // Nothing highlighted before hover.
    expect(bottom).not.toHaveAttribute('data-in-pickup-range');
    expect(top).not.toHaveAttribute('data-in-pickup-range');

    fireEvent.mouseEnter(bottom);
    // Hovering the bottom card highlights it AND every card above it (the whole pile).
    expect(bottom).toHaveAttribute('data-in-pickup-range', 'true');
    expect(top).toHaveAttribute('data-in-pickup-range', 'true');
    // The badge announces the number of cards that would be taken (2).
    const badge = screen.getByTestId('pickup-range-badge');
    expect(badge).toHaveTextContent('2');
    // The aria-label of the hovered card also carries the count.
    expect(bottom).toHaveAttribute('aria-label', expect.stringContaining('2枚引き取る'));

    fireEvent.mouseLeave(bottom);
    expect(bottom).not.toHaveAttribute('data-in-pickup-range');
    expect(top).not.toHaveAttribute('data-in-pickup-range');
  });

  it('previews only the top card when hovering the top of the discard pile', async () => {
    renderWithProviders(<Rummy500Page />);
    const bottom = await screen.findByTestId('disc-card-0');
    const top = screen.getByTestId('disc-card-1');

    fireEvent.focus(top);
    // Hovering/focusing the top card previews just that one card.
    expect(top).toHaveAttribute('data-in-pickup-range', 'true');
    expect(bottom).not.toHaveAttribute('data-in-pickup-range');
    expect(screen.getByTestId('pickup-range-badge')).toHaveTextContent('1');
    expect(top).toHaveAttribute('aria-label', expect.stringContaining('1枚引き取る'));
  });

  it('shows meld + discard buttons in Play phase', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /メルドする/ })).toBeInTheDocument();
    });
    expect(screen.getByRole('button', { name: /^捨てる$/ })).toBeInTheDocument();
  });

  it('selects a lay-off target by clicking a laid meld', async () => {
    const withMeld: Rummy500Response = {
      ...playPhaseState,
      // 置けるかどうかを決めるのはサーバー (Rummy500CanLayoff)。ここでは
      // 「手札 0 枚目はこのメルドに置ける」と答えた状態を作る。
      layoffTargets: [[{ owner: 1, meldIdx: 0 }]],
      players: [
        playPhaseState.players[0],
        {
          ...playPhaseState.players[1],
          laidMelds: [
            {
              cards: [
                { design: 'DIAMOND', value: 4 },
                { design: 'DIAMOND', value: 5 },
                { design: 'DIAMOND', value: 6 },
              ],
            },
          ],
        },
      ],
    };
    mockExec.mockResolvedValue(withMeld);
    renderWithProviders(<Rummy500Page />);
    const meldBtn = await screen.findByTestId('layoff-meld-1-0');
    expect(meldBtn).toHaveAttribute('aria-pressed', 'false');

    // Before selecting, the footer shows the "click a meld" hint and Lay off is disabled.
    expect(screen.getByTestId('r5-layoff-target')).toHaveTextContent(/上のメルドをクリック/);
    expect(screen.getByRole('button', { name: 'レイオフ' })).toBeDisabled();

    fireEvent.click(meldBtn);
    await waitFor(() => expect(screen.getByTestId('layoff-meld-1-0')).toHaveAttribute('aria-pressed', 'true'));
    // The selected meld is now described in the footer text.
    expect(screen.getByTestId('r5-layoff-target')).toHaveTextContent('#0');

    // Clicking the same meld again toggles the selection off.
    fireEvent.click(screen.getByTestId('layoff-meld-1-0'));
    await waitFor(() => expect(screen.getByTestId('layoff-meld-1-0')).toHaveAttribute('aria-pressed', 'false'));
    expect(screen.getByTestId('r5-layoff-target')).toHaveTextContent(/上のメルドをクリック/);
    // Re-select for the lay-off flow below.
    fireEvent.click(screen.getByTestId('layoff-meld-1-0'));
    await waitFor(() => expect(screen.getByTestId('layoff-meld-1-0')).toHaveAttribute('aria-pressed', 'true'));
    // Still disabled until exactly one hand card is chosen.
    expect(screen.getByRole('button', { name: 'レイオフ' })).toBeDisabled();

    const handCard = document.querySelector('[data-tutorial="r5-player-hand"] button') as HTMLButtonElement;
    fireEvent.click(handCard);
    const layoffBtn = screen.getByRole('button', { name: 'レイオフ' });
    await waitFor(() => expect(layoffBtn).toBeEnabled());

    mockExec.mockClear();
    fireEvent.click(layoffBtn);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('layoff', undefined, undefined, undefined, undefined, {
        meldOwner: 1,
        meldIdx: 0,
        cardIndex: 0,
      }),
    );
  });

  it('shows next round button in RoundEnd phase', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /次のラウンド/ })).toBeInTheDocument();
    });
  });

  it('shows game end celebration', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => {
      // The shell renders some game-end indicator
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 500 });
    });
  });

  it('meld button disabled when fewer than 3 selected', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /メルドする/ })).toBeDisabled();
    });
  });

  it('enables the meld button and shows no warning for a valid set selection', async () => {
    // playPhaseState hand is ♠7 ♥7 ♣7 — a valid same-rank set.
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<Rummy500Page />);
    const meldBtn = await screen.findByRole('button', { name: /メルドする/ });
    // No cards selected yet → disabled and no warning.
    expect(meldBtn).toBeDisabled();
    expect(screen.queryByTestId('r5-invalid-meld')).not.toBeInTheDocument();

    const handCards = document.querySelectorAll('[data-tutorial="r5-player-hand"] button');
    for (const c of handCards) fireEvent.click(c);

    await waitFor(() => expect(screen.getByRole('button', { name: /メルドする/ })).toBeEnabled());
    expect(screen.queryByTestId('r5-invalid-meld')).not.toBeInTheDocument();
  });

  it('warns and disables the meld button for an invalid 3-card selection', async () => {
    // playPhaseInvalidHandState hand is ♠3 ♥8 ♣11 — neither a set nor a run.
    mockExec.mockResolvedValue(playPhaseInvalidHandState);
    renderWithProviders(<Rummy500Page />);
    await screen.findByRole('button', { name: /メルドする/ });

    const handCards = document.querySelectorAll('[data-tutorial="r5-player-hand"] button');
    for (const c of handCards) fireEvent.click(c);

    await waitFor(() => expect(screen.getByTestId('r5-invalid-meld')).toBeInTheDocument());
    expect(screen.getByTestId('r5-invalid-meld')).toHaveTextContent('セットまたはランになっていません');
    expect(screen.getByRole('button', { name: /メルドする/ })).toBeDisabled();
  });

  it('drawstock button triggers exec', async () => {
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /山札から引く/ })).toBeInTheDocument();
    });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: /山札から引く/ }));
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith('drawstock');
    });
  });

  it('renders the hand-penalty badge with the summed value when the player holds cards', async () => {
    // drawPhaseState has [♠7, ♥7] in hand → penalty = 7 + 7 = 14.
    mockExec.mockResolvedValue(drawPhaseState);
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => {
      const badge = screen.getByTestId('hand-penalty-badge');
      expect(badge).toBeInTheDocument();
      expect(badge).toHaveTextContent('14');
    });
  });

  it('hides the hand-penalty badge when the player has no cards', async () => {
    const emptyHandState: Rummy500Response = {
      ...drawPhaseState,
      players: [{ ...drawPhaseState.players[0], cardCount: 0, cards: [] }, drawPhaseState.players[1]],
    };
    mockExec.mockResolvedValue(emptyHandState);
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, expect.any(Object)));
    expect(screen.queryByTestId('hand-penalty-badge')).not.toBeInTheDocument();
  });

  // **押せるボタンは必ず通る。**どのメルドのボタンも常に押せて、置けるかどうかは
  // サーバー応答で初めて分かる状態だった (#4832)。
  it('only enables the melds the selected card can go onto', async () => {
    const withMelds = {
      ...playPhaseState,
      players: [
        playPhaseState.players[0],
        {
          ...playPhaseState.players[1],
          laidMelds: [
            {
              cards: [
                { design: 'SPADE' as const, value: 7 },
                { design: 'HEART' as const, value: 7 },
                { design: 'CLOVER' as const, value: 7 },
              ],
            },
            {
              cards: [
                { design: 'SPADE' as const, value: 2 },
                { design: 'SPADE' as const, value: 3 },
                { design: 'SPADE' as const, value: 4 },
              ],
            },
          ],
        },
      ],
      // 手札 0 枚目は 2 つ目のメルドにだけ置ける。
      layoffTargets: [[{ owner: 1, meldIdx: 1 }], []],
    };
    mockExec.mockResolvedValue(withMelds);
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => expect(screen.getByTestId('layoff-meld-1-0')).toBeInTheDocument());

    // 何も選んでいなければ従来どおり両方押せる (選び直しのため)。
    expect(screen.getByTestId('layoff-meld-1-0')).not.toBeDisabled();

    const handCard = document.querySelector('[data-tutorial="r5-player-hand"] button') as HTMLButtonElement;
    fireEvent.click(handCard);

    expect(screen.getByTestId('layoff-meld-1-0')).toBeDisabled();
    expect(screen.getByTestId('layoff-meld-1-1')).not.toBeDisabled();
    expect(screen.getByTestId('layoff-meld-1-1')).toHaveAttribute('data-layoff-legal', 'true');
  });

  // **選び直しで不正になった組み合わせを弾く。**カード A に合う先を選んだあと
  // 選択を B に変えても送信できてしまっていた (#4832 のレビュー指摘)。
  it('disables the layoff button when the chosen meld does not fit the newly selected card', async () => {
    const withMelds = {
      ...playPhaseState,
      players: [
        playPhaseState.players[0],
        {
          ...playPhaseState.players[1],
          laidMelds: [
            {
              cards: [
                { design: 'SPADE' as const, value: 7 },
                { design: 'HEART' as const, value: 7 },
                { design: 'CLOVER' as const, value: 7 },
              ],
            },
          ],
        },
      ],
      // 手札 0 枚目は置ける、1 枚目は置けない。
      layoffTargets: [[{ owner: 1, meldIdx: 0 }], []],
    };
    mockExec.mockResolvedValue(withMelds);
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => expect(screen.getByTestId('layoff-meld-1-0')).toBeInTheDocument());

    const handCards = document.querySelectorAll('[data-tutorial="r5-player-hand"] button');
    fireEvent.click(handCards[0] as HTMLElement);
    fireEvent.click(screen.getByTestId('layoff-meld-1-0'));
    expect(screen.getByRole('button', { name: 'レイオフ' })).toBeEnabled();

    // 選択を「置けない札」に変える。ボタンは無効に戻る。
    fireEvent.click(handCards[0] as HTMLElement);
    fireEvent.click(handCards[1] as HTMLElement);
    expect(screen.getByRole('button', { name: 'レイオフ' })).toBeDisabled();
  });
});

describe('Rummy500Page RoundScoreAnnouncement', () => {
  const scoredPlayers = [
    {
      id: 0,
      isHuman: true,
      cardCount: 0,
      cards: [],
      roundScore: 45,
      cumulativeScore: 120,
      laidMelds: [],
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 5,
      cards: [],
      roundScore: -10,
      cumulativeScore: 75,
      laidMelds: [],
    },
  ];

  beforeEach(() => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(drawPhaseState);
  });

  it('announces scores with exact formatted string in ROUND_END phase', async () => {
    const customRoundEndState: Rummy500Response = {
      ...roundEndState,
      players: scoredPlayers,
    };
    mockExec.mockResolvedValue(customRoundEndState);
    renderWithProviders(<Rummy500Page />);

    await waitFor(() => {
      const status = screen.getByRole('status');
      expect(status).toBeInTheDocument();
      expect(status).toHaveTextContent('ラウンド終了。あなた: +45 (合計 120), CPU 1: +-10 (合計 75)');
    });
  });

  it('announces scores with exact formatted string in GAME_END phase', async () => {
    const customGameEndState: Rummy500Response = {
      ...gameEndState,
      players: scoredPlayers,
    };
    mockExec.mockResolvedValue(customGameEndState);
    renderWithProviders(<Rummy500Page />);

    await waitFor(() => {
      const status = screen.getByRole('status');
      expect(status).toBeInTheDocument();
      expect(status).toHaveTextContent('ラウンド終了。あなた: +45 (合計 120), CPU 1: +-10 (合計 75)');
    });
  });

  it('renders empty live region in DRAW phase (negative control)', async () => {
    mockExec.mockResolvedValue({
      ...drawPhaseState,
      players: scoredPlayers,
    });
    renderWithProviders(<Rummy500Page />);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /山札から引く/ })).toBeInTheDocument();
    });
    const status = screen.getByRole('status');
    expect(status).toBeInTheDocument();
    expect(status).toBeEmptyDOMElement();
  });

  it('renders empty live region in PLAY phase (negative control)', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      players: scoredPlayers,
    });
    renderWithProviders(<Rummy500Page />);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /メルドする/ })).toBeInTheDocument();
    });
    const status = screen.getByRole('status');
    expect(status).toBeInTheDocument();
    expect(status).toBeEmptyDOMElement();
  });
});
