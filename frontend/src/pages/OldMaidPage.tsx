import { useCallback, useEffect, useRef, useState } from 'react';
import { oldmaidApi } from '../api/gameApi';
import { CardBack, CardImage } from '../components/CardImage';
import { ErrorAlert } from '../components/ErrorAlert';
import { StatusBadge } from '../components/StatusBadge';
import { useGameApi } from '../hooks/useGameApi';
import { btnPrimary, btnSecondary, btnWarning } from '../styles/buttonStyles';
import { playerAreaBase } from '../styles/gameStyles';
import type { Card, CpuAction, OldMaidPlayerData, OldMaidResponse } from '../types/card';
import { cardLabel } from '../utils/cardUtils';
import { findPlayerName, playerName } from '../utils/playerUtils';

const REPLAY_DELAY_MS = 800;

const OldMaidMode = {
  Normal: 0,
  JijiNuki: 1,
} as const;

const playerAreaClass = `${playerAreaBase} p-2 flex-[1_1_140px] min-w-[120px]`;

const delay = (ms: number) => new Promise<void>((resolve) => setTimeout(resolve, ms));

/** Compute intermediate player card counts by reversing all CPU actions from the final state,
 *  then replay forward. Returns one OldMaidResponse per CPU action (state after each action). */
function buildReplayStates(finalState: OldMaidResponse): OldMaidResponse[] {
  const actions = finalState.cpuActions;

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
      nextDrawTargetIdx: isLastAction ? finalState.nextDrawTargetIdx : actions[i + 1].drawFromIdx,
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

  const [firstCpuAction] = finalState.cpuActions;

  return {
    ...finalState,
    players: finalState.players.map((p, idx) => ({
      ...p,
      cardCount: Math.max(0, counts[idx]),
      isFinished: counts[idx] <= 0,
    })),
    hasDrawn: true,
    lastDrawPlayerIdx: ha.drawPlayerIdx,
    lastDrawFromIdx: ha.drawFromIdx,
    lastDrawCard: ha.drawnCard,
    lastDiscardedPairs: ha.discardedPairs,
    lastDiscardedCards: ha.discardedCards ?? [],
    cpuActions: [],
    ...(firstCpuAction && {
      currentTurn: firstCpuAction.drawPlayerIdx,
      gameEndFlag: false,
      message: '',
      nextDrawTargetIdx: firstCpuAction.drawFromIdx,
    }),
  };
}

interface PlayerAreaProps {
  player: OldMaidPlayerData;
  isTarget: boolean;
  isHumanTurn: boolean;
  gameEndFlag: boolean;
  loading: boolean;
  highlightedCardIdx: number;
  onDraw: (drawIdx: number) => void;
  onReorder?: (indices: number[]) => void;
}

