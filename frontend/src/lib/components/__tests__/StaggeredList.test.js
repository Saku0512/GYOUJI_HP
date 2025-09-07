import { render, screen, waitFor } from '@testing-library/svelte';
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import StaggeredList from '../StaggeredList.svelte';

describe('StaggeredList', () => {
  const mockItems = [
    { id: 1, name: 'Item 1' },
    { id: 2, name: 'Item 2' },
    { id: 3, name: 'Item 3' }
  ];

  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('アイテムが順次表示される', async () => {
    render(StaggeredList, {
      props: {
        items: mockItems,
        staggerDelay: 100
      },
      $$slots: {
        default: [
          {
            component: 'div',
            props: { 'let:item': true },
            children: '{{item.name}}'
          }
        ]
      }
    });

    // 初期状態では何も表示されていない
    expect(screen.queryByText('Item 1')).not.toBeInTheDocument();

    // 100ms後に最初のアイテムが表示される
    vi.advanceTimersByTime(100);
    await waitFor(() => {
      expect(screen.getByText('Item 1')).toBeInTheDocument();
    });

    // 200ms後に2番目のアイテムが表示される
    vi.advanceTimersByTime(100);
    await waitFor(() => {
      expect(screen.getByText('Item 2')).toBeInTheDocument();
    });

    // 300ms後に3番目のアイテムが表示される
    vi.advanceTimersByTime(100);
    await waitFor(() => {
      expect(screen.getByText('Item 3')).toBeInTheDocument();
    });
  });

  it('カスタムスタガー遅延が機能する', async () => {
    render(StaggeredList, {
      props: {
        items: mockItems,
        staggerDelay: 200
      },
      $$slots: {
        default: [
          {
            component: 'div',
            props: { 'let:item': true },
            children: '{{item.name}}'
          }
        ]
      }
    });

    // 200ms後に最初のアイテムが表示される
    vi.advanceTimersByTime(200);
    await waitFor(() => {
      expect(screen.getByText('Item 1')).toBeInTheDocument();
    });

    // 400ms後に2番目のアイテムが表示される
    vi.advanceTimersByTime(200);
    await waitFor(() => {
      expect(screen.getByText('Item 2')).toBeInTheDocument();
    });
  });

  it('カスタムクラス名が適用される', () => {
    const { container } = render(StaggeredList, {
      props: {
        items: mockItems,
        className: 'custom-list',
        itemClassName: 'custom-item'
      },
      $$slots: {
        default: [
          {
            component: 'div',
            props: { 'let:item': true },
            children: '{{item.name}}'
          }
        ]
      }
    });

    const list = container.querySelector('.staggered-list');
    expect(list).toHaveClass('custom-list');
  });

  it('カスタムタグが使用される', () => {
    const { container } = render(StaggeredList, {
      props: {
        items: mockItems,
        tag: 'ul',
        itemTag: 'li'
      },
      $$slots: {
        default: [
          {
            component: 'div',
            props: { 'let:item': true },
            children: '{{item.name}}'
          }
        ]
      }
    });

    const list = container.querySelector('ul');
    expect(list).toBeInTheDocument();
  });

  it('空のアイテム配列でエラーが発生しない', () => {
    expect(() => {
      render(StaggeredList, {
        props: {
          items: []
        },
        $$slots: {
          default: [
            {
              component: 'div',
              props: { 'let:item': true },
              children: '{{item.name}}'
            }
          ]
        }
      });
    }).not.toThrow();
  });
});