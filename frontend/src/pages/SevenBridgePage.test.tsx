import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sevenBridgeApi } from '../api/gameApi';
import { gameTheme } from '../styles/gameTheme';
import { renderWithProviders } from '../test/renderWithProviders';
import type { SevenBridgeResponse } from '../types/card';
import { SevenBridgePage } from './SevenBridgePage';

vi.mock('../api/gameApi', () => ({
  sevenBridgeApi: { exec: vi.fn() },
  actionLogApi: { sevenbridge: vi.fn() },
  sessionId: 'test-session',
}));

const mockExec = vi.mocked(sevenBridgeApi.exec);

const drawState: SevenBridgeResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 3,
      cards: [
        { design: 'SPADE', value: 9 },
        { design: 'CLOVER', value: 9 },
        { design: 'DIAMOND', value: 3 },
      ],
      melds: [],
      roundScore: 0,
      cumulativeScore: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 7,
      cards: [],
      melds: [],
      roundScore: 0,
      cumulativeScore: 10,
    },
  ],
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'HEART', value: 9 },
  drawPileCount: 30,
  gameEndFlag: false,
  winnerIdx: -1,
  roundWinnerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 100 },
};

const playState: SevenBridgeResponse = { ...drawState, phase: 1 };

describe('SevenBridgePage', () => {
  beforeEach(() => {
    mockExec.mockReset();
  });

  it('calls reset on mount', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<SevenBridgePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, expect.any(Object)));
  });

  it('applies the shared gameTheme background instead of a hardcoded class', async () => {
    mockExec.mockResolvedValue(drawState);
    const { container } = renderWithProviders(<SevenBridgePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(container.querySelector(`.${gameTheme.sevenbridge.bg}`)).toBeInTheDocument();
    expect(container.querySelector('.bg-ds-bg')).not.toBeInTheDocument();
  });

  it('renders draw phase controls', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<SevenBridgePage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /山札から引く|Draw from stock/i })).toBeInTheDocument(),
    );
    expect(screen.getByRole('button', { name: /ポン|Pon/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /チー|Chi/i })).toBeInTheDocument();
  });

  it('renders play phase controls', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<SevenBridgePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /メルド|Meld/i })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /捨てる|Discard/i })).toBeInTheDocument();
  });

  it('describes the pon/chi selection requirement and flips it to "met" via aria-describedby', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<SevenBridgePage />);
    const pon = await screen.findByRole('button', { name: /ポン|Pon/i });
    const chi = screen.getByRole('button', { name: /チー|Chi/i });
    expect(pon).toHaveAttribute('aria-describedby', 'sb-select-two-hint');
    expect(chi).toHaveAttribute('aria-describedby', 'sb-select-two-hint');
    // With nothing selected the hint states the 2-card requirement and pon is disabled.
    expect(screen.getByTestId('sb-select-two-hint')).toHaveTextContent('2枚');
    expect(pon).toBeDisabled();
    // Selecting exactly two cards satisfies the requirement.
    fireEvent.click(screen.getByRole('button', { name: '♠ 9' }));
    fireEvent.click(screen.getByRole('button', { name: '♣ 9' }));
    await waitFor(() => expect(screen.getByTestId('sb-select-two-hint')).toHaveTextContent('実行'));
    expect(pon).not.toBeDisabled();
  });

  it('describes the meld selection requirement and flips it to "met" via aria-describedby', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<SevenBridgePage />);
    const meld = await screen.findByRole('button', { name: /メルド|Meld/i });
    expect(meld).toHaveAttribute('aria-describedby', 'sb-meld-hint');
    // With fewer than 3 selected the hint states the requirement and meld is disabled.
    expect(screen.getByTestId('sb-meld-hint')).toHaveTextContent('3枚以上');
    expect(meld).toBeDisabled();
    // Selecting three cards satisfies the requirement.
    fireEvent.click(screen.getByRole('button', { name: '♠ 9' }));
    fireEvent.click(screen.getByRole('button', { name: '♣ 9' }));
    fireEvent.click(screen.getByRole('button', { name: '♦ 3' }));
    await waitFor(() => expect(screen.getByTestId('sb-meld-hint')).toHaveTextContent('実行'));
    expect(meld).not.toBeDisabled();
  });

  it('renders layoff target melds as clickable card-row buttons and highlights the selection', async () => {
    const meldState: SevenBridgeResponse = {
      ...playState,
      players: [
        {
          ...playState.players[0],
          melds: [
            {
              cards: [
                { design: 'SPADE', value: 5 },
                { design: 'HEART', value: 5 },
                { design: 'DIAMOND', value: 5 },
              ],
            },
          ],
        },
        playState.players[1],
      ],
    };
    mockExec.mockResolvedValue(meldState);
    renderWithProviders(<SevenBridgePage />);
    const meldBtn = await screen.findByTestId('sb-layoff-meld-0-0');
    fireEvent.click(meldBtn);
    expect(meldBtn).toHaveAttribute('aria-pressed', 'true');
    expect(meldBtn.className).toContain('ring-ds-info');
  });

  it('fires drawstock when clicking stock draw', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<SevenBridgePage />);
    const btn = await screen.findByRole('button', { name: /山札から引く|Draw from stock/i });
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('executes a typed CLI command (no longer a no-op stub)', async () => {
    localStorage.setItem('cli-mode-sevenbridge', 'true');
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<SevenBridgePage />);
    const input = await screen.findByRole('textbox');
    mockExec.mockClear();
    fireEvent.change(input, { target: { value: 'd' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
    localStorage.removeItem('cli-mode-sevenbridge');
  });
});

