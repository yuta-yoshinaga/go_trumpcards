import { useCallback, useEffect, useState } from 'react';
import { blackjackApi } from '../api/gameApi';
import { CardBack, CardImage } from '../components/CardImage';
import { ErrorAlert } from '../components/ErrorAlert';
import { btnDanger, btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import type { BlackJackResponse } from '../types/card';
import { BjPhase } from '../types/phases';

// suggestedAction constants (must match domain BJSuggestedAction)
const BJ_SUGGEST_NONE = 0;
const BJ_SUGGEST_HIT = 1;
const BJ_SUGGEST_STAND = 2;
const BJ_SUGGEST_DOUBLE = 3;
const BJ_SUGGEST_SPLIT = 4;
const BJ_SUGGEST_SURRENDER = 5;
const BJ_SUGGEST_DECLINE_INSURANCE = 6;

const VALID_DECK_COUNTS = [1, 2, 4, 6, 8] as const;

const SUGGESTION_LABELS: Record<number, string> = {
  [BJ_SUGGEST_HIT]: 'ヒット',
  [BJ_SUGGEST_STAND]: 'スタンド',
  [BJ_SUGGEST_DOUBLE]: 'ダブルダウン',
  [BJ_SUGGEST_SPLIT]: 'スプリット',
  [BJ_SUGGEST_SURRENDER]: 'サレンダー',
  [BJ_SUGGEST_DECLINE_INSURANCE]: '辞退',
};

function highlightClass(base: string, isHighlighted: boolean): string {
  return isHighlighted ? `${base} ring-2 ring-white ring-offset-1` : base;
}

export function BlackJackPage() {
  const [state, setState] = useState<BlackJackResponse | null>(null);
  const [message, setMessage] = useState('');
  const [betAmount, setBetAmount] = useState(10);
  const [deckCount, setDeckCount] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const exec = useCallback(
    async (
      command:
        | 'reset'
        | 'hit'
        | 'stand'
        | 'bet'
        | 'doubledown'
        | 'split'
        | 'insurance'
        | 'declineinsurance'
        | 'surrender'
        | 'togglehint'
        | 'setdeckcount',
      amount?: number,
    ) => {
      setLoading(true);
      try {
        setError(null);
        const res = await blackjackApi.exec(command, amount);
        setState(res);
        setMessage(res.message);
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

  const phase = state?.phase ?? BjPhase.BET;
  const hands = state?.hands ?? [];
  const currentHandIdx = state?.currentHandIdx ?? 0;
  const currentHand = hands[currentHandIdx];
  const playerChips = state?.player?.chips ?? 0;
  const hintEnabled = state?.hintEnabled ?? false;
  const suggestedAction = state?.suggestedAction ?? BJ_SUGGEST_NONE;

  const handleDeckCountChange = (newCount: number) => {
    setDeckCount(newCount);
    exec('setdeckcount', newCount);
  };

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#008000]">
      {/* Chip info bar */}
      {state && (
        <div className="shrink-0 bg-black/40 text-white text-sm px-4 py-1.5 flex justify-between">
          <span>プレイヤー: {state.player.chips} chips</span>
          <span>デッキ: {state.deckCount}デッキ</span>
          <span>ディーラー: {state.dealer.chips} chips</span>
        </div>
      )}

      {/* Scrollable: dealer area */}
      <div className="flex-1 overflow-y-auto p-4">
        {state && phase !== BjPhase.BET && (
          <div>
            <h3 className="text-white">ディーラー手札</h3>
            <h3 className="text-white">スコア {state.dealer.score ? state.dealer.score : ''}</h3>
            <div className="flex flex-wrap gap-2">
              {state.dealer.cards?.map((card) => (
                <CardImage key={`${card.design}-${card.value}`} card={card} width={60} />
              ))}
              {!state.dealer.score && <CardBack width={60} />}
            </div>
          </div>
        )}
      </div>

      {/* Sticky footer: player hand + result + buttons */}
      <div
        className="shrink-0 bg-[#005a00] border-t border-white/15 px-4 py-3"
        style={{ paddingBottom: 'calc(env(safe-area-inset-bottom) + 12px)' }}
      >
        {/* Player hands */}
        {state && phase !== BjPhase.BET && hands.length > 0 && (
          <div className="mb-2">
            {hands.map((hand, handIndex) => (
              <div key={`hand-${hand.score}-${hand.bet}-${hand.cards.length}`} className="mb-2">
                <h3 className="text-white mt-0 mb-0.5">
                  {hands.length > 1 ? `ハンド ${handIndex + 1}` : 'プレイヤー手札'}
                  {handIndex === currentHandIdx && phase === BjPhase.ACTION && ' (*)'}
                  {hand.busted && ' [BUST]'}
                  {hand.doubled && ' [DD]'}
                  {hand.isBlackJack && ' [BJ]'}
                  {hand.surrendered && (
                    <span className="ml-1 text-xs bg-gray-500 text-white px-1 rounded">SURRENDER</span>
                  )}
                </h3>
                <h3 className="text-white mt-0 mb-0.5">
                  スコア {hand.score} / ベット {hand.bet}
                </h3>
                <div className="flex flex-wrap gap-1.5">
                  {hand.cards.map((card) => (
                    <CardImage key={`${card.design}-${card.value}-${handIndex}`} card={card} width={60} />
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Insurance info */}
        {state && state.insuranceBet > 0 && (
          <div className="text-yellow-300 text-sm mb-1">インシュランス: {state.insuranceBet}</div>
        )}

        {/* Hint banner */}
        {hintEnabled && suggestedAction !== BJ_SUGGEST_NONE && (
          <div className="bg-yellow-300/90 text-gray-900 text-center text-sm font-bold px-3 py-1 rounded mb-2">
            推奨: {SUGGESTION_LABELS[suggestedAction]}
          </div>
        )}

        {/* Result message */}
        {message && (
          <div className="bg-black/70 text-white text-center px-4 py-2 text-[1.1em] font-bold mb-2 rounded-lg">
            {message}
          </div>
        )}

        <ErrorAlert message={error} />

        {/* Phase-based buttons */}
        <div className="text-center">
          {phase === BjPhase.BET && (
            <>
              <div className="flex items-center justify-center gap-2 mb-2">
                <label htmlFor="bj-bet-amount" className="text-white text-sm">
                  ベット額:
                </label>
                <input
                  id="bj-bet-amount"
                  type="number"
                  min={10}
                  step={10}
                  value={betAmount}
                  onChange={(e) => setBetAmount(Number(e.target.value))}
                  className="w-20 px-2 py-1 rounded text-sm"
                />
              </div>
              <div className="flex items-center justify-center gap-2 mb-2">
                <label htmlFor="bj-deck-count" className="text-white text-sm">
                  デッキ数:
                </label>
                <select
                  id="bj-deck-count"
                  value={deckCount}
                  onChange={(e) => handleDeckCountChange(Number(e.target.value))}
                  className="px-2 py-1 rounded text-sm"
                  disabled={loading}
                >
                  {VALID_DECK_COUNTS.map((d) => (
                    <option key={d} value={d}>
                      {d}デッキ
                    </option>
                  ))}
                </select>
              </div>
              <div className="flex items-center justify-center gap-2 mb-2">
                <button
                  type="button"
                  className={hintEnabled ? btnSuccess : btnWarning}
                  disabled={loading}
                  onClick={() => exec('togglehint')}
                >
                  ヒント {hintEnabled ? 'ON' : 'OFF'}
                </button>
              </div>
              <button type="button" className={btnPrimary} disabled={loading} onClick={() => exec('bet', betAmount)}>
                ベット
              </button>
            </>
          )}

          {phase === BjPhase.INSURANCE && (
            <>
              <button type="button" className={btnWarning} disabled={loading} onClick={() => exec('insurance')}>
                インシュランス
              </button>
              <button
                type="button"
                className={highlightClass(btnDanger, suggestedAction === BJ_SUGGEST_DECLINE_INSURANCE && hintEnabled)}
                disabled={loading}
                onClick={() => exec('declineinsurance')}
              >
                辞退
              </button>
            </>
          )}

          {phase === BjPhase.ACTION && (
            <>
              <button
                type="button"
                className={highlightClass(btnPrimary, suggestedAction === BJ_SUGGEST_HIT && hintEnabled)}
                disabled={loading}
                onClick={() => exec('hit')}
              >
                ヒット
              </button>
              <button
                type="button"
                className={highlightClass(btnPrimary, suggestedAction === BJ_SUGGEST_STAND && hintEnabled)}
                disabled={loading}
                onClick={() => exec('stand')}
              >
                スタンド
              </button>
              {currentHand && currentHand.cards.length === 2 && playerChips >= currentHand.bet && (
                <button
                  type="button"
                  className={highlightClass(btnWarning, suggestedAction === BJ_SUGGEST_DOUBLE && hintEnabled)}
                  disabled={loading}
                  onClick={() => exec('doubledown')}
                >
                  ダブルダウン
                </button>
              )}
              {currentHand?.canSplit && playerChips >= currentHand.bet && (
                <button
                  type="button"
                  className={highlightClass(btnSuccess, suggestedAction === BJ_SUGGEST_SPLIT && hintEnabled)}
                  disabled={loading}
                  onClick={() => exec('split')}
                >
                  スプリット
                </button>
              )}
              {currentHand?.canSurrender && (
                <button
                  type="button"
                  className={highlightClass(btnDanger, suggestedAction === BJ_SUGGEST_SURRENDER && hintEnabled)}
                  disabled={loading}
                  onClick={() => exec('surrender')}
                >
                  サレンダー
                </button>
              )}
            </>
          )}

          {phase === BjPhase.END && (
            <button type="button" className={btnPrimary} disabled={loading} onClick={() => exec('reset')}>
              リセット
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
