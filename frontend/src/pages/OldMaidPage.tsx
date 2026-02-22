import { useCallback, useEffect, useRef, useState } from 'react';
import { oldmaidApi } from '../api/gameApi';
import { CardBack, CardImage } from '../components/CardImage';
import type { Card, CpuAction, OldMaidPlayerData, OldMaidResponse } from '../types/card';

const REPLAY_DELAY_MS = 800;

const playerAreaBaseClass = 'bg-black/35 rounded-[10px] p-2 border-2 border-transparent flex-[1_1_140px] min-w-[120px]';
const btnPrimary =
  'px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed mx-1';
const btnWarning =
  'px-3 py-1.5 text-sm font-medium text-gray-900 bg-yellow-400 rounded hover:bg-yellow-500 disabled:opacity-50 disabled:cursor-not-allowed mx-1';

function playerName(idx: number): string {
  return idx === 0 ? 'あなた' : `CPU ${idx}`;
}

function cardLabel(card: OldMaidResponse['lastDrawCard']): string {
  if (!card) return '';
  if (card.design === 'JOKER') return 'JOKER';
  return `${card.design} ${card.value}`;
}

const delay = (ms: number) => new Promise<void>((resolve) => setTimeout(resolve, ms));

/** Compute intermediate player card counts by reversing all CPU actions from the final state,
 *  then replay forward. Returns one OldMaidResponse per CPU action (state after each action). */
function buildReplayStates(finalState: OldMaidResponse): OldMaidResponse[] {
  const actions = finalState.cpuActions;
  if (actions.length === 0) return [];

  // Work backwards to get counts before all CPU actions
  const counts = finalState.players.map((p) => p.cardCount);
  for (let i = actions.length - 1; i >= 0; i--) {
    const a = actions[i];
    counts[a.drawPlayerIdx] = counts[a.drawPlayerIdx] + 2 * a.discardedPairs - 1;
    counts[a.drawFromIdx] = counts[a.drawFromIdx] + 1;
  }

  // Play forward, building a display state after each CPU action
  const states: OldMaidResponse[] = [];
  for (let i = 0; i < actions.length; i++) {
    const a = actions[i];
    counts[a.drawFromIdx] -= 1;
    counts[a.drawPlayerIdx] += 1 - 2 * a.discardedPairs;

    const isLastAction = i === actions.length - 1;
    states.push({
      ...finalState,
      players: finalState.players.map((p, idx) => ({
        ...p,
        cardCount: Math.max(0, counts[idx]),
        isFinished: counts[idx] <= 0,
      })),
      currentTurn: a.drawPlayerIdx,
      hasDrawn: true,
      lastDrawPlayerIdx: a.drawPlayerIdx,
      lastDrawFromIdx: a.drawFromIdx,
      lastDrawCard: a.drawnCard,
      lastDiscardedPairs: a.discardedPairs,
      lastDiscardedCards: a.discardedCards ?? [],
      cpuActions: actions.slice(0, i + 1),
      gameEndFlag: isLastAction ? finalState.gameEndFlag : false,
      message: isLastAction ? finalState.message : '',
      nextDrawTargetIdx: isLastAction
        ? finalState.nextDrawTargetIdx
        : i + 1 < actions.length
          ? actions[i + 1].drawFromIdx
          : finalState.nextDrawTargetIdx,
    });
  }
  return states;
}

