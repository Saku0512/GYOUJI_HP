<script>
  import { onMount, createEventDispatcher } from 'svelte';
  import { fade, fly, scale, slide } from 'svelte/transition';
  import { quintOut, elasticOut, bounceOut } from 'svelte/easing';

  // アニメーション付きトランジションコンポーネント
  export let show = true;
  export let type = 'fade'; // fade, fly, scale, slide, custom
  export let direction = 'up'; // up, down, left, right (flyとslideで使用)
  export let duration = 300;
  export let delay = 0;
  export let easing = quintOut;
  export let distance = 20; // flyアニメーションの距離
  export let start = 0.95; // scaleアニメーションの開始値
  export let className = '';
  export let tag = 'div';

  const dispatch = createEventDispatcher();

  // イージング関数のマッピング
  const easingMap = {
    quintOut,
    elasticOut,
    bounceOut
  };

  // 方向のマッピング（flyとslide用）
  const directionMap = {
    up: { y: distance },
    down: { y: -distance },
    left: { x: distance },
    right: { x: -distance }
  };

  // トランジション設定の計算
  $: transitionConfig = {
    duration,
    delay,
    easing: typeof easing === 'string' ? easingMap[easing] || quintOut : easing
  };

  // flyトランジション用の設定
  $: flyConfig = {
    ...transitionConfig,
    ...directionMap[direction]
  };

  // scaleトランジション用の設定
  $: scaleConfig = {
    ...transitionConfig,
    start
  };

  // slideトランジション用の設定
  $: slideConfig = {
    ...transitionConfig,
    axis: direction === 'left' || direction === 'right' ? 'x' : 'y'
  };

  // トランジション関数の選択
  function getTransition(node) {
    switch (type) {
      case 'fly':
        return fly(node, flyConfig);
      case 'scale':
        return scale(node, scaleConfig);
      case 'slide':
        return slide(node, slideConfig);
      case 'fade':
      default:
        return fade(node, transitionConfig);
    }
  }

  // アニメーション開始時のイベント
  function handleIntroStart() {
    dispatch('introstart');
  }

  // アニメーション終了時のイベント
  function handleIntroEnd() {
    dispatch('introend');
  }

  // アニメーション開始時のイベント（退場）
  function handleOutroStart() {
    dispatch('outrostart');
  }

  // アニメーション終了時のイベント（退場）
  function handleOutroEnd() {
    dispatch('outroend');
  }
</script>

{#if show}
  <svelte:element 
    this={tag}
    class="animated-transition {className}"
    transition:getTransition
    on:introstart={handleIntroStart}
    on:introend={handleIntroEnd}
    on:outrostart={handleOutroStart}
    on:outroend={handleOutroEnd}
  >
    <slot />
  </svelte:element>
{/if}

<style>
  .animated-transition {
    display: contents;
  }
</style>