import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { handandfootApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { HandAndFootPlayerData, HandAndFootResponse, HandAndFootTeamData } from '../types/card';
import { HandAndFootPage } from './HandAndFootPage';

vi.mock('../api/gameApi', () => ({
  handandfootApi: { exec: vi.fn() },
  actionLogApi: { handandfoot: vi.fn() },
}));
const mockExec = vi.mocked(handandfootApi.exec);

const basePlayers: HandAndFootPlayerData[] = [
  {
    id: 0,
    team: 0,
    isHuman: true,
    cardCount: 15,
    cards: [
      { design: 'SPADE', value: 7 },
      { design: 'CLOVER', value: 7 },
      { design: 'HEART', value: 7 },
      { design: 'SPADE', value: 10 },
      { design: 'CLOVER', value: 10 },
    ],
    footCount: 11,
    inFoot: false,
    roundScore: 0,
    cumulativeScore: 0,
  },
  {
    id: 1,
    team: 1,
    isHuman: false,
    cardCount: 15,
    cards: [],
    footCount: 11,
    inFoot: false,
    roundScore: 0,
    cumulativeScore: 0,
  },
];

const baseTeams: HandAndFootTeamData[] = [
  { team: 0, melds: [], red3Count: 0, red3s: [] },
  { team: 1, melds: [], red3Count: 0, red3s: [] },
];

const drawPhaseState: HandAndFootResponse = {
  players: basePlayers,
  teams: baseTeams,
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'SPADE', value: 5 },
  drawPileCount: 67,
  discardPileCount: 1,
  isFrozen: false,
  gameEndFlag: false,
  winnerTeam: -1,
  message: '',
  messageCode: 'handandfoot.drawPhase',
  config: { cpuDifficulty: 1, pointLimit: 5000 },
};

const meldPhaseState: HandAndFootResponse = {
  ...drawPhaseState,
  phase: 1,
  messageCode: 'handandfoot.meldPhase',
};

const discardPhaseState: HandAndFootResponse = {
  ...drawPhaseState,
  phase: 2,
  messageCode: 'handandfoot.discardPhase',
};

// A discard-phase state where the human has met every go-out requirement:
// entered their foot and their team holds both a natural and a mixed canasta.
const goOutReadyState: HandAndFootResponse = {
  ...discardPhaseState,
  players: [{ ...basePlayers[0], inFoot: true }, basePlayers[1]],
  teams: [
    {
      team: 0,
      melds: [
        { cards: [], isNatural: true, isCanasta: true, rank: 7 },
        { cards: [], isNatural: false, isCanasta: true, rank: 10 },
      ],
      red3Count: 0,
      red3s: [],
    },
    { team: 1, melds: [], red3Count: 0, red3s: [] },
  ],
};

// Human is in their foot and has a natural canasta but no mixed canasta yet.
const goOutNeedBlackState: HandAndFootResponse = {
  ...discardPhaseState,
  players: [{ ...basePlayers[0], inFoot: true }, basePlayers[1]],
  teams: [
    {
      team: 0,
      melds: [{ cards: [], isNatural: true, isCanasta: true, rank: 7 }],
      red3Count: 0,
      red3s: [],
    },
    { team: 1, melds: [], red3Count: 0, red3s: [] },
  ],
};

const roundEndState: HandAndFootResponse = {
  ...drawPhaseState,
  phase: 3,
  messageCode: 'handandfoot.roundEnd',
};

const gameEndState: HandAndFootResponse = {
  ...drawPhaseState,
  phase: 4,
  gameEndFlag: true,
  winnerTeam: 0,
};

