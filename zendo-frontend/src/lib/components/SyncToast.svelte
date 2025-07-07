<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { offlineSync } from '../offline-sync';
	import { toast } from 'svelte-sonner';

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

	let lastSyncResult = $state<{ success: number; failed: number } | null>(null);

	onMount(async () => {
		// Get initial status
		syncStatus = await offlineSync.getStatus();
		
		// Listen for status changes
		offlineSync.onStatusChange(async (status) => {
			const previousStatus = { ...syncStatus };
			syncStatus = status;
			
			// Show appropriate notifications
			await handleStatusChange(previousStatus, status);
		});

		// Initial sync check if online
		if (syncStatus.isOnline && syncStatus.pendingActions > 0) {
			await triggerSync();
		}
	});

	async function handleStatusChange(previous: SyncStatus, current: SyncStatus) {
		// Network status change
		if (previous.isOnline !== current.isOnline) {
			if (current.isOnline) {
				toast.success('Connection restored', {
					description: 'Syncing offline changes...'
				});
				
				// Trigger sync when coming back online
				if (current.pendingActions > 0) {
					await triggerSync();
				}
			} else {
				toast.warning('Connection lost', {
					description: 'Changes will be saved locally and synced when connection returns'
				});
			}
		}

		// Pending actions change
		if (previous.pendingActions !== current.pendingActions) {
			if (current.pendingActions > 0 && !current.isOnline) {
				toast.info('Offline mode', {
					description: `${current.pendingActions} action${current.pendingActions > 1 ? 's' : ''} queued for sync`
				});
			}
		}

		// Sync completion
		if (previous.isSyncing && !current.isSyncing && lastSyncResult) {
			const { success, failed } = lastSyncResult;
			
			if (success > 0 && failed === 0) {
				toast.success('Sync completed', {
					description: `Successfully synced ${success} change${success > 1 ? 's' : ''}`
				});
			} else if (success > 0 && failed > 0) {
				toast.warning('Partial sync', {
					description: `Synced ${success} changes, ${failed} failed`
				});
			} else if (failed > 0) {
				toast.error('Sync failed', {
					description: `Failed to sync ${failed} change${failed > 1 ? 's' : ''}`
				});
			}
			
			lastSyncResult = null;
		}
	}

	async function triggerSync() {
		if (syncStatus.isOnline && syncStatus.pendingActions > 0 && !syncStatus.isSyncing) {
			toast.info('Syncing...', {
				description: 'Syncing offline changes with server'
			});
			
			lastSyncResult = await offlineSync.syncPendingActions();
		}
	}

	// Manual sync function that can be called from parent components
	export async function manualSync() {
		await triggerSync();
	}
</script>

<!-- This component doesn't render anything visible, it just manages toast notifications --> 