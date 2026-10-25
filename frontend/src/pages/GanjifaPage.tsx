import { useEffect } from 'react';
import type { ganjifaApi } from '../api/gameApi';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, TARGET_ROUNDS_OPTIONS, useGanjifaGame } from '../hooks/useGanjifaGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { formatGanjifaSuit, type GanjifaResponse, isGanjifaStrongSuit } from '../types/card';
import { GanjifaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { GANJIFA_HELP, parseGanjifaCommand } from '../utils/cli/commands/ganjifaCommands';
import { formatGanjifaState } from '../utils/cli/formatters/ganjifaFormatter';
import { ganjifaHandCardStatus } from '../utils/ganjifaHandStatus';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Ganjifa tutorial step definitions. */
const GANJIFA_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="ganjifa-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="ganjifa-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ganjifa-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ganjifa-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const GANJIFA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [GanjifaPhase.PLAY]: 'play',
  [GanjifaPhase.TRICK_END]: 'trickEnd',
  [GanjifaPhase.ROUND_END]: 'roundEnd',
  [GanjifaPhase.GAME_END]: 'gameEnd',
};

type GanjifaGame = ReturnType<typeof useGanjifaGame>;
type GanjifaView = ReturnType<typeof useGanjifaView>;
type GanjifaContext = GamePageContext<GanjifaResponse, GanjifaGame, GanjifaView>;

