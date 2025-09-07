/**
 * 画像最適化ユーティリティのテスト
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { 
  checkWebPSupport, 
  generateSrcSet, 
  getOptimizedImageUrl,
  preloadImage,
  compressImage,
  getImageMetadata,
  convertImageFormat
} from './imageOptimization.js';

// モック設定
const mockImage = {
  onload: null,
  onerror: null,
  src: '',
  width: 800,
  height: 600,
  naturalWidth: 800,
  naturalHeight: 600
};

const mockCanvas = {
  width: 0,
  height: 0,
  getContext: vi.fn(() => ({
    drawImage: vi.fn()
  })),
  toBlob: vi.fn()
};

const mockFile = {
  size: 1024 * 100, // 100KB
  type: 'image/jpeg',
  name: 'test.jpg'
};

// グローバルモック
global.Image = vi.fn(() => mockImage);
global.document = {
  createElement: vi.fn((tag) => {
    if (tag === 'canvas') return mockCanvas;
    return {};
  })
};
global.URL = {
  createObjectURL: vi.fn(() => 'blob:mock-url')
};

describe('画像最適化ユーティリティ', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('checkWebPSupport', () => {
    it('WebP対応をチェックできる', async () => {
      // WebP対応の場合
      setTimeout(() => {
        mockImage.height = 2;
        mockImage.onload();
      }, 0);

      const supported = await checkWebPSupport();
      expect(supported).toBe(true);
    });

    it('WebP非対応の場合はfalseを返す', async () => {
      // WebP非対応の場合
      setTimeout(() => {
        mockImage.height = 0;
        mockImage.onerror();
      }, 0);

      const supported = await checkWebPSupport();
      expect(supported).toBe(false);
    });
  });

  describe('generateSrcSet', () => {
    it('レスポンシブ画像のsrcsetを生成できる', () => {
      const basePath = '/images/test.jpg';
      const sizes = [320, 640, 960];
      
      const srcset = generateSrcSet(basePath, sizes);
      
      expect(srcset).toBe('/images/test-320w.jpg 320w, /images/test-640w.jpg 640w, /images/test-960w.jpg 960w');
    });

    it('デフォルトサイズでsrcsetを生成できる', () => {
      const basePath = '/images/test.png';
      
      const srcset = generateSrcSet(basePath);
      
      expect(srcset).toContain('320w');
      expect(srcset).toContain('640w');
      expect(srcset).toContain('960w');
      expect(srcset).toContain('1280w');
      expect(srcset).toContain('1920w');
    });
  });

  describe('getOptimizedImageUrl', () => {
    it('最適化されたURLを生成できる', async () => {
      const originalUrl = '/images/test.jpg';
      const options = {
        width: 800,
        height: 600,
        quality: 80
      };

      const optimizedUrl = await getOptimizedImageUrl(originalUrl, options);
      
      expect(optimizedUrl).toContain('w=800');
      expect(optimizedUrl).toContain('h=600');
      expect(optimizedUrl).toContain('q=80');
    });

    it('WebP対応時はWebPフォーマットを指定する', async () => {
      // WebP対応をモック
      vi.mocked(checkWebPSupport).mockResolvedValue(true);
      
      const originalUrl = '/images/test.jpg';
      const optimizedUrl = await getOptimizedImageUrl(originalUrl);
      
      expect(optimizedUrl).toContain('f=webp');
    });
  });

  describe('preloadImage', () => {
    it('画像をプリロードできる', async () => {
      const src = '/images/test.jpg';
      
      setTimeout(() => {
        mockImage.onload();
      }, 0);

      const img = await preloadImage(src);
      expect(img).toBeDefined();
      expect(mockImage.src).toBe(src);
    });

    it('プリロード失敗時はエラーを投げる', async () => {
      const src = '/images/nonexistent.jpg';
      
      setTimeout(() => {
        mockImage.onerror(new Error('Failed to load'));
      }, 0);

      await expect(preloadImage(src)).rejects.toThrow();
    });

    it('crossOrigin設定を適用できる', async () => {
      const src = '/images/test.jpg';
      const options = { crossOrigin: 'anonymous' };
      
      setTimeout(() => {
        mockImage.onload();
      }, 0);

      await preloadImage(src, options);
      expect(mockImage.crossOrigin).toBe('anonymous');
    });
  });

  describe('compressImage', () => {
    it('画像を圧縮できる', async () => {
      const options = {
        maxWidth: 800,
        maxHeight: 600,
        quality: 0.8
      };

      mockCanvas.toBlob.mockImplementation((callback) => {
        callback(new Blob(['compressed'], { type: 'image/jpeg' }));
      });

      setTimeout(() => {
        mockImage.onload();
      }, 0);

      const compressedBlob = await compressImage(mockFile, options);
      
      expect(compressedBlob).toBeInstanceOf(Blob);
      expect(mockCanvas.width).toBe(800);
      expect(mockCanvas.height).toBe(600);
    });

    it('アスペクト比を保持して圧縮する', async () => {
      mockImage.width = 1600;
      mockImage.height = 900;
      
      const options = {
        maxWidth: 800,
        maxHeight: 600
      };

      mockCanvas.toBlob.mockImplementation((callback) => {
        callback(new Blob(['compressed'], { type: 'image/jpeg' }));
      });

      setTimeout(() => {
        mockImage.onload();
      }, 0);

      await compressImage(mockFile, options);
      
      // アスペクト比16:9を保持して800x450になるはず
      expect(mockCanvas.width).toBe(800);
      expect(mockCanvas.height).toBe(450);
    });
  });

  describe('getImageMetadata', () => {
    it('画像のメタデータを取得できる', async () => {
      setTimeout(() => {
        mockImage.onload();
      }, 0);

      const metadata = await getImageMetadata(mockFile);
      
      expect(metadata).toEqual({
        width: 800,
        height: 600,
        aspectRatio: 800 / 600,
        size: 1024 * 100,
        type: 'image/jpeg',
        name: 'test.jpg'
      });
    });

    it('メタデータ取得失敗時はエラーを投げる', async () => {
      setTimeout(() => {
        mockImage.onerror(new Error('Failed to load'));
      }, 0);

      await expect(getImageMetadata(mockFile)).rejects.toThrow();
    });
  });

  describe('convertImageFormat', () => {
    it('画像フォーマットを変換できる', async () => {
      const imageUrl = '/images/test.jpg';
      const targetFormat = 'webp';
      const quality = 0.8;

      mockCanvas.toBlob.mockImplementation((callback) => {
        callback(new Blob(['converted'], { type: 'image/webp' }));
      });

      setTimeout(() => {
        mockImage.onload();
      }, 0);

      const convertedBlob = await convertImageFormat(imageUrl, targetFormat, quality);
      
      expect(convertedBlob).toBeInstanceOf(Blob);
      expect(mockCanvas.toBlob).toHaveBeenCalledWith(
        expect.any(Function),
        'image/webp',
        0.8
      );
    });

    it('変換失敗時はエラーを投げる', async () => {
      const imageUrl = '/images/nonexistent.jpg';

      setTimeout(() => {
        mockImage.onerror(new Error('Failed to load'));
      }, 0);

      await expect(convertImageFormat(imageUrl)).rejects.toThrow();
    });
  });
});

describe('LazyImageLoader', () => {
  let mockIntersectionObserver;
  let mockObserve;
  let mockUnobserve;
  let mockDisconnect;

  beforeEach(() => {
    mockObserve = vi.fn();
    mockUnobserve = vi.fn();
    mockDisconnect = vi.fn();

    mockIntersectionObserver = vi.fn((callback, options) => ({
      observe: mockObserve,
      unobserve: mockUnobserve,
      disconnect: mockDisconnect
    }));

    global.IntersectionObserver = mockIntersectionObserver;
  });

  it('Intersection Observerが利用可能な場合は初期化される', () => {
    const { LazyImageLoader } = require('./imageOptimization.js');
    const loader = new LazyImageLoader();
    
    expect(mockIntersectionObserver).toHaveBeenCalled();
  });

  it('Intersection Observer非対応の場合は即座に読み込む', () => {
    global.IntersectionObserver = undefined;
    
    const { LazyImageLoader } = require('./imageOptimization.js');
    const loader = new LazyImageLoader();
    
    const mockImg = { dataset: { src: '/test.jpg' } };
    const loadImageSpy = vi.spyOn(loader, 'loadImage');
    
    loader.observe(mockImg);
    
    expect(loadImageSpy).toHaveBeenCalledWith(mockImg);
  });
});