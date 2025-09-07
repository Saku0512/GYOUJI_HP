<script>
  // レスポンシブグリッドコンポーネント
  export let cols = {
    mobile: 1,
    tablet: 2,
    desktop: 3,
    large: 4
  };
  
  export let gap = '1rem';
  export let alignItems = 'stretch';
  export let justifyItems = 'stretch';
  export let className = '';
  export let autoFit = false;
  export let minItemWidth = '250px';

  // CSS Grid プロパティの計算
  $: gridTemplateColumns = autoFit 
    ? `repeat(auto-fit, minmax(${minItemWidth}, 1fr))`
    : 'var(--grid-columns)';

  $: gridClass = [
    'responsive-grid',
    autoFit ? 'auto-fit' : '',
    className
  ].filter(Boolean).join(' ');

  $: gridStyle = `
    --grid-gap: ${gap};
    --grid-align-items: ${alignItems};
    --grid-justify-items: ${justifyItems};
    --grid-columns: repeat(var(--cols-mobile), 1fr);
    --cols-mobile: ${cols.mobile || 1};
    --cols-tablet: ${cols.tablet || cols.mobile || 2};
    --cols-desktop: ${cols.desktop || cols.tablet || cols.mobile || 3};
    --cols-large: ${cols.large || cols.desktop || cols.tablet || cols.mobile || 4};
    grid-template-columns: ${gridTemplateColumns};
  `;
</script>

<div 
  class={gridClass}
  style={gridStyle}
>
  <slot />
</div>

<style>
  .responsive-grid {
    display: grid;
    gap: var(--grid-gap);
    align-items: var(--grid-align-items);
    justify-items: var(--grid-justify-items);
    width: 100%;
  }

  .responsive-grid.auto-fit {
    grid-template-columns: repeat(auto-fit, minmax(var(--min-item-width, 250px), 1fr));
  }

  /* モバイル (< 768px) */
  @media (max-width: 767px) {
    .responsive-grid:not(.auto-fit) {
      --grid-columns: repeat(var(--cols-mobile), 1fr);
    }
  }

  /* タブレット (768px - 1023px) */
  @media (min-width: 768px) and (max-width: 1023px) {
    .responsive-grid:not(.auto-fit) {
      --grid-columns: repeat(var(--cols-tablet), 1fr);
    }
  }

  /* デスクトップ (1024px - 1199px) */
  @media (min-width: 1024px) and (max-width: 1199px) {
    .responsive-grid:not(.auto-fit) {
      --grid-columns: repeat(var(--cols-desktop), 1fr);
    }
  }

  /* 大画面 (>= 1200px) */
  @media (min-width: 1200px) {
    .responsive-grid:not(.auto-fit) {
      --grid-columns: repeat(var(--cols-large), 1fr);
    }
  }

  /* プリント用 */
  @media print {
    .responsive-grid {
      grid-template-columns: repeat(2, 1fr) !important;
      gap: 0.5rem;
    }
  }
</style>