import { render, screen } from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import AnimatedTransition from '../AnimatedTransition.svelte';

// Svelteトランジションのモック
vi.mock('svelte/transition', () => ({
  fade: vi.fn(() => ({ duration: 300 })),
  fly: vi.fn(() => ({ duration: 300 })),
  scale: vi.fn(() => ({ duration: 300 })),
  slide: vi.fn(() => ({ duration: 300 }))
}));

vi.mock('svelte/easing', () => ({
  quintOut: vi.fn(),
  elasticOut: vi.fn(),
  bounceOut: vi.fn()
}));

describe('AnimatedTransition', () => {
  it('show=trueの時にコンテンツが表示される', () => {
    render(AnimatedTransition, {
      props: {
        show: true
      },
      $$slots: {
        default: [{ component: 'div', props: {}, children: 'Test Content' }]
      }
    });

    expect(screen.getByText('Test Content')).toBeInTheDocument();
  });

  it('show=falseの時にコンテンツが非表示になる', () => {
    render(AnimatedTransition, {
      props: {
        show: false
      },
      $$slots: {
        default: [{ component: 'div', props: {}, children: 'Test Content' }]
      }
    });

    expect(screen.queryByText('Test Content')).not.toBeInTheDocument();
  });

  it('カスタムクラス名が適用される', () => {
    render(AnimatedTransition, {
      props: {
        show: true,
        className: 'custom-class'
      },
      $$slots: {
        default: [{ component: 'div', props: {}, children: 'Test Content' }]
      }
    });

    const element = screen.getByText('Test Content').parentElement;
    expect(element).toHaveClass('custom-class');
  });

  it('カスタムタグが使用される', () => {
    render(AnimatedTransition, {
      props: {
        show: true,
        tag: 'section'
      },
      $$slots: {
        default: [{ component: 'div', props: {}, children: 'Test Content' }]
      }
    });

    const element = screen.getByText('Test Content').parentElement;
    expect(element.tagName.toLowerCase()).toBe('section');
  });

  it('デフォルトでdivタグが使用される', () => {
    render(AnimatedTransition, {
      props: {
        show: true
      },
      $$slots: {
        default: [{ component: 'div', props: {}, children: 'Test Content' }]
      }
    });

    const element = screen.getByText('Test Content').parentElement;
    expect(element.tagName.toLowerCase()).toBe('div');
  });
});