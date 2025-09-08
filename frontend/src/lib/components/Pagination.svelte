<script>
	import { createEventDispatcher } from 'svelte';

	// Props
	export let currentPage = 1;
	export let totalPages = 1;
	export let totalItems = 0;
	export let pageSize = 20;
	export let showInfo = true;
	export let showSizeSelector = true;
	export let pageSizeOptions = [10, 20, 50, 100];
	export let maxVisiblePages = 5;
	export let disabled = false;

	const dispatch = createEventDispatcher();

	// 計算されたプロパティ
	$: hasNext = currentPage < totalPages;
	$: hasPrev = currentPage > 1;
	$: startItem = totalItems === 0 ? 0 : (currentPage - 1) * pageSize + 1;
	$: endItem = Math.min(currentPage * pageSize, totalItems);

	// 表示するページ番号の配列を計算
	$: visiblePages = calculateVisiblePages(currentPage, totalPages, maxVisiblePages);

	function calculateVisiblePages(current, total, maxVisible) {
		if (total <= maxVisible) {
			return Array.from({ length: total }, (_, i) => i + 1);
		}

		const half = Math.floor(maxVisible / 2);
		let start = Math.max(1, current - half);
		let end = Math.min(total, start + maxVisible - 1);

		// 終端に近い場合の調整
		if (end - start + 1 < maxVisible) {
			start = Math.max(1, end - maxVisible + 1);
		}

		return Array.from({ length: end - start + 1 }, (_, i) => start + i);
	}

	function goToPage(page) {
		if (disabled || page < 1 || page > totalPages || page === currentPage) {
			return;
		}
		dispatch('pageChange', { page, pageSize });
	}

	function changePageSize(newSize) {
		if (disabled || newSize === pageSize) {
			return;
		}
		// ページサイズ変更時は1ページ目に戻る
		dispatch('pageChange', { page: 1, pageSize: newSize });
	}

	function goToFirst() {
		goToPage(1);
	}

	function goToLast() {
		goToPage(totalPages);
	}

	function goToPrev() {
		goToPage(currentPage - 1);
	}

	function goToNext() {
		goToPage(currentPage + 1);
	}
</script>

