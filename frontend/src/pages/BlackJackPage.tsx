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
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0, background: '#008000' }}>
      {/* Scrollable: dealer area */}
      <div style={{ flex: 1, overflowY: 'auto', padding: '16px' }}>
        {state && (
          <div>
            <h3 className="text-white">ディーラー手札</h3>
            <h3 className="text-white">スコア {state.dealer.score !== 0 ? state.dealer.score : ''}</h3>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
              {state.dealer.cards.map((card) => (
                <CardImage key={`${card.design}-${card.value}`} card={card} />
              ))}
              {state.dealer.score === 0 && <CardBack />}
            </div>
          </div>
        )}
      </div>

      {/* Sticky footer: player hand + result + buttons */}
      <div
        style={{
          flexShrink: 0,
          background: '#005a00',
          borderTop: '1px solid rgba(255,255,255,0.15)',
          padding: '12px 16px',
          paddingBottom: 'calc(env(safe-area-inset-bottom) + 12px)',
        }}
      >
        {state && (
          <div style={{ marginBottom: 8 }}>
            <h3 className="text-white" style={{ margin: '0 0 2px' }}>
              プレイヤー手札
            </h3>
            <h3 className="text-white" style={{ margin: '0 0 8px' }}>
              スコア {state.player.score}
            </h3>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
              {state.player.cards.map((card) => (
                <CardImage key={`${card.design}-${card.value}`} card={card} width={60} />
              ))}
            </div>
          </div>
        )}

        {/* Result message */}
        {message && (
          <div
            style={{
              background: 'rgba(0,0,0,0.7)',
              color: '#fff',
              textAlign: 'center',
              padding: '8px 16px',
              fontSize: '1.1em',
              fontWeight: 'bold',
              marginBottom: 8,
              borderRadius: 8,
            }}
          >
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
