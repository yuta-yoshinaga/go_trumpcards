import { useCallback, useEffect, useState } from 'react'
import { oldmaidApi } from '../api/gameApi'
import { CardImage, CardBack } from '../components/CardImage'
import type { OldMaidResponse, OldMaidPlayerData, CpuAction } from '../types/card'

const tableStyle: React.CSSProperties = {
  backgroundColor: '#1a5c1a',
  borderRadius: 16,
  padding: 20,
  margin: '10px auto',
  maxWidth: 960,
}

const playerAreaBase: React.CSSProperties = {
  background: 'rgba(0,0,0,0.35)',
  borderRadius: 10,
  padding: 10,
  border: '2px solid transparent',
  flex: '1 1 180px',
  minWidth: 150,
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
      <div style={{ color: '#fff', fontWeight: 'bold', marginBottom: 4 }}>
        {playerName(player.id)}
        {player.isFinished && (
          <span style={{
            background: '#5cb85c', color: '#fff', borderRadius: 6,
            padding: '1px 8px', marginLeft: 6, fontSize: '0.8em',
          }}>上がり</span>
        )}
        {isTarget && !player.isHuman && !player.isFinished && !gameEndFlag && (
          <span style={{
            background: '#f0ad4e', color: '#222', borderRadius: 6,
            padding: '1px 8px', marginLeft: 6, fontSize: '0.8em', fontWeight: 'bold',
          }}>← 引く相手</span>
        )}
      </div>
      {!player.isFinished && (
        <div style={{ color: '#ccc', fontSize: '0.85em', marginBottom: 4 }}>
          {player.cardCount}枚
        </div>
      )}
      {showSelectable && !player.isFinished && (
        <div style={{ color: '#cfc', fontSize: '0.8em', marginBottom: 4 }}>
          クリックして引いてください
        </div>
      )}
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
        {player.isFinished ? null : player.isHuman ? (
          player.cards?.map((card, i) => (
            <CardImage
              key={i}
              card={card}
              style={{ width: 60 }}
            />
          ))
        ) : showSelectable ? (
          <>
            {Array.from({ length: showCount }).map((_, i) => (
              <CardBack
                key={i}
                style={{
                  width: 60,
                  border: '2px solid transparent',
                  borderRadius: 6,
                  cursor: 'pointer',
                }}
                onClick={() => onDraw(i)}
              />
            ))}
            {player.cardCount > 10 && (
              <span style={{ color: '#fff', alignSelf: 'center', marginLeft: 4 }}>
                +{player.cardCount - 10}
              </span>
            )}
          </>
        ) : (
          <>
            {Array.from({ length: showCount }).map((_, i) => (
              <CardBack key={i} style={{ width: 60 }} />
            ))}
            {player.cardCount > 10 && (
              <span style={{ color: '#fff', alignSelf: 'center', marginLeft: 4 }}>
                +{player.cardCount - 10}
              </span>
            )}
          </>
        )}
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
      <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', marginBottom: 12 }}>
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

      <div style={{ borderTop: '1px solid rgba(255,255,255,0.2)', margin: '12px 0' }} />

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
          padding: '10px 16px',
          margin: '10px 0',
          whiteSpace: 'pre-line',
          fontSize: '0.95em',
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
          padding: '8px 14px',
          margin: '8px 0',
          whiteSpace: 'pre-line',
          fontSize: '0.85em',
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
          background: 'rgba(0,0,0,0.55)',
          borderRadius: 10,
          color: '#fff',
          textAlign: 'center',
          padding: '12px 20px',
          fontSize: '1.3em',
          fontWeight: 'bold',
          margin: '10px 0',
        }}>
          {state.message}
        </div>
      )}

      {/* Buttons */}
      <div style={{ textAlign: 'center', margin: '14px 0 4px 0' }}>
        <button className="btn btn-primary btn-sm" style={{ margin: '0 6px', minWidth: 90 }}
          onClick={() => exec('reset')}>
          リセット
        </button>
        <button className="btn btn-warning btn-sm" style={{ margin: '0 6px', minWidth: 120 }}
          disabled={!isHumanTurn || state.gameEndFlag}
          onClick={() => exec('draw')}>
          ランダムに引く
        </button>
      </div>
    </div>
  )
}
