import { useMemo, useState } from 'react';
import type { zwickerApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardBack } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useZwickerGame } from '../hooks/useZwickerGame';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ZwickerResponse } from '../types/card';
import { ZwickerPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseZwickerCommand, ZWICKER_HELP } from '../utils/cli/commands/zwickerCommands';
import { formatZwickerState } from '../utils/cli/formatters/zwickerFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const ZW_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="zw-rule"]', messageKey: 'tutorial.values', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="zw-rule"]', messageKey: 'tutorial.jokers', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="zw-table"]', messageKey: 'tutorial.zwick', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="zw-seats"]', messageKey: 'tutorial.scoring', placement: 'bottom', advanceOn: 'next' },
];

/** Renders the Zwicker page: 55 cards, two-valued courts, Zwick clears the table. */
export const ZwickerPage = withTutorial(ZwickerPageContent, 'zwicker', ZW_TUTORIAL_STEPS);

function ZwickerPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('zwicker');
  const game = useZwickerGame();
  const { state, loading, error, retry } = game;

  // 出す 1 枚と、その札をどの値で使うか。**A と絵札は 2 択を持つ**ので、
  // 札を選んだだけでは捕獲が決まらない。
  const [handIdx, setHandIdx] = useState<number | null>(null);
  const [playedValue, setPlayedValue] = useState<number | null>(null);
  const [tableSel, setTableSel] = useState<number[]>([]);
  const [buildSel, setBuildSel] = useState<number[]>([]);
  const [buildValue, setBuildValue] = useState('');

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('zwicker');
  const cliConfig: CliGameConfig<ZwickerResponse, Parameters<typeof zwickerApi.exec>> = useMemo(
    () => ({
      gameName: 'zwicker',
      parseCommand: parseZwickerCommand,
      formatResponse: formatZwickerState,
      helpText: ZWICKER_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('zwicker', state);

  if (!state) {
    return <GameSkeleton gameKey="zwicker" layout={{ kind: 'tableau', topRow: 3, tableau: 4 }} />;
  }

  const ended = state.phase === ZwickerPhase.GAME_END;
  const roundOver = state.phase === ZwickerPhase.ROUND_END;
  const playing = state.phase === ZwickerPhase.PLAY;
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && playing && state.currentPlayerIdx === 0;
  const chosen = handIdx !== null ? human?.cards[handIdx] : undefined;

  const clearSelection = () => {
    setHandIdx(null);
    setPlayedValue(null);
    setTableSel([]);
    setBuildSel([]);
    setBuildValue('');
  };

  const pickHand = (i: number) => {
    if (handIdx === i) {
      clearSelection();
      return;
    }
    setHandIdx(i);
    // 値が 1 つしかない札なら選ばせる必要はない。
    const values = human?.cards[i]?.values ?? [];
    setPlayedValue(values.length === 1 ? (values[0] ?? null) : null);
  };

  const toggle = (list: number[], set: (v: number[]) => void, i: number) => {
    set(list.includes(i) ? list.filter((n) => n !== i) : [...list, i]);
  };

  const phaseName = ended ? t('phase.end') : roundOver ? t('phase.roundEnd') : t('phase.play');
  const canTake = handIdx !== null && playedValue !== null && (tableSel.length > 0 || buildSel.length > 0);
  const canBuild = handIdx !== null && tableSel.length > 0 && Number(buildValue) > 0;

  return (
    <GamePageShell
      title={tc('nav.zwicker')}
      gameThemeBg={gameTheme.zwicker.bg}
      phaseName={phaseName}
      gamePath="/zwicker"
      gameEndFlag={ended}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted">
            {t('stock')}: {state.stockCount} /{' '}
            {t('score', {
              us: state.teamScores[0] ?? 0,
              them: state.teamScores[1] ?? 0,
              target: state.targetScore,
            })}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      <LandscapeBanner message={t('landscapeBanner')} />

      <SettingsPanel
        title={tc('settings.title')}
        groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
      />

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            {/* Permanent, not tutorial-only: the two-valued courts and what a
                Zwick actually is are the two rules a player gets wrong. */}
            <div className="text-center text-xs text-ds-warning mb-3 font-medium" data-tutorial="zw-rule">
              {t('ruleLine')}
            </div>

            <div className="flex justify-center gap-4 mb-3 flex-wrap" data-tutorial="zw-seats">
              {opponents.map((o) => (
                <div key={`opp-${o.id.toString()}`} className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">
                    {t('seat', { name: `CPU${o.id.toString()}`, team: o.team })}
                    {' · '}
                    {t('taken', { n: o.capturedCount, z: o.zwicks })}
                  </div>
                  <div
                    className="flex gap-1 justify-center flex-wrap"
                    role="img"
                    aria-label={t('opponentHandAriaLabel', { name: `CPU${o.id.toString()}`, n: o.cardCount })}
                  >
                    {Array.from({ length: o.cardCount }, (_, i) => (
                      <CardBack key={`opp-${o.id.toString()}-c${i.toString()}`} width={cardWidth} />
                    ))}
                  </div>
                </div>
              ))}
            </div>

            <div className="mb-4" data-tutorial="zw-table">
              <div className="text-game-text-muted text-xs mb-1 text-center">{t('table')}</div>
              {state.tableCards.length === 0 ? (
                <div className="text-center text-game-text-muted text-xs">{t('emptyTable')}</div>
              ) : (
                <div className="flex gap-1 justify-center flex-wrap">
                  {state.tableCards.map((card, i) => {
                    const isHinted = frontendHintEnabled && state.hint?.tableIndices?.includes(i) === true;
                    return (
                      <button
                        key={`tbl-${i.toString()}`}
                        type="button"
                        data-testid="zwicker-table-card"
                        aria-pressed={tableSel.includes(i)}
                        aria-disabled={!isHumanTurn}
                        onClick={() => isHumanTurn && toggle(tableSel, setTableSel, i)}
                        // The hint says which table cards to take with the card it
                        // names; only the hand card was ever highlighted, leaving the
                        // multi-card capture for the player to work out (#4898).
                        data-hinted-table={(frontendHintEnabled && state.hint?.tableIndices?.includes(i)) || undefined}
                        className={[
                          'rounded min-h-11 flex flex-col items-center',
                          tableSel.includes(i) ? 'ring-2 ring-ds-accent' : '',
                          isHinted ? 'ring-2 ring-ds-warning' : '',
                        ].join(' ')}
                      >
                        <AnimatedCard card={card} width={cardWidth} draggable={false} />
                        {/* 値を出さないと何と取れるか判らない。 */}
                        <span className="text-[10px] text-ds-text-muted">{card.values.join('/')}</span>
                      </button>
                    );
                  })}
                </div>
              )}

              {state.builds.length > 0 && (
                <div className="flex flex-col gap-1 items-center mt-2">
                  {state.builds.map((b, i) => (
                    <button
                      key={`bld-${i.toString()}`}
                      type="button"
                      data-testid="zwicker-build"
                      aria-pressed={buildSel.includes(i)}
                      aria-disabled={!isHumanTurn}
                      onClick={() => isHumanTurn && toggle(buildSel, setBuildSel, i)}
                      className={[
                        'flex items-center gap-1 rounded px-2 py-1 min-h-11',
                        buildSel.includes(i) ? 'ring-2 ring-ds-accent' : '',
                      ].join(' ')}
                    >
                      <span className="text-[10px] text-ds-text-muted">
                        {t('buildLabel', { value: b.value, owner: b.owner })}
                      </span>
                      {b.cards.map((card, j) => (
                        <AnimatedCard
                          key={`bld-${i.toString()}-c${j.toString()}`}
                          card={card}
                          width={Math.round(cardWidth * 0.7)}
                          draggable={false}
                        />
                      ))}
                    </button>
                  ))}
                </div>
              )}
            </div>

            {roundOver && state.lastRound && (
              <div className="text-center text-sm mb-3" data-testid="zwicker-round-result">
                {t('roundResult', {
                  us: state.lastRound.total[0] ?? 0,
                  them: state.lastRound.total[1] ?? 0,
                })}
                {/* 同数だと 3 点が宙に浮く。黙っていると合計が合わないように見える。 */}
                {state.lastRound.majorityTeam < 0 && ` · ${t('majorityTied')}`}
              </div>
            )}

            <div className="text-center" data-tutorial="zw-hand">
              <div className="text-game-text-muted text-xs mb-1">
                {t('yourHand')}
                {' · '}
                {t('taken', { n: human?.capturedCount ?? 0, z: human?.zwicks ?? 0 })}
              </div>
              <div className="flex gap-1 justify-center flex-wrap">
                {(human?.cards ?? []).map((card, i) => (
                  <button
                    key={`hand-${i.toString()}`}
                    type="button"
                    data-hint-action="discard"
                    aria-pressed={handIdx === i}
                    aria-disabled={!isHumanTurn}
                    onClick={() => isHumanTurn && pickHand(i)}
                    className={[
                      'rounded transition-transform flex flex-col items-center',
                      isHumanTurn ? 'hover:-translate-y-2' : 'opacity-60',
                      handIdx === i ? 'ring-2 ring-ds-accent -translate-y-2' : '',
                      frontendHintEnabled && state.hint?.cardIndex === i ? 'ring-2 ring-ds-warning' : '',
                    ].join(' ')}
                  >
                    <AnimatedCard card={card} width={cardWidth} draggable={false} />
                    <span className="text-[10px] text-ds-text-muted">{card.values.join('/')}</span>
                  </button>
                ))}
              </div>

              {/* A と絵札は 2 択。どちらで使うかを選ばせる。 */}
              {isHumanTurn && chosen && chosen.values.length > 1 && (
                <div className="flex gap-2 justify-center mt-2" data-testid="zwicker-value-picker">
                  <span className="text-[10px] text-ds-text-muted self-center">{t('pickValue')}</span>
                  {chosen.values.map((v) => (
                    <button
                      key={`val-${v.toString()}`}
                      type="button"
                      data-testid="zwicker-value-option"
                      aria-pressed={playedValue === v}
                      onClick={() => setPlayedValue(v)}
                      className={[
                        'px-3 min-h-11 rounded text-sm',
                        playedValue === v ? 'ring-2 ring-ds-accent' : 'ring-1 ring-ds-border',
                      ].join(' ')}
                    >
                      {v}
                    </button>
                  ))}
                </div>
              )}

              {isHumanTurn && <div className="text-[10px] text-ds-text-muted mt-1">{t('selectHint')}</div>}
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={ended}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.zwicker.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isHumanTurn && (
                <>
                  <button
                    type="button"
                    data-hint-action="take"
                    className={`${btnPrimary} min-h-11`}
                    disabled={!canTake}
                    onClick={() => {
                      if (handIdx !== null && playedValue !== null) {
                        game.handleTake(handIdx, playedValue, tableSel, buildSel);
                        clearSelection();
                      }
                    }}
                  >
                    {t('take')}
                  </button>
                  <input
                    type="number"
                    min={1}
                    aria-label={t('buildValue')}
                    placeholder={t('buildValue')}
                    value={buildValue}
                    onChange={(e) => setBuildValue(e.target.value)}
                    className="w-20 min-h-11 rounded px-2 text-sm ring-1 ring-ds-border bg-transparent"
                  />
                  <button
                    type="button"
                    className={`${btnSecondary} min-h-11`}
                    disabled={!canBuild}
                    onClick={() => {
                      if (handIdx !== null) {
                        game.handleBuild(handIdx, tableSel, Number(buildValue));
                        clearSelection();
                      }
                    }}
                  >
                    {t('build')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-h-11`}
                    disabled={handIdx === null}
                    onClick={() => {
                      if (handIdx !== null) {
                        game.handleTrail(handIdx);
                        clearSelection();
                      }
                    }}
                  >
                    {t('trail')}
                  </button>
                </>
              )}
              {roundOver && !ended && (
                <button type="button" className={`${btnPrimary} min-h-11`} onClick={game.handleNextRound}>
                  {t('nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={ended}
                onReset={game.handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="zw-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
