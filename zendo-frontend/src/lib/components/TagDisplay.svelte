<script lang="ts">
	import { stringToTags } from '$lib/utils';

	const { tags = '', completed = false } = $props<{
		tags?: string;
		completed?: boolean;
	}>();

	const tagList = $derived(stringToTags(tags));
</script>

{#if tagList.length > 0}
	<span class="inline-flex flex-wrap gap-1 ml-2">
		{#each tagList as tag}
			<span 
				class="inline-flex items-center px-2 py-0.5 rounded-sm text-xs font-medium transition-colors
					{completed 
						? 'bg-gray-200/20 text-gray-400' 
						: 'bg-magnum-400 text-black hover:bg-magnum-500'
					}"
			>
				#{tag}
			</span>
		{/each}
	</span>
{/if} 