import { useCallback, useEffect, useState } from 'react'
import { oldmaidApi } from '../api/gameApi'
import { CardImage, CardBack } from '../components/CardImage'
import type { OldMaidResponse, OldMaidPlayerData, CpuAction, Card } from '../types/card'

const tableStyle: React.CSSProperties = {
  backgroundColor: '#1a5c1a',
  borderRadius: 16,
  padding: 16,
  margin: '10px auto',
  maxWidth: 800,
  boxShadow: '0 4px 12px rgba(0,0,0,0.5)',
}

const playerAreaBase: React.CSSProperties = {
  background: 'rgba(0,0,0,0.35)',
  borderRadius: 10,
  padding: 8,
  border: '2px solid transparent',
  flex: '1 1 140px',
  minWidth: 120,
}

function playerName(idx: number): string {
  return idx === 0 ? 'あなた' : `CPU ${idx}`
}

function cardLabel(card: OldMaidResponse['lastDrawCard']): string {
  if (!card) return ''
  if (card.design === 'JOKER') return 'JOKER'
  return `${card.design} ${card.value}`
}

interface PlayerAreaProps {
  player: OldMaidPlayerData
  isTarget: boolean
  isHumanTurn: boolean
  gameEndFlag: boolean
  onDraw: (drawIdx: number) => void
}

function PlayerArea({ player, isTarget, isHumanTurn, gameEndFlag, onDraw }: PlayerAreaProps) {
  const areaStyle: React.CSSProperties = {
    ...playerAreaBase,
    ...(player.isFinished
      ? { opacity: 0.5 }
      : isTarget && !gameEndFlag
      ? { border: '2px solid #f0ad4e', boxShadow: '0 0 12px #f0ad4e' }
      : {}),
  }

  const showSelectable = isHumanTurn && isTarget && !player.isFinished && !player.isHuman && !gameEndFlag
  const showCount = Math.min(player.cardCount, 10)

  return (
    <div id={`player-area-${player.id}`} style={areaStyle}>
      <div style={{ color: '#fff', fontWeight: 'bold', marginBottom: 4, fontSize: '0.9em' }}>
        {playerName(player.id)}
        {player.isFinished && (
          <span style={{
            background: '#5cb85c', color: '#fff', borderRadius: 6,
            padding: '1px 6px', marginLeft: 6, fontSize: '0.8em',
          }}>上がり</span>
        )}
        {isTarget && !player.isHuman && !player.isFinished && !gameEndFlag && (
          <span style={{
            background: '#f0ad4e', color: '#222', borderRadius: 6,
            padding: '1px 6px', marginLeft: 6, fontSize: '0.8em', fontWeight: 'bold',
          }}>← 引く相手</span>
        )}
      </div>
      {!player.isFinished && (
        <div style={{ color: '#ccc', fontSize: '0.8em', marginBottom: 4 }}>
          {player.cardCount}枚
        </div>
      )}
      {showSelectable && !player.isFinished && (
        <div style={{ color: '#cfc', fontSize: '0.75em', marginBottom: 4 }}>
          引く
        </div>
      )}
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 2, justifyContent: 'center' }}>
        {player.isFinished ? null : player.isHuman ? (
          player.cards?.map((card, i) => (
            <CardImage
              key={i}
              card={card}
              style={{ width: 50 }}
            />
          ))
        ) : showSelectable ? (
          <>
            {Array.from({ length: showCount }).map((_, i) => (
              <CardBack
                key={i}
                style={{
                  width: 40,
                  border: '2px solid transparent',
                  borderRadius: 4,
                  cursor: 'pointer',
                }}
                onClick={() => onDraw(i)}
              />
            ))}
            {player.cardCount > 10 && (
              <span style={{ color: '#fff', alignSelf: 'center', marginLeft: 2, fontSize: '0.8em' }}>
                +{player.cardCount - 10}
              </span>
            )}
          </>
        ) : (
          <>
            {Array.from({ length: showCount }).map((_, i) => (
              <CardBack key={i} style={{ width: 40 }} />
            ))}
            {player.cardCount > 10 && (
              <span style={{ color: '#fff', alignSelf: 'center', marginLeft: 2, fontSize: '0.8em' }}>
                +{player.cardCount - 10}
              </span>
            )}
          </>
        )}
      </div>
    </div>
  )
}

function DiscardedArea({ cards }: { cards: Card[] | undefined }) {
  if (!cards || cards.length === 0) {
    return (
      <div style={{
        height: 90, display: 'flex', alignItems: 'center', justifyContent: 'center',
        border: '2px dashed rgba(255,255,255,0.15)', borderRadius: 10, margin: '8px 0',
        color: 'rgba(255,255,255,0.3)', fontSize: '0.9em'
      }}>
        捨て札エリア
      </div>
    )
  }
  return (
    <div style={{
      margin: '8px 0', padding: '8px',
      background: 'rgba(0,0,0,0.2)', borderRadius: 10,
      textAlign: 'center', minHeight: 90
    }}>
      <div style={{ color: '#ccc', fontSize: '0.8em', marginBottom: 6 }}>直前に捨てられたカード</div>
      <div style={{ display: 'flex', justifyContent: 'center', gap: 12 }}>
        {cards.map((c, i) => (
          <CardImage key={i} card={c} style={{ width: 55 }} />
        ))}
      </div>
    </div>
  )
}

