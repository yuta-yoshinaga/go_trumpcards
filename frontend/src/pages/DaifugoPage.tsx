import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { daifugoApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { CpuTurnArea } from '../components/CpuTurnArea';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { StatusBadge } from '../components/StatusBadge';
import { useCardSelection } from '../hooks/useCardSelection';
import { useGameApi } from '../hooks/useGameApi';
import { btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { playerAreaBase } from '../styles/gameStyles';
import type {
  DaifugoAction,
  DaifugoConfigInput,
  DaifugoExchangeAction,
  DaifugoPlayerData,
  DaifugoResponse,
} from '../types/card';
import { cardLabel } from '../utils/cardUtils';
import { findPlayerName, playerName } from '../utils/playerUtils';

const playerAreaClass = `${playerAreaBase} p-[10px] flex-[1_1_180px] min-w-[150px]`;

function CpuPlayerAreaWrapper({ player, isCurrentTurn }: { player: DaifugoPlayerData; isCurrentTurn: boolean }) {
  const { t } = useTranslation('daifugo');
  const finishedLabel = player.isFinished ? t('finishedWithRank', { rank: t(`rank.${player.rank}`) }) : undefined;
  return (
    <CpuTurnArea
      id={`player-area-${player.id}`}
      playerId={player.id}
      isHuman={player.isHuman}
      isCurrentTurn={isCurrentTurn}
      isFinished={player.isFinished}
      finishedLabel={finishedLabel}
      className={playerAreaClass}
    >
      {!player.isFinished && (
        <div className="text-[#ccc] text-[0.85em]">{t('cardCount', { count: player.cardCount })}</div>
      )}
    </CpuTurnArea>
  );
}

interface HumanPlayerAreaProps {
  player: DaifugoPlayerData;
  selectedIndices: number[];
  onToggle: (idx: number) => void;
  isCurrentTurn: boolean;
  onDragCard: (idx: number) => void;
}

function HumanPlayerArea({ player, selectedIndices, onToggle, isCurrentTurn, onDragCard }: HumanPlayerAreaProps) {
  const { t } = useTranslation('daifugo');
  const conditionalStyle: React.CSSProperties = player.isFinished
    ? { opacity: 0.5 }
    : isCurrentTurn
      ? { border: '2px solid #5cb85c', boxShadow: '0 0 12px #5cb85c' }
      : {};
  return (
    <div id={`player-area-${player.id}`} className={playerAreaClass} style={conditionalStyle}>
      <div className="text-white font-bold mb-1">
        {playerName(player.id, player.isHuman)}
        {player.isFinished && (
          <StatusBadge variant="success">{t('finishedWithRank', { rank: t(`rank.${player.rank}`) })}</StatusBadge>
        )}
      </div>
      {!player.isFinished && (
        <div style={{ color: '#ccc', fontSize: '0.85em', marginBottom: 4 }}>
          {t('cardCount', { count: player.cardCount })}
          {isCurrentTurn && <span style={{ marginLeft: 8, color: '#cfc' }}>{t('selectToPlay')}</span>}
        </div>
      )}
      <div className="flex flex-wrap gap-1">
        {player.cards?.map((card, i) => (
          <button
            key={`${card.design}-${card.value}`}
            type="button"
            aria-pressed={selectedIndices.includes(i)}
            disabled={!isCurrentTurn}
            draggable={isCurrentTurn}
            onClick={() => onToggle(i)}
            onDragStart={(e) => {
              e.dataTransfer.setData('cardIndex', String(i));
              onDragCard(i);
            }}
            style={{
              background: 'none',
              padding: 0,
              cursor: isCurrentTurn ? 'pointer' : 'default',
              borderRadius: 8,
              border: selectedIndices.includes(i) ? '3px solid #f0ad4e' : '3px solid transparent',
              boxSizing: 'border-box',
            }}
          >
            <CardImage card={card} width={52} />
          </button>
        ))}
      </div>
    </div>
  );
}

const badgeStyle: React.CSSProperties = {
  display: 'inline-block',
  borderRadius: 6,
  padding: '2px 10px',
  marginRight: 6,
  marginBottom: 4,
  fontSize: '0.8em',
  fontWeight: 'bold',
};

function RulesBadges({ state }: { state: DaifugoResponse }) {
  const { t } = useTranslation('daifugo');
  const badges: { label: string; bg: string; color: string }[] = [];
  if (state.revolutionActive) {
    badges.push({ label: t('badge.revolution'), bg: '#d9534f', color: '#fff' });
  }
  if (state.elevenBackActive) {
    badges.push({ label: t('badge.elevenBack'), bg: '#f0ad4e', color: '#222' });
  }
  if (state.suitLocked) {
    badges.push({ label: t('badge.suitLock', { suit: state.lockedSuit }), bg: '#5bc0de', color: '#222' });
  }
  if (state.tableIsSequence) {
    badges.push({ label: t('badge.sequence'), bg: '#9b59b6', color: '#fff' });
  }
  if (state.reverseDirection) {
    badges.push({ label: t('badge.nineReverse'), bg: '#e67e22', color: '#fff' });
  }
  if (state.numberLocked) {
    badges.push({ label: t('badge.numberLock'), bg: '#1abc9c', color: '#fff' });
  }
  if (badges.length === 0) return null;
  return (
    <div className="my-1 px-1">
      {badges.map((b) => (
        <span key={b.label} style={{ ...badgeStyle, background: b.bg, color: b.color }}>
          {b.label}
        </span>
      ))}
    </div>
  );
}

function ExchangeLog({
  players,
  actions,
}: {
  players: { id: number; isHuman: boolean }[];
  actions: DaifugoExchangeAction[];
}) {
  const { t } = useTranslation('daifugo');
  return (
    <div className="bg-black/40 rounded-lg text-[#ffd] py-2 px-3.5 my-2 whitespace-pre-line text-[0.85em]">
      {[
        t('exchange.title'),
        ...actions.map((a) => {
          const from = findPlayerName(players, a.fromPlayerIdx);
          const to = findPlayerName(players, a.toPlayerIdx);
          const cards = a.cards.map(cardLabel).join(', ');
          return t('exchange.entry', { from, to, cards });
        }),
      ].join('\n')}
    </div>
  );
}

interface SettingsPanelProps {
  config: DaifugoConfigInput;
  onChange: (key: keyof DaifugoConfigInput, value: boolean | number) => void;
}

function SettingsPanel({ config, onChange }: SettingsPanelProps) {
  const { t } = useTranslation('daifugo');
  const boolRules: { key: keyof DaifugoConfigInput; label: string }[] = [
    { key: 'eightCutEnabled', label: t('settings.eightCut') },
    { key: 'suitLockEnabled', label: t('settings.suitLock') },
    { key: 'elevenBackEnabled', label: t('settings.elevenBack') },
    { key: 'sequenceEnabled', label: t('settings.sequence') },
    { key: 'cardExchangeEnabled', label: t('settings.cardExchange') },
    { key: 'fiveSkipEnabled', label: t('settings.fiveSkip') },
    { key: 'sevenPassEnabled', label: t('settings.sevenPass') },
    { key: 'tenDiscardEnabled', label: t('settings.tenDiscard') },
    { key: 'spadeThreeEnabled', label: t('settings.spadeThree') },
    { key: 'capitalFallEnabled', label: t('settings.capitalFall') },
    { key: 'nineReverseEnabled', label: t('settings.nineReverse') },
    { key: 'coupDetatEnabled', label: t('settings.coupDetat') },
    { key: 'intenseLockEnabled', label: t('settings.intenseLock') },
    { key: 'sandstormEnabled', label: t('settings.sandstorm') },
    { key: 'emperorEnabled', label: t('settings.emperor') },
  ];
  return (
    <details className="mb-2">
      <summary className="cursor-pointer text-[#ccc] text-[0.85em] select-none">{t('settings.title')}</summary>
      <div className="bg-black/40 rounded-lg p-2 mt-1 text-[0.82em] text-white">
        <div className="mb-1">
          <label htmlFor="joker-count" className="mr-2">
            {t('settings.jokerCount')}
          </label>
          <select
            id="joker-count"
            value={config.jokerCount}
            onChange={(e) => onChange('jokerCount', Number(e.target.value))}
            className="bg-black/50 text-white rounded px-1"
          >
            <option value={0}>0</option>
            <option value={1}>1</option>
            <option value={2}>2</option>
          </select>
        </div>
        <div className="flex flex-wrap gap-x-4 gap-y-1">
          {boolRules.map(({ key, label }) => (
            <label key={key} className="flex items-center gap-1 cursor-pointer">
              <input
                type="checkbox"
                checked={config[key] as boolean}
                onChange={(e) => onChange(key, e.target.checked)}
              />
              {label}
            </label>
          ))}
        </div>
      </div>
    </details>
  );
}

const defaultConfigInput: DaifugoConfigInput = {
  jokerCount: 2,
  eightCutEnabled: true,
  suitLockEnabled: true,
  elevenBackEnabled: true,
  sequenceEnabled: true,
  cardExchangeEnabled: true,
  fiveSkipEnabled: false,
  sevenPassEnabled: false,
  tenDiscardEnabled: false,
  spadeThreeEnabled: false,
  capitalFallEnabled: false,
  nineReverseEnabled: false,
  coupDetatEnabled: false,
  intenseLockEnabled: false,
  sandstormEnabled: false,
  emperorEnabled: false,
};

export function DaifugoPage() {
  const { t } = useTranslation('daifugo');
  const { t: tc } = useTranslation('common');
  const {
    selected: selectedIndices,
    toggle: toggleCardSelection,
    clear: clearSelection,
    setSelected: setSelectedIndices,
  } = useCardSelection();
  const [configInput, setConfigInput] = useState<DaifugoConfigInput>(defaultConfigInput);

  const onSuccess = useCallback(() => {
    clearSelection();
  }, [clearSelection]);
  const { state, loading, error, exec } = useGameApi(daifugoApi.exec, { onSuccess });

  useEffect(() => {
    exec('reset');
  }, [exec]);

  if (!state) return null;

  const pendingAction = state.pendingAction ?? 'none';
  const isHumanTurn = !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const humanPlayer = state.players.find((p) => p.isHuman);

  const handleDragCard = (idx: number) => {
    // Ensure dragged card is in selection
    setSelectedIndices((prev) => (prev.includes(idx) ? prev : [...prev, idx]));
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    const draggedIdx = parseInt(e.dataTransfer.getData('cardIndex'), 10);
    if (Number.isNaN(draggedIdx)) {
      return;
    }
    const toPlay = selectedIndices.includes(draggedIdx) ? selectedIndices : [draggedIdx];
    exec(
      'play',
      [...toPlay].sort((a, b) => a - b),
    );
  };

  const handleConfigChange = (key: keyof DaifugoConfigInput, value: boolean | number) => {
    setConfigInput((prev) => ({ ...prev, [key]: value }));
  };

  // Pending action UI
  let playButtonLabel = t('playButton');
  let pendingBanner: string | null = null;
  if (pendingAction === 'sevenPass') {
    playButtonLabel = t('passButton');
    const targetName = findPlayerName(state.players, state.pendingActionTarget);
    pendingBanner = t('sevenPassBanner', { target: targetName });
  } else if (pendingAction === 'tenDiscard') {
    playButtonLabel = t('discardButton');
    pendingBanner = t('tenDiscardBanner');
  }

  const actionDescription = (players: { id: number; isHuman: boolean }[], action: DaifugoAction): string => {
    if (!action.playedCards || action.playedCards.length === 0) {
      return t('actionPassed', { name: findPlayerName(players, action.playerIdx) });
    }
    const cards = action.playedCards.map(cardLabel).join(', ');
    return t('actionPlayed', { name: findPlayerName(players, action.playerIdx), cards });
  };

  const sortModes = [
    { mode: 0, label: t('sort.strength') },
    { mode: 1, label: t('sort.suit') },
    { mode: 2, label: t('sort.number') },
  ] as const;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a5c1a]" aria-busy={loading} aria-live="polite">
      {loading && <span className="sr-only">{tc('status.loading')}</span>}
      {/* Scrollable: CPU rows + table cards + action logs + result */}
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {/* CPU row */}
        <div className="flex gap-2.5 flex-wrap mb-2.5">
          {cpuPlayers.map((player) => (
            <CpuPlayerAreaWrapper key={player.id} player={player} isCurrentTurn={state.currentTurn === player.id} />
          ))}
        </div>

        {/* Table cards (drop zone) */}
        {/* biome-ignore lint/a11y/noStaticElementInteractions: drag-and-drop target; keyboard play uses select+button */}
        <div
          className="bg-black/30 rounded-[10px] p-2.5 my-2"
          onDragOver={(e) => e.preventDefault()}
          onDrop={handleDrop}
        >
          <div className="text-white font-bold mb-1.5">{t('tableCards')}</div>
          <div className="flex flex-wrap gap-1">
            {!state.tableCards || state.tableCards.length === 0 ? (
              <span style={{ color: '#aaa' }}>{t('tableEmpty')}</span>
            ) : (
              state.tableCards.map((card) => <CardImage key={`${card.design}-${card.value}`} card={card} width={52} />)
            )}
          </div>
        </div>

        {/* Pending action banner */}
        {pendingBanner && (
          <div className="bg-yellow-700/80 rounded-[10px] text-white text-center py-2 px-4 text-[0.95em] font-bold my-2">
            {pendingBanner}
          </div>
        )}

        {/* Local rules status badges */}
        <RulesBadges state={state} />

        {/* Card exchange actions */}
        {state.exchangeActions && state.exchangeActions.length > 0 && (
          <ExchangeLog players={state.players} actions={state.exchangeActions} />
        )}

        {/* Human action log */}
        {state.humanAction && (
          <div className="bg-black/40 rounded-lg text-[#cfc] py-2 px-3.5 my-2 text-[0.85em]">
            {actionDescription(state.players, state.humanAction)}
          </div>
        )}

        {/* CPU action log */}
        {state.cpuActions && state.cpuActions.length > 0 && (
          <div className="bg-black/40 rounded-lg text-[#ccc] py-2 px-3.5 my-2 whitespace-pre-line text-[0.85em]">
            {[tc('label.cpuActions'), ...state.cpuActions.map((a) => actionDescription(state.players, a))].join('\n')}
          </div>
        )}

        {/* Result message */}
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
      </div>

      {/* Sticky footer: human player hand + buttons */}
      <GameFooter className="bg-[#163e16] border-white/20 px-4 py-2.5">
        {/* Settings panel */}
        <SettingsPanel config={configInput} onChange={handleConfigChange} />

        {/* Sort buttons */}
        <div className="text-center mb-1">
          {sortModes.map(({ mode, label }) => (
            <button
              key={mode}
              type="button"
              className={state.sortMode === mode ? `${btnPrimary} min-w-[70px]` : `${btnSecondary} min-w-[70px]`}
              disabled={loading}
              onClick={() => exec('sort', undefined, undefined, mode)}
            >
              {label}
            </button>
          ))}
        </div>

        {/* Human player */}
        {humanPlayer && (
          <div className="mb-2">
            <HumanPlayerArea
              player={humanPlayer}
              selectedIndices={selectedIndices}
              onToggle={toggleCardSelection}
              isCurrentTurn={isHumanTurn}
              onDragCard={handleDragCard}
            />
          </div>
        )}

        <ErrorAlert message={error} />

        {/* Buttons */}
        <div className="text-center">
          <button
            type="button"
            className={`${btnPrimary} min-w-[90px]`}
            disabled={loading}
            onClick={() => exec('reset', [], configInput)}
          >
            {tc('button.reset')}
          </button>
          <button
            type="button"
            className={`${btnWarning} min-w-[90px]`}
            disabled={loading || !isHumanTurn || state.gameEndFlag || pendingAction !== 'none'}
            onClick={() => exec('play', [])}
          >
            {tc('button.pass')}
          </button>
          <button
            type="button"
            className={`${btnSuccess} min-w-[120px]`}
            disabled={loading || !isHumanTurn || state.gameEndFlag || selectedIndices.length === 0}
            onClick={() =>
              exec(
                'play',
                [...selectedIndices].sort((a, b) => a - b),
              )
            }
          >
            {playButtonLabel}
          </button>
        </div>
      </GameFooter>
    </div>
  );
}
