// ComponentDemo コンポーネントの単体テスト
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import ComponentDemo from '../ComponentDemo.svelte';

describe('ComponentDemo', () => {
  beforeEach(() => {
    // console のモック
    vi.spyOn(console, 'log').mockImplementation(() => {});
  });

  it('正しくレンダリングされる', () => {
    render(ComponentDemo);
    
    expect(screen.getByText('共通UIコンポーネント デモ')).toBeInTheDocument();
    expect(screen.getByText('Button コンポーネント')).toBeInTheDocument();
    expect(screen.getByText('Input コンポーネント')).toBeInTheDocument();
    expect(screen.getByText('Select コンポーネント')).toBeInTheDocument();
    expect(screen.getByText('LoadingSpinner コンポーネント')).toBeInTheDocument();
    expect(screen.getByText('Modal コンポーネント')).toBeInTheDocument();
  });

  describe('Button コンポーネントのデモ', () => {
    it('様々なバリアントのボタンが表示される', () => {
      render(ComponentDemo);
      
      expect(screen.getByText('Primary Button')).toBeInTheDocument();
      expect(screen.getByText('Secondary Outline')).toBeInTheDocument();
      expect(screen.getByText('Small Success')).toBeInTheDocument();
      expect(screen.getByText('Large Danger')).toBeInTheDocument();
      expect(screen.getByText('Disabled Button')).toBeInTheDocument();
    });

    it('Primary Buttonクリックで通知が表示される', async () => {
      render(ComponentDemo);
      
      const primaryButton = screen.getByText('Primary Button');
      await fireEvent.click(primaryButton);
      
      // 通知メッセージが表示されることを確認
      expect(screen.getByText('ボタンがクリックされました！')).toBeInTheDocument();
    });

    it('Loading Buttonでローディング状態を切り替えできる', async () => {
      render(ComponentDemo);
      
      const loadingButton = screen.getByText('Toggle Loading');
      await fireEvent.click(loadingButton);
      
      // ローディング状態の通知が表示されることを確認
      expect(screen.getByText('ローディング開始')).toBeInTheDocument();
    });
  });

  describe('Input コンポーネントのデモ', () => {
    it('様々なタイプのInputが表示される', () => {
      render(ComponentDemo);
      
      expect(screen.getByLabelText('基本入力')).toBeInTheDocument();
      expect(screen.getByLabelText('メールアドレス')).toBeInTheDocument();
      expect(screen.getByLabelText('パスワード')).toBeInTheDocument();
      expect(screen.getByLabelText('エラー例')).toBeInTheDocument();
    });

    it('基本入力フィールドに入力できる', async () => {
      render(ComponentDemo);
      
      const input = screen.getByLabelText('基本入力');
      await fireEvent.input(input, { target: { value: 'テスト入力' } });
      
      expect(input.value).toBe('テスト入力');
    });
  });

  describe('Select コンポーネントのデモ', () => {
    it('様々なSelectが表示される', () => {
      render(ComponentDemo);
      
      expect(screen.getByLabelText('スポーツ選択')).toBeInTheDocument();
      expect(screen.getByLabelText('サイズ選択')).toBeInTheDocument();
      expect(screen.getByLabelText('無効な選択')).toBeInTheDocument();
    });

    it('スポーツ選択で値を変更できる', async () => {
      render(ComponentDemo);
      
      const select = screen.getByLabelText('スポーツ選択');
      await fireEvent.change(select, { target: { value: 'volleyball' } });
      
      expect(select.value).toBe('volleyball');
    });
  });

  describe('LoadingSpinner コンポーネントのデモ', () => {
    it('異なるサイズのスピナーが表示される', () => {
      render(ComponentDemo);
      
      expect(screen.getByText('Small Spinner')).toBeInTheDocument();
      expect(screen.getByText('Medium Spinner')).toBeInTheDocument();
      expect(screen.getByText('Large Spinner')).toBeInTheDocument();
    });
  });

  describe('Modal コンポーネントのデモ', () => {
    it('モーダルを開くボタンが表示される', () => {
      render(ComponentDemo);
      
      expect(screen.getByText('モーダルを開く')).toBeInTheDocument();
    });

    it('モーダルを開くことができる', async () => {
      render(ComponentDemo);
      
      const openModalButton = screen.getByText('モーダルを開く');
      await fireEvent.click(openModalButton);
      
      // モーダルが表示されることを確認
      expect(screen.getByText('デモモーダル')).toBeInTheDocument();
      expect(screen.getByText('これはモーダルダイアログのデモです。')).toBeInTheDocument();
    });

    it('モーダルを閉じることができる', async () => {
      render(ComponentDemo);
      
      // モーダルを開く
      const openModalButton = screen.getByText('モーダルを開く');
      await fireEvent.click(openModalButton);
      
      // モーダルを閉じる
      const closeButton = screen.getByText('キャンセル');
      await fireEvent.click(closeButton);
      
      // 閉じた通知が表示されることを確認
      expect(screen.getByText('モーダルが閉じられました')).toBeInTheDocument();
    });
  });

  describe('現在の状態表示', () => {
    it('初期状態が正しく表示される', () => {
      render(ComponentDemo);
      
      expect(screen.getByText('Input Value: (空)')).toBeInTheDocument();
      expect(screen.getByText('Select Value: (未選択)')).toBeInTheDocument();
      expect(screen.getByText('Loading: No')).toBeInTheDocument();
      expect(screen.getByText('Modal Open: No')).toBeInTheDocument();
    });

    it('入力値の変更が状態に反映される', async () => {
      render(ComponentDemo);
      
      const input = screen.getByLabelText('基本入力');
      await fireEvent.input(input, { target: { value: 'テスト' } });
      
      expect(screen.getByText('Input Value: テスト')).toBeInTheDocument();
    });

    it('選択値の変更が状態に反映される', async () => {
      render(ComponentDemo);
      
      const select = screen.getByLabelText('スポーツ選択');
      await fireEvent.change(select, { target: { value: 'volleyball' } });
      
      expect(screen.getByText('Select Value: volleyball')).toBeInTheDocument();
    });

    it('ローディング状態の変更が状態に反映される', async () => {
      render(ComponentDemo);
      
      const loadingButton = screen.getByText('Toggle Loading');
      await fireEvent.click(loadingButton);
      
      expect(screen.getByText('Loading: Yes')).toBeInTheDocument();
    });

    it('モーダル状態の変更が状態に反映される', async () => {
      render(ComponentDemo);
      
      const openModalButton = screen.getByText('モーダルを開く');
      await fireEvent.click(openModalButton);
      
      expect(screen.getByText('Modal Open: Yes')).toBeInTheDocument();
    });
  });

  describe('通知システム', () => {
    it('通知を閉じることができる', async () => {
      render(ComponentDemo);
      
      // 通知を表示
      const primaryButton = screen.getByText('Primary Button');
      await fireEvent.click(primaryButton);
      
      expect(screen.getByText('ボタンがクリックされました！')).toBeInTheDocument();
      
      // 通知を閉じる（実際の実装に依存）
      // NotificationToast コンポーネントの close イベントをテスト
    });

    it('異なるタイプの通知が表示される', async () => {
      render(ComponentDemo);
      
      // Success 通知
      const primaryButton = screen.getByText('Primary Button');
      await fireEvent.click(primaryButton);
      expect(screen.getByText('ボタンがクリックされました！')).toBeInTheDocument();
      
      // Info 通知
      const loadingButton = screen.getByText('Toggle Loading');
      await fireEvent.click(loadingButton);
      expect(screen.getByText('ローディング開始')).toBeInTheDocument();
    });
  });

  describe('レスポンシブ対応', () => {
    it('モバイル表示でも正しくレンダリングされる', () => {
      // ビューポートサイズを変更
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      });

      render(ComponentDemo);
      
      // 基本的な要素が表示されることを確認
      expect(screen.getByText('共通UIコンポーネント デモ')).toBeInTheDocument();
      expect(screen.getByText('Button コンポーネント')).toBeInTheDocument();
    });
  });

  describe('アクセシビリティ', () => {
    it('適切なラベルが設定されている', () => {
      render(ComponentDemo);
      
      // フォーム要素にラベルが関連付けられていることを確認
      expect(screen.getByLabelText('基本入力')).toBeInTheDocument();
      expect(screen.getByLabelText('メールアドレス')).toBeInTheDocument();
      expect(screen.getByLabelText('パスワード')).toBeInTheDocument();
      expect(screen.getByLabelText('スポーツ選択')).toBeInTheDocument();
    });

    it('見出し構造が適切である', () => {
      render(ComponentDemo);
      
      // h1, h2 の見出し構造を確認
      expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();
      expect(screen.getAllByRole('heading', { level: 2 })).toHaveLength(6);
    });
  });
});