<div class="pagination-container" class:disabled>
	<!-- ページ情報表示 -->
	{#if showInfo && totalItems > 0}
		<div class="pagination-info">
			<span class="info-text">
				{startItem}〜{endItem}件 / 全{totalItems}件
			</span>
		</div>
	{/if}

	<!-- ページサイズ選択 -->
	{#if showSizeSelector && totalItems > 0}
		<div class="page-size-selector">
			<label for="page-size">表示件数:</label>
			<select
				id="page-size"
				bind:value={pageSize}
				on:change={(e) => changePageSize(parseInt(e.target.value))}
				{disabled}
			>
				{#each pageSizeOptions as option}
					<option value={option}>{option}件</option>
				{/each}
			</select>
		</div>
	{/if}

	<!-- ページネーション -->
	{#if totalPages > 1}
		<nav class="pagination-nav" aria-label="ページネーション">
			<ul class="pagination-list">
				<!-- 最初のページ -->
				<li class="pagination-item">
					<button
						class="pagination-button first"
						on:click={goToFirst}
						disabled={disabled || !hasPrev}
						aria-label="最初のページ"
						title="最初のページ"
					>
						<span class="pagination-icon">⟪</span>
					</button>
				</li>

				<!-- 前のページ -->
				<li class="pagination-item">
					<button
						class="pagination-button prev"
						on:click={goToPrev}
						disabled={disabled || !hasPrev}
						aria-label="前のページ"
						title="前のページ"
					>
						<span class="pagination-icon">‹</span>
					</button>
				</li>

				<!-- ページ番号 -->
				{#each visiblePages as page}
					<li class="pagination-item">
						<button
							class="pagination-button page"
							class:active={page === currentPage}
							on:click={() => goToPage(page)}
							{disabled}
							aria-label="ページ {page}"
							aria-current={page === currentPage ? 'page' : undefined}
						>
							{page}
						</button>
					</li>
				{/each}

				<!-- 次のページ -->
				<li class="pagination-item">
					<button
						class="pagination-button next"
						on:click={goToNext}
						disabled={disabled || !hasNext}
						aria-label="次のページ"
						title="次のページ"
					>
						<span class="pagination-icon">›</span>
					</button>
				</li>

				<!-- 最後のページ -->
				<li class="pagination-item">
					<button
						class="pagination-button last"
						on:click={goToLast}
						disabled={disabled || !hasNext}
						aria-label="最後のページ"
						title="最後のページ"
					>
						<span class="pagination-icon">⟫</span>
					</button>
				</li>
			</ul>
		</nav>
	{/if}
</div>

<style>
	.pagination-container {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 1rem 0;
		flex-wrap: wrap;
	}

	.pagination-container.disabled {
		opacity: 0.6;
		pointer-events: none;
	}

	.pagination-info {
		color: var(--text-secondary, #666);
		font-size: 0.875rem;
	}

	.page-size-selector {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
	}

	.page-size-selector label {
		color: var(--text-secondary, #666);
	}

	.page-size-selector select {
		padding: 0.25rem 0.5rem;
		border: 1px solid var(--border-color, #ddd);
		border-radius: 4px;
		background: var(--bg-primary, white);
		color: var(--text-primary, #333);
		font-size: 0.875rem;
	}

	.page-size-selector select:focus {
		outline: none;
		border-color: var(--primary-color, #007bff);
		box-shadow: 0 0 0 2px var(--primary-color-alpha, rgba(0, 123, 255, 0.25));
	}

	.pagination-nav {
		margin-left: auto;
	}

	.pagination-list {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		list-style: none;
		margin: 0;
		padding: 0;
	}

	.pagination-item {
		margin: 0;
	}

	.pagination-button {
		display: flex;
		align-items: center;
		justify-content: center;
		min-width: 2.5rem;
		height: 2.5rem;
		padding: 0.5rem;
		border: 1px solid var(--border-color, #ddd);
		background: var(--bg-primary, white);
		color: var(--text-primary, #333);
		font-size: 0.875rem;
		font-weight: 500;
		text-decoration: none;
		border-radius: 4px;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.pagination-button:hover:not(:disabled) {
		background: var(--bg-hover, #f8f9fa);
		border-color: var(--primary-color, #007bff);
		color: var(--primary-color, #007bff);
	}

	.pagination-button:focus {
		outline: none;
		box-shadow: 0 0 0 2px var(--primary-color-alpha, rgba(0, 123, 255, 0.25));
	}

	.pagination-button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
		background: var(--bg-disabled, #f8f9fa);
		color: var(--text-disabled, #999);
	}

	.pagination-button.active {
		background: var(--primary-color, #007bff);
		border-color: var(--primary-color, #007bff);
		color: white;
	}

	.pagination-button.active:hover {
		background: var(--primary-color-dark, #0056b3);
		border-color: var(--primary-color-dark, #0056b3);
		color: white;
	}

	.pagination-icon {
		font-size: 1rem;
		line-height: 1;
	}

	/* レスポンシブ対応 */
	@media (max-width: 768px) {
		.pagination-container {
			flex-direction: column;
			align-items: stretch;
			gap: 0.75rem;
		}

		.pagination-info,
		.page-size-selector {
			justify-content: center;
		}

		.pagination-nav {
			margin-left: 0;
		}

		.pagination-list {
			justify-content: center;
			flex-wrap: wrap;
		}

		.pagination-button {
			min-width: 2.25rem;
			height: 2.25rem;
			font-size: 0.8rem;
		}
	}

	@media (max-width: 480px) {
		.pagination-button {
			min-width: 2rem;
			height: 2rem;
			padding: 0.25rem;
			font-size: 0.75rem;
		}

		.pagination-icon {
			font-size: 0.875rem;
		}
	}
</style>