import { useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { allfoursApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { AllFoursResponse } from '../types/card';
import { AllFoursPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { ALLFOURS_HELP, parseAllFoursCommand } from '../utils/cli/commands/allfoursCommands';
import { formatAllFoursState } from '../utils/cli/formatters/allfoursFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { findPlayerName, playerName } from '../utils/playerUtils';

const AF_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="af-beg-controls"]',
    messageKey: 'tutorial.begControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="af-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="af-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="af-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="af-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="af-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const PHASE_KEYS: Readonly<Record<number, string>> = {
  [AllFoursPhase.BEG]: 'beg',
  [AllFoursPhase.GIFT]: 'gift',
  [AllFoursPhase.PLAY]: 'play',
  [AllFoursPhase.TRICK_END]: 'trickEnd',
  [AllFoursPhase.ROUND_END]: 'roundEnd',
  [AllFoursPhase.GAME_END]: 'gameEnd',
};

const SUIT_LABELS: Readonly<Record<number, string>> = {
  0: '—',
  1: '♠',
  2: '♣',
  3: '♥',
  4: '♦',
};

/** Suit-name i18n key suffixes indexed by trump-suit number (1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_NAME_KEYS: Readonly<Record<number, string>> = {
  1: 'spade',
  2: 'club',
  3: 'heart',
  4: 'diamond',
};

/** Suit-name i18n key suffixes indexed by card design string. */
const SUIT_NAME_BY_DESIGN: Readonly<Record<string, string>> = {
  SPADE: 'spade',
  CLOVER: 'club',
  HEART: 'heart',
  DIAMOND: 'diamond',
};

const SUIT_DESIGNS: Readonly<Record<string, string>> = {
  SPADE: '♠',
  CLOVER: '♣',
  HEART: '♥',
  DIAMOND: '♦',
};

const VALUE_LABELS: Readonly<Record<number, string>> = {
  1: 'A',
  11: 'J',
  12: 'Q',
  13: 'K',
};

function cardLabel(c: { design: string; value: number }): string {
  const v = VALUE_LABELS[c.value] ?? String(c.value);
  return `${SUIT_DESIGNS[c.design] ?? c.design}${v}`;
}

/** All Fours (Seven Up) game page. */
export const AllFoursPage = withTutorial(AllFoursPageContent, 'allfours', AF_TUTORIAL_STEPS);

function AllFoursPageContent() {
  const { t } = useTranslation('allfours');
  const { tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('allfours');
  const { playSound } = useSound();
  const { state, loading, error, exec: execApi, retry } = useGameApi(allfoursApi.exec);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('allfours', state);
  const [selectedCardIdx, setSelectedCardIdx] = useState<number | null>(null);
  const { cardWidth } = useCardDimensions();

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('allfours');
  const cliConfig: CliGameConfig<AllFoursResponse, Parameters<typeof allfoursApi.exec>> = useMemo(
    () => ({
      gameName: 'allfours',
      parseCommand: parseAllFoursCommand,
      formatResponse: formatAllFoursState,
      helpText: ALLFOURS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  // biome-ignore lint/correctness/useExhaustiveDependencies: phase/trick deps drive re-evaluation
  useEffect(() => {
    setSelectedCardIdx(null);
  }, [state?.trickNumber, state?.phase]);

  const isBegPhase = state?.phase === AllFoursPhase.BEG;
  const isGiftPhase = state?.phase === AllFoursPhase.GIFT;
  const isPlayPhase = state?.phase === AllFoursPhase.PLAY;
  const isTrickEnd = state?.phase === AllFoursPhase.TRICK_END;
  const isRoundEnd = state?.phase === AllFoursPhase.ROUND_END;
  const isGameEnd = state?.phase === AllFoursPhase.GAME_END || state?.gameEndFlag === true;
  const humanIdx = state?.players.findIndex((p) => p.isHuman) ?? -1;
  const isHumanBegTurn = isBegPhase && state?.nonDealerIdx === humanIdx;
  const isHumanGiftTurn = isGiftPhase && state?.dealerIdx === humanIdx;
  const isHumanPlayTurn = isPlayPhase && state?.currentPlayerIdx === humanIdx;
  const human = state?.players.find((p) => p.isHuman);

  const cpuDifficulty = state?.config.cpuDifficulty ?? 1;
  const pointLimit = state?.config.pointLimit ?? 7;

  const handleBeg = useCallback((beg: boolean) => execApi('beg', beg), [execApi]);
  const handleRespond = useCallback((run: boolean) => execApi('respond', undefined, run), [execApi]);
  const handlePlay = useCallback(() => {
    if (selectedCardIdx === null) return;
    void execApi('play', undefined, undefined, selectedCardIdx);
    setSelectedCardIdx(null);
  }, [execApi, selectedCardIdx]);
  const handleNextTrick = useCallback(() => execApi('next'), [execApi]);
  const handleNextRound = useCallback(() => execApi('nextround'), [execApi]);
  const handleManualReset = useCallback(() => {
    hideActionLog();
    void execApi('reset', undefined, undefined, undefined, { cpuDifficulty, pointLimit });
  }, [execApi, hideActionLog, cpuDifficulty, pointLimit]);
  const handleConfigChange = useCallback(
    (key: 'cpuDifficulty' | 'pointLimit', value: number) => {
      void execApi('reset', undefined, undefined, undefined, {
        cpuDifficulty: key === 'cpuDifficulty' ? value : cpuDifficulty,
        pointLimit: key === 'pointLimit' ? value : pointLimit,
      });
    },
    [execApi, cpuDifficulty, pointLimit],
  );

  if (!state) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.allfours.bg}`}>
        <div className="text-ds-text-primary">Loading...</div>
      </div>
    );
  }

  const phaseName = t(`phase.${PHASE_KEYS[state.phase] ?? 'beg'}`);

  return (
    <GamePageShell
      title={tc('nav.allfours')}
      gameThemeBg={gameTheme.allfours.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanBegTurn || isHumanGiftTurn || isHumanPlayTurn}
      gamePath="/allfours"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerIdx === humanIdx}
      onCelebrate={() => playSound('winFanfare')}
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
          <div className="overflow-y-auto pt-3 px-4 lg:px-8 flex-1">
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <div className="flex flex-wrap items-center gap-3 text-sm text-ds-text-primary mb-3">
              <span>{t('round', { n: state.roundNumber })}</span>
              <span>{t('trick', { n: state.trickNumber })}</span>
              <span>{t('dealer', { name: findPlayerName(state.players, state.dealerIdx) })}</span>
              {/* Trump: symbol shown, suit read out by name to avoid "black spade symbol". */}
              <span
                role="img"
                aria-label={t('trumpSuit', {
                  suit: state.trumpSuit === 0 ? t('trumpUnset') : t(`suitName.${SUIT_NAME_KEYS[state.trumpSuit]}`, '?'),
                })}
              >
                <span aria-hidden="true">
                  {t('trumpSuit', {
                    suit: state.trumpSuit === 0 ? t('trumpUnset') : (SUIT_LABELS[state.trumpSuit] ?? '?'),
                  })}
                </span>
              </span>
              {state.turnUp && (
                <span
                  role="img"
                  aria-label={t('turnUp', {
                    card: `${t(`suitName.${SUIT_NAME_BY_DESIGN[state.turnUp.design] ?? ''}`, state.turnUp.design)}${VALUE_LABELS[state.turnUp.value] ?? state.turnUp.value}`,
                  })}
                >
                  <span aria-hidden="true">{t('turnUp', { card: cardLabel(state.turnUp) })}</span>
                </span>
              )}
            </div>

            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

            <div data-tutorial="af-score-table" className="overflow-x-auto mb-3">
              <table className="text-sm w-full border-collapse text-ds-text-primary">
                <thead>
                  <tr className="border-b border-white/20">
                    <th className="text-left p-1">{t('scoresPlayer')}</th>
                    <th className="text-right p-1">{t('scoresTricks')}</th>
                    <th className="text-right p-1">{t('scoresRound')}</th>
                    <th className="text-right p-1">{t('scoresTotal')}</th>
                  </tr>
                </thead>
                <tbody>
                  {state.players.map((p) => (
                    <tr key={p.id} className={p.isHuman ? 'font-semibold' : ''}>
                      <td className="p-1">{playerName(p.id, p.isHuman)}</td>
                      <td className="text-right p-1">{p.trickCount}</td>
                      <td className="text-right p-1">{p.roundScore}</td>
                      <td className="text-right p-1">{p.cumulativeScore}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div
              data-tutorial="af-trick-display"
              className="border border-ds-border-subtle rounded p-2 min-h-[80px] mb-3 text-ds-text-primary"
            >
              <div className="text-xs uppercase opacity-60 mb-1">{t('currentTrick')}</div>
              {state.currentTrick.length === 0 ? (
                <div className="opacity-50">—</div>
              ) : (
                <div className="flex flex-wrap gap-2">
                  {state.currentTrick.map((tk) => (
                    <div
                      key={`${tk.playerIdx}-${tk.card.design}-${tk.card.value}`}
                      className="flex flex-col items-center"
                    >
                      <span className="text-[10px] opacity-60">{findPlayerName(state.players, tk.playerIdx)}</span>
                      <CardImage card={tk.card} width={cardWidth} />
                    </div>
                  ))}
                </div>
              )}
            </div>

            {human && human.cards.length > 0 && (
              <div data-tutorial="af-player-hand" className="flex flex-col gap-2 mb-3 text-ds-text-primary">
                <div className="text-xs uppercase opacity-60">
                  {findPlayerName(state.players, humanIdx)} ({t('cards', { count: human.cards.length })})
                </div>
                <div className="flex flex-wrap gap-2">
                  {human.cards.map((c, idx) => {
                    const isValid = state.validPlayIndices.includes(idx);
                    const isSelected = selectedCardIdx === idx;
                    const canSelect = isHumanPlayTurn && isValid && !loading;
                    return (
                      <button
                        key={`${c.design}-${c.value}-${idx}`}
                        type="button"
                        onClick={() => canSelect && setSelectedCardIdx(idx)}
                        disabled={!canSelect}
                        aria-label={cardLabel(c)}
                        className={`min-w-[44px] min-h-[44px] rounded transition-all
                          ${isSelected ? 'ring-2 ring-ds-accent' : ''}
                          ${canSelect ? 'opacity-100' : 'opacity-50 cursor-not-allowed'}
                        `}
                      >
                        <CardImage card={c} width={cardWidth} />
                      </button>
                    );
                  })}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.allfours.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[
                {
                  items: [
                    {
                      type: 'select',
                      id: 'cpuDifficulty',
                      label: t('settings.cpuDifficulty'),
                      value: cpuDifficulty,
                      options: [
                        { value: 0, label: t('settings.easy') },
                        { value: 1, label: t('settings.normal') },
                        { value: 2, label: t('settings.hard') },
                      ],
                      onSelect: (v) => handleConfigChange('cpuDifficulty', Number(v)),
                    },
                    {
                      type: 'select',
                      id: 'pointLimit',
                      label: t('settings.pointLimit'),
                      value: pointLimit,
                      options: [5, 7, 9, 11, 15, 21].map((v) => ({ value: v, label: String(v) })),
                      onSelect: (v) => handleConfigChange('pointLimit', Number(v)),
                    },
                    {
                      type: 'checkbox',
                      id: 'frontendHint',
                      label: tc('hint.toggle', { ns: 'tutorial' }),
                      checked: hintEnabled,
                      onToggle: setHintEnabled,
                    },
                  ],
                },
              ]}
            />

            {isBegPhase && isHumanBegTurn && (
              <div data-tutorial="af-beg-controls" className="flex flex-wrap gap-2 items-center justify-center pb-2">
                <span className="text-ds-text-primary text-sm">{t('begPhase')}</span>
                <button type="button" className={btnPrimary} disabled={loading} onClick={() => handleBeg(false)}>
                  {t('standButton')}
                </button>
                <button type="button" className={btnSecondary} disabled={loading} onClick={() => handleBeg(true)}>
                  {t('begButton')}
                </button>
              </div>
            )}

            {isGiftPhase && isHumanGiftTurn && (
              <div className="flex flex-wrap gap-2 items-center justify-center pb-2">
                <span className="text-ds-text-primary text-sm">{t('giftPhase')}</span>
                <button type="button" className={btnPrimary} disabled={loading} onClick={() => handleRespond(false)}>
                  {t('giftButton')}
                </button>
                <button type="button" className={btnSecondary} disabled={loading} onClick={() => handleRespond(true)}>
                  {t('runButton')}
                </button>
              </div>
            )}

            {isPlayPhase && isHumanPlayTurn && (
              <div className="flex justify-center pb-2">
                <button
                  type="button"
                  data-tutorial="af-play-button"
                  className={btnSuccess}
                  onClick={handlePlay}
                  disabled={selectedCardIdx === null || loading}
                >
                  {t('playButton')}
                </button>
              </div>
            )}

            {isTrickEnd && (
              <div className="flex justify-center pb-2">
                <button type="button" className={btnPrimary} onClick={handleNextTrick} disabled={loading}>
                  {t('nextTrick')}
                </button>
              </div>
            )}

            {isRoundEnd && !isGameEnd && (
              <div className="flex justify-center pb-2">
                <button type="button" className={btnPrimary} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              </div>
            )}

            {isGameEnd && (
              <div className="flex flex-col items-center gap-2 pb-2">
                <div className="text-lg font-bold text-ds-text-primary">
                  {state.winnerIdx >= 0 && state.players[state.winnerIdx]?.isHuman
                    ? t('result.humanWin')
                    : t('result.cpuWin', { cpuId: state.winnerIdx })}
                </div>
                <div className="flex justify-center gap-2">
                  <GameResetButton
                    isGameEnd={isGameEnd}
                    onReset={handleManualReset}
                    requestConfirm={requestConfirm}
                    loading={loading}
                  />
                </div>
              </div>
            )}

            <div className="flex justify-center gap-2 pb-2">
              <button
                type="button"
                data-tutorial="af-reset-button"
                className={btnDanger}
                onClick={() => requestConfirm(handleManualReset)}
                disabled={loading}
              >
                {tc('button.reset')}
              </button>
              <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
                {tc('actionLog.view')}
              </button>
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
