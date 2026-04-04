import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { AdaptationLevel, StrategyStyle } from '../utils/metaAiAdaptation';
import { MetaAiIndicator } from './MetaAiIndicator';

describe('MetaAiIndicator', () => {
  const strategies: StrategyStyle[] = ['aggressive', 'defensive', 'balanced', 'cautious', 'observing'];

  describe('adaptation levels', () => {
    it('renders learning state in gray', () => {
      render(<MetaAiIndicator adaptationLevel="learning" strategyStyle="balanced" />);
      const el = screen.getByTestId('meta-ai-indicator');
      expect(el).toHaveTextContent('AI: 学習中...');
      expect(el.className).toContain('text-gray-400');
    });

    it('renders adapting state in yellow', () => {
      render(<MetaAiIndicator adaptationLevel="adapting" strategyStyle="balanced" />);
      const el = screen.getByTestId('meta-ai-indicator');
      expect(el).toHaveTextContent('AI: 適応中');
      expect(el.className).toContain('text-yellow-300');
    });

    it('renders adapted state in green', () => {
      render(<MetaAiIndicator adaptationLevel="adapted" strategyStyle="balanced" />);
      const el = screen.getByTestId('meta-ai-indicator');
      expect(el).toHaveTextContent('AI: 適応済');
      expect(el.className).toContain('text-green-300');
    });
  });

  describe('strategy styles', () => {
    const expectedText: Record<StrategyStyle, string> = {
      aggressive: '攻撃的',
      defensive: '守備的',
      balanced: 'バランス型',
      cautious: '慎重',
      observing: '観察中',
    };

    it.each(strategies)('renders strategy "%s" with correct translation', (strategy) => {
      render(<MetaAiIndicator adaptationLevel="adapted" strategyStyle={strategy} />);
      const el = screen.getByTestId('meta-ai-indicator');
      expect(el).toHaveTextContent(`[${expectedText[strategy]}]`);
    });
  });

  describe('combined rendering', () => {
    const cases: { level: AdaptationLevel; strategy: StrategyStyle }[] = [
      { level: 'learning', strategy: 'aggressive' },
      { level: 'adapting', strategy: 'defensive' },
      { level: 'adapted', strategy: 'observing' },
    ];

    it.each(cases)('renders level=$level with strategy=$strategy', ({ level, strategy }) => {
      render(<MetaAiIndicator adaptationLevel={level} strategyStyle={strategy} />);
      const el = screen.getByTestId('meta-ai-indicator');
      expect(el).toBeInTheDocument();
    });
  });
});