/** Build the display state right after human's draw, before any CPU actions. */
function buildHumanDrawState(finalState: OldMaidResponse): OldMaidResponse | null {
  const ha = finalState.humanAction;
  if (!ha) return null;

  const counts = finalState.players.map((p) => p.cardCount);
  for (let i = finalState.cpuActions.length - 1; i >= 0; i--) {
    const a = finalState.cpuActions[i];
    counts[a.drawPlayerIdx] = counts[a.drawPlayerIdx] + 2 * a.discardedPairs - 1;
    counts[a.drawFromIdx] = counts[a.drawFromIdx] + 1;
  }

  return {
    ...finalState,
    players: finalState.players.map((p, idx) => ({
      ...p,
      cardCount: Math.max(0, counts[idx]),
      isFinished: counts[idx] <= 0,
    })),
    currentTurn: finalState.cpuActions.length > 0 ? finalState.cpuActions[0].drawPlayerIdx : finalState.currentTurn,
    hasDrawn: true,
    lastDrawPlayerIdx: ha.drawPlayerIdx,
    lastDrawFromIdx: ha.drawFromIdx,
    lastDrawCard: ha.drawnCard,
    lastDiscardedPairs: ha.discardedPairs,
    lastDiscardedCards: ha.discardedCards ?? [],
    cpuActions: [],
    gameEndFlag: false,
    message: '',
    nextDrawTargetIdx:
      finalState.cpuActions.length > 0 ? finalState.cpuActions[0].drawFromIdx : finalState.nextDrawTargetIdx,
  };
}

interface PlayerAreaProps {
  player: OldMaidPlayerData;
  isTarget: boolean;
  isHumanTurn: boolean;
  gameEndFlag: boolean;
  onDraw: (drawIdx: number) => void;
}

