<script>
  import { onMount } from 'svelte';
  import { writable } from 'svelte/store';

  // スタガードアニメーション付きリストコンポーネント
  export let items = [];
  export let staggerDelay = 100; // ミリ秒
  export let animationType = 'fadeInUp';
  export let className = '';
  export let itemClassName = '';
  export let tag = 'div';
  export let itemTag = 'div';

  const visibleItems = writable([]);
  let mounted = false;

  // アイテムを順次表示
  function showItemsWithStagger() {
    visibleItems.set([]);
    
    items.forEach((item, index) => {
      setTimeout(() => {
        visibleItems.update(visible => [...visible, item]);
      }, index * staggerDelay);
    });
  }

  // アイテムが変更されたときに再アニメーション
  $: if (mounted && items.length > 0) {
    showItemsWithStagger();
  }

  onMount(() => {
    mounted = true;
    showItemsWithStagger();
  });

  // アニメーションクラスの取得
  function getAnimationClass(index) {
    return `animate-${animationType} stagger-item-${index}`;
  }
</script>

<svelte:element this={tag} class="staggered-list {className}">
  {#each $visibleItems as item, index (item.id || index)}
    <svelte:element 
      this={itemTag} 
      class="stagger-item {itemClassName} {getAnimationClass(index)}"
      style="animation-delay: {index * (staggerDelay / 1000)}s"
    >
      <slot {item} {index} />
    </svelte:element>
  {/each}
</svelte:element>

<style>
  .staggered-list {
    display: contents;
  }

  .stagger-item {
    opacity: 0;
    animation-fill-mode: forwards;
  }

  /* アニメーション定義 */
  :global(.animate-fadeInUp) {
    animation-name: fadeInUp;
    animation-duration: 0.6s;
    animation-timing-function: cubic-bezier(0, 0, 0.2, 1);
  }

  :global(.animate-fadeInLeft) {
    animation-name: fadeInLeft;
    animation-duration: 0.6s;
    animation-timing-function: cubic-bezier(0, 0, 0.2, 1);
  }

  :global(.animate-fadeInRight) {
    animation-name: fadeInRight;
    animation-duration: 0.6s;
    animation-timing-function: cubic-bezier(0, 0, 0.2, 1);
  }

  :global(.animate-scaleIn) {
    animation-name: scaleIn;
    animation-duration: 0.6s;
    animation-timing-function: cubic-bezier(0, 0, 0.2, 1);
  }

  @keyframes fadeInUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  @keyframes fadeInLeft {
    from {
      opacity: 0;
      transform: translateX(-20px);
    }
    to {
      opacity: 1;
      transform: translateX(0);
    }
  }

  @keyframes fadeInRight {
    from {
      opacity: 0;
      transform: translateX(20px);
    }
    to {
      opacity: 1;
      transform: translateX(0);
    }
  }

  @keyframes scaleIn {
    from {
      opacity: 0;
      transform: scale(0.9);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }

  /* アクセシビリティ: アニメーション無効化 */
  @media (prefers-reduced-motion: reduce) {
    .stagger-item {
      opacity: 1 !important;
      animation: none !important;
      transform: none !important;
    }
  }
</style>