import { render, screen } from '@testing-library/svelte';
import { describe, it, expect } from 'vitest';
import ResponsiveGrid from '../ResponsiveGrid.svelte';

describe('ResponsiveGrid', () => {
  it('デフォルトプロパティでレンダリングされる', () => {
    render(ResponsiveGrid, {
      props: {},
      $$slots: {
        default: [{ component: 'div', props: {}, children: 'Grid Item' }]
      }
    });

    const grid = screen.getByText('Grid Item').parentElement;
    expect(grid).toHaveClass('responsive-grid');
  });

  it('カスタムカラム設定が適用される', () => {
    const cols = {
      mobile: 1,
      tablet: 2,
      desktop: 3,
      large: 4
    };

    render(ResponsiveGrid, {
      props: { cols },
      $$slots: {
        default: [{ component: 'div', props: {}, children: 'Grid Item' }]
      }
    });

    const grid = screen.getByText('Grid Item').parentElement;
    expect(grid.style.getPropertyValue('--cols-mobile')).toBe('1');
    expect(grid.style.getPropertyValue('--cols-tablet')).toBe('2');
    expect(grid.style.getPropertyValue('--cols-desktop')).toBe('3');
    expect(grid.style.getPropertyValue('--cols-large')).toBe('4');
  });

  it('カスタムギャップが適用される', () => {
    render(ResponsiveGrid, {
      props: { gap: '2rem' },
      $$slots: {
        default: [{ component: 'div', props: {}, children: 'Grid Item' }]
      }
    });

    const grid = screen.getByText('Grid Item').parentElement;
    expect(grid.style.getPropertyValue('--grid-gap')).toBe('2rem');
  });

  it('autoFitモードが機能する', () => {
    render(ResponsiveGrid, {
      props: { 
        autoFit: true,
        minItemWidth: '300px'
      },
      $$slots: {
        default: [{ component: 'div', props: {}, children: 'Grid Item' }]
      }
    });

    const grid = screen.getByText('Grid Item').parentElement;
    expect(grid).toHaveClass('auto-fit');
    expect(grid.style.gridTemplateColumns).toContain('repeat(auto-fit, minmax(300px, 1fr))');
  });

  it('alignItemsとjustifyItemsが適用される', () => {
    render(ResponsiveGrid, {
      props: { 
        alignItems: 'center',
        justifyItems: 'start'
      },
      $$slots: {
        default: [{ component: 'div', props: {}, children: 'Grid Item' }]
      }
    });

    const grid = screen.getByText('Grid Item').parentElement;
    expect(grid.style.getPropertyValue('--grid-align-items')).toBe('center');
    expect(grid.style.getPropertyValue('--grid-justify-items')).toBe('start');
  });

  it('カスタムクラス名が適用される', () => {
    render(ResponsiveGrid, {
      props: { className: 'custom-grid' },
      $$slots: {
        default: [{ component: 'div', props: {}, children: 'Grid Item' }]
      }
    });

    const grid = screen.getByText('Grid Item').parentElement;
    expect(grid).toHaveClass('custom-grid');
  });
});