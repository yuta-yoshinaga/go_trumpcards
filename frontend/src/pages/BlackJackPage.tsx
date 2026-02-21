import { useCallback, useEffect, useState } from 'react'
import { blackjackApi } from '../api/gameApi'
import { CardImage, CardBack } from '../components/CardImage'
import type { BlackJackResponse } from '../types/card'

export function BlackJackPage() {
  const [state, setState] = useState<BlackJackResponse | null>(null)
  const [message, setMessage] = useState('')

  const exec = useCallback(async (command: 'reset' | 'hit' | 'stand') => {
    try {
      const res = await blackjackApi.exec(command)
      setState(res)
      setMessage(res.message ?? '')
    } catch {
      console.error('blackjack request failed')
    }
  }, [])

  useEffect(() => { exec('reset') }, [exec])

  return (
    <div>
      <div
        
        style={{ margin: 'auto', backgroundColor: '#008000', padding: 16 }}
      >
        {state && (
          <>
            {/* Dealer area */}
            <div style={{ marginBottom: 16 }}>
              <h3  style={{ color: '#fff' }}>ディーラー手札</h3>
              <h3  style={{ color: '#fff' }}>
                スコア {state.dealer.score !== 0 ? state.dealer.score : ''}
              </h3>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                {state.dealer.cards.map((card, i) => (
                  <CardImage key={i} card={card} />
                ))}
                {state.dealer.score === 0 && <CardBack />}
              </div>
            </div>

            {/* Player area */}
            <div>
              <h3  style={{ color: '#fff' }}>プレイヤー手札</h3>
              <h3  style={{ color: '#fff' }}>
                スコア {state.player.score}
              </h3>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                {state.player.cards.map((card, i) => (
                  <CardImage key={i} card={card} />
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
      <div  style={{ textAlign: 'center', margin: '20px auto' }}>
        <button className="px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed" style={{ margin: '0 6px' }} onClick={() => exec('reset')}>
          リセット
        </button>
        <button className="px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed" style={{ margin: '0 6px' }} onClick={() => exec('hit')}>
          ヒット
        </button>
        <button className="px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed" style={{ margin: '0 6px' }} onClick={() => exec('stand')}>
          スタンド
        </button>
      </div>
    </div>
  )
}
