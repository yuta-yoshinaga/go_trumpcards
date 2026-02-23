import { useCallback, useEffect, useState } from 'react';
import { blackjackApi } from '../api/gameApi';
import { CardBack, CardImage } from '../components/CardImage';
import type { BlackJackResponse } from '../types/card';

const PHASE_BET = 1;
const PHASE_INSURANCE = 3;
const PHASE_ACTION = 4;
const PHASE_END = 5;

const btnPrimary =
  'px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed mx-1.5';
const btnWarning =
  'px-3 py-1.5 text-sm font-medium text-gray-900 bg-yellow-400 rounded hover:bg-yellow-500 disabled:opacity-50 disabled:cursor-not-allowed mx-1.5';
const btnDanger =
  'px-3 py-1.5 text-sm font-medium text-white bg-red-600 rounded hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed mx-1.5';
const btnSuccess =
  'px-3 py-1.5 text-sm font-medium text-white bg-green-600 rounded hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed mx-1.5';

export function BlackJackPage() {
  const [state, setState] = useState<BlackJackResponse | null>(null);
  const [message, setMessage] = useState('');
  const [betAmount, setBetAmount] = useState(10);

  const exec = useCallback(
    async (
      command: 'reset' | 'hit' | 'stand' | 'bet' | 'doubledown' | 'split' | 'insurance' | 'declineinsurance',
      amount?: number,
    ) => {
      try {
        const res = await blackjackApi.exec(command, amount);
        setState(res);
        setMessage(res.message ?? '');
      } catch {
        console.error('blackjack request failed');
      }
    },
    [],
  );

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const phase = state?.phase ?? PHASE_BET;
  const hands = state?.hands ?? [];
  const currentHandIdx = state?.currentHandIdx ?? 0;
  const currentHand = hands[currentHandIdx];
  const playerChips = state?.player?.chips ?? 0;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#008000]">
      {/* Chip info bar */}
      {state && (
        <div className="shrink-0 bg-black/40 text-white text-sm px-4 py-1.5 flex justify-between">
          <span>プレイヤー: {state.player.chips} chips</span>
          <span>ディーラー: {state.dealer.chips} chips</span>
        </div>
      )}

      {/* Scrollable: dealer area */}
      <div className="flex-1 overflow-y-auto p-4">
        {state && phase !== PHASE_BET && (
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
        {state && phase !== PHASE_BET && hands.length > 0 && (
          <div className="mb-2">
            {hands.map((hand, handIndex) => (
              <div key={`hand-${hand.score}-${hand.bet}-${hand.cards.length}`} className="mb-2">
                <h3 className="text-white mt-0 mb-0.5">
                  {hands.length > 1 ? `ハンド ${handIndex + 1}` : 'プレイヤー手札'}
                  {handIndex === currentHandIdx && phase === PHASE_ACTION && ' (*)'}
                  {hand.busted && ' [BUST]'}
                  {hand.doubled && ' [DD]'}
                  {hand.isBlackJack && ' [BJ]'}
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

        {/* Result message */}
        {message && (
          <div className="bg-black/70 text-white text-center px-4 py-2 text-[1.1em] font-bold mb-2 rounded-lg">
            {message}
          </div>
        )}

        {/* Phase-based buttons */}
        <div className="text-center">
          {phase === PHASE_BET && (
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
              <button type="button" className={btnPrimary} onClick={() => exec('bet', betAmount)}>
                ベット
              </button>
            </>
          )}

          {phase === PHASE_INSURANCE && (
            <>
              <button type="button" className={btnWarning} onClick={() => exec('insurance')}>
                インシュランス
              </button>
              <button type="button" className={btnDanger} onClick={() => exec('declineinsurance')}>
                辞退
              </button>
            </>
          )}

          {phase === PHASE_ACTION && (
            <>
              <button type="button" className={btnPrimary} onClick={() => exec('hit')}>
                ヒット
              </button>
              <button type="button" className={btnPrimary} onClick={() => exec('stand')}>
                スタンド
              </button>
              {currentHand && currentHand.cards.length === 2 && playerChips >= currentHand.bet && (
                <button type="button" className={btnWarning} onClick={() => exec('doubledown')}>
                  ダブルダウン
                </button>
              )}
              {currentHand?.canSplit && playerChips >= currentHand.bet && (
                <button type="button" className={btnSuccess} onClick={() => exec('split')}>
                  スプリット
                </button>
              )}
            </>
          )}

          {phase === PHASE_END && (
            <button type="button" className={btnPrimary} onClick={() => exec('reset')}>
              リセット
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
