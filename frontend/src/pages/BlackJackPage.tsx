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
    <div>
      <div className="mx-auto bg-[#008000] p-4">
        {state && (
          <>
            {/* Dealer area */}
            <div className="mb-4">
              <h3 className="text-white">ディーラー手札</h3>
              <h3 className="text-white">スコア {state.dealer.score !== 0 ? state.dealer.score : ''}</h3>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                {state.dealer.cards.map((card) => (
                  <CardImage key={`${card.design}-${card.value}`} card={card} />
                ))}
                {state.dealer.score === 0 && <CardBack />}
              </div>
            </div>

            {/* Player area */}
            <div>
              <h3 className="text-white">プレイヤー手札</h3>
              <h3 className="text-white">スコア {state.player.score}</h3>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                {state.player.cards.map((card) => (
                  <CardImage key={`${card.design}-${card.value}`} card={card} />
                ))}
              </div>
            </div>
          </>
        )}
      </div>

      {/* Result message */}
      {message && (
        <div
          style={{
            background: 'rgba(0,0,0,0.7)',
            color: '#fff',
            textAlign: 'center',
            padding: '12px 20px',
            fontSize: '1.3em',
            fontWeight: 'bold',
            margin: '10px auto',
            maxWidth: 600,
            borderRadius: 10,
          }}
        >
          {message}
        </div>
      )}

      {/* Buttons */}
      <div className="text-center my-5">
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
  );
}
