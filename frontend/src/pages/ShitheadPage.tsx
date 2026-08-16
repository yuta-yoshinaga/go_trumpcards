import { useCallback, useMemo } from 'react';
import type { shitheadApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, useShitheadGame } from '../hooks/useShitheadGame';
import { badgeInfoColors, badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { focusRingCard } from '../styles/cardStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, ShitheadConfig, ShitheadResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { parseShitheadCommand, SHITHEAD_HELP } from '../utils/cli/commands/shitheadCommands';
import { formatShitheadState } from '../utils/cli/formatters/shitheadFormatter';
import { hintCliText, isHintCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';

/** Shithead tutorial step definitions. */
const SHITHEAD_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sh-discard"]',
    messageKey: 'tutorial.discardPile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sh-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sh-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Source identifier strings (sync with internal/domain/Shithead.go). */
const SOURCE_HAND = 'hand';
const SOURCE_FACE_UP = 'faceup';
const SOURCE_FACE_DOWN = 'facedown';

/** Renders the Shithead / Karma game page. */
export const ShitheadPage = withTutorial(ShitheadPageContent, 'shithead', SHITHEAD_TUTORIAL_STEPS);
/**
 * ルールを切り替えられるマジックカード。値は `ShitheadConfig` のキーで、
 * 文言は `settings.<key>` を引く。
 */
const MAGIC_TOGGLES = ['magicTwo', 'magicSeven', 'magicEight', 'magicTen', 'fourOfAKindBurn'] as const;

/** Inner content of the Shithead page. */
function ShitheadPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('shithead');
  const {
    state,
    loading,
    error,
    selectedCardIndices,
    toggleCard,
    handlePlay,
    handlePickup,
    handleReset,
    retry,
    dispatch,
    shitheadConfig,
    handleConfigChange,
    handleMagicToggle,
    handleResetWithConfig,
  } = useShitheadGame();
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('shithead', state);
  const cliMode = useCliMode('shithead');
  const cliConfig: CliGameConfig<ShitheadResponse, Parameters<typeof shitheadApi.exec>> = useMemo(
    () => ({
      gameName: 'shithead',
      parseCommand: parseShitheadCommand,
      formatResponse: formatShitheadState,
      helpText: SHITHEAD_HELP,
      localCommand: (input: string) => (isHintCommand(input) ? hintCliText(hint) : null),
    }),
    [hint],
  );
  const { handleCommand } = useCliGame(dispatch, cliConfig, state, {
    addInput: cliMode.addInput,
    addOutput: cliMode.addOutput,
    addError: cliMode.addError,
    clearLog: cliMode.clearLog,
  });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    handleReset();
  }, [hideActionLog, handleReset]);

  // Hook order must stay stable across renders — compute the lookup above the
  // skeleton guard. buildMagicLookup safely handles a null config.
  const magicLookup = useMemo(() => (state ? buildMagicLookup(state.config, t) : EMPTY_MAGIC_LOOKUP), [state, t]);

  if (!state) {
    return (
      <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.shithead.bg}`}>
        <div className="flex-1 flex items-center justify-center text-ds-text-primary">
          <p>{tc('skeleton.loading')}</p>
        </div>
      </div>
    );
  }

  const isGameEnd = state.gameEndFlag;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = state.players[state.currentTurn]?.isHuman === true && !isGameEnd;
  const phaseName = isGameEnd ? t('phase.gameEnd') : t('phase.play');

  return (
    <GamePageShell
      title={tc('nav.shithead')}
      gameThemeBg={gameTheme.shithead.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/shithead"
      gameEndFlag={isGameEnd}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliMode.cliEnabled} onToggle={cliMode.toggleCli} />}
    >
      {cliMode.cliEnabled ? (
        <CliTerminal logEntries={cliMode.logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="flex-1 overflow-y-auto px-4 py-2 space-y-4">
            {error && <ErrorAlert message={error} onRetry={retry} />}

            {/* CPU player rows */}
            <div className="bg-black/30 text-ds-text-primary p-3 rounded text-sm space-y-1">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="flex justify-between items-center">
                    <span>
                      CPU {p.id}
                      {p.isFinished && ` — ${t('labels.rank')} ${p.rank}`}
                      {p.rank === state.players.length && ` (${t('labels.shithead')})`}
                    </span>
                    <span>
                      {t('labels.hand')}:{p.handCount} / {t('labels.faceUp')}:{p.faceUpCards.length} /{' '}
                      {t('labels.faceDown')}:{p.faceDownCount}
                    </span>
                  </div>
                ))}
            </div>

            {isHumanTurn && (
              <div
                key={state.currentSource}
                role="status"
                data-testid="sh-source-banner"
                className={`mb-2 animate-pulse-once rounded-lg px-3 py-1.5 text-center font-medium text-sm ${badgeInfoColors}`}
              >
                {t(`sourceBanner.${state.currentSource}`)}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* Discard pile and stock */}
            <div data-tutorial="sh-discard" className="bg-black/30 text-ds-text-primary p-3 rounded">
              <div className="flex flex-wrap items-center gap-3 text-sm">
                <span className="inline-flex items-center gap-1">
                  {t('labels.discardPile')}:{' '}
                  {state.discardPile.length === 0 ? (
                    '—'
                  ) : (
                    <>
                      <AnimatedCard
                        card={state.discardPile[state.discardPile.length - 1]}
                        width={Math.round(cardWidth * 0.7)}
                      />
                      {(() => {
                        const topValue = state.discardPile[state.discardPile.length - 1]?.value ?? 0;
                        const badge = magicLookup.badgeFor(topValue);
                        const title = magicLookup.titleFor(topValue);
                        if (!badge) return null;
                        return (
                          <span
                            data-testid="sh-discard-magic-badge"
                            data-magic-rank={topValue}
                            title={title || undefined}
                            className={`inline-flex items-center rounded-full px-1.5 py-0 text-xs leading-tight motion-safe:animate-pulse ${badgeWarningColors}`}
                          >
                            {badge}
                          </span>
                        );
                      })()}
                    </>
                  )}
                </span>
                <span>
                  {t('labels.stock')}: {state.stockSize}
                </span>
                {state.sevenActive && <span className="text-ds-warning">{t('rules.magicSeven')}</span>}
                {state.skipNext && <span className="text-ds-warning">{t('rules.magicEight')}</span>}
              </div>
            </div>

            {/* Human player area */}
            {humanPlayer && (
              <div data-tutorial="sh-player-hand" className="bg-black/30 text-ds-text-primary p-3 rounded space-y-2">
                <div className="text-sm">
                  {t('labels.you')} —{' '}
                  {humanPlayer.isFinished
                    ? `${t('labels.rank')} ${humanPlayer.rank}`
                    : `${t('labels.currentSource')}: ${state.currentSource}`}
                </div>
                {humanPlayer.handCards.length > 0 && (
                  <CardRow
                    label={t('labels.hand')}
                    cards={humanPlayer.handCards}
                    selectable={isHumanTurn && state.currentSource === SOURCE_HAND}
                    selected={selectedCardIndices}
                    onToggle={toggleCard}
                    magic={magicLookup}
                    rowKey="hand"
                  />
                )}
                {humanPlayer.faceUpCards.length > 0 && (
                  <CardRow
                    label={t('labels.faceUp')}
                    cards={humanPlayer.faceUpCards}
                    selectable={isHumanTurn && state.currentSource === SOURCE_FACE_UP}
                    selected={selectedCardIndices}
                    onToggle={toggleCard}
                    magic={magicLookup}
                    rowKey="faceUp"
                  />
                )}
                {humanPlayer.faceDownCount > 0 && (
                  <FaceDownRow
                    label={t('labels.faceDown')}
                    count={humanPlayer.faceDownCount}
                    selectable={isHumanTurn && state.currentSource === SOURCE_FACE_DOWN}
                    selected={selectedCardIndices}
                    onToggle={toggleCard}
                  />
                )}
              </div>
            )}

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.shithead.footer} px-4 py-2.5`}>
            <div className="flex flex-wrap gap-2 items-center">
              <label className="flex items-center gap-1 text-ds-text-primary text-xs min-h-[44px]">
                <input
                  type="checkbox"
                  checked={hintEnabled}
                  onChange={(e) => setHintEnabled(e.target.checked)}
                  aria-label={tc('hint.toggle', { ns: 'tutorial' })}
                />
                {tc('hint.toggle', { ns: 'tutorial' })}
              </label>
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="sh-reset-button"
              />

              {isHumanTurn && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handlePlay}
                    disabled={loading || selectedCardIndices.length === 0}
                  >
                    {state.currentSource === SOURCE_FACE_DOWN ? t('actions.blindPlay') : t('actions.play')}
                  </button>
                  <button type="button" className={btnSecondary} onClick={handlePickup} disabled={loading}>
                    {t('actions.pickup')}
                  </button>
                </>
              )}
            </div>
          </GameFooter>
          {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'shitheadCpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: shitheadConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => {
                      handleConfigChange('cpuDifficulty', String(v));
                      handleResetWithConfig({ ...shitheadConfig, cpuDifficulty: Number(v) });
                    },
                  },
                  ...MAGIC_TOGGLES.map((key) => ({
                    type: 'checkbox' as const,
                    id: `shithead-${key}`,
                    label: t(`settings.${key}`),
                    checked: shitheadConfig[key],
                    onToggle: (checked: boolean) => {
                      handleMagicToggle(key, checked);
                      handleResetWithConfig({ ...shitheadConfig, [key]: checked });
                    },
                  })),
                ],
              },
            ]}
          />

          {/* **フックは設定機構を全部エクスポートしていたのに、ページが一切
              使っておらず常に既定値固定だった。**マジックカードの有無は
              ゲーム性そのものを変えるのに、Web からは触れなかった (#4747)。
              変更は即リセットが必要 (配りが変わる) ので handleResetWithConfig を通す。 */}
        </>
      )}
    </GamePageShell>
  );
}

