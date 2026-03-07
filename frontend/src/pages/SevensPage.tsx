import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { sevensApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { CpuTurnArea } from '../components/CpuTurnArea';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { StatusBadge } from '../components/StatusBadge';
import { useGameApi } from '../hooks/useGameApi';
import { btnPrimary, btnWarning } from '../styles/buttonStyles';
import { playerAreaBase } from '../styles/gameStyles';
import type { Card, CardDesign, SevensAction, SevensPlayerData, SevensResponse } from '../types/card';
import { suitName, valueName } from '../utils/cardUtils';
import { findPlayerName, playerName } from '../utils/playerUtils';

// Design → suit index (matches Go backend: 1=SPADE, 2=CLOVER, 3=HEART, 4=DIAMOND)
const designToSuit: Record<CardDesign, number> = {
  SPADE: 1,
  CLOVER: 2,
  HEART: 3,
  DIAMOND: 4,
  JOKER: 0,
};

const SUITS = [
  { idx: 1, name: 'SPADE', label: '♠', color: '#e0e0e0' },
  { idx: 2, name: 'CLOVER', label: '♣', color: '#e0e0e0' },
  { idx: 3, name: 'HEART', label: '♥', color: '#f87171' },
  { idx: 4, name: 'DIAMOND', label: '♦', color: '#f87171' },
];

function isPositionPlaced(tablePlaced: number[], suit: number, value: number): boolean {
  return (tablePlaced[suit] & (1 << value)) !== 0;
}

function isEndStopped(tablePlaced: number[], suit: number, value: number, endStopEnabled: boolean): boolean {
  if (!endStopEnabled) return false;
  if (value === 7) return false;
  if (value > 7 && isPositionPlaced(tablePlaced, suit, 1)) return true;
  if (value < 7 && isPositionPlaced(tablePlaced, suit, 13)) return true;
  return false;
}

function isPositionPlayable(
  tablePlaced: number[],
  suit: number,
  value: number,
  tunnelEnabled: boolean,
  endStopEnabled: boolean,
): boolean {
  if (isPositionPlaced(tablePlaced, suit, value)) return false;
  if (isEndStopped(tablePlaced, suit, value, endStopEnabled)) return false;
  if (isPositionPlaced(tablePlaced, suit, value + 1)) return true;
  if (isPositionPlaced(tablePlaced, suit, value - 1)) return true;
  if (tunnelEnabled) {
    if (value === 1 && isPositionPlaced(tablePlaced, suit, 13)) return true;
    if (value === 13 && isPositionPlaced(tablePlaced, suit, 1)) return true;
  }
  return false;
}

function hasAnyPlayablePosition(tablePlaced: number[], tunnelEnabled: boolean, endStopEnabled: boolean): boolean {
  for (let suit = 1; suit <= 4; suit++) {
    for (let v = 1; v <= 13; v++) {
      if (isPositionPlayable(tablePlaced, suit, v, tunnelEnabled, endStopEnabled)) return true;
    }
  }
  return false;
}

function hasOnlyJokers(cards: Card[]): boolean {
  return cards.length > 0 && cards.every((c) => c.design === 'JOKER');
}

function isCardPlayable(
  card: Card,
  tablePlaced: number[],
  tunnelEnabled: boolean,
  noJokerFinish: boolean,
  allCards: Card[],
  endStopEnabled: boolean,
  jokerConsecutiveBanned: boolean,
  lastPlayedJoker: boolean,
): boolean {
  if (card.design === 'JOKER') {
    if (noJokerFinish && hasOnlyJokers(allCards)) return false;
    if (jokerConsecutiveBanned && lastPlayedJoker) return false;
    return hasAnyPlayablePosition(tablePlaced, tunnelEnabled, endStopEnabled);
  }
  const suit = designToSuit[card.design];
  return isPositionPlayable(tablePlaced, suit, card.value, tunnelEnabled, endStopEnabled);
}

function actionDesc(
  players: { id: number; isHuman: boolean }[],
  action: SevensAction,
  t: (key: string, opts?: Record<string, unknown>) => string,
): string {
  if (!action.playedCard) {
    const base = t('actionPassed', { name: findPlayerName(players, action.playerIdx) });
    return action.forcedPass ? t('actionForcedPass', { base }) : base;
  }
  const c = action.playedCard;
  if (c.design === 'JOKER' && action.targetSuit > 0) {
    return t('actionPlayedJoker', {
      name: findPlayerName(players, action.playerIdx),
      design: c.design,
      value: valueName(c.value),
      targetSuit: suitName(action.targetSuit),
      targetValue: valueName(action.targetValue),
    });
  }
  return t('actionPlayed', {
    name: findPlayerName(players, action.playerIdx),
    design: c.design,
    value: valueName(c.value),
  });
}

// ── styles ──────────────────────────────────────────────────────────────────

const playerAreaClass = `${playerAreaBase} p-[10px] flex-[1_1_180px] min-w-[150px]`;

// ── Board component ──────────────────────────────────────────────────────────

interface BoardProps {
  tablePlaced: number[];
  tunnelEnabled: boolean;
  endStopEnabled: boolean;
  jokerSelecting: boolean;
  onJokerPlace?: (suit: number, value: number) => void;
}

function Board({ tablePlaced, tunnelEnabled, endStopEnabled, jokerSelecting, onJokerPlace }: BoardProps) {
  const { t } = useTranslation('sevens');
  return (
    <div className="bg-black/30 rounded-[10px] py-2.5 px-3.5 my-2">
      <div className="text-white font-bold mb-2">
        {t('board')}
        {tunnelEnabled && <span className="text-yellow-400 text-xs ml-2">{t('tunnelTag')}</span>}
        {jokerSelecting && <span className="text-green-400 text-xs ml-2">{t('jokerSelectHint')}</span>}
      </div>
      <div className="grid grid-cols-2 gap-2">
        {SUITS.map(({ idx, name, label, color }) => (
          <div key={name} className="bg-white/[0.08] rounded-lg py-1.5 px-2.5 flex items-center gap-2">
            <span style={{ color, fontWeight: 'bold', fontSize: '1.1em', minWidth: 18 }}>{label}</span>
            <div className="flex flex-wrap gap-[3px] items-center">
              {Array.from({ length: 13 }, (_, i) => i + 1).map((v) => {
                const placed = isPositionPlaced(tablePlaced, idx, v);
                const isCenter = v === 7;
                const canPlace =
                  jokerSelecting && isPositionPlayable(tablePlaced, idx, v, tunnelEnabled, endStopEnabled);
                const tunnelHighlight =
                  tunnelEnabled &&
                  !placed &&
                  ((v === 1 && isPositionPlaced(tablePlaced, idx, 13)) ||
                    (v === 13 && isPositionPlaced(tablePlaced, idx, 1)));
                const cellStyle: React.CSSProperties = {
                  display: 'inline-block',
                  width: 22,
                  height: 22,
                  lineHeight: '22px',
                  textAlign: 'center',
                  borderRadius: 4,
                  fontSize: '0.7em',
                  fontWeight: isCenter ? 'bold' : 'normal',
                  background: canPlace
                    ? '#3b82f6'
                    : placed
                      ? isCenter
                        ? '#f0ad4e'
                        : '#5cb85c'
                      : 'rgba(255,255,255,0.1)',
                  color: canPlace ? '#fff' : placed ? '#000' : '#555',
                  border: tunnelHighlight ? '1px solid #f59e0b' : undefined,
                  boxSizing: 'border-box',
                };
                if (canPlace) {
                  return (
                    <button
                      key={v}
                      type="button"
                      onClick={() => onJokerPlace?.(idx, v)}
                      aria-label={t('placeAriaLabel', { suit: suitName(idx), value: valueName(v) })}
                      style={{ ...cellStyle, border: '1px solid #60a5fa', cursor: 'pointer', padding: 0 }}
                    >
                      {valueName(v)}
                    </button>
                  );
                }
                return (
                  <span key={v} style={cellStyle}>
                    {valueName(v)}
                  </span>
                );
              })}
              {tunnelEnabled && (
                <span role="img" className="text-yellow-400 text-[0.65em] ml-0.5" aria-label={t('tunnelConnection')}>
                  ↔
                </span>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ── CPU player area ──────────────────────────────────────────────────────────

function SevCpuArea({ player, isCurrentTurn }: { player: SevensPlayerData; isCurrentTurn: boolean }) {
  const { t } = useTranslation('sevens');
  return (
    <CpuTurnArea
      playerId={player.id}
      isHuman={player.isHuman}
      isCurrentTurn={isCurrentTurn}
      isFinished={player.isFinished}
      finishedLabel={player.isFinished ? t('rankLabel', { rank: player.rank }) : undefined}
      className={playerAreaClass}
    >
      {!player.isFinished && (
        <div className="text-[#ccc] text-[0.85em]">
          {t('cardCount', { count: player.cardCount })}
          {'　'}
          {t('passCount', {
            used: player.passesUsed,
            max: player.maxPasses === 0 ? t('passUnlimited') : player.maxPasses,
          })}
        </div>
      )}
    </CpuTurnArea>
  );
}

// ── Human player area ────────────────────────────────────────────────────────

interface HumanAreaProps {
  player: SevensPlayerData;
  isCurrentTurn: boolean;
  tablePlaced: number[];
  tunnelEnabled: boolean;
  noJokerFinish: boolean;
  endStopEnabled: boolean;
  jokerConsecutiveBanned: boolean;
  loading: boolean;
  onPlay: (idx: number) => void;
}

function HumanArea({
  player,
  isCurrentTurn,
  tablePlaced,
  tunnelEnabled,
  noJokerFinish,
  endStopEnabled,
  jokerConsecutiveBanned,
  loading,
  onPlay,
}: HumanAreaProps) {
  const { t } = useTranslation('sevens');
  const conditionalStyle: React.CSSProperties = player.isFinished
    ? { opacity: 0.5 }
    : isCurrentTurn
      ? { border: '2px solid #5cb85c', boxShadow: '0 0 12px #5cb85c' }
      : {};
  return (
    <div className={playerAreaClass} style={conditionalStyle}>
      <div className="text-white font-bold mb-1">
        {playerName(player.id, player.isHuman)}
        {player.isFinished && <StatusBadge variant="success">{t('rankLabel', { rank: player.rank })}</StatusBadge>}
      </div>
      {!player.isFinished && (
        <div className="text-[#ccc] text-[0.85em] mb-1">
          {t('cardCount', { count: player.cardCount })}
          {'　'}
          {t('passCount', {
            used: player.passesUsed,
            max: player.maxPasses === 0 ? t('passUnlimited') : player.maxPasses,
          })}
          {isCurrentTurn && <span style={{ marginLeft: 8, color: '#cfc' }}>{t('clickPlayable')}</span>}
        </div>
      )}
      <div className="flex flex-wrap gap-1">
        {player.cards?.map((card, i) => {
          const playable =
            isCurrentTurn &&
            !loading &&
            isCardPlayable(
              card,
              tablePlaced,
              tunnelEnabled,
              noJokerFinish,
              player.cards,
              endStopEnabled,
              jokerConsecutiveBanned,
              player.lastPlayedJoker,
            );
          return (
            <button
              key={`${card.design}-${card.value}`}
              type="button"
              disabled={!playable}
              onClick={() => onPlay(i)}
              title={playable ? `出す: ${card.design} ${valueName(card.value)}` : undefined}
              data-testid={playable ? 'playable-card' : undefined}
              style={{
                background: 'none',
                padding: 0,
                cursor: playable ? 'pointer' : 'default',
                borderRadius: 8,
                border: playable ? '3px solid #5cb85c' : '3px solid transparent',
                opacity: isCurrentTurn && !playable ? 0.5 : 1,
                boxSizing: 'border-box',
              }}
            >
              <CardImage card={card} width={52} />
            </button>
          );
        })}
      </div>
    </div>
  );
}

// ── Main page ────────────────────────────────────────────────────────────────

export function SevensPage() {
  const { t } = useTranslation('sevens');
  const { t: tc } = useTranslation('common');
  const [jokerCardIdx, setJokerCardIdx] = useState<number | null>(null);
  const [cfgTunnel, setCfgTunnel] = useState(false);
  const [cfgJokerCount, setCfgJokerCount] = useState(0);
  const [cfgCpuStrategy, setCfgCpuStrategy] = useState(false);
  const [cfgMaxPasses, setCfgMaxPasses] = useState(5);
  const [cfgNoJokerFinish, setCfgNoJokerFinish] = useState(false);
  const [cfgJokerReclaim, setCfgJokerReclaim] = useState(false);
  const [cfgEndStop, setCfgEndStop] = useState(false);
  const [cfgJokerConsBan, setCfgJokerConsBan] = useState(false);

  const onSuccess = useCallback((res: SevensResponse) => {
    setJokerCardIdx(null);
    setCfgTunnel(res.config.tunnelEnabled);
    setCfgJokerCount(res.config.jokerCount);
    setCfgCpuStrategy(res.config.cpuStrategy);
    setCfgMaxPasses(res.config.maxPasses);
    setCfgNoJokerFinish(res.config.noJokerFinish);
    setCfgJokerReclaim(res.config.jokerReclaimEnabled);
    setCfgEndStop(res.config.endStopEnabled);
    setCfgJokerConsBan(res.config.jokerConsecutiveBanned);
  }, []);
  const { state, loading, error, exec } = useGameApi(sevensApi.exec, { onSuccess });

  useEffect(() => {
    exec('reset');
  }, [exec]);

  if (!state) return null;

  const tablePlaced = state.tablePlaced;
  const tunnelEnabled = state.config.tunnelEnabled;
  const isHumanTurn = !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const canPass =
    isHumanTurn &&
    humanPlayer != null &&
    (humanPlayer.maxPasses === 0 || humanPlayer.passesUsed < humanPlayer.maxPasses);

  const handleCardPlay = (idx: number) => {
    const card = humanPlayer?.cards?.[idx];
    if (card?.design === 'JOKER') {
      setJokerCardIdx(idx);
    } else {
      exec('play', idx);
    }
  };

  const handleJokerPlace = (suit: number, value: number) => {
    exec('joker', jokerCardIdx as number, suit, value);
  };

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a5c1a]" aria-busy={loading} aria-live="polite">
      {loading && <span className="sr-only">{tc('status.loading')}</span>}
      {/* Scrollable: CPU rows + board + action logs + result */}
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {/* Config rules */}
        {state.config &&
          (state.config.tunnelEnabled ||
            state.config.jokerCount > 0 ||
            state.config.cpuStrategy ||
            state.config.maxPasses !== 5 ||
            state.config.noJokerFinish ||
            state.config.jokerReclaimEnabled ||
            state.config.endStopEnabled ||
            state.config.jokerConsecutiveBanned) && (
            <div className="bg-black/30 rounded-lg text-yellow-300 py-1.5 px-3 mb-2 text-[0.85em]">
              {t('rules.title')}
              {state.config.tunnelEnabled && ` ${t('rules.tunnelTag')}`}
              {state.config.jokerCount > 0 && ` ${t('rules.jokerTag', { count: state.config.jokerCount })}`}
              {state.config.cpuStrategy && ` ${t('rules.cpuStrategy')}`}
              {state.config.maxPasses === 0 && ` ${t('rules.passUnlimited')}`}
              {state.config.maxPasses !== 5 &&
                state.config.maxPasses !== 0 &&
                ` ${t('rules.passCount', { count: state.config.maxPasses })}`}
              {state.config.noJokerFinish && ` ${t('rules.noJokerFinish')}`}
              {state.config.jokerReclaimEnabled && ` ${t('rules.jokerReclaim')}`}
              {state.config.endStopEnabled && ` ${t('rules.endStop')}`}
              {state.config.jokerConsecutiveBanned && ` ${t('rules.jokerConsecutiveBanned')}`}
            </div>
          )}

        {/* CPU row */}
        <div className="flex gap-2.5 flex-wrap mb-2.5">
          {cpuPlayers.map((player) => (
            <SevCpuArea key={player.id} player={player} isCurrentTurn={state.currentTurn === player.id} />
          ))}
        </div>

        {/* Board */}
        <Board
          tablePlaced={tablePlaced}
          tunnelEnabled={tunnelEnabled}
          endStopEnabled={state.config.endStopEnabled}
          jokerSelecting={jokerCardIdx !== null}
          onJokerPlace={handleJokerPlace}
        />

        {/* Human action log */}
        {state.humanAction && (
          <div
            data-testid={state.humanAction.forcedPass ? 'human-action-forced-pass' : 'human-action'}
            className={`rounded-lg py-2 px-3.5 my-2 text-[0.85em] ${state.humanAction.forcedPass ? 'bg-red-900/50 text-[#fca] border border-red-500/50' : 'bg-black/40 text-[#cfc]'}`}
          >
            {actionDesc(state.players, state.humanAction, t)}
          </div>
        )}

        {/* CPU action log */}
        {state.cpuActions && state.cpuActions.length > 0 && (
          <div className="bg-black/40 rounded-lg py-2 px-3.5 my-2 text-[0.85em]">
            <span className="text-[#ccc]">{tc('label.cpuActions')}</span>
            {state.cpuActions.map((a, i) => (
              <div
                key={`cpu-action-${a.playerIdx}-${i}`}
                data-testid={a.forcedPass ? `cpu-action-forced-pass-${i}` : `cpu-action-${i}`}
                className={a.forcedPass ? 'text-[#fca]' : 'text-[#ccc]'}
              >
                {actionDesc(state.players, a, t)}
              </div>
            ))}
          </div>
        )}

        {/* Result message */}
        <GameMessageBox
          message={
            state.gameEndFlag
              ? `${t('resultPrefix')} ${state.players
                  .filter((p) => p.rank > 0)
                  .sort((a, b) => a.rank - b.rank)
                  .map((p) =>
                    t('resultEntry', { name: playerName(p.id, p.isHuman), rank: t('rankLabel', { rank: p.rank }) }),
                  )
                  .join(' ')}`
              : state.message
          }
          messageCode={state.gameEndFlag ? undefined : state.messageCode}
          messageParams={state.gameEndFlag ? undefined : state.messageParams}
        />
      </div>

      {/* Sticky footer: human player hand + buttons */}
      <GameFooter className="bg-[#163e16] border-white/20 px-4 py-2.5">
        {/* Human player */}
        {humanPlayer && (
          <div className="mb-2">
            <HumanArea
              player={humanPlayer}
              isCurrentTurn={isHumanTurn}
              tablePlaced={tablePlaced}
              tunnelEnabled={tunnelEnabled}
              noJokerFinish={state.config.noJokerFinish}
              endStopEnabled={state.config.endStopEnabled}
              jokerConsecutiveBanned={state.config.jokerConsecutiveBanned}
              loading={loading}
              onPlay={handleCardPlay}
            />
          </div>
        )}

        {/* Config panel */}
        <div className="bg-black/30 rounded-lg py-1.5 px-3 mb-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-[0.85em] text-white/80">
          <span className="text-yellow-300 font-bold">{t('config.title')}</span>
          <label className="flex items-center gap-1 cursor-pointer">
            <input type="checkbox" checked={cfgTunnel} onChange={(e) => setCfgTunnel(e.target.checked)} />
            {t('config.tunnel')}
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            {t('config.joker')}
            <select
              value={cfgJokerCount}
              onChange={(e) => setCfgJokerCount(Number(e.target.value))}
              className="bg-black/50 text-white rounded px-1 py-0.5"
            >
              <option value={0}>0</option>
              <option value={1}>1</option>
              <option value={2}>2</option>
            </select>
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            <input type="checkbox" checked={cfgCpuStrategy} onChange={(e) => setCfgCpuStrategy(e.target.checked)} />
            {t('config.cpuStrategy')}
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            {t('config.passCount')}
            <select
              value={cfgMaxPasses}
              onChange={(e) => setCfgMaxPasses(Number(e.target.value))}
              className="bg-black/50 text-white rounded px-1 py-0.5"
            >
              <option value={3}>3</option>
              <option value={5}>5</option>
              <option value={10}>10</option>
              <option value={0}>{t('config.passUnlimited')}</option>
            </select>
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            <input type="checkbox" checked={cfgNoJokerFinish} onChange={(e) => setCfgNoJokerFinish(e.target.checked)} />
            {t('config.noJokerFinish')}
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            <input type="checkbox" checked={cfgJokerReclaim} onChange={(e) => setCfgJokerReclaim(e.target.checked)} />
            {t('config.jokerReclaim')}
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            <input type="checkbox" checked={cfgEndStop} onChange={(e) => setCfgEndStop(e.target.checked)} />
            {t('config.endStop')}
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            <input type="checkbox" checked={cfgJokerConsBan} onChange={(e) => setCfgJokerConsBan(e.target.checked)} />
            {t('config.jokerConsecutiveBanned')}
          </label>
        </div>

        <ErrorAlert message={error} />

        {/* Buttons */}
        <div className="text-center">
          <button
            type="button"
            className={`${btnPrimary} min-w-[90px]`}
            disabled={loading}
            onClick={() =>
              exec('reset', -1, 0, 0, {
                tunnelEnabled: cfgTunnel,
                jokerCount: cfgJokerCount,
                cpuStrategy: cfgCpuStrategy,
                maxPasses: cfgMaxPasses,
                noJokerFinish: cfgNoJokerFinish,
                jokerReclaim: cfgJokerReclaim,
                endStop: cfgEndStop,
                jokerConsecutiveBanned: cfgJokerConsBan,
              })
            }
          >
            {tc('button.reset')}
          </button>
          <button
            type="button"
            className={`${btnWarning} min-w-[90px]`}
            disabled={loading || !canPass}
            onClick={() => exec('play', -1)}
          >
            {tc('button.pass')}
          </button>
          {jokerCardIdx !== null && (
            <button type="button" className={`${btnWarning} min-w-[90px]`} onClick={() => setJokerCardIdx(null)}>
              {tc('button.cancel')}
            </button>
          )}
        </div>
      </GameFooter>
    </div>
  );
}