// #5547: ポン/チーで割り込んで取ったターンか、山から引いたターンかが
// 画面から判別できなかった。値は保存までされている。
describe('SevenBridgePage claim badge', () => {
  it('marks the turn that took the discard', async () => {
    mockExec.mockResolvedValue({ ...playState, claimedThisTurn: true });
    renderWithProviders(<SevenBridgePage />);
    await waitFor(() => expect(screen.getByTestId('sb-claimed-badge')).toBeInTheDocument());
  });

  // **山から引いたターンでは出さない。**毎ターン出ると区別にならない。
  it('shows nothing after an ordinary draw', async () => {
    mockExec.mockResolvedValue({ ...playState, claimedThisTurn: false });
    renderWithProviders(<SevenBridgePage />);
    await waitFor(() => expect(screen.getByTestId('sb-meld-hint')).toBeInTheDocument());
    expect(screen.queryByTestId('sb-claimed-badge')).not.toBeInTheDocument();
  });
});

// ポン/チーとメルドは押せない理由を読み上げるのに、レイオフだけ無言だった
// (#6343)。拒否の条件は「1枚だけ選ぶ」と「出せるメルドがある」の2つあるので、
// **どちらで止まっているか**を言い分けられていることを見る。
describe('SevenBridgePage layoff hint', () => {
  const hint = () => screen.getByTestId('sb-layoff-hint').textContent ?? '';

  it('names the missing meld when nobody has melded yet', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<SevenBridgePage />);
    await waitFor(() => expect(screen.getByTestId('sb-layoff-hint')).toBeInTheDocument());

    // fixture は全員 melds: [] なので、まず「出せるメルドが無い」で止まる。
    expect(hint()).toBe('出せるメルドがまだありません');
    // 選択枚数の話にすり替わっていないこと。
    expect(hint()).not.toContain('1枚');
  });

  it('asks for exactly one card once a meld exists', async () => {
    const withMeld: SevenBridgeResponse = {
      ...playState,
      // layoffTarget の初期値は 0。その席にメルドが無いと「メルドが無い」で
      // 止まってしまうので、0 番の席に持たせる。
      players: playState.players.map((p, i) =>
        i === 0 ? { ...p, melds: [{ cards: [], kind: 0 }] } : p,
      ) as SevenBridgeResponse['players'],
    };
    mockExec.mockResolvedValue(withMeld);
    renderWithProviders(<SevenBridgePage />);
    await waitFor(() => expect(screen.getByTestId('sb-layoff-hint')).toBeInTheDocument());

    expect(hint()).toBe('カードを1枚だけ選択してください');
  });

  it('describes the layoff button, so the reason is announced with it', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<SevenBridgePage />);
    const button = await screen.findByRole('button', { name: 'レイオフ' });
    expect(button).toHaveAttribute('aria-describedby', 'sb-layoff-hint');
  });
});
