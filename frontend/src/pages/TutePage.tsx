import { useEffect } from 'react';
import type { tuteApi } from '../api/gameApi';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useTuteGame } from '../hooks/useTuteGame';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { TuteResponse } from '../types/card';
import { TutePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { suitSymbolAt } from '../utils/cardAlt';
import { suitName } from '../utils/cardUtils';
import { parseTuteCommand, TUTE_HELP } from '../utils/cli/commands/tuteCommands';
import { formatTuteState } from '../utils/cli/formatters/tuteFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit id → `suitName.*` i18n key (1=♠ .. 4=♦). */
const SUIT_KEYS = ['', 'spade', 'club', 'heart', 'diamond'] as const;

/** Tute tutorial step definitions. */
const TUTE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tute-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tute-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tute-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="tute-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="tute-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const TUTE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TutePhase.PLAY]: 'play',
  [TutePhase.TRICK_END]: 'trickEnd',
  [TutePhase.ROUND_END]: 'roundEnd',
  [TutePhase.GAME_END]: 'gameEnd',
};

type TuteGame = ReturnType<typeof useTuteGame>;
type TuteView = ReturnType<typeof useTuteView>;
type TuteContext = GamePageContext<TuteResponse, TuteGame, TuteView>;

function useTuteView({ game }: { game: TuteGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('tute', TUTE_PHASE_KEYS);
  return { cardWidth, isMobile, phaseNames };
}

function Board({ state, t, view }: TuteContext) {
  const { cardWidth, isMobile } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const teamLabels = [t('team.a'), t('team.b')] as const;

  const isPlayPhase = state.phase === TutePhase.PLAY;
  const isTrickEnd = state.phase === TutePhase.TRICK_END;
  const isRoundEnd = state.phase === TutePhase.ROUND_END;
  const isGameEnd = state.phase === TutePhase.GAME_END || state.gameEndFlag;

  // Rendered both during play (so a marriage is seen to score) and at round end.
  const teamPointLines = (
    <>
      <div>{t('roundResult.teamA', { points: state.roundTeamPoints[0] })}</div>
      <div>{t('roundResult.teamB', { points: state.roundTeamPoints[1] })}</div>
    </>
  );
  const trumpSymbol = suitSymbolAt(state.trumpSuit, '');

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span>{t('trump', { suit: trumpSymbol })}</span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players.map((p) => ({ ...p, team: p.teamId }))}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            winnerIdx={isTrickEnd || isRoundEnd ? state.leadPlayerIdx : undefined}
            winnerLabel={t('trickWinner')}
            dataTutorial="tute-trick-display"
          />
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="tute-info">
          {/* Team scores */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-xs whitespace-nowrap">
            {t('teamScores', {
              a: state.teamScores[0],
              b: state.teamScores[1],
              target: state.config.targetPoints,
            })}
          </div>

          <details
            open={!isMobile || undefined}
            className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
            data-testid="tute-reference-info"
          >
            <summary className="cursor-pointer select-none min-h-[44px] flex items-center text-ds-text-primary">
              {t('referenceInfo.title')}
            </summary>
            <div className="mt-1">
              <div className="text-ds-text-primary">{t('scoringReference.title')}</div>
              <div>{t('scoringReference.ace')}</div>
              <div>{t('scoringReference.three')}</div>
              <div>{t('scoringReference.king')}</div>
              <div>{t('scoringReference.queen')}</div>
              <div>{t('scoringReference.jack')}</div>
              <div className="mt-1">{t('scoringReference.lastTrick')}</div>
              <div data-testid="tute-declared-marriages">
                <div className="mt-2 text-ds-text-primary">{t('declaredMarriages.title')}</div>
                <div className="mt-2">
                  {([1, 2, 3, 4] as const).map((suit) => {
                    const declared = state.declaredSuits[suit] ?? false;
                    return (
                      <span key={suit} className="mr-3">
                        <span className="mr-1">{suitSymbolAt(suit, '')}</span>
                        {suit === state.trumpSuit && <span className="mr-1 text-ds-warning">★</span>}
                        <span className={`whitespace-nowrap ${declared ? 'text-ds-text-primary' : ''}`}>
                          {declared ? t('declaredMarriages.declared') : t('declaredMarriages.undeclared')}
                        </span>
                      </span>
                    );
                  })}
                </div>
                <div className="mt-1 text-xs">{t('declaredMarriages.trumpNote')}</div>
              </div>
              {humanPlayer && (
                <div className="mt-2" data-testid="tute-human-info">
                  {t('playerTeam', {
                    name: playerName(humanPlayer.id, true),
                    team: teamLabels[humanPlayer.teamId],
                  })}
                  : {t('cards', { count: humanPlayer.cardCount })} | {t('tricks', { count: humanPlayer.trickCount })}
                </div>
              )}
              {(isPlayPhase || isTrickEnd) && (
                <div className="mt-2" data-testid="tute-running-points" role="status" aria-live="polite">
                  <div className="text-ds-text-primary">{t('roundResult.runningTitle')}</div>
                  {teamPointLines}
                </div>
              )}
              {(isRoundEnd || isGameEnd) && (
                <div className="mt-2">
                  <div className="text-ds-text-primary">{t('roundResult.title')}</div>
                  {teamPointLines}
                  {state.lastTrickBonusTeam >= 0 && (
                    <div data-testid="tute-last-trick-bonus">
                      {t('roundResult.lastTrickBonus', {
                        team: state.lastTrickBonusTeam === 0 ? t('team.a') : t('team.b'),
                        points: state.lastTrickBonusPoints,
                      })}
                    </div>
                  )}
                </div>
              )}
            </div>
          </details>
          <CpuAccordion playerCount={state.players.filter((p) => !p.isHuman).length}>
            <div className="flex flex-wrap gap-x-3">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                    {t('playerTeam', { name: playerName(p.id, false), team: teamLabels[p.teamId] })}:{' '}
                    {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
                  </div>
                ))}
            </div>
          </CpuAccordion>
        </div>
      </div>

      {isMobile && (
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
      )}
    </>
  );
}

