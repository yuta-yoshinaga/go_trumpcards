import { fireEvent, screen, waitFor } from '@testing-library/react';
import i18n from 'i18next';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { escobaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeEscobaState } from '../test/stateFactories';
import { EscobaPage, escobaCardValue, escobaSelectionSum } from './EscobaPage';

vi.mock('../api/gameApi', () => ({
  escobaApi: { exec: vi.fn() },
  actionLogApi: { escoba: vi.fn() },
}));

const mockExec = vi.mocked(escobaApi.exec);

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeEscobaState());
});

afterEach(async () => {
  await i18n.changeLanguage('ja');
});

describe('EscobaPage', () => {
  it('calls reset on mount with the short "r" command', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('r'));
  });

  it('renders CPU difficulty options with localized labels', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // Difficulty options are localized (ja), not the hardcoded Easy/Normal/Hard.
    expect(screen.getByRole('option', { name: 'かんたん' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'ふつう' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'むずかしい' })).toBeInTheDocument();
  });

  it('renders the human hand', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('hand-card-1')).toBeInTheDocument();
    expect(screen.getByTestId('hand-card-2')).toBeInTheDocument();
  });

  it('renders table cards', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('table-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('table-card-1')).toBeInTheDocument();
  });

  it('renders per-player scores and stock', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('player-score-0')).toBeInTheDocument());
    expect(screen.getByTestId('player-score-1')).toBeInTheDocument();
    expect(screen.getByTestId('stock-remaining')).toBeInTheDocument();
  });

  it('take button is disabled until both hand and table are selected', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('take-button')).toBeInTheDocument());
    expect(screen.getByTestId('take-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('hand-card-0'));
    expect(screen.getByTestId('take-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('table-card-0'));
    await waitFor(() => expect(screen.getByTestId('take-button')).not.toBeDisabled());
  });

  it('lay button is enabled when a hand card is selected and no table card', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('lay-button')).toBeInTheDocument());
    expect(screen.getByTestId('lay-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('hand-card-0'));
    await waitFor(() => expect(screen.getByTestId('lay-button')).not.toBeDisabled());
  });

  it('plays "p" with sorted table indices when Take is clicked', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('table-card-1'));
    fireEvent.click(screen.getByTestId('table-card-0'));
    fireEvent.click(screen.getByTestId('take-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('p', { handIndex: 0, tableIndices: [0, 1] }));
  });

  it('plays "p" with empty table indices when Lay is clicked', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('lay-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('p', { handIndex: 0, tableIndices: [] }));
  });

  it('disables actions when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeEscobaState({ currentTurn: 1, isHumanTurn: false }));
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('take-button')).toBeDisabled());
    expect(screen.getByTestId('lay-button')).toBeDisabled();
  });

  it('shows the round-end breakdown and a next-round button on roundEnd', async () => {
    mockExec.mockResolvedValue(
      makeEscobaState({
        phase: 'roundEnd',
        isHumanTurn: false,
        lastRoundDetail: {
          cards: [1, 0, 0, 0],
          espadas: [1, 0, 0, 0],
          sevens: [0, 1, 0, 0],
          oros: [1, 0, 0, 0],
          escobas: [1, 0, 0, 0],
          gained: [4, 1, 0, 0],
          aceEspada: 0,
          seteEspada: 1,
        },
      }),
    );
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('round-detail')).toBeInTheDocument());
    expect(screen.getByTestId('next-round-button')).toBeInTheDocument();
  });

  it('next-round button dispatches "n"', async () => {
    mockExec.mockResolvedValue(makeEscobaState({ phase: 'roundEnd', isHumanTurn: false }));
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('next-round-button')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('next-round-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('n'));
  });

  it('resets with config and passes it to the API', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('r', {
        config: { targetScore: 10, cpuDifficulty: 1 },
      }),
    );
  });

  it('changes CPU difficulty and includes it in the reset config', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    const select = screen.getByLabelText(/CPU難易度|CPU Difficulty/);
    fireEvent.change(select, { target: { value: '2' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'r',
        expect.objectContaining({ config: expect.objectContaining({ cpuDifficulty: 2 }) }),
      ),
    );
  });

  it('shows loading state when state has fewer than 4 players', async () => {
    mockExec.mockResolvedValue(
      makeEscobaState({
        players: [
          {
            id: 0,
            isHuman: true,
            handCount: 0,
            cards: [],
            capturedCount: 0,
            capturedCards: [],
            escobaCount: 0,
            score: 0,
          },
        ],
      }),
    );
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.queryByTestId('hand-card-0')).not.toBeInTheDocument());
  });

  it('renders the human captured-cards viewer with the captured pile', async () => {
    mockExec.mockResolvedValue(
      makeEscobaState({
        players: [
          {
            id: 0,
            isHuman: true,
            handCount: 3,
            cards: [
              { design: 'SPADE', value: 7 },
              { design: 'HEART', value: 5 },
              { design: 'DIAMOND', value: 11 },
            ],
            capturedCount: 2,
            capturedCards: [
              { design: 'SPADE', value: 4 },
              { design: 'HEART', value: 3 },
            ],
            escobaCount: 0,
            score: 0,
          },
          {
            id: 1,
            isHuman: false,
            handCount: 3,
            cards: [],
            capturedCount: 5,
            capturedCards: [],
            escobaCount: 0,
            score: 0,
          },
          {
            id: 2,
            isHuman: false,
            handCount: 3,
            cards: [],
            capturedCount: 0,
            capturedCards: [],
            escobaCount: 0,
            score: 0,
          },
          {
            id: 3,
            isHuman: false,
            handCount: 3,
            cards: [],
            capturedCount: 0,
            capturedCards: [],
            escobaCount: 0,
            score: 0,
          },
        ],
      }),
    );
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('captured-viewer')).toBeInTheDocument());
    // Summary reflects the human's captured count.
    expect(screen.getByText('獲得した札を見る（2枚）')).toBeInTheDocument();
    // The captured pile renders the actual cards.
    expect(screen.getByTestId('captured-cards')).toBeInTheDocument();
  });

  it('shows the empty-pile message when the human has captured nothing', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('captured-viewer')).toBeInTheDocument());
    expect(screen.getByText('まだ札を獲得していません')).toBeInTheDocument();
    expect(screen.queryByTestId('captured-cards')).not.toBeInTheDocument();
  });

  it('renders CLI terminal when CLI mode is enabled via localStorage', async () => {
    localStorage.setItem('cli-mode-escoba', 'true');
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    localStorage.removeItem('cli-mode-escoba');
  });

  it('hides the 15-counter until a hand card is selected, then shows a running sum', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.queryByTestId('escoba-sum-indicator')).not.toBeInTheDocument();
    // hand[0] = ♠7 (value 7); no table cards selected yet.
    fireEvent.click(screen.getByTestId('hand-card-0'));
    const counter = await screen.findByTestId('escoba-sum-indicator');
    expect(counter).toHaveTextContent('7 / 15');
    expect(counter.className).toContain('text-ds-text-muted');
  });

  it('turns the counter success-green when the selection sums to exactly 15', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('hand-card-0')); // 7
    fireEvent.click(screen.getByTestId('table-card-0')); // +4
    fireEvent.click(screen.getByTestId('table-card-1')); // +4 => 15
    const counter = await screen.findByTestId('escoba-sum-indicator');
    expect(counter).toHaveTextContent('15 / 15');
    expect(counter.className).toContain('text-ds-success');
  });

  it('turns the counter error-red when the selection exceeds 15', async () => {
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-2')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('hand-card-2')); // ♦11 => value 8
    fireEvent.click(screen.getByTestId('table-card-0')); // +4
    fireEvent.click(screen.getByTestId('table-card-1')); // +4 => 16
    const counter = await screen.findByTestId('escoba-sum-indicator');
    expect(counter).toHaveTextContent('16 / 15');
    expect(counter.className).toContain('text-ds-error');
  });

  it('does not show the escoba badge on initial load', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(makeEscobaState());
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.queryByTestId('escoba-celebration')).not.toBeInTheDocument();
  });

  it('flashes the badge when a sweep lands, and clears it when the round resets', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(makeEscobaState());
    renderWithProviders(<EscobaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    // The human sweeps: escobaCount rises from 0 to 1.
    const swept = makeEscobaState();
    swept.players = swept.players.map((p) => (p.isHuman ? { ...p, escobaCount: 1 } : p));
    mockExec.mockResolvedValue(swept);
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('table-card-0'));
    fireEvent.click(screen.getByTestId('take-button'));
    const badge = await screen.findByTestId('escoba-celebration');
    // Escoba has no teams, so the emphasised label is the human's own sweep.
    expect(badge).toHaveTextContent('エスコバ！ あなた');

    // A new round drops the count back to 0; the badge must not linger or re-fire.
    mockExec.mockResolvedValue(makeEscobaState());
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('table-card-0'));
    fireEvent.click(screen.getByTestId('take-button'));
    await waitFor(() => expect(screen.queryByTestId('escoba-celebration')).not.toBeInTheDocument());
  });
});

