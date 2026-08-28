import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { callBreakApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeCallBreakState } from '../test/stateFactories';
import type { CallBreakResponse } from '../types/card';
import { CallBreakPage, fmtScore } from './CallBreakPage';

vi.mock('../api/gameApi', () => ({
  callBreakApi: { exec: vi.fn() },
  actionLogApi: { callbreak: vi.fn() },
}));

const mockExec = vi.mocked(callBreakApi.exec);

const playPhaseState = makeCallBreakState({ validPlayIndices: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12] });

const bidPhaseState = makeCallBreakState({
  phase: 0,
  bidPlayerIdx: 0,
  players: makeCallBreakState().players.map((p) => ({ ...p, bid: -1 })),
});

const bidPhaseCpuTurnState = makeCallBreakState({
  ...bidPhaseState,
  bidPlayerIdx: 1,
});

const trickEndState = makeCallBreakState({
  phase: 2,
  currentTrick: [
    { playerIdx: 0, card: { design: 'DIAMOND', value: 3 } },
    { playerIdx: 1, card: { design: 'HEART', value: 5 } },
  ],
});

const roundEndState = makeCallBreakState({ phase: 3 });

const gameEndState = makeCallBreakState({
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
});

const gameEndByFlagState = makeCallBreakState({
  phase: 1,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
});

const spadesBrokenState = makeCallBreakState({ spadesBroken: true });

const cpuTurnState = makeCallBreakState({ currentPlayerIdx: 1 });

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('fmtScore', () => {
  it('formats int×10 scores as X.Y for en/ja locales (period separator)', () => {
    expect(fmtScore(41, 'en')).toBe('4.1');
    expect(fmtScore(40, 'en')).toBe('4.0');
    expect(fmtScore(5, 'en')).toBe('0.5');
    expect(fmtScore(-41, 'en')).toBe('-4.1');
    expect(fmtScore(105, 'ja')).toBe('10.5');
  });

  it('uses the locale decimal separator for comma-decimal locales', () => {
    expect(fmtScore(41, 'de-DE')).toBe('4,1');
    expect(fmtScore(-5, 'de-DE')).toBe('-0,5');
  });
});

