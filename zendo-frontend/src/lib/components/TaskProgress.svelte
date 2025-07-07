<script lang="ts">
	import { type Task } from '$lib/stores/taskStore';

	const { tasks = [] } = $props<{
		tasks?: Task[];
	}>();

	const totalTasks = $derived(tasks.length);
	const completedTasks = $derived(tasks.filter((task: Task) => task.completed).length);
	const progressPercentage = $derived(totalTasks > 0 ? Math.round((completedTasks / totalTasks) * 100) : 0);
	const progressColor = $derived(
		progressPercentage === 100 ? 'bg-green-400' : 
		progressPercentage >= 50 ? 'bg-blue-400' : 'bg-gray-400'
	);
</script>

<div class="flex items-center gap-2 text-xs text-gray-400">
	<span>{completedTasks}/{totalTasks}</span>
	<div class="w-12 h-1 bg-white/20 rounded-full overflow-hidden">
		<div 
			class="h-full {progressColor} transition-all duration-300 ease-out"
			style="width: {progressPercentage}%"
		></div>
	</div>
	<span>{progressPercentage}%</span>
</div> 