function useGanjifaView({ game }: { game: GanjifaGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('ganjifa', GANJIFA_PHASE_KEYS);

  return { cardWidth, isMobile, phaseNames };
}

function Board({ state, t, view }: GanjifaContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isTrickEnd = state.phase === GanjifaPhase.TRICK_END;
  const isRoundEnd = state.phase === GanjifaPhase.ROUND_END;
  const isGameEnd = state.phase === GanjifaPhase.GAME_END || state.gameEndFlag;

  const trumpLabel = formatGanjifaSuit(state.trumpSuit);
  const trumpIsStrong = isGanjifaStrongSuit(state.trumpSuit);
  const dealer = state.players[state.dealerIdx];

  return (
    <>
      <div
        className="flex flex-wrap items-center justify-center gap-x-2 gap-y-0 text-xs sm:text-base text-ds-text-primary text-center mb-2"
        data-tutorial="ganjifa-info"
      >
        <span className="sm:mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="sm:mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span className="sm:mr-4">{t('trump', { suit: trumpLabel })}</span>
        <span>{t('target', { rounds: state.config.targetRounds })}</span>
        {/* 強さの向きがこのゲームの要 */}
        <span
          className={`text-xs font-semibold ${trumpIsStrong ? 'text-ds-info' : 'text-ds-warning'}`}
          data-testid="ganjifa-trump-group"
          role="status"
          aria-live="polite"
        >
          {trumpIsStrong ? t('trumpGroupStrong') : t('trumpGroupWeak')}
        </span>
        {dealer && (
          <span className="text-ds-text-muted text-sm" data-testid="ganjifa-dealer">
            {t('dealer', { name: playerName(dealer.id, dealer.isHuman) })}
          </span>
        )}
        <details className="basis-full text-sm sm:ml-auto sm:basis-auto" open={!view.isMobile || undefined}>
          <summary className="cursor-pointer select-none">{t('rules')}</summary>
          <div className="text-ds-text-muted w-full" data-testid="ganjifa-trump-auto-note">
            {t('trumpAutoNote')}
          </div>
        </details>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={view.cardWidth}
            label={t('currentTrick')}
            dataTutorial="ganjifa-trick-display"
            // トリック確定と同時に ResolveTrick が走るので、TRICK_END の
            // 間は leadPlayerIdx がそのトリックの勝者 (OhHell と同じ)。
            winnerIdx={isTrickEnd ? state.leadPlayerIdx : undefined}
            winnerLabel={t('trickWinnerBadge')}
          />
        </div>

        {/* Right: info sidebar */}
        <div>
          <div className="mb-2 px-2 py-1 rounded bg-black/30 text-ds-text-muted text-sm text-center sm:text-left">
            <div className="sm:hidden whitespace-nowrap" data-testid="ganjifa-mobile-scores">
              {t('scores')}{' '}
              {state.players.map((p, index) => (
                <span key={p.id}>
                  {index > 0 && ' · '}
                  {playerName(p.id, p.isHuman)} {p.score}
                </span>
              ))}
            </div>
            <div className="hidden sm:flex sm:flex-col sm:gap-x-0">
              {state.players.map((p) => (
                <span key={p.id} className="py-0.5">
                  {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                </span>
              ))}
            </div>
          </div>

          {humanPlayer && (
            <div className="mb-2 rounded bg-black/30 px-2 py-1 text-ds-text-muted text-sm">
              <div className="py-0.5" data-testid="ganjifa-human-player" data-tutorial="ganjifa-human-player">
                {playerName(humanPlayer.id, humanPlayer.isHuman)}: {t('cards', { count: humanPlayer.cardCount })} |{' '}
                {t('tricks', { count: humanPlayer.trickCount })}
              </div>
            </div>
          )}

          <CpuAccordion playerCount={state.players.filter((p) => !p.isHuman).length}>
            <div className="mb-2 p-2 rounded bg-black/30">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                    {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                    {t('tricks', { count: p.trickCount })}
                  </div>
                ))}
            </div>
          </CpuAccordion>

          {(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              {state.players.map((p) => (
                <div key={p.id}>
                  {t('roundResult.tricks', {
                    name: playerName(p.id, p.isHuman),
                    count: state.roundTricks[p.id] ?? 0,
                  })}
                  {t('roundResult.separator')}
                  {t('roundResult.score', { score: state.roundTricks[p.id] ?? 0 })}
                  {t('roundResult.separator')}
                  {t('roundResult.cumulativeScore', { score: p.score })}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ game, state, t, view }: GanjifaContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;

  const isPlayPhase = state.phase === GanjifaPhase.PLAY;
  const isHumanTurn = state.isHumanTurn;
  const canPlay = isPlayPhase && isHumanTurn;

  return (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="ganjifa"
      stackRows
      validIndices={canPlay ? state.playableIndices : undefined}
      restrictedTooltip={t('mustFollowSuit')}
      cardStatusFor={(idx) => {
        const rankDirection = ganjifaHandCardStatus(humanPlayer.cards[idx], t);
        const mustFollow = canPlay && state.currentTrick.length > 0 && !state.playableIndices.includes(idx);
        if (!mustFollow) return rankDirection;
        return [rankDirection, t('mustFollowSuit')].filter(Boolean).join(' — ');
      }}
    />
  );
}

function FooterStatus({ state, t, frontendHint }: GanjifaContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="ganjifa-hint-live" role="status" aria-live="polite">
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

function primaryActions({ game, state, t, busy }: GanjifaContext) {
  const isHumanTurn = state.isHumanTurn;
  const isPlayPhase = state.phase === GanjifaPhase.PLAY;
  const isTrickEnd = state.phase === GanjifaPhase.TRICK_END;
  const isRoundEnd = state.phase === GanjifaPhase.ROUND_END;
  const canPlay = isPlayPhase && isHumanTurn;

  // The hand-wired page always rendered this row (it carries the tutorial anchor), even empty on CPU turns.
  return (
    <div className="flex flex-wrap gap-2 items-center" data-tutorial="ganjifa-action-buttons">
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
    </div>
  );
}

const ganjifaDefinition: GameDefinition<
  GanjifaResponse,
  GanjifaGame,
  Parameters<typeof ganjifaApi.exec>,
  GanjifaView
> = {
  key: 'ganjifa',
  useGame: useGanjifaGame,
  useView: useGanjifaView,
  tutorial: GANJIFA_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 10 },
  cli: () => ({
    parseCommand: parseGanjifaCommand,
    formatResponse: formatGanjifaState,
    helpText: GANJIFA_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === GanjifaPhase.GAME_END || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerPlayer === humanIdx,
    };
  },
  actionLogDuringPlay: true,
  shell: ({ state }) => {
    const isGameEnd = state.phase === GanjifaPhase.GAME_END || state.gameEndFlag;
    return {
      isHumanTurn: state.isHumanTurn && !isGameEnd,
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
          value: game.ganjifaConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetRounds',
          label: t('settings.targetRounds'),
          value: game.ganjifaConfig.targetRounds,
          options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetRounds', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  footerTall: true,
  resetTutorialId: 'ganjifa-reset-button',
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the Ganjifa game page: a Mughal-era Indian 3-player trick-taker on a 96-card 8-suit deck. */
export const GanjifaPage = createGamePage(ganjifaDefinition);