describe('CallBreakPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<CallBreakPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with default config', async () => {
    renderWithProviders(<CallBreakPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        maxRounds: 5,
      }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  // **バッグの式はドメインの GetBags() が唯一の出どころ (#4752)。**以前はこの
  // ページが bid と trickCount から自前で組み立てており、CUI 側と二重化していた。
  // ここで確かめるのは「サーバーの値をそのまま出すこと」— わざと式と矛盾する
  // bags を送り、ページが再計算していないことを踏む。
  it('renders the bags value the server sent instead of recomputing it', async () => {
    const bagsState = makeCallBreakState({
      players: makeCallBreakState().players.map((p, i) =>
        // bid 3 / 5 トリックなら式の上では 2 だが、サーバーは 7 と言っている。
        i === 0 ? { ...p, bid: 3, trickCount: 5, bags: 7 } : { ...p, bid: 4, trickCount: 1, bags: 0 },
      ),
    });
    mockExec.mockResolvedValue(bagsState);
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => expect(screen.getByTestId('cb-bags-counter')).toBeInTheDocument());
    expect(screen.getByTestId('cb-bags-0')).toHaveTextContent('7');
    expect(screen.getByTestId('cb-bags-1')).toHaveTextContent('0');
  });

  it('renders bid phase with a 1-13 bid button group', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'ビッド' })).toBeInTheDocument();
    });
    // 13 selectable bid options, defaulting to 1 pressed.
    expect(screen.getByTestId('bid-option-1')).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('bid-option-13')).toBeInTheDocument();
    expect(screen.queryByLabelText('bid-input')).not.toBeInTheDocument();
  });

  it('shows bid phase instruction when human bid turn', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => {
      expect(screen.getByText('ビッド宣言 (1-13)')).toBeInTheDocument();
    });
  });

  it('does not show bid instruction on CPU bid turn', async () => {
    mockExec.mockResolvedValue(bidPhaseCpuTurnState);
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => expect(screen.getAllByText(/CPU 1/).length).toBeGreaterThan(0));
    expect(screen.queryByText(/ビッド宣言/)).not.toBeInTheDocument();
  });

  it('calls bid command when bid button is clicked', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ビッド' })).toBeInTheDocument());

    // Select bid 5 from the button group.
    fireEvent.click(screen.getByTestId('bid-option-5'));
    expect(screen.getByTestId('bid-option-5')).toHaveAttribute('aria-pressed', 'true');
    // The live region reflects the current selection for screen readers.
    expect(screen.getByTestId('cb-bid-selected')).toHaveTextContent('5');

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ビッド' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 5));
  });

  it('play button disabled when not 1 card selected', async () => {
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeDisabled());
  });

  it('play button enabled when 1 card selected', async () => {
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '出す' })).not.toBeDisabled();
  });

  it('marks cards outside validPlayIndices as aria-disabled with tooltip on human play turn', async () => {
    const must = makeCallBreakState({ validPlayIndices: [0] });
    mockExec.mockResolvedValue(must);
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const allowed = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    const blocked = screen.getByAltText('♥ J').closest('button') as HTMLButtonElement;

    expect(allowed).not.toHaveAttribute('aria-disabled');
    expect(blocked).toHaveAttribute('aria-disabled', 'true');
    // Critical accessibility constraint: aria-disabled cards must stay focusable
    // so keyboard / screen-reader users can reach the tooltip explaining the rule.
    expect(blocked).not.toBeDisabled();
    expect(blocked).toHaveAttribute(
      'title',
      'このカードは出せません (リードスートに従うか、ボイドならスペードで切ってください)',
    );

    fireEvent.click(blocked);
    expect(screen.getByRole('button', { name: '出す' })).toBeDisabled();
  });

  it('does not show play button when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => expect(screen.getAllByText(/CPU 1/).length).toBeGreaterThan(0));
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('shows next trick button on trick end', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('shows next round button on round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('shows game end with action log button (phase 4)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-4 phase', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('renders spades-broken status text', async () => {
    mockExec.mockResolvedValue(spadesBrokenState);
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => expect(screen.getAllByText('スペードブレイク済').length).toBeGreaterThan(0));
  });

  it('shows a persistent spades-break indicator in the footer hand area', async () => {
    mockExec.mockResolvedValue(spadesBrokenState);
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => {
      expect(screen.getByTestId('cb-spades-break-footer')).toHaveTextContent('スペードブレイク済');
    });
  });

  it('shows decimal score in the score table (cumulativeScore 41 → 4.1)', async () => {
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => {
      expect(screen.getAllByText(/4\.1/).length).toBeGreaterThan(0);
    });
  });

  it('shows error alert when API rejects', async () => {
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('settings panel changes cpuDifficulty', async () => {
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => expect(screen.getByText('コールブレイク')).toBeInTheDocument());

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
        maxRounds: 5,
      }),
    );
  });

  it('settings panel changes maxRounds', async () => {
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => expect(screen.getByText('コールブレイク')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: '10' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        maxRounds: 10,
      }),
    );
  });

  // このページのヒントは `hintAvailable` を使わないので、読み上げガードが
  // 一度も見ておらず aria-live が無いまま出荷されていた (#6663)。領域は
  // **常設**で、分岐ごとにその分岐だけが出す語を見る。
  describe('CallBreakPage hint live region', () => {
    it('is mounted and empty before any hint arrives', async () => {
      mockExec.mockResolvedValue(playPhaseState);
      renderWithProviders(<CallBreakPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());

      const region = screen.getByTestId('callbreak-hint-live');
      expect(region).toHaveAttribute('role', 'status');
      expect(region).toHaveAttribute('aria-live', 'polite');
      expect(region).toBeEmptyDOMElement();
    });

    it('names a bid recommendation', async () => {
      mockExec.mockResolvedValue(bidPhaseState);
      renderWithProviders(<CallBreakPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

      mockExec.mockResolvedValue({
        ...bidPhaseState,
        hint: { bid: 4, reason: 'strategic_bid' },
      } as unknown as CallBreakResponse);
      fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

      const region = await screen.findByTestId('callbreak-hint-live');
      await waitFor(() => expect(region).toHaveTextContent('推奨ビッド'));
      expect(region).toHaveTextContent('4');
      expect(region).not.toHaveTextContent('{{');
    });

    it('names the card to play', async () => {
      mockExec.mockResolvedValue(playPhaseState);
      renderWithProviders(<CallBreakPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

      mockExec.mockResolvedValue({
        ...playPhaseState,
        hint: { cardIndex: 2, reason: 'lead_strong' },
      } as unknown as CallBreakResponse);
      fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

      const region = await screen.findByTestId('callbreak-hint-live');
      await waitFor(() => expect(region).toHaveTextContent('推奨'));
      expect(region).toHaveTextContent('[2]');
      expect(region).not.toHaveTextContent('推奨ビッド');
      expect(region).not.toHaveTextContent('{{');
    });
  });
});
