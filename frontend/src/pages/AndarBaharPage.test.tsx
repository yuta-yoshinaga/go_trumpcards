import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { andarbaharApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { AndarBaharResponse } from '../types/card';
import { AndarBaharColumn, AndarBaharPhase, AndarBaharSideBand } from '../types/phases';
import { AndarBaharPage } from './AndarBaharPage';

vi.mock('../api/gameApi', () => ({
  andarbaharApi: { exec: vi.fn() },
  actionLogApi: { andarbahar: vi.fn() },
}));

vi.mock('../hooks/useCliMode', () => ({
  useCliMode: vi.fn(() => ({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  })),
}));

const mockApi = vi.mocked(andarbaharApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const betState: AndarBaharResponse = {
  joker: { design: 'SPADE', value: 7 },
  andarCards: [],
  baharCards: [],
  firstColumn: AndarBaharColumn.ANDAR,
  dealtCount: 0,
  phase: AndarBaharPhase.BET,
  chips: 1000,
  betAmount: 0,
  betTarget: AndarBaharColumn.ANDAR,
  sideAmount: 0,
  sideBand: AndarBaharSideBand.NONE,
  winner: -1,
  result: 0,
  payout: 0,
  history: [],
  message: '',
};

const andarWinState: AndarBaharResponse = {
  ...betState,
  andarCards: [
    { design: 'CLOVER', value: 3 },
    { design: 'HEART', value: 7 },
  ],
  baharCards: [{ design: 'DIAMOND', value: 12 }],
  dealtCount: 3,
  phase: AndarBaharPhase.END,
  chips: 1090,
  betAmount: 100,
  betTarget: AndarBaharColumn.ANDAR,
  winner: AndarBaharColumn.ANDAR,
  result: 1,
  payout: 190,
  history: [AndarBaharColumn.ANDAR],
};

beforeEach(() => {
  vi.clearAllMocks();
  mockUseCliMode.mockReturnValue({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  });
});

describe('AndarBaharPage', () => {
  it('resets on mount', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<AndarBaharPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('shows the joker and which column is dealt first', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<AndarBaharPage />);

    // **先に配る列は賭ける前に見えていなければならない。** 配当が下がる側だからです。
    await waitFor(() => expect(screen.getByTestId('andarbahar-first-column')).toBeInTheDocument());
    expect(screen.getByTestId('andarbahar-first-column')).toHaveTextContent('アンダー');
  });

  it('places a bet on Andar with no side bet', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<AndarBaharPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /アンダーに賭ける/ })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /アンダーに賭ける/ }));

    // **賭けていない帯は NONE で送る。** 0 は「1枚目の帯」という有効な値なので、
    // そのまま送るとサーバに賭け金 0 のサイドベットとして拒否されます。
    await waitFor(() =>
      expect(mockApi).toHaveBeenCalledWith('bet', 100, AndarBaharColumn.ANDAR, 0, AndarBaharSideBand.NONE),
    );
  });

  it('places a bet on Bahar', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<AndarBaharPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /バハールに賭ける/ })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /バハールに賭ける/ }));
    await waitFor(() =>
      expect(mockApi).toHaveBeenCalledWith('bet', 100, AndarBaharColumn.BAHAR, 0, AndarBaharSideBand.NONE),
    );
  });

  it('sends the side bet once a band is chosen', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<AndarBaharPage />);
    await waitFor(() => expect(screen.getByLabelText('サイドベット')).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('サイドベット'), {
      target: { value: String(AndarBaharSideBand.SIX_TO_TEN) },
    });
    // 帯を選ぶと金額入力が現れる。
    const sideInput = await screen.findByLabelText('サイドベット額');
    fireEvent.change(sideInput, { target: { value: '50' } });

    fireEvent.click(screen.getByRole('button', { name: /アンダーに賭ける/ }));
    await waitFor(() =>
      expect(mockApi).toHaveBeenCalledWith('bet', 100, AndarBaharColumn.ANDAR, 50, AndarBaharSideBand.SIX_TO_TEN),
    );
  });

  it('keeps the side band NONE when the band is chosen but the stake stays 0', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<AndarBaharPage />);
    await waitFor(() => expect(screen.getByLabelText('サイドベット')).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('サイドベット'), {
      target: { value: String(AndarBaharSideBand.FIRST) },
    });
    fireEvent.click(screen.getByRole('button', { name: /アンダーに賭ける/ }));

    // 金額が 0 のままなら帯は送らない。
    await waitFor(() =>
      expect(mockApi).toHaveBeenCalledWith('bet', 100, AndarBaharColumn.ANDAR, 0, AndarBaharSideBand.NONE),
    );
  });

  it('renders both columns and the settled result', async () => {
    mockApi.mockResolvedValue(andarWinState);
    renderWithProviders(<AndarBaharPage />);

    await waitFor(() => expect(screen.getByTestId('payout-result')).toBeInTheDocument());
    expect(screen.getByTestId('payout-result')).toHaveTextContent('アンダーの勝ち');
    expect(screen.getByTestId('andarbahar-dealt-count')).toHaveTextContent('3');
    // 2 列の合計が配った枚数と一致する。
    const andar = screen.getByTestId(`andarbahar-column-${AndarBaharColumn.ANDAR}`);
    const bahar = screen.getByTestId(`andarbahar-column-${AndarBaharColumn.BAHAR}`);
    expect(andar.querySelectorAll('img').length + bahar.querySelectorAll('img').length).toBe(andarWinState.dealtCount);
  });

  it('shows the profit net of both stakes', async () => {
    mockApi.mockResolvedValue({ ...andarWinState, sideAmount: 50, sideBand: AndarBaharSideBand.TWO_TO_FIVE });
    renderWithProviders(<AndarBaharPage />);

    // 払戻 190 - メイン 100 - サイド 50 = +40
    await waitFor(() => expect(screen.getByTestId('payout-diff')).toHaveTextContent('40'));
  });

  it('replays the last bet without its side bet', async () => {
    mockApi.mockResolvedValueOnce(betState).mockResolvedValue(andarWinState);
    renderWithProviders(<AndarBaharPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /アンダーに賭ける/ })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /アンダーに賭ける/ }));
    const rebet = await screen.findByTestId('ab-rebet-button');
    mockApi.mockClear();
    fireEvent.click(rebet);

    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
    await waitFor(() =>
      expect(mockApi).toHaveBeenCalledWith('bet', 100, AndarBaharColumn.ANDAR, 0, AndarBaharSideBand.NONE),
    );
  });

  it('clears the road history', async () => {
    mockApi.mockResolvedValue(andarWinState);
    renderWithProviders(<AndarBaharPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '履歴を消す' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '履歴を消す' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('clear'));
  });

  it('renders the road only once there is history', async () => {
    mockApi.mockResolvedValue(betState);
    const { unmount } = renderWithProviders(<AndarBaharPage />);
    await waitFor(() => expect(screen.queryByTestId('andarbahar-road')).not.toBeInTheDocument());
    unmount();

    mockApi.mockResolvedValue(andarWinState);
    renderWithProviders(<AndarBaharPage />);
    await waitFor(() => expect(screen.getByTestId('andarbahar-road')).toBeInTheDocument());
  });

  // **サーバが実際に返す形を踏む。** fixture が自分で配列を渡していると、
  // 空の列が null で返るケースを一度も通らないまま緑になります (#5321)。
  it('survives a response whose columns arrive empty', async () => {
    mockApi.mockResolvedValue({ ...betState, andarCards: [], baharCards: [], history: [] });
    renderWithProviders(<AndarBaharPage />);

    await waitFor(() => expect(screen.getByTestId('card-area')).toBeInTheDocument());
    expect(screen.getByTestId(`andarbahar-column-${AndarBaharColumn.ANDAR}`)).toBeInTheDocument();
  });

  it('renders the CLI terminal when CLI mode is on', async () => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<AndarBaharPage />);

    await waitFor(() => expect(screen.getByRole('log')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /アンダーに賭ける/ })).not.toBeInTheDocument();
  });
});