describe('HandAndFootPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(drawPhaseState);
  });

  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<HandAndFootPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 1,
        pointLimit: 5000,
      }),
    );
  });

  it('shows draw phase buttons', async () => {
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument();
  });

  it('pulses the draw-from-discard button only when frozen and two cards are selected', async () => {
    mockExec.mockResolvedValue({ ...drawPhaseState, isFrozen: true });
    renderWithProviders(<HandAndFootPage />);
    const btn = await screen.findByRole('button', { name: '捨て札を取る' });
    // Disabled (no selection) → no misleading pulse yet.
    expect(btn).not.toHaveAttribute('data-frozen');
    // Select two hand cards to enable the action; the pulse then warns about the freeze.
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    const enabled = screen.getByRole('button', { name: '捨て札を取る' });
    expect(enabled).toHaveAttribute('data-frozen', 'true');
    expect(enabled.className).toMatch(/animate-pulse/);
  });

  it('does not pulse the draw-from-discard button when not frozen', async () => {
    renderWithProviders(<HandAndFootPage />);
    const btn = await screen.findByRole('button', { name: '捨て札を取る' });
    expect(btn).not.toHaveAttribute('data-frozen');
    expect(btn.className).not.toMatch(/animate-pulse/);
  });

  it('calls drawstock command when button clicked', async () => {
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('shows meld phase buttons', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'メルドする' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument();
  });

  it('shows the initial-meld minimum and updates the running total as cards are selected', async () => {
    // Team 0 has no melds and cumulative score 0 -> initial-meld minimum is 50.
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByTestId('hf-meld-points')).toBeInTheDocument());
    const readout = screen.getByTestId('hf-meld-points');
    // Nothing selected: total 0, below the 50 minimum -> warning styling.
    expect(readout).toHaveTextContent('初回メルド最低点: 50 / 選択合計: 0');
    expect(readout.className).toContain('text-ds-warning');

    // Select the two 10s (10 points each) -> running total 20.
    fireEvent.click(screen.getByRole('button', { name: '♠ 10' }));
    fireEvent.click(screen.getByRole('button', { name: '♣ 10' }));
    await waitFor(() => expect(screen.getByTestId('hf-meld-points')).toHaveTextContent('選択合計: 20'));
    expect(screen.getByTestId('hf-meld-points')).toHaveTextContent('初回メルド最低点: 50 / 選択合計: 20');
  });

  it('highlights the readout in success when the selection meets the minimum', async () => {
    // Negative cumulative score -> minimum 15; selecting the two 10s totals 20.
    mockExec.mockResolvedValue({
      ...meldPhaseState,
      players: [{ ...basePlayers[0], cumulativeScore: -10 }, basePlayers[1]],
    });
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByTestId('hf-meld-points')).toBeInTheDocument());
    // Nothing selected yet: 0 < 15 -> warning.
    expect(screen.getByTestId('hf-meld-points').className).toContain('text-ds-warning');
    fireEvent.click(screen.getByRole('button', { name: '♠ 10' }));
    fireEvent.click(screen.getByRole('button', { name: '♣ 10' }));
    await waitFor(() => expect(screen.getByTestId('hf-meld-points')).toHaveTextContent('選択合計: 20'));
    expect(screen.getByTestId('hf-meld-points').className).toContain('text-ds-success');
  });

  it('shows only the selected total once the team has already opened', async () => {
    mockExec.mockResolvedValue({
      ...meldPhaseState,
      teams: [
        { team: 0, melds: [{ cards: [], isNatural: true, isCanasta: false, rank: 7 }], red3Count: 0, red3s: [] },
        { team: 1, melds: [], red3Count: 0, red3s: [] },
      ],
    });
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByTestId('hf-meld-points')).toBeInTheDocument());
    const readout = screen.getByTestId('hf-meld-points');
    expect(readout).toHaveTextContent('選択合計: 0');
    expect(readout).not.toHaveTextContent('初回メルド最低点');
    expect(readout.className).toContain('text-ds-text-muted');
  });

  it('calls skipmeld command when skip button clicked', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skipmeld'));
  });

  it('shows discard phase buttons', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '上がる' })).toBeInTheDocument();
  });

  it('disables go out and explains the unmet requirement when the player is not in their foot', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '上がる' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '上がる' })).toBeDisabled();
    expect(screen.getByTestId('hf-go-out-guidance')).toHaveTextContent('まだフットに入っていないため上がれません');
  });

  it('shows the missing canasta reason when the player is in foot but lacks a mixed canasta', async () => {
    mockExec.mockResolvedValue(goOutNeedBlackState);
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByTestId('hf-go-out-guidance')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '上がる' })).toBeDisabled();
    expect(screen.getByTestId('hf-go-out-guidance')).toHaveTextContent('黒（ミックス）カナスタが不足しています');
  });

  it('enables go out and confirms readiness once every requirement is met', async () => {
    mockExec.mockResolvedValue(goOutReadyState);
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByTestId('hf-go-out-guidance')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '上がる' })).not.toBeDisabled();
    expect(screen.getByTestId('hf-go-out-guidance')).toHaveTextContent('上がれます');
  });

  it('calls goout command when go out is clicked while ready', async () => {
    mockExec.mockResolvedValue(goOutReadyState);
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '上がる' })).not.toBeDisabled());
    mockExec.mockClear();
    mockExec.mockResolvedValue(roundEndState);
    fireEvent.click(screen.getByRole('button', { name: '上がる' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('goout'));
  });

  it('shows next round button at round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('shows win celebration at game end', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
  });

  it('shows 次のゲーム at game-end and fires reset directly (no confirm)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 1,
        pointLimit: 5000,
      }),
    );
    expect(screen.queryByRole('button', { name: '確認' })).not.toBeInTheDocument();
  });

  it('shows the disabled reason for draw-from-discard when no cards are selected', async () => {
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const reason = screen.getByTestId('hf-draw-discard-reason');
    expect(reason).toHaveTextContent('手札からトップカードと同ランクの2枚を選択してください');
  });

  it('renders the frozen badge and reason when the discard pile is frozen', async () => {
    mockExec.mockResolvedValue({ ...drawPhaseState, isFrozen: true });
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByTestId('hf-frozen-badge')).toBeInTheDocument());
    expect(screen.getByTestId('hf-draw-discard-reason')).toHaveTextContent(
      'フリーズ中はワイルドカードでの代用ができません',
    );
  });

  it('switches the reason to selectOneMore once the player selects one card', async () => {
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByTestId('hf-draw-discard-reason')).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    expect(screen.getByTestId('hf-draw-discard-reason')).toHaveTextContent('もう1枚選択してください');
  });

  it('clears the reason and enables the draw button when exactly 2 cards are selected', async () => {
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    expect(screen.queryByTestId('hf-draw-discard-reason')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '捨て札を取る' })).not.toBeDisabled();
    expect(screen.getByRole('button', { name: '捨て札を取る' })).not.toHaveAttribute('aria-describedby');
  });

  it('warns when more than 2 cards are selected', async () => {
    renderWithProviders(<HandAndFootPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    fireEvent.click(handCards[2]);
    expect(screen.getByTestId('hf-draw-discard-reason')).toHaveTextContent('選択は2枚までです');
    expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeDisabled();
  });

  afterEach(() => {
    localStorage.clear();
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も
  // 実行されない（#4596 / #4600 のレビュー指摘）。
  it('turns the frontend hint on from the settings panel', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(drawPhaseState);
    renderWithProviders(<HandAndFootPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });
});