describe('escoba capture-sum helpers', () => {
  it('values pip cards as their rank and face cards as 8/9/10', () => {
    expect(escobaCardValue({ design: 'SPADE', value: 1 })).toBe(1);
    expect(escobaCardValue({ design: 'SPADE', value: 7 })).toBe(7);
    expect(escobaCardValue({ design: 'SPADE', value: 11 })).toBe(8);
    expect(escobaCardValue({ design: 'SPADE', value: 12 })).toBe(9);
    expect(escobaCardValue({ design: 'SPADE', value: 13 })).toBe(10);
  });

  it('sums the hand card plus the selected table cards', () => {
    const table = [
      { design: 'SPADE' as const, value: 4 },
      { design: 'HEART' as const, value: 4 },
      { design: 'CLOVER' as const, value: 6 },
    ];
    expect(escobaSelectionSum({ design: 'SPADE', value: 7 }, table, [0, 1])).toBe(15);
    expect(escobaSelectionSum({ design: 'SPADE', value: 7 }, table, [])).toBe(7);
    expect(escobaSelectionSum(null, table, [0, 1])).toBe(8);
    // Out-of-range indices are ignored.
    expect(escobaSelectionSum({ design: 'SPADE', value: 1 }, table, [9])).toBe(1);
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も
  // 実行されない（#4596 / #4600 のレビュー指摘）。
  it('turns the frontend hint on from the settings panel', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(makeEscobaState());
    renderWithProviders(<EscobaPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });
});