function PlayerArea({
  player,
  isTarget,
  isHumanTurn,
  gameEndFlag,
  loading,
  highlightedCardIdx,
  onDraw,
  onReorder,
}: PlayerAreaProps) {
  const conditionalStyle: React.CSSProperties = player.isFinished
    ? { opacity: 0.5 }
    : isTarget && !gameEndFlag
      ? { border: '2px solid #f0ad4e', boxShadow: '0 0 12px #f0ad4e' }
      : {};

  const showSelectable = isHumanTurn && !loading && isTarget && !player.isFinished && !player.isHuman && !gameEndFlag;
  const showCount = Math.min(player.cardCount, 10);

  return (
    <div id={`player-area-${player.id}`} className={playerAreaClass} style={conditionalStyle}>
      <div className="text-white font-bold mb-1 text-[0.9em]">
        {playerName(player.id, player.isHuman)}
        {player.isFinished && <StatusBadge variant="success">上がり</StatusBadge>}
        {isTarget && !player.isHuman && !player.isFinished && !gameEndFlag && (
          <StatusBadge variant="warning">← 引く相手</StatusBadge>
        )}
      </div>
      {!player.isFinished && <div className="text-[#ccc] text-[0.8em] mb-1">{player.cardCount}枚</div>}
      {showSelectable && !player.isFinished && <div className="text-[#cfc] text-[0.75em] mb-1">引く</div>}
      <div className="flex flex-wrap gap-0.5 justify-center">
        {player.isFinished ? null : player.isHuman ? (
          player.cards?.map((card, i) => (
            <CardImage
              key={`${card.design}-${card.value}`}
              card={card}
              width={50}
              draggable={!gameEndFlag && !!onReorder}
              onDragStart={(e: React.DragEvent) => {
                e.dataTransfer.setData('oldmaidCardIndex', String(i));
              }}
              onDragOver={(e: React.DragEvent) => e.preventDefault()}
              onDrop={(e: React.DragEvent) => {
                e.preventDefault();
                const fromStr = e.dataTransfer.getData('oldmaidCardIndex');
                if (!fromStr || !onReorder || !player.cards) return;
                const from = Number(fromStr);
                if (from === i) return;
                const indices = player.cards.map((_, idx) => idx);
                indices.splice(from, 1);
                indices.splice(i, 0, from);
                onReorder(indices);
              }}
            />
          ))
        ) : showSelectable ? (
          <>
            {Array.from({ length: showCount }, (_, i) => {
              const isHighlighted = isTarget && !player.isHuman && i === highlightedCardIdx;
              const cardStyle: React.CSSProperties = {
                border: '2px solid transparent',
                borderRadius: 4,
                cursor: 'pointer',
                ...(isHighlighted ? { transform: 'translateY(-8px)', transition: 'transform 0.2s' } : {}),
              };
              return (
                <CardBack
                  // biome-ignore lint/suspicious/noArrayIndexKey: placeholder array with no card identity
                  key={i}
                  width={40}
                  style={cardStyle}
                  onClick={() => onDraw(i)}
                  ariaLabel={`カード ${i + 1} 枚目を引く`}
                />
              );
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

interface SetupScreenProps {
  mode: number;
  cpuPlacementStrategy: boolean;
  onModeChange: (m: number) => void;
  onStrategyChange: (v: boolean) => void;
  onStart: () => void;
  loading: boolean;
}

function SetupScreen({
  mode,
  cpuPlacementStrategy,
  onModeChange,
  onStrategyChange,
  onStart,
  loading,
}: SetupScreenProps) {
  return (
    <div className="flex-1 flex flex-col items-center justify-center bg-[#1a5c1a] p-6 gap-4" aria-busy={loading}>
      <div className="text-white text-2xl font-bold mb-2">Old Maid 設定</div>
      <div className="bg-black/40 rounded-xl p-4 w-full max-w-sm flex flex-col gap-3">
        <div className="text-white font-bold mb-1">モード選択</div>
        <label className="flex items-center gap-2 text-white cursor-pointer">
          <input
            type="radio"
            name="oldmaid-mode"
            value={OldMaidMode.Normal}
            checked={mode === OldMaidMode.Normal}
            onChange={() => onModeChange(OldMaidMode.Normal)}
          />
          ババ抜き（ジョーカーが奇数カード）
        </label>
        <label className="flex items-center gap-2 text-white cursor-pointer">
          <input
            type="radio"
            name="oldmaid-mode"
            value={OldMaidMode.JijiNuki}
            checked={mode === OldMaidMode.JijiNuki}
            onChange={() => onModeChange(OldMaidMode.JijiNuki)}
          />
          ジジ抜き（ランダム1枚除外）
        </label>
        <div className="border-t border-white/20 my-1" />
        <label className="flex items-center gap-2 text-white cursor-pointer">
          <input type="checkbox" checked={cpuPlacementStrategy} onChange={(e) => onStrategyChange(e.target.checked)} />
          CPU心理戦（奇数カードを端に配置）
        </label>
      </div>
      <button type="button" className={`${btnPrimary} min-w-[120px] mt-2`} disabled={loading} onClick={onStart}>
        ゲーム開始
      </button>
    </div>
  );
}

export function OldMaidPage() {
  const [displayState, setDisplayState] = useState<OldMaidResponse | null>(null);
  const [setupMode, setSetupMode] = useState<number>(OldMaidMode.Normal);
  const [setupStrategy, setSetupStrategy] = useState(false);
  const [gameSettings, setGameSettings] = useState<{ mode: number; cpuPlacementStrategy: boolean } | null>(null);
  const [shakeKey, setShakeKey] = useState(0);
  const [revealedCard, setRevealedCard] = useState<Card | null>(null);
  const revealTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Card reveal suspense: show card-back for 600ms, then flip to actual card
  useEffect(() => {
    const card = displayState?.lastDrawCard;
    if (!card) {
      setRevealedCard(null);
      return;
    }
    setRevealedCard(null);
    revealTimerRef.current = setTimeout(() => {
      setRevealedCard(card);
      revealTimerRef.current = null;
      if (card.design === 'JOKER') {
        setShakeKey((k) => k + 1);
      }
    }, 600);
    return () => {
      if (revealTimerRef.current !== null) {
        clearTimeout(revealTimerRef.current);
        revealTimerRef.current = null;
      }
    };
  }, [displayState?.lastDrawCard]);

  const onSuccess = useCallback(async (res: OldMaidResponse) => {
    const humanDrawState = buildHumanDrawState(res);
    if (humanDrawState) {
      setDisplayState(humanDrawState);
      await delay(REPLAY_DELAY_MS);
    }
    const replayStates = buildReplayStates(res);
    if (replayStates.length === 0) {
      setDisplayState(res);
      return;
    }
    for (const step of replayStates) {
      setDisplayState(step);
      await delay(REPLAY_DELAY_MS);
    }
    setDisplayState(res);
  }, []);

  const { loading, error, exec } = useGameApi(oldmaidApi.exec, { onSuccess });

  const handleStart = useCallback(() => {
    const settings = { mode: setupMode, cpuPlacementStrategy: setupStrategy };
    setGameSettings(settings);
    exec('reset', undefined, settings.mode, settings.cpuPlacementStrategy);
  }, [exec, setupMode, setupStrategy]);

  const handleReset = useCallback(() => {
    if (gameSettings) {
      exec('reset', undefined, gameSettings.mode, gameSettings.cpuPlacementStrategy);
    }
  }, [exec, gameSettings]);

  const handleReorder = useCallback(
    (indices: number[]) => {
      exec('reorder', undefined, undefined, undefined, indices);
    },
    [exec],
  );

  if (!gameSettings) {
    return (
      <SetupScreen
        mode={setupMode}
        cpuPlacementStrategy={setupStrategy}
        onModeChange={setSetupMode}
        onStrategyChange={setSetupStrategy}
        onStart={handleStart}
        loading={loading}
      />
    );
  }

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
    <div
      key={shakeKey}
      className={`flex-1 flex flex-col min-h-0 bg-[#1a5c1a]${shakeKey > 0 ? ' animate-shake' : ''}`}
      aria-busy={loading}
      aria-live="polite"
    >
      {loading && <span className="sr-only">処理中...</span>}
      {/* Scrollable: CPU rows + discard + status + logs + result */}
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {/* Mode badge */}
        {state.mode === OldMaidMode.JijiNuki && (
          <div className="text-center mb-1">
            <span className="inline-block rounded-md bg-red-600 px-2.5 py-0.5 text-sm font-bold text-white">
              ジジ抜き
            </span>
          </div>
        )}

        {/* CPU row */}
        <div className="flex gap-2 flex-wrap mb-2 justify-center">
          {cpuPlayers.map((player) => (
            <PlayerArea
              key={player.id}
              player={player}
              isTarget={state.nextDrawTargetIdx === player.id}
              isHumanTurn={isHumanTurn}
              gameEndFlag={state.gameEndFlag}
              loading={loading}
              highlightedCardIdx={state.nextDrawTargetIdx === player.id ? state.cpuHighlightedCardIdx : -1}
              onDraw={(drawIdx) => exec('draw', drawIdx)}
            />
          ))}
        </div>

        {/* Discarded Area */}
        <DiscardedArea cards={state.lastDiscardedCards} />

        {/* Card reveal area */}
        {state.lastDrawCard && !state.gameEndFlag && (
          <div className="flex justify-center my-2" data-testid="card-reveal-area">
            {revealedCard ? (
              <div className="animate-flipIn">
                <CardImage card={revealedCard} width={60} />
              </div>
            ) : (
              <CardBack width={60} />
            )}
          </div>
        )}

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

        {/* JijiNuki: show removed card at game end */}
        {state.gameEndFlag && state.removedCard && (
          <div className="text-center my-2 text-white text-[0.9em]">除外カード: {cardLabel(state.removedCard)}</div>
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
              loading={loading}
              highlightedCardIdx={-1}
              onDraw={(drawIdx) => exec('draw', drawIdx)}
              onReorder={handleReorder}
            />
          </div>
        )}

        <ErrorAlert message={error} />

        {/* Buttons */}
        <div className="text-center">
          <button
            type="button"
            className={`${btnSecondary} min-w-[80px]`}
            disabled={loading}
            onClick={() => setGameSettings(null)}
          >
            設定
          </button>
          <button type="button" className={`${btnPrimary} min-w-[80px]`} disabled={loading} onClick={handleReset}>
            リセット
          </button>
          <button
            type="button"
            className={`${btnWarning} min-w-[110px]`}
            disabled={loading || !isHumanTurn || state.gameEndFlag}
            onClick={() => exec('draw')}
          >
            ランダムに引く
          </button>
          <button
            type="button"
            className={`${btnSecondary} min-w-[110px]`}
            disabled={loading || state.gameEndFlag}
            onClick={() => exec('shuffle')}
          >
            シャッフル
          </button>
        </div>
      </div>
    </div>
  );
}