function PlayerArea({ player, isTarget, isHumanTurn, gameEndFlag, onDraw }: PlayerAreaProps) {
  const conditionalStyle: React.CSSProperties = player.isFinished
    ? { opacity: 0.5 }
    : isTarget && !gameEndFlag
      ? { border: '2px solid #f0ad4e', boxShadow: '0 0 12px #f0ad4e' }
      : {};

  const showSelectable = isHumanTurn && isTarget && !player.isFinished && !player.isHuman && !gameEndFlag;
  const showCount = Math.min(player.cardCount, 10);

  return (
    <div id={`player-area-${player.id}`} className={playerAreaBaseClass} style={conditionalStyle}>
      <div style={{ color: '#fff', fontWeight: 'bold', marginBottom: 4, fontSize: '0.9em' }}>
        {playerName(player.id)}
        {player.isFinished && (
          <span
            style={{
              background: '#5cb85c',
              color: '#fff',
              borderRadius: 6,
              padding: '1px 6px',
              marginLeft: 6,
              fontSize: '0.8em',
            }}
          >
            上がり
          </span>
        )}
        {isTarget && !player.isHuman && !player.isFinished && !gameEndFlag && (
          <span
            style={{
              background: '#f0ad4e',
              color: '#222',
              borderRadius: 6,
              padding: '1px 6px',
              marginLeft: 6,
              fontSize: '0.8em',
              fontWeight: 'bold',
            }}
          >
            ← 引く相手
          </span>
        )}
      </div>
      {!player.isFinished && (
        <div style={{ color: '#ccc', fontSize: '0.8em', marginBottom: 4 }}>{player.cardCount}枚</div>
      )}
      {showSelectable && !player.isFinished && (
        <div style={{ color: '#cfc', fontSize: '0.75em', marginBottom: 4 }}>引く</div>
      )}
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 2, justifyContent: 'center' }}>
        {player.isFinished ? null : player.isHuman ? (
          player.cards?.map((card) => <CardImage key={`${card.design}-${card.value}`} card={card} width={50} />)
        ) : showSelectable ? (
          <>
            {Array.from({ length: showCount }, (_, i) => {
              const cardStyle = { border: '2px solid transparent', borderRadius: 4, cursor: 'pointer' };
              // biome-ignore lint/suspicious/noArrayIndexKey: placeholder array with no card identity
              return <CardBack key={i} width={40} style={cardStyle} onClick={() => onDraw(i)} />;
            })}
            {player.cardCount > 10 && (
              <span style={{ color: '#fff', alignSelf: 'center', marginLeft: 2, fontSize: '0.8em' }}>
                +{player.cardCount - 10}
              </span>
            )}
          </>
        ) : (
          <>
            {Array.from({ length: showCount }).map((_, i) => (
              // biome-ignore lint/suspicious/noArrayIndexKey: placeholder array with no card identity
              <CardBack key={i} width={40} />
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
  );
}

/** Show discarded card pairs stacked (overlapping) to represent a pair being set aside. */
function DiscardedArea({ cards }: { cards: Card[] | undefined }) {
  if (!cards || cards.length === 0) {
    return (
      <div
        style={{
          height: 90,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          border: '2px dashed rgba(255,255,255,0.15)',
          borderRadius: 10,
          margin: '8px 0',
          color: 'rgba(255,255,255,0.3)',
          fontSize: '0.9em',
        }}
      >
        捨て札エリア
      </div>
    );
  }

  // Group cards into pairs (every 2 cards is one discarded pair)
  const pairs: [Card, Card][] = [];
  for (let i = 0; i + 1 < cards.length; i += 2) {
    pairs.push([cards[i], cards[i + 1]]);
  }
  // If odd number of cards, show the last one alone
  const remainder = cards.length % 2 === 1 ? cards[cards.length - 1] : null;

  return (
    <div
      style={{
        margin: '8px 0',
        padding: '8px',
        background: 'rgba(0,0,0,0.2)',
        borderRadius: 10,
        textAlign: 'center',
        minHeight: 90,
      }}
    >
      <div style={{ color: '#ccc', fontSize: '0.8em', marginBottom: 6 }}>直前に捨てられたカード</div>
      <div style={{ display: 'flex', justifyContent: 'center', gap: 20, alignItems: 'flex-end' }}>
        {pairs.map(([c1, c2]) => (
          <div key={`${c1.design}-${c1.value}`} style={{ position: 'relative', width: 65, height: 82 }}>
            <CardImage card={c1} width={55} style={{ position: 'absolute', left: 0, top: 0 }} />
            <CardImage card={c2} width={55} style={{ position: 'absolute', left: 10, top: 6 }} />
          </div>
        ))}
        {remainder && <CardImage card={remainder} width={55} />}
      </div>
    </div>
  );
}

export function OldMaidPage() {
  const [displayState, setDisplayState] = useState<OldMaidResponse | null>(null);
  const replayGenRef = useRef(0);

  const exec = useCallback(async (command: 'reset' | 'draw', drawIdx?: number) => {
    const myGen = ++replayGenRef.current;
    try {
      const res = await oldmaidApi.exec(command, drawIdx);
      if (myGen !== replayGenRef.current) return;

      if (command === 'reset' || res.cpuActions.length === 0) {
        setDisplayState(res);
        return;
      }

      // Show human's draw result first (if humanAction is available)
      const humanDrawState = buildHumanDrawState(res);
      if (humanDrawState) {
        setDisplayState(humanDrawState);
        await delay(REPLAY_DELAY_MS);
        if (myGen !== replayGenRef.current) return;
      }

      // Replay each CPU action step by step
      const replayStates = buildReplayStates(res);
      for (const state of replayStates) {
        setDisplayState(state);
        await delay(REPLAY_DELAY_MS);
        if (myGen !== replayGenRef.current) return;
      }
      // Restore the actual final state so currentTurn reflects the server response
      setDisplayState(res);
    } catch {
      console.error('oldmaid request failed');
    }
  }, []);

  useEffect(() => {
    exec('reset');
  }, [exec]);

  if (!displayState) return null;

  const state = displayState;
  const isHumanTurn = !state.gameEndFlag && state.currentTurn === 0;
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const humanPlayer = state.players.find((p) => p.isHuman);

  const statusLines: string[] = [];
  if (!state.gameEndFlag && state.hasDrawn) {
    let msg = `${playerName(state.lastDrawPlayerIdx)}が${playerName(state.lastDrawFromIdx)}から1枚引きました`;
    if (state.lastDrawCard) msg += ` (${cardLabel(state.lastDrawCard)})`;
    if (state.lastDiscardedPairs > 0) msg += `。${state.lastDiscardedPairs}組捨てました`;
    statusLines.push(msg);
  }
  if (!state.gameEndFlag && state.currentTurn === 0) {
    statusLines.push(`あなたの番！ ${playerName(state.nextDrawTargetIdx)}のカードをクリックして引いてください。`);
  }

  return (
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0, background: '#1a5c1a' }}>
      {/* Scrollable: CPU rows + discard + status + logs + result */}
      <div style={{ flex: 1, overflowY: 'auto', padding: '12px 16px 0' }}>
        {/* CPU row */}
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginBottom: 8, justifyContent: 'center' }}>
          {cpuPlayers.map((player) => (
            <PlayerArea
              key={player.id}
              player={player}
              isTarget={state.nextDrawTargetIdx === player.id}
              isHumanTurn={isHumanTurn}
              gameEndFlag={state.gameEndFlag}
              onDraw={(drawIdx) => exec('draw', drawIdx)}
            />
          ))}
        </div>

        {/* Discarded Area */}
        <DiscardedArea cards={state.lastDiscardedCards} />

        {/* Status */}
        {statusLines.length > 0 && (
          <div
            style={{
              background: 'rgba(0,0,0,0.5)',
              borderRadius: 8,
              color: '#fff',
              padding: '8px 12px',
              margin: '8px 0',
              whiteSpace: 'pre-line',
              fontSize: '0.9em',
            }}
          >
            {statusLines.join('\n')}
          </div>
        )}

        {/* CPU log */}
        {state.cpuActions && state.cpuActions.length > 0 && (
          <div
            style={{
              background: 'rgba(0,0,0,0.4)',
              borderRadius: 8,
              color: '#ccc',
              padding: '6px 10px',
              margin: '6px 0',
              whiteSpace: 'pre-line',
              fontSize: '0.8em',
              maxHeight: 120,
              overflowY: 'auto',
            }}
          >
            {[
              '[CPUの行動]',
              ...state.cpuActions.map((action: CpuAction) => {
                let msg = `${playerName(action.drawPlayerIdx)}が${playerName(action.drawFromIdx)}から1枚引きました`;
                // CPU drawn card is intentionally hidden to preserve game fairness
                if (action.discardedPairs > 0) msg += `。${action.discardedPairs}組捨てました`;
                return msg;
              }),
            ].join('\n')}
          </div>
        )}

        {/* Result */}
        {state.message && (
          <div
            style={{
              background: 'rgba(0,0,0,0.6)',
              borderRadius: 10,
              color: '#fff',
              textAlign: 'center',
              padding: '10px 16px',
              fontSize: '1.2em',
              fontWeight: 'bold',
              margin: '8px 0',
            }}
          >
            {state.message}
          </div>
        )}
      </div>

      {/* Sticky footer: human player hand + buttons */}
      <div
        style={{
          flexShrink: 0,
          background: '#163e16',
          borderTop: '1px solid rgba(255,255,255,0.2)',
          padding: '10px 16px',
          paddingBottom: 'calc(env(safe-area-inset-bottom) + 10px)',
        }}
      >
        {/* Human player */}
        {humanPlayer && (
          <div style={{ marginBottom: 8 }}>
            <PlayerArea
              player={humanPlayer}
              isTarget={false}
              isHumanTurn={isHumanTurn}
              gameEndFlag={state.gameEndFlag}
              onDraw={(drawIdx) => exec('draw', drawIdx)}
            />
          </div>
        )}

        {/* Buttons */}
        <div className="text-center">
          <button type="button" className={`${btnPrimary} min-w-[80px]`} onClick={() => exec('reset')}>
            リセット
          </button>
          <button
            type="button"
            className={`${btnWarning} min-w-[110px]`}
            disabled={!isHumanTurn || state.gameEndFlag}
            onClick={() => exec('draw')}
          >
            ランダムに引く
          </button>
        </div>
      </div>
    </div>
  );
}
