<script lang="ts">
	import { onMount } from 'svelte';
	import { offlineSync } from '../offline-sync';
	import { Button } from '$lib/components/ui/button';
	import { RefreshCw, Wifi, WifiOff } from 'lucide-svelte';

	interface SyncStatus {
		isOnline: boolean;
		pendingActions: number;
		isSyncing: boolean;
		lastSyncTime?: number;
	}

	let syncStatus = $state<SyncStatus>({
		isOnline: navigator.onLine,
		pendingActions: 0,
		isSyncing: false
	});

	onMount(async () => {
		// Get initial status
		syncStatus = await offlineSync.getStatus();
		
		// Listen for status changes
		offlineSync.onStatusChange((status) => {
			syncStatus = status;
		});
	});

	async function manualSync() {
		if (syncStatus.isOnline && syncStatus.pendingActions > 0 && !syncStatus.isSyncing) {
			await offlineSync.syncPendingActions();
		}
	}

	const statusColor = $derived(syncStatus.isOnline ? 'text-green-500' : 'text-red-500');
	const showPendingBadge = $derived(syncStatus.pendingActions > 0);
	const isSpinning = $derived(syncStatus.isSyncing);
</script>

<div class="flex items-center gap-2">
	<!-- Network Status Icon -->
	<div class="flex items-center gap-1">
		{#if syncStatus.isOnline}
			<Wifi class="h-4 w-4 {statusColor}" />
		{:else}
			<WifiOff class="h-4 w-4 {statusColor}" />
		{/if}
	</div>

	<!-- Pending Actions Badge -->
	{#if showPendingBadge}
		<div class="flex items-center gap-1">
			<span class="text-xs bg-yellow-500 text-white px-2 py-1 rounded-full font-medium">
				{syncStatus.pendingActions}
			</span>
		</div>
	{/if}

	<!-- Sync Button -->
	{#if syncStatus.isOnline && syncStatus.pendingActions > 0}
		<Button
			variant="ghost"
			size="sm"
			onclick={manualSync}
			disabled={syncStatus.isSyncing}
			class="h-6 w-6 p-0"
		>
			<RefreshCw class="h-3 w-3 {isSpinning ? 'animate-spin' : ''}" />
		</Button>
	{/if}
</div> 