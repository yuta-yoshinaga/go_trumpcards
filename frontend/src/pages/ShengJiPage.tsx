import { useEffect, useMemo } from 'react';
import { shengjiApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardSelection } from '../hooks/useCardSelection';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ShengJiResponse } from '../types/card';
import { ShengJiPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseShengJiCommand, SHENGJI_HELP } from '../utils/cli/commands/shengjiCommands';
import { formatShengJiState } from '../utils/cli/formatters/shengjiFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Combination names by wire code (sync: `ShengJiComboKind`). */
const COMBO_KEYS: Readonly<Record<number, string>> = {
  1: 'comboSingle',
  2: 'comboPair',
  3: 'comboTractor',
};

/** Sheng Ji tutorial step definitions. */
const SHENGJI_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="shengji-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="shengji-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="shengji-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="shengji-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="shengji-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SHENGJI_PHASE_KEYS: Readonly<Record<number, string>> = {
  [ShengJiPhase.DECLARE]: 'declare',
  [ShengJiPhase.KITTY]: 'kitty',
  [ShengJiPhase.PLAY]: 'play',
  [ShengJiPhase.HAND_END]: 'handEnd',
  [ShengJiPhase.GAME_END]: 'gameEnd',
};

/** Renders the Sheng Ji (升级 / 拖拉机) game page: a two-pack point-trick game played at rising levels. */
export const ShengJiPage = withTutorial(ShengJiPageContent, 'shengji', SHENGJI_TUTORIAL_STEPS);