export function OldMaidPage() {
  const [state, setState] = useState<OldMaidResponse | null>(null)

  const exec = useCallback(async (command: 'reset' | 'draw', drawIdx?: number) => {
    try {
      const res = await oldmaidApi.exec(command, drawIdx)
      setState(res)
    } catch {
      console.error('oldmaid request failed')
    }
  }, [])

  useEffect(() => { exec('reset') }, [exec])

  if (!state) return null

  const isHumanTurn = !state.gameEndFlag && state.currentTurn === 0
  const cpuPlayers = state.players.filter(p => !p.isHuman)
  const humanPlayer = state.players.find(p => p.isHuman)

  const statusLines: string[] = []
  if (!state.gameEndFlag && state.hasDrawn) {
    let msg = `${playerName(state.lastDrawPlayerIdx)}が${playerName(state.lastDrawFromIdx)}から1枚引きました`
    if (state.lastDrawCard) msg += ` (${cardLabel(state.lastDrawCard)})`
    if (state.lastDiscardedPairs > 0) msg += `。${state.lastDiscardedPairs}組捨てました`
    statusLines.push(msg)
  }
  if (!state.gameEndFlag && state.currentTurn === 0) {
    statusLines.push(`あなたの番！ ${playerName(state.nextDrawTargetIdx)}のカードをクリックして引いてください。`)
  }

  return (
    <div style={tableStyle}>
      {/* CPU row */}
      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginBottom: 8, justifyContent: 'center' }}>
        {cpuPlayers.map(player => (
          <PlayerArea
            key={player.id}
            player={player}
            isTarget={state.nextDrawTargetIdx === player.id}
            isHumanTurn={isHumanTurn}
            gameEndFlag={state.gameEndFlag}
            onDraw={drawIdx => exec('draw', drawIdx)}
          />
        ))}
      </div>

      {/* Discarded Area */}
      <DiscardedArea cards={state.lastDiscardedCards} />

      {/* Human player */}
      {humanPlayer && (
        <PlayerArea
          player={humanPlayer}
          isTarget={false}
          isHumanTurn={isHumanTurn}
          gameEndFlag={state.gameEndFlag}
          onDraw={drawIdx => exec('draw', drawIdx)}
        />
      )}

      {/* Status */}
      {statusLines.length > 0 && (
        <div style={{
          background: 'rgba(0,0,0,0.5)',
          borderRadius: 8,
          color: '#fff',
          padding: '8px 12px',
          margin: '8px 0',
          whiteSpace: 'pre-line',
          fontSize: '0.9em',
        }}>
          {statusLines.join('\n')}
        </div>
      )}

      {/* CPU log */}
      {state.cpuActions && state.cpuActions.length > 0 && (
        <div style={{
          background: 'rgba(0,0,0,0.4)',
          borderRadius: 8,
          color: '#ccc',
          padding: '6px 10px',
          margin: '6px 0',
          whiteSpace: 'pre-line',
          fontSize: '0.8em',
          maxHeight: 120,
          overflowY: 'auto'
        }}>
          {['[CPUの行動]', ...state.cpuActions.map((action: CpuAction) => {
            let msg = `${playerName(action.drawPlayerIdx)}が${playerName(action.drawFromIdx)}から1枚引きました`
            if (action.drawnCard) {
              msg += ` (${cardLabel(action.drawnCard)})`
            }
            if (action.discardedPairs > 0) msg += `。${action.discardedPairs}組捨てました`
            return msg
          })].join('\n')}
        </div>
      )}

      {/* Result */}
      {state.message && (
        <div style={{
          background: 'rgba(0,0,0,0.6)',
          borderRadius: 10,
          color: '#fff',
          textAlign: 'center',
          padding: '10px 16px',
          fontSize: '1.2em',
          fontWeight: 'bold',
          margin: '8px 0',
        }}>
          {state.message}
        </div>
      )}

      {/* Buttons */}
      <div style={{ textAlign: 'center', margin: '12px 0 4px 0' }}>
        <button className="btn btn-primary btn-sm" style={{ margin: '0 4px', minWidth: 80 }}
          onClick={() => exec('reset')}>
          リセット
        </button>
        <button className="btn btn-warning btn-sm" style={{ margin: '0 4px', minWidth: 110 }}
          disabled={!isHumanTurn || state.gameEndFlag}
          onClick={() => exec('draw')}>
          ランダムに引く
        </button>
      </div>
    </div>
  )
}
