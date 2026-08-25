import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { boliviaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeBoliviaState } from '../test/stateFactories';
import { BoliviaPage } from './BoliviaPage';

vi.mock('../api/gameApi', () => ({
  boliviaApi: { exec: vi.fn() },
  actionLogApi: { bolivia: vi.fn() },
}));
const mockExec = vi.mocked(boliviaApi.exec);

const drawPhaseState = makeBoliviaState();
const meldPhaseState = makeBoliviaState({ phase: 1, messageCode: 'bolivia.meldPhase' });
const discardPhaseState = makeBoliviaState({ phase: 2, messageCode: 'bolivia.discardPhase' });
const roundEndState = makeBoliviaState({ phase: 3, messageCode: 'bolivia.roundEnd' });
const gameEndState = makeBoliviaState({ phase: 4, gameEndFlag: true, winnerIdx: 0 });
// Not the human's turn — CPU (seat 1) is active during the draw phase.
const cpuTurnState = makeBoliviaState({ currentPlayerIdx: 1 });

describe('BoliviaPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(drawPhaseState);
  });

  afterEach(() => {
    localStorage.clear();
  });

  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BoliviaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<BoliviaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 10000 }),
    );
  });

  it('shows draw phase buttons and team scores', async () => {
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument();
    expect(screen.getByTestId('sa-team-scores')).toBeInTheDocument();
  });

  it('announces when the discard pile becomes frozen', async () => {
    renderWithProviders(<BoliviaPage />);
    const announce = await screen.findByTestId('sa-frozen-announce');
    expect(announce).toHaveAttribute('role', 'status');
    expect(announce).toHaveAttribute('aria-live', 'polite');
    expect(announce).toHaveTextContent(''); // no transition yet
    // A draw resolves to a frozen state → isFrozen false→true triggers the announcement.
    mockExec.mockResolvedValue(makeBoliviaState({ isFrozen: true }));
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(screen.getByTestId('sa-frozen-announce')).toHaveTextContent('捨札が凍結されました'));
  });

  it('calls drawstock command when button clicked', async () => {
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('hides action controls when it is not the human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '山札から引く' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '捨て札を取る' })).not.toBeInTheDocument();
  });

  it('shows meld phase buttons', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'メルドする' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument();
  });

  it('shows the initial-meld minimum and selected total in the meld phase', async () => {
    mockExec.mockResolvedValue(meldPhaseState); // team score 0 → min 50; hasInitMeld false
    renderWithProviders(<BoliviaPage />);
    const info = await screen.findByTestId('sa-meld-points');
    expect(info).toHaveTextContent('初回メルド最低点: 50');
    expect(info).toHaveTextContent('選択合計: 0');
  });

  it('calls skipmeld command when skip button clicked', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skipmeld'));
  });

  it('shows discard phase buttons', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '上がる' })).toBeInTheDocument();
  });

  it('shows next round button at round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('renders hint toggle checkbox', async () => {
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled in draw phase', async () => {
    localStorage.setItem('hint_enabled_bolivia', 'true');
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('shows 次のゲーム at game-end and fires reset directly (no confirm)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 10000 }),
    );
    expect(screen.queryByRole('button', { name: '確認' })).not.toBeInTheDocument();
  });

  it('shows the disabled reason for draw-from-discard when no cards are selected', async () => {
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    expect(screen.getByTestId('sa-draw-discard-reason')).toHaveTextContent(
      '手札からトップカードと同ランクの2枚を選択してください',
    );
  });

  it('renders the frozen badge and reason when the discard pile is frozen', async () => {
    mockExec.mockResolvedValue(makeBoliviaState({ isFrozen: true }));
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByTestId('sa-frozen-badge')).toBeInTheDocument());
    expect(screen.getByTestId('sa-draw-discard-reason')).toHaveTextContent(
      'フリーズ中はワイルドカードでの代用ができません',
    );
  });

  it('enables the draw button and clears the reason when exactly 2 cards are selected', async () => {
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    expect(screen.queryByTestId('sa-draw-discard-reason')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '捨て札を取る' })).not.toBeDisabled();
  });

  it('shows canasta/bolivia progress and a completion pulse per meld', async () => {
    const base = makeBoliviaState();
    const human = {
      ...base.players[0],
      melds: [
        // Incomplete set: 5 cards → 2 more for a canasta.
        {
          cards: Array.from({ length: 5 }, () => ({ design: 'SPADE' as const, value: 4 })),
          kind: 0,
          isNatural: true,
          isCanasta: false,
          isEscalera: false,
          isBolivia: false,
          rank: 4,
        },
        // Completed 7-card sequence → bolivia, with pulse emphasis.
        {
          cards: Array.from({ length: 7 }, (_, i) => ({ design: 'HEART' as const, value: i + 3 })),
          kind: 1,
          isNatural: true,
          isCanasta: false,
          isEscalera: false,
          isBolivia: true,
          rank: 3,
        },
      ],
    };
    mockExec.mockResolvedValue(makeBoliviaState({ players: [human, ...base.players.slice(1)] }));
    renderWithProviders(<BoliviaPage />);

    const setProgress = await screen.findByTestId('sa-meld-progress-0-0');
    expect(setProgress).toHaveTextContent('あと2枚でカナスタ');
    expect(setProgress.className).not.toContain('animate-pulse');

    const escaleraProgress = screen.getByTestId('sa-meld-progress-0-1');
    // **7 枚のシーケンスはエスカレラ。** ボリビアはワイルド 7 枚のほう。
    expect(escaleraProgress).toHaveTextContent('エスカレラ成立！');
    expect(escaleraProgress.className).toContain('animate-pulse');
  });

  it('localizes the CPU hand-count label at round end', async () => {
    // The label is split across text nodes, so match on the row's textContent.
    const cpuRows = (re: RegExp) =>
      screen.queryAllByText((_, el) => el?.tagName === 'DIV' && re.test(el.textContent ?? ''));
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<BoliviaPage />);
    await waitFor(() => expect(cpuRows(/^CPU \d+: \d+枚$/).length).toBeGreaterThan(0));
    expect(cpuRows(/^CPU \d+: \d+ cards$/)).toHaveLength(0);
  });

  // **メルドは 3 種類あり、進捗の言い回しも 3 つ。**
  //
  // クローン元 (Samba) はセットとシーケンスの 2 分岐しか持たないので、
  // そのままだと**ワイルドだけのメルドが「カナスタ」と表示される** ── この
  // ゲームの名前になっている役が、別の役の名前で出てしまう。
  it('names each of the three meld kinds, including the wild-only one', async () => {
    const base = makeBoliviaState();
    const seven = (design: 'SPADE' | 'HEART', from: number) =>
      Array.from({ length: 7 }, (_, i) => ({ design, value: from + i }));
    mockExec.mockResolvedValue(
      makeBoliviaState({
        players: base.players.map((p, i) =>
          i === 0
            ? {
                ...p,
                melds: [
                  {
                    cards: seven('SPADE', 4),
                    kind: 1,
                    isNatural: true,
                    isCanasta: false,
                    isEscalera: true,
                    isBolivia: false,
                    rank: 4,
                  },
                  {
                    cards: [
                      { design: 'SPADE', value: 2 },
                      { design: 'HEART', value: 2 },
                      { design: 'JOKER', value: 0 },
                      { design: 'CLOVER', value: 2 },
                      { design: 'DIAMOND', value: 2 },
                      { design: 'JOKER', value: 0 },
                      { design: 'JOKER', value: 0 },
                    ],
                    kind: 2,
                    isNatural: false,
                    isCanasta: false,
                    isEscalera: false,
                    isBolivia: true,
                    rank: 0,
                  },
                ],
              }
            : p,
        ),
      }),
    );
    renderWithProviders(<BoliviaPage />);
    expect(await screen.findByTestId('sa-meld-progress-0-0')).toHaveTextContent('エスカレラ成立！');
    const wild = screen.getByTestId('sa-meld-progress-0-1');
    expect(wild).toHaveTextContent('ボリビア成立！');
    // 負のコントロール: ワイルドのメルドをカナスタと呼ばない。
    expect(wild).not.toHaveTextContent('カナスタ');
  });

  // **ラベルの側も 3 分岐であること (レビュー指摘)。** 進捗表示だけを直して
  // ラベルを 2 分岐のまま残していたので、完成したボリビアが枚数表示に
  // 落ちて、画面のどこにも「ボリビア」と出ていなかった。
  it('labels a completed wild meld a bolivia in the meld label, not just the progress', async () => {
    const base = makeBoliviaState();
    mockExec.mockResolvedValue(
      makeBoliviaState({
        players: base.players.map((p, i) =>
          i === 0
            ? {
                ...p,
                melds: [
                  {
                    cards: [
                      { design: 'SPADE', value: 2 },
                      { design: 'HEART', value: 2 },
                      { design: 'CLOVER', value: 2 },
                      { design: 'DIAMOND', value: 2 },
                      { design: 'JOKER', value: 0 },
                      { design: 'JOKER', value: 0 },
                      { design: 'JOKER', value: 0 },
                    ],
                    kind: 2,
                    isNatural: false,
                    isCanasta: false,
                    isEscalera: false,
                    isBolivia: true,
                    rank: 0,
                  },
                ],
              }
            : p,
        ),
      }),
    );
    renderWithProviders(<BoliviaPage />);
    const area = await screen.findByTestId('sa-meld-progress-0-0');
    const row = area.parentElement as HTMLElement;
    expect(row).toHaveTextContent('ボリビア');
    // 負のコントロール: 枚数だけの表示に落ちていないこと。
    expect(row).not.toHaveTextContent('(7)');
  });
});
