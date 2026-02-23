import { useCallback, useEffect, useRef, useState } from 'react';
import { oldmaidApi } from '../api/gameApi';
import { CardBack, CardImage } from '../components/CardImage';
import { btnPrimary, btnWarning } from '../styles/buttonStyles';
import type { Card, CpuAction, OldMaidPlayerData, OldMaidResponse } from '../types/card';
import { findPlayerName, playerName } from '../utils/playerUtils';

const REPLAY_DELAY_MS = 800;

const playerAreaBaseClass = 'bg-black/35 rounded-[10px] p-2 border-2 border-transparent flex-[1_1_140px] min-w-[120px]';

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
      <div className="text-white font-bold mb-1 text-[0.9em]">
        {playerName(player.id, player.isHuman)}
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
      {!player.isFinished && <div className="text-[#ccc] text-[0.8em] mb-1">{player.cardCount}枚</div>}
      {showSelectable && !player.isFinished && <div className="text-[#cfc] text-[0.75em] mb-1">引く</div>}
      <div className="flex flex-wrap gap-0.5 justify-center">
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
      <div className="h-[90px] flex items-center justify-center border-2 border-dashed border-white/15 rounded-[10px] my-2 text-white/30 text-[0.9em]">
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
    <div className="my-2 p-2 bg-black/20 rounded-[10px] text-center min-h-[90px]">
      <div className="text-[#ccc] text-[0.8em] mb-1.5">直前に捨てられたカード</div>
      <div className="flex justify-center gap-5 items-end">
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
  const isHumanTurn = !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const humanPlayer = state.players.find((p) => p.isHuman);

  const statusLines: string[] = [];
  if (!state.gameEndFlag && state.hasDrawn) {
    let msg = `${findPlayerName(state.players, state.lastDrawPlayerIdx)}が${findPlayerName(state.players, state.lastDrawFromIdx)}から1枚引きました`;
    if (state.lastDrawCard) msg += ` (${cardLabel(state.lastDrawCard)})`;
    if (state.lastDiscardedPairs > 0) msg += `。${state.lastDiscardedPairs}組捨てました`;
    statusLines.push(msg);
  }
  if (isHumanTurn) {
    statusLines.push(
      `あなたの番！ ${findPlayerName(state.players, state.nextDrawTargetIdx)}のカードをクリックして引いてください。`,
    );
  }

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a5c1a]">
      {/* Scrollable: CPU rows + discard + status + logs + result */}
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {/* CPU row */}
        <div className="flex gap-2 flex-wrap mb-2 justify-center">
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
          <div className="bg-black/50 rounded-lg text-white py-2 px-3 my-2 whitespace-pre-line text-[0.9em]">
            {statusLines.join('\n')}
          </div>
        )}

        {/* CPU log */}
        {state.cpuActions && state.cpuActions.length > 0 && (
          <div className="bg-black/40 rounded-lg text-[#ccc] py-1.5 px-2.5 my-1.5 whitespace-pre-line text-[0.8em] max-h-[120px] overflow-y-auto">
            {[
              '[CPUの行動]',
              ...state.cpuActions.map((action: CpuAction) => {
                let msg = `${findPlayerName(state.players, action.drawPlayerIdx)}が${findPlayerName(state.players, action.drawFromIdx)}から1枚引きました`;
                // CPU drawn card is intentionally hidden to preserve game fairness
                if (action.discardedPairs > 0) msg += `。${action.discardedPairs}組捨てました`;
                return msg;
              }),
            ].join('\n')}
          </div>
        )}

        {/* Result */}
        {state.message && (
          <div className="bg-black/60 rounded-[10px] text-white text-center py-2.5 px-4 text-[1.2em] font-bold my-2">
            {state.message}
          </div>
        )}
      </div>

      {/* Sticky footer: human player hand + buttons */}
      <div
        className="shrink-0 bg-[#163e16] border-t border-white/20 px-4 py-2.5"
        style={{ paddingBottom: 'calc(env(safe-area-inset-bottom) + 10px)' }}
      >
        {/* Human player */}
        {humanPlayer && (
          <div className="mb-2">
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
