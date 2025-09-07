#!/usr/bin/env node

/**
 * パフォーマンステストスクリプト
 * バンドルサイズ分析とパフォーマンステストを実行
 */

import { execSync } from 'child_process';
import { readFileSync, existsSync, writeFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const projectRoot = join(__dirname, '..');

// カラー出力用のヘルパー
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m'
};

function colorLog(color, message) {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

function formatBytes(bytes) {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

async function runCommand(command, description) {
  colorLog('blue', `\n🔄 ${description}...`);
  try {
    const output = execSync(command, { 
      cwd: projectRoot, 
      encoding: 'utf8',
      stdio: 'pipe'
    });
    colorLog('green', `✅ ${description} 完了`);
    return output;
  } catch (error) {
    colorLog('red', `❌ ${description} 失敗: ${error.message}`);
    throw error;
  }
}

async function analyzeBundleSize() {
  colorLog('cyan', '\n📦 バンドルサイズ分析を開始...');
  
  // 本番ビルドを実行
  await runCommand('npm run build', '本番ビルド');
  
  // ビルド結果のディレクトリを確認
  const buildDir = join(projectRoot, '.svelte-kit/output/client');
  if (!existsSync(buildDir)) {
    throw new Error('ビルドディレクトリが見つかりません');
  }

  // バンドルファイルのサイズを分析
  const clientDir = join(buildDir, '_app', 'immutable');
  const chunks = [];
  
  try {
    const chunksDir = join(clientDir, 'chunks');
    if (existsSync(chunksDir)) {
      const chunkFiles = execSync(`find "${chunksDir}" -name "*.js" -type f`, { encoding: 'utf8' })
        .split('\n')
        .filter(Boolean);
      
      for (const file of chunkFiles) {
        const stats = execSync(`stat -c%s "${file}"`, { encoding: 'utf8' }).trim();
        const size = parseInt(stats);
        const name = file.split('/').pop();
        chunks.push({ name, size, path: file });
      }
    }
  } catch (error) {
    colorLog('yellow', '⚠️ チャンクファイルの分析をスキップしました');
  }

  // 結果を表示
  colorLog('bright', '\n📊 バンドルサイズ分析結果:');
  
  if (chunks.length > 0) {
    chunks.sort((a, b) => b.size - a.size);
    
    const totalSize = chunks.reduce((sum, chunk) => sum + chunk.size, 0);
    colorLog('cyan', `\n合計サイズ: ${formatBytes(totalSize)}`);
    
    colorLog('yellow', '\n📁 チャンク別サイズ (上位10件):');
    chunks.slice(0, 10).forEach((chunk, index) => {
      const percentage = ((chunk.size / totalSize) * 100).toFixed(1);
      console.log(`  ${index + 1}. ${chunk.name}: ${formatBytes(chunk.size)} (${percentage}%)`);
    });

    // サイズ警告
    const largeChunks = chunks.filter(chunk => chunk.size > 500 * 1024); // 500KB以上
    if (largeChunks.length > 0) {
      colorLog('yellow', '\n⚠️ 大きなチャンク (500KB以上):');
      largeChunks.forEach(chunk => {
        console.log(`  - ${chunk.name}: ${formatBytes(chunk.size)}`);
      });
    }

    // 推奨事項
    if (totalSize > 2 * 1024 * 1024) { // 2MB以上
      colorLog('yellow', '\n💡 推奨事項:');
      console.log('  - コード分割をさらに細かく行うことを検討してください');
      console.log('  - 未使用のライブラリやコードを削除してください');
      console.log('  - 動的インポートを活用してください');
    }
  }

  return { chunks, totalSize: chunks.reduce((sum, chunk) => sum + chunk.size, 0) };
}

async function runLighthouseTest() {
  colorLog('cyan', '\n🚀 Lighthouse パフォーマンステストを開始...');
  
  try {
    // Lighthouseがインストールされているかチェック
    execSync('which lighthouse', { stdio: 'pipe' });
  } catch (error) {
    colorLog('yellow', '⚠️ Lighthouse がインストールされていません。npm install -g lighthouse でインストールしてください。');
    return null;
  }

  // 開発サーバーを起動（バックグラウンド）
  colorLog('blue', '🔄 開発サーバーを起動中...');
  
  const serverProcess = execSync('npm run preview > /dev/null 2>&1 & echo $!', { 
    encoding: 'utf8',
    cwd: projectRoot 
  }).trim();

  // サーバーの起動を待つ
  await new Promise(resolve => setTimeout(resolve, 5000));

  try {
    // Lighthouseテストを実行
    const lighthouseOutput = await runCommand(
      'lighthouse http://localhost:4173 --output=json --output-path=./lighthouse-report.json --chrome-flags="--headless --no-sandbox"',
      'Lighthouse パフォーマンステスト'
    );

    // レポートを読み込み
    const reportPath = join(projectRoot, 'lighthouse-report.json');
    if (existsSync(reportPath)) {
      const report = JSON.parse(readFileSync(reportPath, 'utf8'));
      
      colorLog('bright', '\n📊 Lighthouse スコア:');
      const categories = report.lhr.categories;
      
      Object.entries(categories).forEach(([key, category]) => {
        const score = Math.round(category.score * 100);
        const color = score >= 90 ? 'green' : score >= 50 ? 'yellow' : 'red';
        colorLog(color, `  ${category.title}: ${score}/100`);
      });

      // Core Web Vitals
      const audits = report.lhr.audits;
      colorLog('bright', '\n🎯 Core Web Vitals:');
      
      if (audits['largest-contentful-paint']) {
        const lcp = audits['largest-contentful-paint'].numericValue;
        console.log(`  LCP (Largest Contentful Paint): ${(lcp / 1000).toFixed(2)}s`);
      }
      
      if (audits['first-input-delay']) {
        const fid = audits['first-input-delay'].numericValue;
        console.log(`  FID (First Input Delay): ${fid.toFixed(2)}ms`);
      }
      
      if (audits['cumulative-layout-shift']) {
        const cls = audits['cumulative-layout-shift'].numericValue;
        console.log(`  CLS (Cumulative Layout Shift): ${cls.toFixed(3)}`);
      }

      return report;
    }
  } finally {
    // サーバーを停止
    try {
      execSync(`kill ${serverProcess}`, { stdio: 'pipe' });
    } catch (error) {
      // プロセスが既に終了している場合は無視
    }
  }

  return null;
}

async function analyzeImageOptimization() {
  colorLog('cyan', '\n🖼️ 画像最適化分析を開始...');
  
  const buildDir = join(projectRoot, '.svelte-kit/output/client');
  const imageStats = {
    totalImages: 0,
    totalSize: 0,
    webpImages: 0,
    largeImages: 0,
    unoptimizedImages: []
  };

  try {
    // 画像ファイルを検索
    const imageFiles = execSync(`find "${buildDir}" -type f \\( -name "*.jpg" -o -name "*.jpeg" -o -name "*.png" -o -name "*.gif" -o -name "*.webp" -o -name "*.svg" \\)`, { encoding: 'utf8' })
      .split('\n')
      .filter(Boolean);

    for (const file of imageFiles) {
      const stats = execSync(`stat -c%s "${file}"`, { encoding: 'utf8' }).trim();
      const size = parseInt(stats);
      const name = file.split('/').pop();
      
      imageStats.totalImages++;
      imageStats.totalSize += size;
      
      if (name.includes('.webp')) {
        imageStats.webpImages++;
      }
      
      if (size > 100 * 1024) { // 100KB以上
        imageStats.largeImages++;
        imageStats.unoptimizedImages.push({
          name,
          size,
          sizeFormatted: formatBytes(size)
        });
      }
    }
    
    colorLog('bright', '\n📊 画像最適化分析結果:');
    colorLog('cyan', `総画像数: ${imageStats.totalImages}`);
    colorLog('cyan', `総サイズ: ${formatBytes(imageStats.totalSize)}`);
    colorLog('cyan', `WebP画像: ${imageStats.webpImages}/${imageStats.totalImages}`);
    colorLog('cyan', `大きな画像 (100KB以上): ${imageStats.largeImages}`);
    
    if (imageStats.unoptimizedImages.length > 0) {
      colorLog('yellow', '\n⚠️ 最適化が推奨される画像:');
      imageStats.unoptimizedImages.slice(0, 5).forEach(img => {
        console.log(`  - ${img.name}: ${img.sizeFormatted}`);
      });
    }
    
  } catch (error) {
    colorLog('yellow', '⚠️ 画像分析をスキップしました');
  }

  return imageStats;
}

async function generatePerformanceReport(bundleAnalysis, lighthouseReport, imageAnalysis) {
  colorLog('cyan', '\n📝 パフォーマンスレポートを生成中...');
  
  const report = {
    timestamp: new Date().toISOString(),
    bundle: {
      totalSize: bundleAnalysis.totalSize,
      totalSizeFormatted: formatBytes(bundleAnalysis.totalSize),
      chunks: bundleAnalysis.chunks.length,
      largeChunks: bundleAnalysis.chunks.filter(chunk => chunk.size > 500 * 1024).length
    },
    images: imageAnalysis,
    lighthouse: null
  };

  if (lighthouseReport) {
    const categories = lighthouseReport.lhr.categories;
    const audits = lighthouseReport.lhr.audits;
    
    report.lighthouse = {
      performance: Math.round(categories.performance.score * 100),
      accessibility: Math.round(categories.accessibility.score * 100),
      bestPractices: Math.round(categories['best-practices'].score * 100),
      seo: Math.round(categories.seo.score * 100),
      lcp: audits['largest-contentful-paint']?.numericValue / 1000,
      fid: audits['first-input-delay']?.numericValue,
      cls: audits['cumulative-layout-shift']?.numericValue
    };
  }

  // レポートをファイルに保存
  const reportPath = join(projectRoot, 'performance-report.json');
  writeFileSync(reportPath, JSON.stringify(report, null, 2));
  
  colorLog('green', `✅ パフォーマンスレポートを保存しました: ${reportPath}`);
  
  // 推奨事項を表示
  colorLog('bright', '\n💡 パフォーマンス改善の推奨事項:');
  
  if (report.bundle.totalSize > 2 * 1024 * 1024) {
    console.log('  📦 バンドルサイズが大きいです (2MB以上)');
    console.log('    - コード分割を検討してください');
    console.log('    - 未使用のライブラリを削除してください');
  }
  
  if (report.bundle.largeChunks > 0) {
    console.log(`  📁 大きなチャンクが${report.bundle.largeChunks}個あります (500KB以上)`);
    console.log('    - 動的インポートを使用してください');
    console.log('    - チャンクをさらに細分化してください');
  }
  
  if (report.lighthouse) {
    if (report.lighthouse.performance < 90) {
      console.log('  🚀 パフォーマンススコアが低いです');
      console.log('    - 画像の最適化を検討してください');
      console.log('    - 不要なJavaScriptを削除してください');
    }
    
    if (report.lighthouse.lcp > 2.5) {
      console.log('  ⏱️ LCP (Largest Contentful Paint) が遅いです');
      console.log('    - 重要なリソースの優先読み込みを検討してください');
    }
  }

  return report;
}

async function main() {
  colorLog('bright', '🎯 フロントエンドパフォーマンステストを開始します\n');
  
  try {
    // バンドルサイズ分析
    const bundleAnalysis = await analyzeBundleSize();
    
    // 画像最適化分析
    const imageAnalysis = await analyzeImageOptimization();
    
    // Lighthouseテスト
    const lighthouseReport = await runLighthouseTest();
    
    // レポート生成
    await generatePerformanceReport(bundleAnalysis, lighthouseReport, imageAnalysis);
    
    colorLog('green', '\n🎉 パフォーマンステストが完了しました！');
    
  } catch (error) {
    colorLog('red', `\n❌ パフォーマンステストでエラーが発生しました: ${error.message}`);
    process.exit(1);
  }
}

// スクリプトが直接実行された場合のみmain関数を実行
if (import.meta.url === `file://${process.argv[1]}`) {
  main();
}