interface CardRowProps {
  label: string;
  cards: Card[];
  selectable: boolean;
  selected: number[];
  onToggle: (index: number) => void;
  magic: MagicLookup;
  /** Stable prefix for React keys so hand and face-up rows don't share semantics. */
  rowKey: string;
}

/** Lookup helpers for Shithead magic-card affordances. */
interface MagicLookup {
  /** Returns the emoji badge for the rank, or empty string if the card is not magic. */
  badgeFor: (value: number) => string;
  /** Returns the localized tooltip describing the card's effect, or empty string if none. */
  titleFor: (value: number) => string;
}

/** Lookup that exposes no magic affordances — used while state is loading. */
const EMPTY_MAGIC_LOOKUP: MagicLookup = {
  badgeFor: () => '',
  titleFor: () => '',
};

/** Build a magic-card lookup from the active Shithead config. */
function buildMagicLookup(config: ShitheadConfig, t: (key: string) => string): MagicLookup {
  const map: Record<number, { enabled: boolean; emoji: string; key: string }> = {
    2: { enabled: config.magicTwo, emoji: '🔄', key: 'magicEffect.two' },
    7: { enabled: config.magicSeven, emoji: '⬇️', key: 'magicEffect.seven' },
    8: { enabled: config.magicEight, emoji: '⏭️', key: 'magicEffect.eight' },
    10: { enabled: config.magicTen, emoji: '🔥', key: 'magicEffect.ten' },
  };
  return {
    badgeFor: (v) => (map[v]?.enabled ? map[v].emoji : ''),
    titleFor: (v) => (map[v]?.enabled ? t(map[v].key) : ''),
  };
}

