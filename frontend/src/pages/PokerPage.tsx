import { useCallback, useEffect, useState } from 'react';
import { pokerApi } from '../api/gameApi';
import { CardBack, CardImage } from '../components/CardImage';
import { ErrorAlert } from '../components/ErrorAlert';
import { btnDanger, btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import type { PokerResponse } from '../types/card';
import { PokerPhase } from '../types/phases';

const cardWrapBase: React.CSSProperties = {
  position: 'relative',
  cursor: 'pointer',
  transition: 'transform 0.15s',
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
};

export function PokerPage() {
  const [state, setState] = useState<PokerResponse | null>(null);
  const [selected, setSelected] = useState<number[]>([]);
  const [betAmount, setBetAmount] = useState(10);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const exec = useCallback(
    async (
      command: 'reset' | 'exchange' | 'stand' | 'bet' | 'call' | 'raise' | 'fold' | 'check',
      indices?: number[],
      amount?: number,
    ) => {
      setLoading(true);
      try {
        setError(null);
        const res = await pokerApi.exec(command, indices, amount);
        setState(res);
        setSelected([]);
      } catch {
        setError('通信エラーが発生しました。もう一度お試しください。');
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const phase = state?.phase ?? PokerPhase.INIT;
  const isBettingPhase = phase === PokerPhase.DEAL || phase === PokerPhase.SECOND_BET;
  const isExchangePhase = phase === PokerPhase.EXCHANGE;
  const dealerBet = state?.dealer?.bet ?? 0;
  const playerBet = state?.player?.bet ?? 0;
  const hasOutstandingBet = dealerBet > playerBet;
  /* v8 ignore next */
  const anteAmount = state?.ante ?? 10;

  const toggleSelect = (idx: number) => {
    if (!isExchangePhase) return;
    setSelected((prev) => (prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]));
  };

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a6b1a]">
      {/* Chip/Pot info bar */}
      <div className="shrink-0 bg-black/40 text-white text-sm px-5 py-2 flex flex-wrap gap-x-6 gap-y-1">
        <span>
          ポット: <strong>{state?.pot ?? 0}</strong>
        </span>
        <span>
          プレイヤー チップ: <strong>{state?.player?.chips ?? 0}</strong>
        </span>
        <span>
          ディーラー チップ: <strong>{state?.dealer?.chips ?? 0}</strong>
        </span>
        {dealerBet > 0 && (
          <span>
            ディーラー ベット: <strong>{dealerBet}</strong>
          </span>
        )}
      </div>

      {/* Scrollable: dealer area */}
      <div className="flex-1 overflow-y-auto pt-4 px-5">
        <div className="mb-2">
          <div className="text-white text-[1.1em] mb-1.5">
            ディーラー手札
            {phase === PokerPhase.END && state?.dealer?.handName && (
              <span
                style={{
                  display: 'inline-block',
                  background: '#f0ad4e',
                  color: '#222',
                  fontWeight: 'bold',
                  borderRadius: 8,
                  padding: '2px 12px',
                  marginLeft: 8,
                  fontSize: '0.95em',
                }}
              >
                {state.dealer.handName}
              </span>
            )}
          </div>
          <div className="flex flex-wrap gap-2 mb-2.5">
            {phase === PokerPhase.END && state?.dealer?.cards?.length
              ? state.dealer.cards.map((card) => (
                  <div key={`${card.design}-${card.value}`} style={{ ...cardWrapBase, cursor: 'default' }}>
                    <CardImage card={card} width={60} style={{ border: '3px solid transparent' }} />
                  </div>
                ))
              : Array.from({ length: 5 }).map((_, i) => (
                  // biome-ignore lint/suspicious/noArrayIndexKey: placeholder array with no card identity
                  <CardBack key={i} width={60} />
                ))}
          </div>
        </div>
      </div>

      {/* Sticky footer: player hand + result + buttons */}
      <div
        className="shrink-0 bg-[#155715] border-t border-white/20 px-5 py-3"
        style={{ paddingBottom: 'calc(env(safe-area-inset-bottom) + 12px)' }}
      >
        {/* Player area */}
        <div>
          <div className="text-white text-[1.1em] mb-1">
            プレイヤー手札
            {phase === PokerPhase.END && state?.player?.handName && (
              <span
                style={{
                  display: 'inline-block',
                  background: '#f0ad4e',
                  color: '#222',
                  fontWeight: 'bold',
                  borderRadius: 8,
                  padding: '2px 12px',
                  marginLeft: 8,
                  fontSize: '0.95em',
                }}
              >
                {state.player.handName}
              </span>
            )}
          </div>
          {isExchangePhase && (
            <div className="text-[#cfc] text-[0.85em] mb-1">
              交換したいカードをクリックして選択し、「交換」または「スタンド」を押してください。
            </div>
          )}
          <div className="flex flex-wrap gap-1.5 mb-2">
            {state?.player?.cards?.map((card, i) => {
              const isSelected = selected.includes(i);
              return (
                <button
                  key={`${card.design}-${card.value}`}
                  type="button"
                  onClick={() => toggleSelect(i)}
                  style={{
                    background: 'none',
                    border: 'none',
                    padding: 0,
                    ...cardWrapBase,
                    cursor: isExchangePhase ? 'pointer' : 'default',
                  }}
                >
                  <CardImage
                    card={card}
                    width={60}
                    style={{
                      border: isSelected ? '3px solid #f0ad4e' : '3px solid transparent',
                      transform: isSelected ? 'translateY(-10px)' : undefined,
                      transition: 'transform 0.15s',
                    }}
                  />
                  <div
                    style={{
                      color: '#f0ad4e',
                      fontSize: '0.75em',
                      fontWeight: 'bold',
                      visibility: isSelected ? 'visible' : 'hidden',
                    }}
                  >
                    交換
                  </div>
                </button>
              );
            })}
          </div>
        </div>

        {/* Result */}
        <div className="bg-black/55 rounded-lg text-white text-center px-4 py-2 text-[1.1em] font-bold mb-2 min-h-[36px]">
          {state?.message ?? ''}
        </div>

        <ErrorAlert message={error} />

        {/* Betting controls */}
        {isBettingPhase && (
          <div className="text-center mb-2">
            <div className="flex items-center justify-center gap-2 mb-2">
              <label htmlFor="betAmount" className="text-white text-sm">
                ベット額:
              </label>
              <input
                id="betAmount"
                type="number"
                min={anteAmount}
                step={anteAmount}
                value={betAmount}
                onChange={(e) => setBetAmount(Number(e.target.value))}
                className="w-20 px-2 py-1 text-sm rounded bg-white/90 text-gray-900"
              />
            </div>
            {hasOutstandingBet ? (
              <>
                <button
                  type="button"
                  className={`${btnSuccess} min-w-[80px]`}
                  disabled={loading}
                  onClick={() => exec('call')}
                >
                  コール
                </button>
                <button
                  type="button"
                  className={`${btnWarning} min-w-[80px]`}
                  disabled={loading}
                  onClick={() => exec('raise', undefined, betAmount)}
                >
                  レイズ
                </button>
              </>
            ) : (
              <>
                <button
                  type="button"
                  className={`${btnWarning} min-w-[80px]`}
                  disabled={loading}
                  onClick={() => exec('bet', undefined, betAmount)}
                >
                  ベット
                </button>
                <button
                  type="button"
                  className={`${btnSuccess} min-w-[80px]`}
                  disabled={loading}
                  onClick={() => exec('check')}
                >
                  チェック
                </button>
              </>
            )}
            <button
              type="button"
              className={`${btnDanger} min-w-[80px]`}
              disabled={loading}
              onClick={() => exec('fold')}
            >
              フォールド
            </button>
          </div>
        )}

        {/* Exchange controls */}
        {isExchangePhase && (
          <div className="text-center mb-2">
            <button
              type="button"
              className={`${btnWarning} min-w-[90px]`}
              disabled={loading}
              onClick={() => exec('exchange', selected)}
            >
              交換
            </button>
            <button
              type="button"
              className={`${btnSuccess} min-w-[90px]`}
              disabled={loading}
              onClick={() => exec('stand')}
            >
              スタンド
            </button>
          </div>
        )}

        {/* Always-visible buttons */}
        <div className="text-center">
          <button
            type="button"
            className={`${btnPrimary} min-w-[90px]`}
            disabled={loading}
            onClick={() => exec('reset')}
          >
            リセット
          </button>
        </div>
      </div>
    </div>
  );
}