/** Inner content of the Sheng Ji page, wrapped by TutorialProvider. */
function ShengJiPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('shengji');
  const { state, loading, error, exec, retry } = useGameApi(shengjiApi.exec);
  const { selected, toggle, clear } = useCardSelection();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('shengji');
  const cliConfig: CliGameConfig<ShengJiResponse, Parameters<typeof shengjiApi.exec>> = useMemo(
    () => ({
      gameName: 'shengji',
      parseCommand: parseShengJiCommand,
      formatResponse: formatShengJiState,
      helpText: SHENGJI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('shengji', SHENGJI_PHASE_KEYS);
  // **ヒントは前から算出されていた。**getShengJiHint も hintFactories への
  // 登録もあるのに、このページが一度も読んでいなかった (#4774)。
  // check-hint-coverage はファクトリの有無しか見ないので CI をすり抜けていた。
  //
  // **フックは早期 return より上。**下に置くと初回レンダーだけフック数が
  // 変わってページが骨組みのまま固まる。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('shengji', state);

  if (!state)
    return <GameSkeleton gameKey="shengji" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 12 }} />;

  const human = state.players.find((p) => p.isHuman);
  const isGameEnd = state.phase === ShengJiPhase.GAME_END || state.gameEndFlag;
  const isHandEnd = !isGameEnd && state.phase === ShengJiPhase.HAND_END;
  const isHumanTurn = !isGameEnd && !isHandEnd && state.currentPlayerIdx === 0;
  const humanWon = isGameEnd && state.winnerTeam === 0;
  const isDeclare = state.phase === ShengJiPhase.DECLARE && isHumanTurn;
  const isKitty = state.phase === ShengJiPhase.KITTY && isHumanTurn;
  const isPlay = state.phase === ShengJiPhase.PLAY && isHumanTurn;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));
  const suitLabel = (suit: number): string => (suit === 0 ? t('noTrump') : t(`suitName.${suit}`));
  // **レベルは 2〜A。**数字のままでは J/Q/K/A が読めない。
  const levelLabel = (level: number): string => ({ 11: 'J', 12: 'Q', 13: 'K', 14: 'A' })[level] ?? String(level);
  const comboLabel = (kind: number): string => (COMBO_KEYS[kind] ? t(COMBO_KEYS[kind]) : '-');

  // **点を集めるのは守備側。**宣言側の合計を出しても意味がない。
  const defenders = 1 - state.declarerTeam;
  const defenderPoints = state.teamPoints[defenders] ?? 0;
  const declarableSuits = Object.entries(state.declarableSuits).map(([suit, strength]) => ({
    suit: Number(suit),
    strength,
  }));

  const handlePlay = () => {
    if (selected.length === 0) return;
    exec('play', { cardIndexes: [...selected].sort((a, b) => a - b) });
    clear();
  };

  const handleBury = () => {
    if (selected.length !== state.kittySizeMax) return;
    exec('bury', { cardIndexes: [...selected].sort((a, b) => a - b) });
    clear();
  };

  const handleDeclare = (suit: number) => {
    clear();
    exec('declare', { suit });
  };

  const handleNext = () => {
    clear();
    exec('next');
  };

  const handleManualReset = () => {
    hideActionLog();
    clear();
    exec('reset');
  };

  const canSelect = isPlay || isKitty;

  return (
    <GamePageShell
      title={tc('nav.shengji')}
      gameThemeBg={gameTheme.shengji.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/shengji"
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
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* The trump group and who collects the points are the two things you cannot play without. */}
            <div className="mb-2 p-2 rounded bg-black/20 text-sm" data-testid="shengji-info">
              <div data-tutorial="shengji-info">
                {t('scoreLine', {
                  hand: state.handNumber,
                  level: levelLabel(state.level),
                  trump: suitLabel(state.trumpSuit),
                  t0: levelLabel(state.teamLevels[0]),
                  t1: levelLabel(state.teamLevels[1]),
                })}
              </div>
              <div className="text-xs text-ds-text-muted" data-testid="shengji-trump-note">
                {t('trumpNote', { level: levelLabel(state.level) })}
              </div>
              <div className="text-xs text-ds-text-muted" data-testid="shengji-points-note">
                {t('pointsLine', {
                  team: defenders,
                  points: defenderPoints,
                  target: state.defenderTarget,
                  total: state.totalPoints,
                })}
              </div>
            </div>

            {/* Players — which side you are on decides what you are trying to do. */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="shengji-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => (
                <div
                  key={`player-${p.id}`}
                  data-testid="shengji-player"
                  className={`text-sm py-0.5 flex items-center gap-2 ${
                    p.isCurrentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
                  } ${p.isHuman ? 'font-semibold' : ''}`}
                >
                  <span>{t('seat', { n: p.id })}</span>
                  <span>{playerLabel(p.id, p.isHuman)}</span>
                  <span>({t('team', { n: p.team })})</span>
                  <span className={p.isDeclarer ? 'text-ds-accent' : ''}>
                    {p.isDeclarer ? t('declarer') : t('defender')}
                  </span>
                  <span>{t('cardCount', { count: p.cardCount })}</span>
                  {p.isCurrentTurn && !isGameEnd && <span className="text-ds-accent">[{t('turnTag')}]</span>}
                </div>
              ))}
            </div>

            {/* The trick. */}
            <div className="mb-2 p-2 rounded bg-black/20 text-sm" data-testid="shengji-trick">
              {state.trick.length === 0 ? (
                <span>{t('trickEmpty')}</span>
              ) : (
                <>
                  {state.leadCombo && (
                    <div>{t('trickLead', { combo: comboLabel(state.leadCombo.kind), size: state.leadCombo.size })}</div>
                  )}
                  {state.trick.map((play) => (
                    <div key={`play-${play.seat}`} className="flex items-center gap-1 flex-wrap">
                      <span className="text-xs text-ds-text-muted">{t('seat', { n: play.seat })}</span>
                      {play.cards.map((c, i) => (
                        <CardImage key={`trick-${play.seat}-${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />
                      ))}
                    </div>
                  ))}
                </>
              )}
            </div>

            {/* The settled hand. */}
            {isHandEnd && state.lastResult && (
              <div className="mb-2 p-2 rounded bg-black/20 text-sm" data-testid="shengji-hand-result">
                <div>
                  {t(state.lastResult.declarerHeld ? 'handHeldLine' : 'handTakenLine', {
                    points: state.lastResult.defenderPoints,
                    target: state.defenderTarget,
                    team: state.lastResult.advancingTeam,
                    advance: state.lastResult.advance,
                  })}
                </div>
                {/* **底牌の倍率は最終トリックを取った側にしか掛からない。** */}
                {state.lastResult.kittyMultiplier > 0 && (
                  <div data-testid="shengji-kitty-line">
                    {t('kittyLine', {
                      points: state.lastResult.kittyPoints,
                      mult: state.lastResult.kittyMultiplier,
                    })}
                  </div>
                )}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <label className="flex items-center gap-1 text-ds-text-primary text-xs w-full justify-center cursor-pointer min-h-[44px]">
              <input
                type="checkbox"
                checked={frontendHintEnabled}
                onChange={(e) => setFrontendHintEnabled(e.target.checked)}
              />
              {tc('hint.toggle', { ns: 'tutorial' })}
            </label>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.shengji.footer} px-4 py-2.5`}>
            <div className="mb-2" data-tutorial="shengji-hand">
              <div className="text-ds-text-muted text-xs mb-1">{t('yourHand')}</div>
              <div className="flex flex-wrap gap-1">
                {(human?.cards ?? []).map((c, i) => (
                  <button
                    key={`hand-${c.design}-${c.value}-${i}`}
                    type="button"
                    onClick={() => canSelect && toggle(i)}
                    disabled={!canSelect}
                    className={`rounded transition-all ${selected.includes(i) ? 'ring-2 ring-ds-info -translate-y-2' : ''} ${
                      canSelect ? 'cursor-pointer hover:opacity-90' : 'cursor-default'
                    }`}
                    data-testid={`hand-card-${i}`}
                  >
                    <CardImage card={c} width={cardWidth} />
                  </button>
                ))}
              </div>
            </div>

            <ErrorAlert message={error} onRetry={retry} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="shengji-actions">
              {isDeclare && (
                <>
                  <span className="text-ds-text-muted text-xs" data-testid="shengji-declare-rules">
                    {t('declareRules')}
                  </span>
                  {declarableSuits.map(({ suit, strength }) => (
                    <button
                      key={`declare-${suit}`}
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleDeclare(suit)}
                      disabled={loading}
                      data-testid={`shengji-declare-${suit}`}
                    >
                      {t('declareButton', { suit: suitLabel(suit), strength })}
                    </button>
                  ))}
                  {/* **0 はパス。**宣言できる札が無くても手番は進める。 */}
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={() => handleDeclare(0)}
                    disabled={loading}
                    data-testid="shengji-pass"
                  >
                    {t('passButton')}
                  </button>
                </>
              )}

              {isKitty && (
                <>
                  <span className="text-ds-text-muted text-xs" data-testid="shengji-kitty-rules">
                    {t('kittyRules', { count: state.kittySizeMax })}
                  </span>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleBury}
                    disabled={loading || selected.length !== state.kittySizeMax}
                  >
                    {t('buryButton', { selected: selected.length, count: state.kittySizeMax })}
                  </button>
                </>
              )}

              {isPlay && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handlePlay}
                  disabled={loading || selected.length === 0}
                >
                  {t('playButton')}
                </button>
              )}

              {isHandEnd && (
                <button type="button" className={btnPrimary} onClick={handleNext} disabled={loading}>
                  {t('nextButton')}
                </button>
              )}

              {isGameEnd && (
                <span className="text-ds-text-primary text-sm font-semibold mr-1">
                  {humanWon ? t('win') : t('lose')}
                </span>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="shengji-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
