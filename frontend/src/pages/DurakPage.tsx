import { useCallback, useMemo } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { CPU_DIFFICULTY_OPTIONS, useDurakGame } from '../hooks/useDurakGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnOutline, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { DurakResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { DURAK_HELP, parseDurakCommand } from '../utils/cli/commands/durakCommands';
import { formatDurakState } from '../utils/cli/formatters/durakFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Durak phase constants. */
const PHASE_ATTACK = 0;
const PHASE_DEFEND = 1;
const PHASE_BOUT_END = 2;
/** Total horizontal/vertical padding on each pair button (p-1 = 4px × 2 sides). */
const DK_PAIR_PADDING = 8;

const theme = gameTheme.durak;

/** Durak tutorial step definitions. */
const DURAK_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="dk-trump-info"]',
    messageKey: 'tutorial.trumpInfo',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dk-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dk-table-cards"]',
    messageKey: 'tutorial.tableCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dk-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Durak game page with attack/defense mechanics. */
export const DurakPage = withTutorial(DurakPageContent, 'durak', DURAK_TUTORIAL_STEPS);

/** Inner content of the Durak page, wrapped by TutorialProvider. */
function DurakPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('durak');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    durakConfig,
    selectedCardIdx,
    setSelectedCardIdx,
    selectedAttackIdx,
    setSelectedAttackIdx,
    handleConfigChange,
    handleConfigToggle,
    handleAttack,
    handleDefend,
    handlePass,
    handleTake,
    handleSort,
    hint: serverHint,
    hintError,
    handleHint,
  } = useDurakGame();

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('durak');
  const durakCliConfig: CliGameConfig<DurakResponse, Parameters<typeof gameExec>> = useMemo(
    () => ({
      gameName: 'durak',
      parseCommand: parseDurakCommand,
      formatResponse: formatDurakState,
      helpText: DURAK_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, durakCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('durak', state);

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, undefined, durakConfig);
  }, [gameExec, hideActionLog, durakConfig]);

  // Play a card-place sound when the human commits an attack or defense card, and
  // an error buzz for the disadvantageous "take" action (defender scoops the table).
  const handleAttackWithSound = useCallback(() => {
    handleAttack();
  }, [handleAttack]);
  const handleDefendWithSound = useCallback(() => {
    handleDefend();
  }, [handleDefend]);
  const handleTakeWithSound = useCallback(() => {
    playSound('errorBuzz');
    handleTake();
  }, [playSound, handleTake]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="durak"
        layout={{
          kind: 'trick-taking',
          titleBar: false,
          opponents: 3,
          opponentStyle: 'hand',
          opponentHandSize: 6,
          centerCard: true,
          footerHandSize: 6,
        }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const isAttacker = humanPlayer ? state.attackerIdx === humanPlayer.id : false;
  const isDefender = humanPlayer ? state.defenderIdx === humanPlayer.id : false;
  const isGameEnd = state.gameEndFlag;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : state.phase === PHASE_ATTACK
      ? t('phase.attack')
      : state.phase === PHASE_DEFEND
        ? t('phase.defend')
        : state.phase === PHASE_BOUT_END
          ? t('phase.boutEnd')
          : '';

  const isHumanTurn = !isGameEnd && humanPlayer !== undefined && state.currentTurn === humanPlayer.id;

  const showAttackBtn = !isGameEnd && isAttacker && state.phase === PHASE_ATTACK;
  const showDefendBtn = !isGameEnd && isDefender && state.phase === PHASE_DEFEND;
  const showPassBtn = !isGameEnd && isAttacker && state.phase === PHASE_ATTACK && state.tablePairs.length > 0;
  const showTakeBtn = !isGameEnd && isDefender && state.phase === PHASE_DEFEND;

  const suitSymbol = (suit: string) => {
    switch (suit) {
      case 'SPADE':
        return '\u2660';
      case 'HEART':
        return '\u2665';
      case 'DIAMOND':
        return '\u2666';
      case 'CLUB':
        return '\u2663';
      default:
        return suit;
    }
  };

  const humanWon = isGameEnd && humanPlayer !== undefined && state.loserIdx >= 0 && state.loserIdx !== humanPlayer.id;

  return (
    <GamePageShell
      title={tc('nav.durak')}
      gameThemeBg={theme.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/durak"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          {/* Settings panel */}
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: durakConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((opt) => ({
                      value: opt.value,
                      label: t(`settings.difficulty${opt.label.charAt(0).toUpperCase()}${opt.label.slice(1)}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'checkbox',
                    id: 'transferEnabled',
                    label: t('settings.transferEnabled'),
                    checked: durakConfig.transferEnabled,
                    onToggle: (checked) => handleConfigToggle('transferEnabled', checked),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          {/* Scrollable area */}
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Trump info */}
                <div className="bg-black/30 rounded-[10px] py-2.5 px-3.5 my-2" data-tutorial="dk-trump-info">
                  <div className="text-ds-text-primary font-bold mb-1">
                    {t('trump')}: {suitSymbol(state.trumpSuit)}
                    {state.trumpCard && (
                      <span className="ml-2">
                        <AnimatedCard card={state.trumpCard} width={cardWidth * 0.6} />
                      </span>
                    )}
                  </div>
                  <div className="text-game-text-muted text-sm">
                    {t('stock')}: {state.stockCount}
                  </div>
                </div>

                {/* Table pairs */}
                <div className="bg-black/30 rounded-[10px] py-2.5 px-3.5 my-2" data-tutorial="dk-table-cards">
                  <div className="text-ds-text-primary font-bold mb-1">{t('table')}</div>
                  {state.tablePairs.length === 0 ? (
                    <div className="text-game-text-muted text-sm">{t('tableEmpty')}</div>
                  ) : (
                    <div className="flex flex-wrap gap-3">
                      {state.tablePairs.map((pair, i) => {
                        const pairCardWidth = cardWidth * 0.8;
                        const offset = pairCardWidth * 0.3;
                        const undefended = pair.defense === null;
                        const shouldPulse = undefended && isDefender;
                        const ariaLabel = undefended
                          ? t('pair.undefendedAria', { card: cardAlt(pair.attack) })
                          : t('pair.defendedAria', {
                              attack: cardAlt(pair.attack),
                              defense: pair.defense ? cardAlt(pair.defense) : '',
                            });
                        return (
                          <button
                            type="button"
                            key={`pair-${pair.attack.design}-${pair.attack.value}-${i}`}
                            className={`relative p-1 rounded cursor-pointer ${
                              selectedAttackIdx === i ? 'ring-2 ring-ds-warning' : ''
                            }`}
                            style={{
                              width: pairCardWidth + offset + DK_PAIR_PADDING,
                              height: pairCardWidth * 1.4 + offset + DK_PAIR_PADDING,
                            }}
                            onClick={() => setSelectedAttackIdx(selectedAttackIdx === i ? null : i)}
                            data-testid={`dk-pair-${i}`}
                            aria-label={ariaLabel}
                          >
                            <div
                              className={`absolute top-0 left-0 rounded ${
                                shouldPulse ? 'motion-safe:animate-pulse ring-2 ring-ds-error' : ''
                              }`}
                              data-testid={`dk-attack-${i}`}
                              data-undefended={undefended ? 'true' : 'false'}
                            >
                              <AnimatedCard card={pair.attack} width={pairCardWidth} />
                            </div>
                            {pair.defense ? (
                              <div
                                className="absolute rounded shadow-md"
                                style={{ top: offset, left: offset }}
                                data-testid={`dk-defense-${i}`}
                              >
                                <AnimatedCard card={pair.defense} width={pairCardWidth} />
                              </div>
                            ) : null}
                          </button>
                        );
                      })}
                    </div>
                  )}
                </div>

                {/* CPU actions */}
                {state.cpuActions.length > 0 && (
                  <div
                    role="status"
                    aria-live="polite"
                    data-testid="durak-cpu-actions"
                    className="bg-black/40 rounded-lg text-game-text-muted py-2 px-3.5 my-2 whitespace-pre-line text-xs"
                  >
                    {[
                      tc('label.cpuActions'),
                      ...state.cpuActions.map(
                        (a) =>
                          `${playerName(state.players[a.playerIdx]?.id ?? a.playerIdx, false)}: ${t(`action.${['attack', 'defend', 'pass', 'take'][a.actionType] ?? 'pass'}`)}`,
                      ),
                    ].join('\n')}
                  </div>
                )}

                {/* Result message */}
                <GameMessageBox
                  message={state.message}
                  messageCode={state.messageCode}
                  messageParams={state.messageParams}
                />

                <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

                {/* Action log */}
                <ActionLogSection
                  isEndPhase={isGameEnd}
                  actionLog={actionLog}
                  showActionLog={showActionLog}
                  hideActionLog={hideActionLog}
                />
              </div>

              {/* Right: CPU info sidebar */}
              <div>
                <div className="flex gap-2 flex-wrap mb-3 lg:flex-col">
                  {cpuPlayers.map((player) => (
                    <div
                      key={player.id}
                      className={`p-2 rounded bg-black/30 ${
                        state.currentTurn === player.id ? 'ring-2 ring-ds-warning' : ''
                      }`}
                    >
                      <div className="text-ds-text-primary text-sm font-bold">
                        {playerName(player.id, false)}
                        {state.attackerIdx === player.id && (
                          <span className="text-ds-error text-xs ml-1">({t('attacker')})</span>
                        )}
                        {state.defenderIdx === player.id && (
                          <span className="text-ds-info text-xs ml-1">({t('defender')})</span>
                        )}
                      </div>
                      {/* 上がった相手は CUI では手札欄が「上がり」に差し替わるのに、
                          Web は枚数を出すだけで対局中の CPU と見分けが付かなかった
                          (#5524)。**枚数 0 では代用できない** -- 配り直しの直前など、
                          まだ抜けていないのに 0 枚の瞬間がある。 */}
                      {player.isFinished ? (
                        <div className="text-ds-success text-xs font-bold" data-testid={`cpu-finished-${player.id}`}>
                          {t('finished')}
                        </div>
                      ) : (
                        <div className="text-game-text-muted text-xs">{player.cardCount}</div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* Sticky footer: human player hand + action buttons */}
          <GameFooter className={`${theme.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <div className="mb-2" data-tutorial="dk-player-hand">
                <div className="text-ds-text-primary font-bold text-sm mb-1">
                  {humanPlayer.cardCount}
                  {isAttacker && <span className="text-ds-error text-xs ml-2">({t('attacker')})</span>}
                  {isDefender && <span className="text-ds-info text-xs ml-2">({t('defender')})</span>}
                  {isHumanTurn && <span className="text-ds-success text-xs ml-2">{t('selectCard')}</span>}
                </div>
                <div className="flex flex-wrap gap-1">
                  {humanPlayer.cards.map((card, i) => (
                    <button
                      type="button"
                      key={`${card.design}-${card.value}-${i}`}
                      className={`cursor-pointer transition-transform ${selectedCardIdx === i ? '-translate-y-2' : ''}`}
                      onClick={() => setSelectedCardIdx(selectedCardIdx === i ? null : i)}
                      aria-label={cardAlt(card)}
                      aria-pressed={selectedCardIdx === i}
                    >
                      <AnimatedCard card={card} width={cardWidth} isSelected={selectedCardIdx === i} />
                    </button>
                  ))}
                </div>
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {serverHint && (
              <p className="mt-2 text-sm text-ds-accent" data-testid="durak-server-hint">
                {serverHint.takeCards
                  ? t('hintTake')
                  : serverHint.cardIndex === undefined
                    ? t('hintPass')
                    : t('hintCard', { idx: serverHint.cardIndex })}{' '}
                ({t(`hintReason.${serverHint.reason}`)})
              </p>
            )}

            <ErrorAlert message={hintError} onRetry={undefined} />

            {/* Action buttons */}
            <div className="text-center" data-tutorial="dk-action-buttons">
              <GameResetButton
                isGameEnd={!!isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                className="min-w-[90px]"
              />
              {showAttackBtn && (
                <button
                  type="button"
                  className={`${btnDanger} min-w-[90px]`}
                  disabled={loading || selectedCardIdx === null}
                  onClick={handleAttackWithSound}
                >
                  {t('attackButton')}
                </button>
              )}
              {showDefendBtn && (
                <button
                  type="button"
                  className={`${btnSuccess} min-w-[90px]`}
                  disabled={loading || selectedCardIdx === null || selectedAttackIdx === null}
                  onClick={handleDefendWithSound}
                >
                  {t('defendButton')}
                </button>
              )}
              {showPassBtn && (
                <button
                  type="button"
                  className={`${btnSecondary} min-w-[90px]`}
                  disabled={loading}
                  onClick={handlePass}
                >
                  {t('passButton')}
                </button>
              )}
              {showTakeBtn && (
                <button
                  type="button"
                  className={`${btnPrimary} min-w-[90px]`}
                  disabled={loading}
                  onClick={handleTakeWithSound}
                >
                  {t('takeButton')}
                </button>
              )}
              {/* **他のトリック系はサーバー計算の理由付きヒントを持つのに、
                  Durak には無く、クライアント完結の簡易ヒューリスティックだけ
                  だった (#4740)。** */}
              {!isGameEnd && (
                <button
                  type="button"
                  className={`${btnSuccess} min-w-[90px]`}
                  disabled={loading}
                  onClick={handleHint}
                  data-testid="durak-hint-button"
                >
                  {tc('button.hint')}
                </button>
              )}

              {/* Sort buttons */}
              {!isGameEnd && (
                <>
                  <button
                    type="button"
                    className={`${btnOutline} min-w-[70px]`}
                    disabled={loading}
                    onClick={() => handleSort(0)}
                  >
                    {t('sort.suit')}
                  </button>
                  <button
                    type="button"
                    className={`${btnOutline} min-w-[70px]`}
                    disabled={loading}
                    onClick={() => handleSort(1)}
                  >
                    {t('sort.value')}
                  </button>
                </>
              )}
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