/** Row of selectable cards rendered as real card images. */
function CardRow({ label, cards, selectable, selected, onToggle, magic, rowKey }: CardRowProps) {
  const { cardWidth } = useCardDimensions();
  return (
    <div className="space-y-1">
      <div className="text-xs uppercase tracking-wide text-ds-text-muted">{label}</div>
      <div className="flex flex-wrap gap-2">
        {cards.map((c, i) => {
          const isSelected = selected.includes(i);
          const badge = magic.badgeFor(c.value);
          const title = magic.titleFor(c.value);
          return (
            <button
              key={`${rowKey}-${i}`}
              type="button"
              disabled={!selectable}
              onClick={() => onToggle(i)}
              title={title || undefined}
              aria-pressed={selectable ? isSelected : undefined}
              data-magic-rank={badge ? c.value : undefined}
              className={`relative rounded transition-transform ${focusRingCard} ${
                isSelected ? 'ring-2 ring-ds-warning -translate-y-1' : ''
              } ${selectable ? 'cursor-pointer' : 'cursor-default'}`}
            >
              <AnimatedCard card={c} width={cardWidth} />
              {badge && (
                <>
                  <span
                    aria-hidden="true"
                    className="absolute -top-1 -right-1 text-xs leading-none drop-shadow"
                    data-testid={`sh-magic-badge-${c.value}-${i}`}
                  >
                    {badge}
                  </span>
                  {title && <span className="sr-only">{title}</span>}
                </>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}

interface FaceDownRowProps {
  label: string;
  count: number;
  selectable: boolean;
  selected: number[];
  onToggle: (index: number) => void;
}

/** Row of face-down (hidden) cards rendered as card-back images, selectable for blind play. */
function FaceDownRow({ label, count, selectable, selected, onToggle }: FaceDownRowProps) {
  const { cardWidth } = useCardDimensions();
  return (
    <div className="space-y-1">
      <div className="text-xs uppercase tracking-wide text-ds-text-muted">{label}</div>
      <div className="flex flex-wrap gap-2">
        {Array.from({ length: count }).map((_, i) => {
          const isSelected = selected.includes(i);
          return (
            <button
              key={`fd-${i}`}
              type="button"
              disabled={!selectable}
              onClick={() => onToggle(i)}
              aria-pressed={selectable ? isSelected : undefined}
              data-testid={`sh-facedown-${i}`}
              className={`relative rounded transition-transform ${focusRingCard} ${
                isSelected ? 'ring-2 ring-ds-warning -translate-y-1' : ''
              } ${selectable ? 'cursor-pointer' : 'cursor-default'}`}
            >
              <AnimatedCardBack width={cardWidth} />
            </button>
          );
        })}
      </div>
    </div>
  );
}