function FooterHand({ game, state, t, view }: TuteContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.currentPlayerIdx === humanIdx;
  const canPlay = state.phase === TutePhase.PLAY && isHumanTurn;
  // Trump suit is 0 before the last dealt card establishes it.
  const trumpDesign = suitName(state.trumpSuit);
  const trumpIndices = trumpDesign
    ? humanPlayer.cards.reduce<number[]>((indices, card, index) => {
        if (card.design === trumpDesign) indices.push(index);
        return indices;
      }, [])
    : [];

  return (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="tute"
      validIndices={canPlay ? state.playableIndices : undefined}
      restrictedTooltip={t('playButton')}
      trumpIndices={trumpIndices}
      trumpTitle={t('trumpTitle')}
    />
  );
}

function FooterStatus({ state, t, frontendHint }: TuteContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="tute-hint-live" role="status" aria-live="polite">
        {state.hint && isRequestedHint(state) && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
            {state.hint.cardIndices &&
              state.hint.cardIndices.length > 0 &&
              ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ game, state, t, busy }: TuteContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.currentPlayerIdx === humanIdx;
  const isPlayPhase = state.phase === TutePhase.PLAY;
  const isTrickEnd = state.phase === TutePhase.TRICK_END;
  const isRoundEnd = state.phase === TutePhase.ROUND_END;
  const canPlay = isPlayPhase && isHumanTurn;

  return (
    <>
      {canPlay && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handlePlay}
          disabled={busy || game.selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
      )}
      {canPlay && (state.canDeclareMarriage || state.canDeclareTute) && (
        <fieldset className="flex flex-wrap items-center gap-2 border-0 p-0 m-0" data-testid="tute-declarations">
          <legend className="text-xs text-ds-text-muted mr-1">{t('declarationsLabel')}</legend>
          {state.canDeclareMarriage &&
            humanPlayer &&
            ([1, 2, 3, 4] as const)
              .filter((suit) => {
                // Only offer suits the human actually holds K+Q of and has
                // not already declared, so every button is a legal move.
                const design = (['', 'SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as const)[suit];
                const hasK = humanPlayer.cards.some((c) => c.design === design && c.value === 13);
                const hasQ = humanPlayer.cards.some((c) => c.design === design && c.value === 12);
                return hasK && hasQ && !state.declaredSuits[suit];
              })
              .map((suit) => (
                <button
                  key={suit}
                  type="button"
                  className={btnSecondary}
                  onClick={() => game.handleDeclareMarriage(suit)}
                  disabled={busy}
                  aria-label={t('marriageAria', {
                    suit: t(`suitName.${SUIT_KEYS[suit]}`),
                    points: suit === state.trumpSuit ? 40 : 20,
                  })}
                >
                  {t('declareMarriage', { suit: suitSymbolAt(suit, '') })}
                </button>
              ))}
          {state.canDeclareTute && (
            <>
              <button
                type="button"
                className={btnSecondary}
                onClick={game.handleDeclareTute}
                disabled={busy}
                title={t('tuteHelp')}
                aria-describedby="tute-help-desc"
                data-testid="tute-declare-button"
              >
                {t('declareTute')}
              </button>
              <span id="tute-help-desc" className="sr-only">
                {t('tuteHelp')}
              </span>
            </>
          )}
        </fieldset>
      )}
      {isTrickEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextTrick} disabled={busy}>
          {t('nextTrick')}
        </button>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </>
  );
}

const tuteDefinition: GameDefinition<TuteResponse, TuteGame, Parameters<typeof tuteApi.exec>, TuteView> = {
  key: 'tute',
  useGame: useTuteGame,
  useView: useTuteView,
  tutorial: TUTE_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 6 },
  cli: () => ({
    parseCommand: parseTuteCommand,
    formatResponse: formatTuteState,
    helpText: TUTE_HELP,
  }),
  phase: ({ state, view }) => {
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    const humanTeam = humanIdx >= 0 ? humanIdx % 2 : -1;
    const ended = state.phase === TutePhase.GAME_END || state.gameEndFlag;
    return {
      name: view.phaseNames[state.phase],
      ended,
      won: ended && state.winnerTeam === humanTeam,
    };
  },
  shell: ({ state }) => {
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    return {
      isHumanTurn: state.currentPlayerIdx === humanIdx,
    };
  },
  reset: ({ game }) => {
    game.reset();
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.tuteConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetPoints',
          label: t('settings.targetPoints'),
          value: game.tuteConfig.targetPoints,
          options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetPoints', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  message: ({ state, view }) =>
    !view.isMobile ? (
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    ) : null,
  actionsTutorialId: 'tute-action-buttons',
  resetTutorialId: 'tute-reset-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Tute game page: a Spanish 4-player (2 vs 2) trump trick-taker with marriage and Tute declarations. */
export const TutePage = createGamePage(tuteDefinition);
