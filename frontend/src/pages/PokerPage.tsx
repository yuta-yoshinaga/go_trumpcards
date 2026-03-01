import { useCallback, useEffect, useState } from 'react';
import { pokerApi } from '../api/gameApi';
import { CardBack, CardImage } from '../components/CardImage';
import { ErrorAlert } from '../components/ErrorAlert';
import { useGameApi } from '../hooks/useGameApi';
import { btnDanger, btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';

import { PokerAction, PokerPhase } from '../types/phases';

const ACTION_NAMES: Record<number, string> = {
  [PokerAction.FOLD]: 'フォールド',
  [PokerAction.CHECK]: 'チェック',
  [PokerAction.CALL]: 'コール',
  [PokerAction.BET]: 'ベット',
  [PokerAction.RAISE]: 'レイズ',
  [PokerAction.ALL_IN]: 'オールイン',
};

const cardWrapBase: React.CSSProperties = {
  position: 'relative',
  cursor: 'pointer',
  transition: 'transform 0.15s',
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
};

export function PokerPage() {
  const [selected, setSelected] = useState<number[]>([]);
  const [betAmount, setBetAmount] = useState(10);

  const onSuccess = useCallback(() => {
    setSelected([]);
  }, []);
  const { state, loading, error, exec } = useGameApi(pokerApi.exec, { onSuccess });

  useEffect(() => {
    exec('reset');
  }, [exec]);

  useEffect(() => {
    if (state?.minRaise && state.minRaise > 0) {
      setBetAmount(state.minRaise);
    } else if (state) {
      setBetAmount(10);
    }
  }, [state]);

  const phase = state?.phase ?? PokerPhase.INIT;
  const isBettingPhase = phase === PokerPhase.DEAL || phase === PokerPhase.SECOND_BET;
  const isExchangePhase = phase === PokerPhase.EXCHANGE;
  const isEnd = phase === PokerPhase.END;
  const humanPlayer = state?.players?.find((p) => p.isHuman);
  const humanFolded = humanPlayer?.folded ?? false;
  const humanAllIn = humanPlayer?.allIn ?? false;
  const canAct = isBettingPhase && !humanFolded && !humanAllIn && state?.currentTurn === humanPlayer?.id;
  const canExchange = isExchangePhase && state?.currentTurn === humanPlayer?.id;
  const hasOutstandingBet = (state?.lastBet ?? 0) > (humanPlayer?.currentBet ?? 0);
  const minRaise = state?.minRaise ?? 10;

  const toggleSelect = (idx: number) => {
    if (!canExchange) return;
    setSelected((prev) => (prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]));
  };

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a6b1a]" aria-busy={loading} aria-live="polite">
      {loading && <span className="sr-only">処理中...</span>}
      {/* Info bar */}
      <div className="shrink-0 bg-black/40 text-white text-sm px-5 py-2 flex flex-wrap gap-x-6 gap-y-1">
        <span>
          ポット: <strong>{state?.pot ?? 0}</strong>
        </span>
        <span>
          ディーラー: <strong>Player {state?.dealerIdx ?? 0}</strong>
        </span>
        {(state?.jokerCount ?? 0) > 0 && (
          <span>
            ジョーカー: <strong>{state?.jokerCount}</strong>
          </span>
        )}
      </div>

      {/* Scrollable: CPU players + logs */}
      <div className="flex-1 overflow-y-auto pt-4 px-5">
        {/* CPU players */}
        {state?.players
          ?.filter((p) => !p.isHuman)
          .map((p) => (
            <div key={p.id} className="mb-3">
              <div className="text-white text-[0.95em] mb-1">
                CPU {p.id} <span className="text-gray-300 text-[0.85em]">({p.playStyleName})</span>
                <span className="ml-2 text-[0.85em]">チップ: {p.chips}</span>
                {p.currentBet > 0 && <span className="ml-2 text-[0.85em]">ベット: {p.currentBet}</span>}
                {p.folded && <span className="ml-2 text-red-300 text-[0.85em]">[フォールド]</span>}
                {p.allIn && <span className="ml-2 text-yellow-300 text-[0.85em]">[オールイン]</span>}
                {(phase === PokerPhase.SECOND_BET || isEnd) && p.exchangeCount >= 0 && !p.folded && (
                  <span className="ml-2 text-[0.85em]">交換: {p.exchangeCount}枚</span>
                )}
                {isEnd && !p.folded && p.handName && (
                  <span
                    className="inline-block ml-2 text-[0.85em] font-bold rounded px-2 py-0.5"
                    style={{ background: '#f0ad4e', color: '#222' }}
                  >
                    {p.handName}
                  </span>
                )}
              </div>
              <div className="flex flex-wrap gap-1">
                {isEnd && !p.folded && p.cards?.length
                  ? p.cards.map((card) => (
                      <CardImage
                        key={`${card.design}-${card.value}`}
                        card={card}
                        width={50}
                        style={{ border: '3px solid transparent' }}
                      />
                    ))
                  : Array.from({ length: 5 }).map((_, i) => (
                      // biome-ignore lint/suspicious/noArrayIndexKey: placeholder
                      <CardBack key={i} width={50} />
                    ))}
              </div>
            </div>
          ))}

        {/* CPU actions log */}
        {state?.cpuActions && state.cpuActions.length > 0 && (
          <div className="bg-black/30 rounded p-2 mb-3 text-white text-[0.85em]">
            <div className="font-bold mb-1">CPU行動:</div>
            {state.cpuActions.map((a, i) => (
              <div key={`${i}-${a.playerIdx}-${a.action}`}>
                Player {a.playerIdx}: {ACTION_NAMES[a.action] ?? '不明'}
                {a.amount > 0 && ` (${a.amount})`}
              </div>
            ))}
          </div>
        )}

        {/* CPU exchanges log */}
        {state?.cpuExchanges && state.cpuExchanges.length > 0 && (
          <div className="bg-black/30 rounded p-2 mb-3 text-white text-[0.85em]">
            <div className="font-bold mb-1">CPU交換:</div>
            {state.cpuExchanges.map((ex, i) => (
              <div key={`${i}-${ex.playerIdx}`}>
                Player {ex.playerIdx}: {ex.exchangeCount}枚交換
              </div>
            ))}
          </div>
        )}

        {/* Round results */}
        {isEnd && state?.roundResults && state.roundResults.length > 0 && (
          <div className="bg-black/30 rounded p-2 mb-3 text-white text-[0.85em]">
            <div className="font-bold mb-1">結果:</div>
            {state.roundResults.map((r) => (
              <div key={r.playerIdx}>
                {state.players[r.playerIdx]?.isHuman ? 'あなた' : `CPU ${r.playerIdx}`}
                {r.handName && `: ${r.handName}`}
                {r.wonAmount > 0 && <span className="text-yellow-300 ml-1"> +{r.wonAmount}チップ</span>}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Sticky footer: player hand + buttons */}
      <div
        className="shrink-0 bg-[#155715] border-t border-white/20 px-5 py-3"
        style={{ paddingBottom: 'calc(env(safe-area-inset-bottom) + 12px)' }}
      >
        {/* Human player */}
        {humanPlayer && (
          <div className="mb-2">
            <div className="text-white text-[1.1em] mb-1">
              あなたの手札
              <span className="ml-3 text-[0.85em]">チップ: {humanPlayer.chips}</span>
              {humanPlayer.currentBet > 0 && (
                <span className="ml-2 text-[0.85em]">ベット: {humanPlayer.currentBet}</span>
              )}
              {humanPlayer.folded && <span className="ml-2 text-red-300 text-[0.85em]">[フォールド]</span>}
              {humanPlayer.allIn && <span className="ml-2 text-yellow-300 text-[0.85em]">[オールイン]</span>}
              {isEnd && !humanPlayer.folded && humanPlayer.handName && (
                <span
                  className="inline-block ml-2 text-[0.85em] font-bold rounded px-2 py-0.5"
                  style={{ background: '#f0ad4e', color: '#222' }}
                >
                  {humanPlayer.handName}
                </span>
              )}
            </div>
            {canExchange && (
              <div className="text-[#cfc] text-[0.85em] mb-1">
                交換したいカードをクリックして選択し、「交換」または「スタンド」を押してください。
              </div>
            )}
            <div className="flex flex-wrap gap-1.5 mb-2">
              {humanPlayer.cards?.map((card, i) => {
                const isSelected = selected.includes(i);
                return (
                  <button
                    key={`${card.design}-${card.value}`}
                    type="button"
                    aria-pressed={isSelected}
                    onClick={() => toggleSelect(i)}
                    style={{
                      background: 'none',
                      border: 'none',
                      padding: 0,
                      ...cardWrapBase,
                      cursor: canExchange ? 'pointer' : 'default',
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
        )}

        {/* Message */}
        <div className="bg-black/55 rounded-lg text-white text-center px-4 py-2 text-[1.1em] font-bold mb-2 min-h-[36px]">
          {state?.message ?? ''}
        </div>

        <ErrorAlert message={error} />

        {/* Betting controls */}
        {canAct && (
          <div className="text-center mb-2">
            <div className="flex items-center justify-center gap-2 mb-2">
              <label htmlFor="pokerBetAmount" className="text-white text-sm">
                ベット額:
              </label>
              <input
                id="pokerBetAmount"
                type="number"
                min={minRaise}
                step={10}
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
            <button
              type="button"
              className={`${btnWarning} min-w-[80px]`}
              disabled={loading}
              onClick={() => exec('allin')}
            >
              オールイン
            </button>
          </div>
        )}

        {/* Exchange controls */}
        {canExchange && (
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

        {/* Reset button */}
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
