import { useCallback, useEffect, useState } from 'react';
import { blackjackApi } from '../api/gameApi';
import { CardBack, CardImage } from '../components/CardImage';
import type { BlackJackResponse } from '../types/card';

const btnPrimary =
  'px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed mx-1.5';

export function BlackJackPage() {
  const [state, setState] = useState<BlackJackResponse | null>(null);
  const [message, setMessage] = useState('');

  const exec = useCallback(async (command: 'reset' | 'hit' | 'stand') => {
    try {
      const res = await blackjackApi.exec(command);
      setState(res);
      setMessage(res.message ?? '');
    } catch {
      console.error('blackjack request failed');
    }
  }, []);

  useEffect(() => {
    exec('reset');
  }, [exec]);

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#008000]">
      {/* Scrollable: dealer area */}
      <div className="flex-1 overflow-y-auto p-4">
        {state && (
          <div>
            <h3 className="text-white">ディーラー手札</h3>
            <h3 className="text-white">スコア {state.dealer.score !== 0 ? state.dealer.score : ''}</h3>
            <div className="flex flex-wrap gap-2">
              {state.dealer.cards.map((card) => (
                <CardImage key={`${card.design}-${card.value}`} card={card} width={60} />
              ))}
              {state.dealer.score === 0 && <CardBack width={60} />}
            </div>
          </div>
        )}
      </div>

      {/* Sticky footer: player hand + result + buttons */}
      <div
        className="shrink-0 bg-[#005a00] border-t border-white/15 px-4 py-3"
        style={{ paddingBottom: 'calc(env(safe-area-inset-bottom) + 12px)' }}
      >
        {state && (
          <div className="mb-2">
            <h3 className="text-white mt-0 mb-0.5">プレイヤー手札</h3>
            <h3 className="text-white mt-0 mb-2">スコア {state.player.score}</h3>
            <div className="flex flex-wrap gap-1.5">
              {state.player.cards.map((card) => (
                <CardImage key={`${card.design}-${card.value}`} card={card} width={60} />
              ))}
            </div>
          </div>
        )}

        {/* Result message */}
        {message && (
          <div className="bg-black/70 text-white text-center px-4 py-2 text-[1.1em] font-bold mb-2 rounded-lg">
            {message}
          </div>
        )}

        {/* Buttons */}
        <div className="text-center">
          <button type="button" className={btnPrimary} onClick={() => exec('reset')}>
            リセット
          </button>
          <button type="button" className={btnPrimary} onClick={() => exec('hit')}>
            ヒット
          </button>
          <button type="button" className={btnPrimary} onClick={() => exec('stand')}>
            スタンド
          </button>
        </div>
      </div>
    </div>
  );
}
