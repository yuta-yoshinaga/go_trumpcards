import { useCallback, useEffect, useState } from 'react';
import { holdemApi } from '../api/gameApi';
import { CardBack, CardImage } from '../components/CardImage';
import { ErrorAlert } from '../components/ErrorAlert';
import { btnDanger, btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import type { HoldemResponse } from '../types/card';
import { HoldemAction, HoldemPhase } from '../types/phases';

const ACTION_NAMES: Record<number, string> = {
  [HoldemAction.FOLD]: 'フォールド',
  [HoldemAction.CHECK]: 'チェック',
  [HoldemAction.CALL]: 'コール',
  [HoldemAction.BET]: 'ベット',
  [HoldemAction.RAISE]: 'レイズ',
  [HoldemAction.ALL_IN]: 'オールイン',
};

const PHASE_NAMES: Record<number, string> = {
  [HoldemPhase.PRE_FLOP]: 'プリフロップ',
  [HoldemPhase.FLOP]: 'フロップ',
  [HoldemPhase.TURN]: 'ターン',
  [HoldemPhase.RIVER]: 'リバー',
  [HoldemPhase.SHOWDOWN]: 'ショーダウン',
  [HoldemPhase.END]: '結果',
};

export function HoldemPage() {
  const [state, setState] = useState<HoldemResponse | null>(null);
  const [betAmount, setBetAmount] = useState(20);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const exec = useCallback(
    async (command: 'reset' | 'fold' | 'check' | 'call' | 'bet' | 'raise' | 'allin', amount?: number) => {
      setLoading(true);
      try {
        setError(null);
        const res = await holdemApi.exec(command, amount);
        setState(res);
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

  useEffect(() => {
    if (state?.minRaise && state.minRaise > 0) {
      setBetAmount(state.minRaise);
    } else if (state) {
      setBetAmount(20);
    }
  }, [state]);

  const phase = state?.phase ?? HoldemPhase.INIT;
  const isActive = phase >= HoldemPhase.PRE_FLOP && phase <= HoldemPhase.RIVER;
  const isShowdown = phase === HoldemPhase.SHOWDOWN || phase === HoldemPhase.END;
  const humanPlayer = state?.players?.find((p) => p.isHuman);
  const humanFolded = humanPlayer?.folded ?? false;
  const humanAllIn = humanPlayer?.allIn ?? false;
  const canAct = isActive && !humanFolded && !humanAllIn && state?.currentTurn === humanPlayer?.id;
  const hasOutstandingBet = (state?.lastBet ?? 0) > (humanPlayer?.currentBet ?? 0);
  const minRaise = state?.minRaise ?? 0;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a6b1a]">
      {/* Info bar */}
      <div className="shrink-0 bg-black/40 text-white text-sm px-5 py-2 flex flex-wrap gap-x-6 gap-y-1">
        <span>
          フェーズ: <strong>{PHASE_NAMES[phase] ?? '初期化中'}</strong>
        </span>
        <span>
          ポット: <strong>{state?.pot ?? 0}</strong>
        </span>
        <span>
          ディーラー: <strong>Player {state?.dealerIdx ?? 0}</strong>
        </span>
      </div>

      {/* Scrollable: community cards + CPU players */}
      <div className="flex-1 overflow-y-auto pt-4 px-5">
        {/* Community cards */}
        <div className="mb-4">
          <div className="text-white text-[1.1em] mb-1.5">コミュニティカード</div>
          <div className="flex flex-wrap gap-2">
            {state?.communityCards?.length
              ? state.communityCards.map((card) => (
                  <CardImage
                    key={`${card.design}-${card.value}`}
                    card={card}
                    width={60}
                    style={{ border: '3px solid transparent' }}
                  />
                ))
              : Array.from({ length: 5 }).map((_, i) => (
                  // biome-ignore lint/suspicious/noArrayIndexKey: placeholder
                  <CardBack key={i} width={60} />
                ))}
          </div>
        </div>

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
                {isShowdown && !p.folded && p.handName && (
                  <span
                    className="inline-block ml-2 text-[0.85em] font-bold rounded px-2 py-0.5"
                    style={{ background: '#f0ad4e', color: '#222' }}
                  >
                    {p.handName}
                  </span>
                )}
              </div>
              <div className="flex flex-wrap gap-1">
                {isShowdown && !p.folded && p.cards?.length
                  ? p.cards.map((card) => (
                      <CardImage
                        key={`${card.design}-${card.value}`}
                        card={card}
                        width={50}
                        style={{ border: '3px solid transparent' }}
                      />
                    ))
                  : Array.from({ length: 2 }).map((_, i) => (
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

        {/* Round results */}
        {isShowdown && state?.roundResults && state.roundResults.length > 0 && (
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
              {isShowdown && !humanPlayer.folded && humanPlayer.handName && (
                <span
                  className="inline-block ml-2 text-[0.85em] font-bold rounded px-2 py-0.5"
                  style={{ background: '#f0ad4e', color: '#222' }}
                >
                  {humanPlayer.handName}
                </span>
              )}
            </div>
            <div className="flex flex-wrap gap-1.5 mb-2">
              {humanPlayer.cards?.length
                ? humanPlayer.cards.map((card) => (
                    <CardImage
                      key={`${card.design}-${card.value}`}
                      card={card}
                      width={60}
                      style={{ border: '3px solid transparent' }}
                    />
                  ))
                : !humanPlayer.folded &&
                  Array.from({ length: 2 }).map((_, i) => (
                    // biome-ignore lint/suspicious/noArrayIndexKey: placeholder
                    <CardBack key={i} width={60} />
                  ))}
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
              <label htmlFor="holdemBetAmount" className="text-white text-sm">
                ベット額:
              </label>
              <input
                id="holdemBetAmount"
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
                  onClick={() => exec('raise', betAmount)}
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
                  onClick={() => exec('bet', betAmount)}
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
