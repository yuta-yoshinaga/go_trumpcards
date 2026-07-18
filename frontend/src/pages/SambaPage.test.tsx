import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { sambaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeSambaState } from '../test/stateFactories';
import { SambaPage } from './SambaPage';

vi.mock('../api/gameApi', () => ({
  sambaApi: { exec: vi.fn() },
  actionLogApi: { samba: vi.fn() },
}));
const mockExec = vi.mocked(sambaApi.exec);

const drawPhaseState = makeSambaState();
const meldPhaseState = makeSambaState({ phase: 1, messageCode: 'samba.meldPhase' });
const discardPhaseState = makeSambaState({ phase: 2, messageCode: 'samba.discardPhase' });
const roundEndState = makeSambaState({ phase: 3, messageCode: 'samba.roundEnd' });
const gameEndState = makeSambaState({ phase: 4, gameEndFlag: true, winnerIdx: 0 });
// Not the human's turn — CPU (seat 1) is active during the draw phase.
const cpuTurnState = makeSambaState({ currentPlayerIdx: 1 });

describe('SambaPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(drawPhaseState);
  });

  afterEach(() => {
    localStorage.clear();
  });

  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SambaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<SambaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 10000 }),
    );
  });

  it('shows draw phase buttons and team scores', async () => {
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument();
    expect(screen.getByTestId('sa-team-scores')).toBeInTheDocument();
  });

  it('announces when the discard pile becomes frozen', async () => {
    renderWithProviders(<SambaPage />);
    const announce = await screen.findByTestId('sa-frozen-announce');
    expect(announce).toHaveAttribute('role', 'status');
    expect(announce).toHaveAttribute('aria-live', 'polite');
    expect(announce).toHaveTextContent(''); // no transition yet
    // A draw resolves to a frozen state → isFrozen false→true triggers the announcement.
    mockExec.mockResolvedValue(makeSambaState({ isFrozen: true }));
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(screen.getByTestId('sa-frozen-announce')).toHaveTextContent('捨札が凍結されました'));
  });

  it('calls drawstock command when button clicked', async () => {
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('hides action controls when it is not the human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '山札から引く' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '捨て札を取る' })).not.toBeInTheDocument();
  });

  it('shows meld phase buttons', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'メルドする' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument();
  });

  it('shows the initial-meld minimum and selected total in the meld phase', async () => {
    mockExec.mockResolvedValue(meldPhaseState); // team score 0 → min 50; hasInitMeld false
    renderWithProviders(<SambaPage />);
    const info = await screen.findByTestId('sa-meld-points');
    expect(info).toHaveTextContent('初回メルド最低点: 50');
    expect(info).toHaveTextContent('選択合計: 0');
  });

  it('calls skipmeld command when skip button clicked', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skipmeld'));
  });

  it('shows discard phase buttons', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '上がる' })).toBeInTheDocument();
  });

  it('shows next round button at round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('renders hint toggle checkbox', async () => {
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled in draw phase', async () => {
    localStorage.setItem('hint_enabled_samba', 'true');
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('shows 次のゲーム at game-end and fires reset directly (no confirm)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 10000 }),
    );
    expect(screen.queryByRole('button', { name: '確認' })).not.toBeInTheDocument();
  });

  it('shows the disabled reason for draw-from-discard when no cards are selected', async () => {
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    expect(screen.getByTestId('sa-draw-discard-reason')).toHaveTextContent(
      '手札からトップカードと同ランクの2枚を選択してください',
    );
  });

  it('renders the frozen badge and reason when the discard pile is frozen', async () => {
    mockExec.mockResolvedValue(makeSambaState({ isFrozen: true }));
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByTestId('sa-frozen-badge')).toBeInTheDocument());
    expect(screen.getByTestId('sa-draw-discard-reason')).toHaveTextContent(
      'フリーズ中はワイルドカードでの代用ができません',
    );
  });

  it('enables the draw button and clears the reason when exactly 2 cards are selected', async () => {
    renderWithProviders(<SambaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    expect(screen.queryByTestId('sa-draw-discard-reason')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '捨て札を取る' })).not.toBeDisabled();
  });

  it('shows canasta/samba progress and a completion pulse per meld', async () => {
    const base = makeSambaState();
    const human = {
      ...base.players[0],
      melds: [
        // Incomplete set: 5 cards → 2 more for a canasta.
        {
          cards: Array.from({ length: 5 }, () => ({ design: 'SPADE' as const, value: 4 })),
          kind: 0,
          isNatural: true,
          isCanasta: false,
          isSamba: false,
          rank: 4,
        },
        // Completed 7-card sequence → samba, with pulse emphasis.
        {
          cards: Array.from({ length: 7 }, (_, i) => ({ design: 'HEART' as const, value: i + 3 })),
          kind: 1,
          isNatural: true,
          isCanasta: false,
          isSamba: true,
          rank: 3,
        },
      ],
    };
    mockExec.mockResolvedValue(makeSambaState({ players: [human, ...base.players.slice(1)] }));
    renderWithProviders(<SambaPage />);

    const setProgress = await screen.findByTestId('sa-meld-progress-0-0');
    expect(setProgress).toHaveTextContent('あと2枚でカナスタ');
    expect(setProgress.className).not.toContain('animate-pulse');

    const sambaProgress = screen.getByTestId('sa-meld-progress-0-1');
    expect(sambaProgress).toHaveTextContent('サンバ成立！');
    expect(sambaProgress.className).toContain('animate-pulse');
  });
});
