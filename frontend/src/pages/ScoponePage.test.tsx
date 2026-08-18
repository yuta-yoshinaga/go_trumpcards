import { fireEvent, screen, waitFor } from '@testing-library/react';
import i18n from 'i18next';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { scoponeApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeScoponeState } from '../test/stateFactories';
import { ScoponePage } from './ScoponePage';

vi.mock('../api/gameApi', () => ({
  scoponeApi: { exec: vi.fn() },
  actionLogApi: { scopone: vi.fn() },
}));

const mockExec = vi.mocked(scoponeApi.exec);

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeScoponeState());
});

afterEach(async () => {
  await i18n.changeLanguage('ja');
});

describe('ScoponePage', () => {
  it('renders the GameSkeleton while the initial state is loading', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<ScoponePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the short "r" command', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('r'));
  });

  it('renders CPU difficulty options with localized labels', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // Difficulty options are localized (ja), not the hardcoded Easy/Normal/Hard.
    expect(screen.getByRole('option', { name: 'かんたん' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'ふつう' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'むずかしい' })).toBeInTheDocument();
  });

  it('renders the human hand', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('hand-card-1')).toBeInTheDocument();
    expect(screen.getByTestId('hand-card-2')).toBeInTheDocument();
  });

  it('renders table cards', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('table-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('table-card-1')).toBeInTheDocument();
  });

  it('renders team scores', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('team-score-0')).toBeInTheDocument());
    expect(screen.getByTestId('team-score-1')).toBeInTheDocument();
  });

  it('take button is disabled until both hand and table are selected', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('take-button')).toBeInTheDocument());
    expect(screen.getByTestId('take-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('hand-card-0'));
    expect(screen.getByTestId('take-button')).toBeDisabled();

    fireEvent.click(screen.getByTestId('table-card-0'));
    await waitFor(() => expect(screen.getByTestId('take-button')).not.toBeDisabled());
  });

  it('lay button is enabled when a hand card is selected and no table card', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('lay-button')).toBeInTheDocument());
    expect(screen.getByTestId('lay-button')).toBeDisabled();

    // **手札1 を選ぶ。**手札0 は場札を取れるので、取り札必須で「出す」は選べない (#5661)。
    fireEvent.click(screen.getByTestId('hand-card-1'));
    await waitFor(() => expect(screen.getByTestId('lay-button')).not.toBeDisabled());
  });

  it('plays "p" with sorted table indices when Take is clicked', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('table-card-1'));
    fireEvent.click(screen.getByTestId('table-card-0'));
    fireEvent.click(screen.getByTestId('take-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('p', { handIndex: 0, tableIndices: [0, 1] }));
  });

  it('plays "p" with empty table indices when Lay is clicked', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockClear();
    // 取れない手札 (handCaptures[1] は空) を選ぶ。手札0 は取り札必須で出せない。
    fireEvent.click(screen.getByTestId('hand-card-1'));
    fireEvent.click(screen.getByTestId('lay-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('p', { handIndex: 1, tableIndices: [] }));
  });

  it('disables actions when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeScoponeState({ currentTurn: 1, isHumanTurn: false }));
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('take-button')).toBeDisabled());
    expect(screen.getByTestId('lay-button')).toBeDisabled();
  });

  it('shows the round-end breakdown and a next-round button on roundEnd', async () => {
    mockExec.mockResolvedValue(
      makeScoponeState({
        phase: 'roundEnd',
        isHumanTurn: false,
        lastRoundDetail: {
          cards: [1, 0],
          diamonds: [1, 0],
          sevens: [0, 1],
          scopas: [1, 0],
          gained: [3, 1],
          settebello: 0,
        },
      }),
    );
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('round-detail')).toBeInTheDocument());
    expect(screen.getByTestId('next-round-button')).toBeInTheDocument();
  });

  it('next-round button dispatches "n"', async () => {
    mockExec.mockResolvedValue(makeScoponeState({ phase: 'roundEnd', isHumanTurn: false }));
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('next-round-button')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('next-round-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('n'));
  });

  it('resets with config and passes it to the API', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('r', {
        config: { targetScore: 11, cpuDifficulty: 1 },
      }),
    );
  });

  it('changes CPU difficulty and includes it in the reset config', async () => {
    renderWithProviders(<ScoponePage />);
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
      makeScoponeState({
        players: [{ id: 0, isHuman: true, team: 0, handCount: 0, cards: [], capturedCount: 0, scopaCount: 0 }],
      }),
    );
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.queryByTestId('hand-card-0')).not.toBeInTheDocument());
  });

  it('renders CLI terminal when CLI mode is enabled via localStorage', async () => {
    localStorage.setItem('cli-mode-scopone', 'true');
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
    localStorage.removeItem('cli-mode-scopone');
  });

  // Builds a 4-player Scopone state with a given scopaCount for one player.
  const stateWithScopa = (playerId: number, scopaCount: number) =>
    makeScoponeState({
      players: [
        {
          id: 0,
          isHuman: true,
          team: 0,
          handCount: 3,
          cards: makeScoponeState().players[0].cards,
          capturedCount: 0,
          scopaCount: playerId === 0 ? scopaCount : 0,
        },
        {
          id: 1,
          isHuman: false,
          team: 1,
          handCount: 3,
          cards: [],
          capturedCount: 0,
          scopaCount: playerId === 1 ? scopaCount : 0,
        },
        {
          id: 2,
          isHuman: false,
          team: 0,
          handCount: 3,
          cards: [],
          capturedCount: 0,
          scopaCount: playerId === 2 ? scopaCount : 0,
        },
        {
          id: 3,
          isHuman: false,
          team: 1,
          handCount: 3,
          cards: [],
          capturedCount: 0,
          scopaCount: playerId === 3 ? scopaCount : 0,
        },
      ],
    });

  it('does not show the scopa badge on initial load', async () => {
    mockExec.mockResolvedValue(stateWithScopa(0, 2));
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.queryByTestId('scopone-scopa-celebration')).not.toBeInTheDocument();
  });

  it('shows the scopa badge when a player scopaCount increases', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    // A subsequent play resolves a state where the human (team 0) swept the table.
    mockExec.mockResolvedValue(stateWithScopa(0, 1));
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('table-card-0'));
    fireEvent.click(screen.getByTestId('take-button'));

    const badge = await screen.findByTestId('scopone-scopa-celebration');
    expect(badge).toBeInTheDocument();
    // Own-team sweep uses the emphasised label.
    expect(badge).toHaveTextContent('スコパ！ あなたのチーム');
  });

  it('clears the scopa badge after scopaCount resets to zero', async () => {
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    mockExec.mockResolvedValue(stateWithScopa(0, 1));
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('table-card-0'));
    fireEvent.click(screen.getByTestId('take-button'));
    await screen.findByTestId('scopone-scopa-celebration');

    // A next round drops scopaCount back to 0 — the badge must disappear. Use Lay
    // (no table selection needed) since the prior play cleared the selection.
    // 手札1 を選ぶ: 手札0 は場札を取れるので取り札必須で「出す」が押せない (#5661)。
    mockExec.mockResolvedValue(stateWithScopa(0, 0));
    fireEvent.click(screen.getByTestId('hand-card-1'));
    fireEvent.click(screen.getByTestId('lay-button'));
    await waitFor(() => expect(screen.queryByTestId('scopone-scopa-celebration')).not.toBeInTheDocument());
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も
  // 実行されない（#4596 / #4600 のレビュー指摘）。
  it('turns the frontend hint on from the settings panel', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(makeScoponeState());
    renderWithProviders(<ScoponePage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });

  // #5661: 取れる組み合わせがあるのに「出す」を選ぶと domain が
  // "a capture is available and must be taken" で弾く。画面はその判定に使える
  // handCaptures を既に持っていて場札のハイライトに使っているのに、「出す」
  // ボタンは見ておらず、押してサーバーエラーで初めて気づいた。
  it('blocks laying a card that could capture', async () => {
    localStorage.clear();
    mockExec.mockReset();
    // handCaptures[0] は [[1]] = 手札0 は場札1 を取れる。
    mockExec.mockResolvedValue(makeScoponeState());
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('hand-card-0'));

    expect(screen.getByTestId('lay-button')).toBeDisabled();
    // 理由が読めること (受け入れ条件3)。
    expect(await screen.findByTestId('scopone-must-capture')).toBeInTheDocument();
  });

  it('still lays a card that cannot capture anything', async () => {
    localStorage.clear();
    mockExec.mockReset();
    // handCaptures[1] は [] = 手札1 では何も取れない。
    mockExec.mockResolvedValue(makeScoponeState());
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-1')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('hand-card-1'));

    expect(screen.getByTestId('lay-button')).toBeEnabled();
    expect(screen.queryByTestId('scopone-must-capture')).not.toBeInTheDocument();
  });

  it('totals the selected table cards against the chosen hand card', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(makeScoponeState());
    renderWithProviders(<ScoponePage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.queryByTestId('scopone-sum-indicator')).not.toBeInTheDocument();
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('table-card-0'));
    await waitFor(() => expect(screen.getByTestId('scopone-sum-indicator')).toBeInTheDocument());
  });
});
