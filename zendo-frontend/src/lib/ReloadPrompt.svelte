<script lang="ts">
	import { onMount } from 'svelte';
	
	let needRefresh = $state(false);
	let offlineReady = $state(false);
	
	onMount(() => {
		// Listen for service worker updates
		if ('serviceWorker' in navigator) {
			navigator.serviceWorker.addEventListener('controllerchange', () => {
				needRefresh = true;
			});
			
			navigator.serviceWorker.addEventListener('message', (event) => {
				if (event.data && event.data.type === 'SKIP_WAITING') {
					needRefresh = true;
				}
			});
		}
	});
	
	const close = () => {
		offlineReady = false;
		needRefresh = false;
	};
	
	const reload = () => {
		window.location.reload();
	};
	
	const toast = $derived(offlineReady || needRefresh);
</script>

{#if toast}
	<div class="pwa-toast" role="alert">
		<div class="message">
			{#if offlineReady}
				<span>
					App ready to work offline
				</span>
			{:else}
				<span>
					New content available, click on reload button to update.
				</span>
			{/if}
		</div>
		{#if needRefresh}
			<button onclick={reload}>
				Reload
			</button>
		{/if}
		<button onclick={close}>
			Close
		</button>
	</div>
{/if}

<style>
	.pwa-toast {
		position: fixed;
		right: 0;
		bottom: 0;
		margin: 16px;
		padding: 12px;
		border: 1px solid #8885;
		border-radius: 4px;
		z-index: 9999;
		text-align: left;
		box-shadow: 3px 4px 5px 0 #8885;
		background-color: white;
		color: #333;
	}
	.pwa-toast .message {
		margin-bottom: 8px;
	}
	.pwa-toast button {
		border: 1px solid #8885;
		outline: none;
		margin-right: 5px;
		border-radius: 2px;
		padding: 3px 10px;
		cursor: pointer;
	}
	.pwa-toast button:hover {
		background-color: #f0f0f0;
	}
</style> 