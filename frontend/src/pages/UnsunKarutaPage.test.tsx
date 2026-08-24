import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { unsunKarutaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeUnsunKarutaState } from '../test/stateFactories';
import { UnsunKarutaPage } from './UnsunKarutaPage';

vi.mock('../api/gameApi', () => ({
  unsunKarutaApi: { exec: vi.fn() },
  actionLogApi: { unsunkaruta: vi.fn() },
}));

const mockExec = vi.mocked(unsunKarutaApi.exec);

const leadState = makeUnsunKarutaState();

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(leadState);
});

describe('UnsunKarutaPage', () => {
  it('calls reset on mount with the configured match length', async () => {
    renderWithProviders(<UnsunKarutaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, targetDeals: 4 } }),
    );
  });

  it('shows the deal, trick and trump headers', async () => {
    renderWithProviders(<UnsunKarutaPage />);
    expect(await screen.findByText('ディール 1')).toBeInTheDocument();
    expect(screen.getByText('トリック 1/9')).toBeInTheDocument();
    // **切り札のスートが読めないと数札の強弱が決まらない** ── 長物は 9 が最強、
    // 丸物は 1 が最強なので、どちらが切り札かは盤面の必須情報。
    expect(screen.getByTestId('unsunkaruta-trump')).toHaveTextContent('こつ');
  });

  it('groups the tricks by team rather than by seat', async () => {
    mockExec.mockResolvedValue(makeUnsunKarutaState({ teamTricks: [3, 1], teamScores: [7, 5] }));
    renderWithProviders(<UnsunKarutaPage />);
    const box = await screen.findByTestId('unsunkaruta-teams');
    expect(box).toHaveTextContent('味方 3 / 相手 1');
    expect(box).toHaveTextContent('味方 7 / 相手 5');
  });

  // 味方は 2 つ隣。人間が組 1 に座るディールでは、表示の「味方」も入れ替わる。
  it('reads the team columns from the human seat, not from team 0', async () => {
    mockExec.mockResolvedValue(makeUnsunKarutaState({ humanTeam: 1, teamTricks: [3, 1], teamScores: [7, 5] }));
    renderWithProviders(<UnsunKarutaPage />);
    const box = await screen.findByTestId('unsunkaruta-teams');
    expect(box).toHaveTextContent('味方 1 / 相手 3');
    expect(box).toHaveTextContent('味方 5 / 相手 7');
  });

  it('plays the selected card without declaring', async () => {
    renderWithProviders(<UnsunKarutaPage />);
    fireEvent.click(await screen.findByRole('button', { name: '9 棒' }));
    fireEvent.click(screen.getByTestId('unsunkaruta-play'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0, declare: false }));
  });

  // **宣言は札と一緒に飛ぶ。** 別々に送れると「宣言したが出していない」盤面が
  // 生まれてしまう。
  it('sends the declaration together with the card', async () => {
    renderWithProviders(<UnsunKarutaPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'ウン 杯' }));
    fireEvent.click(screen.getByTestId('unsunkaruta-declare'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2, declare: true }));
  });

  it('offers no declare button when the human is not on lead', async () => {
    mockExec.mockResolvedValue(
      makeUnsunKarutaState({
        canDeclare: false,
        currentTrick: [{ playerIdx: 7, card: leadState.players[0].cards[0] }],
      }),
    );
    renderWithProviders(<UnsunKarutaPage />);
    expect(await screen.findByTestId('unsunkaruta-play')).toBeInTheDocument();
    expect(screen.queryByTestId('unsunkaruta-declare')).not.toBeInTheDocument();
    expect(screen.queryByTestId('unsunkaruta-can-declare')).not.toBeInTheDocument();
  });

  // フォロー義務が出ている理由を書かないと、なぜ札が絞られるのか読めない。
  it('explains the follow obligation once a declaration is standing', async () => {
    mockExec.mockResolvedValue(makeUnsunKarutaState({ mustFollow: true, canDeclare: false, playableIndices: [1, 2] }));
    renderWithProviders(<UnsunKarutaPage />);
    expect(await screen.findByTestId('unsunkaruta-must-follow')).toBeInTheDocument();
  });

  // legalIndices が付けるのは data-legal。validIndices だけでは出せる札に印が付かない。
  it('marks only the legal cards while a declaration stands', async () => {
    mockExec.mockResolvedValue(makeUnsunKarutaState({ mustFollow: true, canDeclare: false, playableIndices: [1, 2] }));
    renderWithProviders(<UnsunKarutaPage />);
    await screen.findByTestId('unsunkaruta-must-follow');
    expect(document.querySelectorAll('[data-legal]')).toHaveLength(2);
    expect(screen.getByRole('button', { name: '9 棒' })).toHaveAttribute('aria-disabled', 'true');
  });

  it('advances the trick', async () => {
    mockExec.mockResolvedValue(makeUnsunKarutaState({ phase: 1, isHumanTurn: false, playableIndices: [] }));
    renderWithProviders(<UnsunKarutaPage />);
    fireEvent.click(await screen.findByTestId('unsunkaruta-next-trick'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('advances the deal', async () => {
    mockExec.mockResolvedValue(
      makeUnsunKarutaState({ phase: 2, isHumanTurn: false, playableIndices: [], teamTricks: [5, 4] }),
    );
    renderWithProviders(<UnsunKarutaPage />);
    fireEvent.click(await screen.findByTestId('unsunkaruta-next-round'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
    expect(screen.getByTestId('unsunkaruta-result')).toHaveTextContent('味方 5 個 / 相手 4 個');
  });

  it('names the winning team at the end of the match', async () => {
    mockExec.mockResolvedValue(
      makeUnsunKarutaState({ phase: 3, gameEndFlag: true, winnerTeam: 0, isHumanTurn: false, playableIndices: [] }),
    );
    renderWithProviders(<UnsunKarutaPage />);
    expect(await screen.findByTestId('unsunkaruta-winner')).toHaveTextContent('組0 の勝ち');
  });

  it('calls a tied match a draw', async () => {
    mockExec.mockResolvedValue(
      makeUnsunKarutaState({ phase: 3, gameEndFlag: true, winnerTeam: -1, isHumanTurn: false, playableIndices: [] }),
    );
    renderWithProviders(<UnsunKarutaPage />);
    expect(await screen.findByTestId('unsunkaruta-winner')).toHaveTextContent('引き分け');
  });

  // ヒントのゲート: 頼んでいないヒントは出さない。
  it('does not render the hint banner unless it was requested', async () => {
    mockExec.mockResolvedValue(
      makeUnsunKarutaState({ hint: { cardIndices: [0], reason: 'lead_strong' }, messageCode: '' }),
    );
    renderWithProviders(<UnsunKarutaPage />);
    await screen.findByText('ディール 1');
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue(
      makeUnsunKarutaState({
        hint: { cardIndices: [0], reason: 'lead_strong' },
        messageCode: 'unsunkaruta.hintRequested',
      }),
    );
    renderWithProviders(<UnsunKarutaPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });
});
