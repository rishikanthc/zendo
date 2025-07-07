<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '$lib/components/ui/accordion';
	import { Calendar } from '$lib/components/ui/calendar';
	import { Dialog, DialogContent, DialogTrigger } from '$lib/components/ui/dialog';
	import { Plus, Trash2, Calendar as CalendarIcon } from 'lucide-svelte';
	import { CalendarDate, getLocalTimeZone, today, type DateValue } from '@internationalized/date';
	import { taskStore, type Task } from '$lib/stores/taskStore';
	import TaskProgress from '$lib/components/TaskProgress.svelte';
	import NaturalLanguageDatePicker from '$lib/components/NaturalLanguageDatePicker.svelte';
	import TagDisplay from '$lib/components/TagDisplay.svelte';
	import { parseTagsFromTitle, tagsToString, stringToTags } from '$lib/utils';
	import { 
		ContextMenu, 
		ContextMenuContent, 
		ContextMenuItem, 
		ContextMenuTrigger 
	} from '$lib/components/ui/context-menu';

	// Timezone support - default to PST
	const TIMEZONE = 'America/Los_Angeles'; // Can be overridden by environment variable

	interface Day {
		name: string;
		tasks: Task[];
		newTaskDescription: string;
		dayBgColor: string;
	}

	const warmGradient = [
		'#0A0F1C', // Sunday - Darkest navy
		'#1A1F2E', // Monday - Dark navy
		'#2A2F3E', // Tuesday - Medium navy
		'#3A3F4E', // Wednesday - Dark blue-gray
		'#4A4F5E', // Thursday - Medium blue-gray
		'#5A5F6E', // Friday - Light blue-gray
		'#6A6F7E', // Saturday - Lighter blue-gray
		'#050A15' // Someday - Almost black navy
	];

	// Date utilities
	function getWeekStart(date: Date): Date {
		const d = new Date(date);
		const day = d.getDay();
		const diff = d.getDate() - day; // Sunday is 0, so this gives us the Sunday of the week
		const weekStart = new Date(d);
		weekStart.setDate(diff);
		return weekStart;
	}

	function getWeekEnd(date: Date): Date {
		const weekStart = getWeekStart(date);
		return new Date(weekStart.getTime() + 6 * 24 * 60 * 60 * 1000);
	}

	function formatDate(date: Date): string {
		// Format date as YYYY-MM-DD in the target timezone
		const year = date.getFullYear();
		const month = String(date.getMonth() + 1).padStart(2, '0');
		const day = String(date.getDate()).padStart(2, '0');
		return `${year}-${month}-${day}`;
	}

	function getWeekDateString(date: Date): string {
		return formatDate(getWeekStart(date));
	}

	function getDayName(date: Date): string {
		const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
		return days[date.getDay()];
	}

	function getWeekDisplayString(weekStart: Date): string {
		const weekEnd = getWeekEnd(weekStart);
		const startMonth = weekStart.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
		const endMonth = weekEnd.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
		return `${startMonth} - ${endMonth}`;
	}

	function getCurrentDateInTimezone(): Date {
		// Get current date in the specified timezone
		const now = new Date();
		const timeZoneDate = new Date(now.toLocaleString("en-US", {timeZone: TIMEZONE}));
		return timeZoneDate;
	}

	function getTodayDayName(): string {
		// Get today's day name in lowercase for accordion value
		const today = getCurrentDateInTimezone();
		return getDayName(today).toLowerCase();
	}

	function getTimezoneOffset(timezone: string): number {
		// Get the current timezone offset for the specified timezone
		const now = new Date();
		const utc = new Date(now.getTime() + (now.getTimezoneOffset() * 60000));
		const targetTime = new Date(now.toLocaleString("en-US", {timeZone: timezone}));
		const targetUtc = new Date(targetTime.getTime() + (targetTime.getTimezoneOffset() * 60000));
		return (targetUtc.getTime() - utc.getTime()) / 60000;
	}

	function getTextColorForDay(dayName: string): string {
		// Always return white text for the dark palette
		return 'text-white';
	}

	// Current week state
	let currentWeekStart = $state(getWeekStart(getCurrentDateInTimezone()));
	let selectedDate = $state<DateValue | undefined>(today(getLocalTimeZone()));


	// View state
	type ViewMode = 'week' | 'today' | 'todayWeek' | 'allTasks';
	let currentView = $state<ViewMode>('week');
	let todayTab = $state<'all' | 'incomplete' | 'completed'>('all');
	let todayWeekTab = $state<'all' | 'incomplete' | 'completed'>('all');
	let tagFilterQuery = $state('');

	// Accordion state for auto-expanding today's day on first load
	let expandedAccordionItem = $state<string | undefined>(undefined);
	let hasAutoExpanded = $state(false);

	// Swipe navigation state
	let touchStartX = $state(0);
	let touchStartY = $state(0);
	let touchEndX = $state(0);
	let touchEndY = $state(0);
	let isSwiping = $state(false);
	let swipeDirection = $state<'left' | 'right' | null>(null);
	let swipeProgress = $state(0);

	// Initialize days array (Sunday to Saturday)
	let days = $state<Day[]>([
		{ name: 'Sunday', tasks: [], newTaskDescription: '', dayBgColor: warmGradient[0] },
		{ name: 'Monday', tasks: [], newTaskDescription: '', dayBgColor: warmGradient[1] },
		{ name: 'Tuesday', tasks: [], newTaskDescription: '', dayBgColor: warmGradient[2] },
		{ name: 'Wednesday', tasks: [], newTaskDescription: '', dayBgColor: warmGradient[3] },
		{ name: 'Thursday', tasks: [], newTaskDescription: '', dayBgColor: warmGradient[4] },
		{ name: 'Friday', tasks: [], newTaskDescription: '', dayBgColor: warmGradient[5] },
		{ name: 'Saturday', tasks: [], newTaskDescription: '', dayBgColor: warmGradient[6] },
		{ name: 'Someday', tasks: [], newTaskDescription: '', dayBgColor: warmGradient[7] }
	]);

	let loading = $state(true);
	let error = $state('');
	let editingTaskId = $state<number | null>(null);
	let editingText = $state('');
	


	// Dynamic API base URL - use current origin or fallback to localhost
	const API_BASE = typeof window !== 'undefined' 
		? `${window.location.origin}/api`
		: 'http://localhost:8080/api';

	onMount(async () => {
		await taskStore.fetchFromAPI();
		loading = false;
		expandedAccordionItem = getTodayDayName();
		hasAutoExpanded = true;
	});

	// Swipe navigation functions
	function handleTouchStart(event: TouchEvent) {
		touchStartX = event.touches[0].clientX;
		touchStartY = event.touches[0].clientY;
		isSwiping = false;
		swipeDirection = null;
		swipeProgress = 0;
	}

	function handleTouchMove(event: TouchEvent) {
		if (!touchStartX || !touchStartY) return;

		touchEndX = event.touches[0].clientX;
		touchEndY = event.touches[0].clientY;

		const deltaX = touchEndX - touchStartX;
		const deltaY = touchEndY - touchStartY;

		// Check if this is a horizontal swipe (more horizontal than vertical)
		if (Math.abs(deltaX) > Math.abs(deltaY) && Math.abs(deltaX) > 10) {
			isSwiping = true;
			swipeDirection = deltaX > 0 ? 'right' : 'left';
			
			// Calculate swipe progress (0 to 1)
			const maxSwipeDistance = window.innerWidth * 0.3; // 30% of screen width
			swipeProgress = Math.min(Math.abs(deltaX) / maxSwipeDistance, 1);
		}
	}

	function handleTouchEnd() {
		if (!isSwiping || !swipeDirection) {
			// Reset swipe state
			isSwiping = false;
			swipeDirection = null;
			swipeProgress = 0;
			return;
		}

		// Determine if swipe was significant enough to trigger view change
		const minSwipeDistance = window.innerWidth * 0.2; // 20% of screen width
		const swipeDistance = Math.abs(touchEndX - touchStartX);

		if (swipeDistance >= minSwipeDistance) {
			// Navigate to next/previous view
			const viewOrder: ViewMode[] = ['week', 'today', 'todayWeek', 'allTasks'];
			const currentIndex = viewOrder.indexOf(currentView);
			
			if (swipeDirection === 'left' && currentIndex < viewOrder.length - 1) {
				// Swipe left - go to next view
				switchView(viewOrder[currentIndex + 1]);
			} else if (swipeDirection === 'right' && currentIndex > 0) {
				// Swipe right - go to previous view
				switchView(viewOrder[currentIndex - 1]);
			}
		}

		// Reset swipe state
		isSwiping = false;
		swipeDirection = null;
		swipeProgress = 0;
		touchStartX = 0;
		touchStartY = 0;
		touchEndX = 0;
		touchEndY = 0;
	}

	async function createTask(dayOfWeek: string) {
		const dayIndex = days.findIndex(d => d.name.toLowerCase() === dayOfWeek);
		const title = days[dayIndex]?.newTaskDescription?.trim();
		if (!title) return;

		try {
			const weekDate = getWeekDateString(currentWeekStart);
			const { cleanTitle, tags } = parseTagsFromTitle(title);
			await taskStore.add({
				title: cleanTitle,
				dayOfWeek,
				weekDate,
				tags: tagsToString(tags),
				completed: false
			});
			days[dayIndex].newTaskDescription = '';
		} catch (err) {
			const errorMessage = err instanceof Error ? err.message : 'Failed to create task';
			error = errorMessage;
		}
	}

	async function updateTask(task: Task) {
		try {
			console.log('Updating task:', task);
			await taskStore.updateTask(task);
			console.log('Updated tasks array:', $taskStore);
		} catch (err) {
			console.error('Update error:', err);
			const errorMessage = err instanceof Error ? err.message : 'Failed to update task';
			error = errorMessage;
		}
	}

	async function deleteTask(taskId: number) {
		try {
			await taskStore.deleteTask(taskId);
		} catch (err) {
			const errorMessage = err instanceof Error ? err.message : 'Failed to delete task';
			error = errorMessage;
		}
	}

	function handleCalendarSelect(date: DateValue | undefined) {
		if (date) {
			selectedDate = date;
			// Convert DateValue to JavaScript Date
			const jsDate = new Date(date.year, date.month - 1, date.day);
			currentWeekStart = getWeekStart(jsDate);
			
			// If we're in week view and the selected week contains today, auto-expand today's day
			// Otherwise, close the accordion to let user choose
			const today = getCurrentDateInTimezone();
			const selectedWeekStart = getWeekStart(jsDate);
			const selectedWeekEnd = getWeekEnd(selectedWeekStart);
			
			if (today >= selectedWeekStart && today <= selectedWeekEnd) {
				expandedAccordionItem = getTodayDayName();
			} else {
				expandedAccordionItem = undefined;
			}
		}
	}

	function switchView(view: ViewMode) {
		currentView = view;
		// If switching to week view and we haven't auto-expanded yet, do it now
		if (view === 'week' && !hasAutoExpanded) {
			expandedAccordionItem = getTodayDayName();
			hasAutoExpanded = true;
		}
	}

	function switchTodayTab(tab: 'all' | 'incomplete' | 'completed') {
		todayTab = tab;
	}

	function switchTodayWeekTab(tab: 'all' | 'incomplete' | 'completed') {
		todayWeekTab = tab;
	}



	function formatTaskDate(weekDate: string, dayOfWeek: string): string {
		// Convert week date to actual task date
		const weekStart = new Date(weekDate);
		const dayOrder = ['sunday', 'monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday'];
		const dayIndex = dayOrder.indexOf(dayOfWeek.toLowerCase());
		const taskDate = new Date(weekStart.getTime() + dayIndex * 24 * 60 * 60 * 1000);
		return taskDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	function handleAccordionChange(value: string | undefined) {
		// Only allow user control after the initial auto-expand
		if (hasAutoExpanded) {
			expandedAccordionItem = value;
		}
	}

	const getTasksForDay = $derived((dayKey: string): Task[] => {
		const currentWeekDate = getWeekDateString(currentWeekStart);
		return $taskStore.filter((task: Task) => 
			task.dayOfWeek === dayKey && task.weekDate === currentWeekDate
		).sort((a: Task, b: Task) => a.id - b.id);
	});

	const getTaskCount = $derived((dayKey: string): number => {
		return getTasksForDay(dayKey).length;
	});

	const getTodayTasks = $derived((): Task[] => {
		const today = getCurrentDateInTimezone();
		const todayDayOfWeek = getDayName(today).toLowerCase();
		const todayWeekDate = getWeekDateString(getWeekStart(today));
		
		const todayTasks = $taskStore.filter((task: Task) => 
			task.dayOfWeek === todayDayOfWeek && task.weekDate === todayWeekDate
		);
		
		if (todayTab === 'all') return todayTasks;
		if (todayTab === 'incomplete') return todayTasks.filter((task: Task) => !task.completed);
		if (todayTab === 'completed') return todayTasks.filter((task: Task) => task.completed);
		return todayTasks;
	});

	const getTodayWeekTasks = $derived((): Task[] => {
		const today = getCurrentDateInTimezone();
		const currentWeekDate = getWeekDateString(getWeekStart(today));
		
		const currentWeekTasks = $taskStore.filter((task: Task) => 
			task.weekDate === currentWeekDate
		);
		
		if (todayWeekTab === 'all') return currentWeekTasks;
		if (todayWeekTab === 'incomplete') return currentWeekTasks.filter((task: Task) => !task.completed);
		if (todayWeekTab === 'completed') return currentWeekTasks.filter((task: Task) => task.completed);
		return currentWeekTasks;
	});

	// Functions for progress display (always show all tasks regardless of tab)
	const getTodayTasksForProgress = $derived((): Task[] => {
		const today = getCurrentDateInTimezone();
		const todayDayOfWeek = getDayName(today).toLowerCase();
		const todayWeekDate = getWeekDateString(getWeekStart(today));
		
		return $taskStore.filter((task: Task) => 
			task.dayOfWeek === todayDayOfWeek && task.weekDate === todayWeekDate
		);
	});

	const getTodayWeekTasksForProgress = $derived((): Task[] => {
		const today = getCurrentDateInTimezone();
		const currentWeekDate = getWeekDateString(getWeekStart(today));
		
		return $taskStore.filter((task: Task) => 
			task.weekDate === currentWeekDate
		);
	});

	const getAllTasks = $derived((): Task[] => {
		const allTasks = $taskStore.sort((a: Task, b: Task) => {
			// Sort by week date first, then by day of week
			if (a.weekDate !== b.weekDate) {
				return new Date(b.weekDate).getTime() - new Date(a.weekDate).getTime();
			}
			const dayOrder = ['sunday', 'monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday'];
			return dayOrder.indexOf(a.dayOfWeek) - dayOrder.indexOf(b.dayOfWeek);
		});
		
		// If no tag filter query, return all tasks
		if (!tagFilterQuery.trim()) {
			return allTasks;
		}
		
		// Parse the tag filter query into individual tags
		const filterTags = tagFilterQuery.trim().toLowerCase().split(/\s+/);
		
		// Filter tasks that have any of the specified tags
		return allTasks.filter((task: Task) => {
			if (!task.tags) return false;
			
			const taskTags = stringToTags(task.tags).map(tag => tag.toLowerCase());
			
			// Return true if any of the filter tags match any of the task tags
			return filterTags.some(filterTag => 
				taskTags.some(taskTag => taskTag.includes(filterTag))
			);
		});
	});

	function handleKeyPress(event: KeyboardEvent, dayOfWeek: string) {
		if (event.key === 'Enter') {
			createTask(dayOfWeek);
		}
	}

	function startEditing(task: Task) {
		editingTaskId = task.id;
		// Show full title with tags when editing
		const tagList = stringToTags(task.tags);
		const tagsString = tagList.map(tag => `#${tag}`).join(' ');
		editingText = task.title + (tagsString ? ` ${tagsString}` : '');
	}

	function stopEditing() {
		editingTaskId = null;
		editingText = '';
	}

	async function saveTaskEdit(task: Task) {
		if (editingText.trim() === '') {
			stopEditing();
			return;
		}

		try {
			const { cleanTitle, tags } = parseTagsFromTitle(editingText.trim());
			const updatedTask = { 
				...task, 
				title: cleanTitle,
				tags: tagsToString(tags)
			};
			await updateTask(updatedTask);
			stopEditing();
		} catch (err) {
			console.error('Failed to save task edit:', err);
			stopEditing();
		}
	}

	function handleEditKeyPress(event: KeyboardEvent, task: Task) {
		if (event.key === 'Enter') {
			saveTaskEdit(task);
		} else if (event.key === 'Escape') {
			stopEditing();
		}
	}


</script>

<svelte:head>
	<title>Zendo - Weekly Todo</title>
</svelte:head>

<div 
	class="min-h-screen bg-gray-800 p-4 pb-20 md:pb-4"
	ontouchstart={handleTouchStart}
	ontouchmove={handleTouchMove}
	ontouchend={handleTouchEnd}
	style="touch-action: pan-y;"
>
	<div class="mx-auto max-w-6xl">
		<!-- Header with Date, View Switcher, and Calendar -->
		<div class="max-w-2xl mx-auto mb-6">
			<div class="flex items-center justify-between mb-4">
				<div class="text-white">
					<h1 class="text-2xl font-[Megrim] uppercase tracking-wider mb-1">Zendo</h1>
					<p class="text-gray-300 text-sm">{getCurrentDateInTimezone().toLocaleDateString('en-US', { month: 'long', day: 'numeric' })}</p>
				</div>
				<!-- Desktop View Switcher -->
				<div class="hidden md:flex justify-center gap-4">
					<button
						onclick={() => switchView('week')}
						class="w-2 h-2 rounded-full transition-all duration-200 {currentView === 'week' ? 'bg-yellow-100 shadow-lg' : 'bg-gray-500 hover:bg-gray-300'}"
						title="Week View"
					>
						<!-- <div class="w-full h-full flex items-center justify-center text-white text-xs font-medium">
							W
						</div> -->
					</button>
					<button
						onclick={() => switchView('today')}
						class="w-2 h-2 rounded-full transition-all duration-200 {currentView === 'today' ? 'bg-yellow-100 shadow-lg' : 'bg-gray-500 hover:bg-gray-300'}"
						title="Today's Tasks"
					>
						<!-- <div class="w-full h-full flex items-center justify-center text-white text-xs font-medium">
							T
						</div> -->
					</button>
					<button
						onclick={() => switchView('todayWeek')}
						class="w-2 h-2 rounded-full transition-all duration-200 {currentView === 'todayWeek' ? 'bg-yellow-100 shadow-lg' : 'bg-gray-500 hover:bg-gray-300'}"
						title="This Week's Tasks"
					>
						<!-- <div class="w-full h-full flex items-center justify-center text-white text-xs font-medium">
							TW
						</div> -->
					</button>
					<button
						onclick={() => switchView('allTasks')}
						class="w-2 h-2 rounded-full transition-all duration-200 {currentView === 'allTasks' ? 'bg-yellow-100 shadow-lg' : 'bg-gray-500 hover:bg-gray-300'}"
						title="All Tasks"
					>
						<!-- <div class="w-full h-full flex items-center justify-center text-white text-xs font-medium">
							A
						</div> -->
					</button>
				</div>
				<div class="flex items-center gap-3">
					<NaturalLanguageDatePicker 
						value={selectedDate}
						onDateChange={handleCalendarSelect}
					/>
				</div>
			</div>
			
			<!-- View Switcher -->
			
		</div>

		<!-- Error Display -->
		{#if error}
			<div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
				{error}
			</div>
		{/if}

		<!-- Offline Indicator -->
		{#if !navigator.onLine}
			<div class="mb-4 p-3 bg-yellow-100 border border-yellow-400 text-yellow-700 rounded flex items-center gap-2">
				<span class="text-sm">📱</span>
				<span class="text-sm">You're offline. Viewing cached data. Changes will sync when connection returns.</span>
			</div>
		{/if}

		<!-- Loading State -->
		{#if loading}
			<div class="flex justify-center items-center h-64">
				<div class="text-white">Loading tasks...</div>
			</div>
		{:else}
			<!-- Main Content Area with Swipe Animation -->
			<div 
				class="max-w-2xl mx-auto overflow-hidden"
				style="transform: translateX({isSwiping && swipeDirection === 'left' ? -swipeProgress * 20 : isSwiping && swipeDirection === 'right' ? swipeProgress * 20 : 0}px); transition: {isSwiping ? 'none' : 'transform 0.3s ease-out'};"
			>
				{#if currentView === 'week'}
					<!-- Week View - Day Cards -->
					
					<!-- Week Progress Summary -->
					<div class="mb-4 flex justify-end gap-6">
						<div class="flex items-center gap-2">
							<span class="text-xs text-gray-400">Today:</span>
							<TaskProgress 
								tasks={getTodayTasks()} 
							/>
						</div>
						<div class="flex items-center gap-2">
							<span class="text-xs text-gray-400">Week:</span>
							<TaskProgress 
								tasks={days.flatMap(day => getTasksForDay(day.name.toLowerCase()))} 
							/>
						</div>
					</div>
					
					<Accordion type="single" class="w-full" bind:value={expandedAccordionItem} onValueChange={handleAccordionChange}>
						{#each days as day, index}
							{@const dayKey = day.name.toLowerCase()}
							{@const dayTasks = getTasksForDay(dayKey)}
							
							<AccordionItem value={dayKey} class="border-none rounded-xs overflow-hidden mb-0" style="background-color: {day.dayBgColor};">
								<AccordionTrigger class="px-6 py-6 hover:no-underline [&[data-state=open]>div>div]:rotate-0">
									<div class="flex items-center justify-between w-full text-xl font-normal">
										<span class="font-[Megrim] text-magnum-400 uppercase tracking-wider">{day.name}</span>
										<div class="flex items-center gap-3">
											<span class="text-sm bg-white/20 px-3 py-1.5 text-gray-50 rounded-full font-medium">
												{getTaskCount(dayKey)}
											</span>
										</div>
									</div>
								</AccordionTrigger>
								<AccordionContent class="px-6 pb-6">
									<!-- Add New Task -->
									<div class="mb-6 flex gap-3 items-center">
										<Input
											bind:value={day.newTaskDescription}
											placeholder="Add new task... Use #tag for tags"
											onkeypress={(e) => handleKeyPress(e, dayKey)}
											class="flex-1 border-none {getTextColorForDay(dayKey)} placeholder:text-gray-300 focus:border-none focus:ring-0 focus:outline-none bg-transparent h-10 text-base"
										/>
										<Button
											size="default"
											onclick={() => createTask(dayKey)}
											disabled={!day.newTaskDescription?.trim()}
											class="bg-white/10 hover:bg-white/20 text-white border-white/20 h-12 px-4"
										>
											<Plus class="h-5 w-5" />
										</Button>
									</div>

																		<!-- Tasks List -->
									<div class="space-y-0">
										{#each dayTasks as task (task.id)}
											<ContextMenu>
												<ContextMenuTrigger>
													<div class="group flex items-start gap-3 p-1 rounded-lg hover:bg-white/10 transition-colors">
														<Checkbox
															bind:checked={task.completed}
															onCheckedChange={async () => {
																console.log('Checkbox clicked for task:', task.id, 'New state:', task.completed);
																await updateTask(task);
															}}
															class="border-white data-[state=checked]:border-none data-[state=checked]:bg-blue-500 rounded-0 rounded-xs mt-0.5"
														/>
														<div class="flex-1 min-w-0">
															{#if editingTaskId === task.id}
																<Input
																	bind:value={editingText}
																	onkeypress={(e) => handleEditKeyPress(e, task)}
																	onblur={() => saveTaskEdit(task)}
																	class="w-full text-sm font-medium border-none bg-transparent p-0 {getTextColorForDay(dayKey)} focus:border-none focus:ring-0 focus:outline-none"
																	autofocus
																/>
															{:else}
																<div 
																	class="cursor-pointer hover:text-gray-700 transition-colors"
																	onclick={() => startEditing(task)}
																>
																	<span class="text-sm font-medium {task.completed ? 'text-gray-200' : getTextColorForDay(dayKey)}">
																		{task.title}
																		{#if task.id > 1000000}
																			<span class="text-xs text-yellow-600 ml-2">(pending sync)</span>
																		{/if}
																		<TagDisplay tags={task.tags} completed={task.completed} />
																	</span>
																</div>
															{/if}
														</div>
														<!-- Desktop delete button (hidden on mobile) -->
														<Button
															variant="ghost"
															size="sm"
															onclick={() => deleteTask(task.id)}
															class="h-6 w-6 p-0 text-red-400 hover:text-red-300 hover:bg-red-500/20 opacity-0 group-hover:opacity-100 transition-opacity md:block hidden"
														>
															<Trash2 class="h-3 w-3" />
														</Button>
													</div>
												</ContextMenuTrigger>
												<ContextMenuContent>
													<ContextMenuItem onclick={() => deleteTask(task.id)} class="text-red-600">
														<Trash2 class="h-4 w-4 mr-2" />
														Delete Task
													</ContextMenuItem>
												</ContextMenuContent>
											</ContextMenu>
										{/each}
									</div>
								</AccordionContent>
							</AccordionItem>
						{/each}
					</Accordion>
				{:else if currentView === 'today'}
					<!-- Today View - Tabbed Interface -->
					
					<Card class="bg-gray-800/50 border-gray-700 shadow-md p-1">
						<CardHeader class="mt-2">
							<div class="flex items-center justify-between">
								<CardTitle class="text-white font-normal font-[Megrim] uppercase tracking-wider">Today's Tasks</CardTitle>
								<TaskProgress tasks={getTodayTasksForProgress()} />
							</div>
						</CardHeader>
						<CardContent class="p-2">
							<!-- Tab Navigation -->
							<div class="flex gap-2 mb-6">
								<Button
									variant={todayTab === 'all' ? 'default' : 'ghost'}
									size="sm"
									onclick={() => switchTodayTab('all')}
									class="text-white {todayTab === 'all' ? 'bg-white/20' : 'hover:bg-white/10'}"
								>
									All
								</Button>
								<Button
									variant={todayTab === 'incomplete' ? 'default' : 'ghost'}
									size="sm"
									onclick={() => switchTodayTab('incomplete')}
									class="text-white {todayTab === 'incomplete' ? 'bg-white/20' : 'hover:bg-white/10'}"
								>
									Incomplete
								</Button>
								<Button
									variant={todayTab === 'completed' ? 'default' : 'ghost'}
									size="sm"
									onclick={() => switchTodayTab('completed')}
									class="text-white {todayTab === 'completed' ? 'bg-white/20' : 'hover:bg-white/10'}"
								>
									Completed
								</Button>
							</div>

							<!-- Tasks List -->
							<div class="space-y-2">
								{#each getTodayTasks() as task (task.id)}
									<ContextMenu>
										<ContextMenuTrigger>
											<div class="group flex items-center gap-3 p-3 rounded-lg bg-gray-700 hover:bg-gray-700/50 transition-colors">
												<Checkbox
													bind:checked={task.completed}
													onCheckedChange={async () => {
														console.log('Checkbox clicked for task:', task.id, 'New state:', task.completed);
														await updateTask(task);
													}}
													class="border-white data-[state=checked]:border-none data-[state=checked]:bg-blue-500 rounded-0 rounded-xs"
												/>
												<div class="flex-1 min-w-0">
													{#if editingTaskId === task.id}
														<Input
															bind:value={editingText}
															onkeypress={(e) => handleEditKeyPress(e, task)}
															onblur={() => saveTaskEdit(task)}
															class="w-full text-sm font-medium border-none bg-transparent p-0 text-white focus:border-none focus:ring-0 focus:outline-none"
															autofocus
														/>
													{:else}
														<div 
															class="cursor-pointer hover:text-gray-300 transition-colors"
															onclick={() => startEditing(task)}
														>
															<span class="text-sm font-normal {task.completed ? 'text-gray-400' : 'text-white'}">
																{task.title}
																<TagDisplay tags={task.tags} completed={task.completed} />
															</span>
														</div>
													{/if}
												</div>
												<!-- Desktop delete button (hidden on mobile) -->
												<Button
													variant="ghost"
													size="sm"
													onclick={() => deleteTask(task.id)}
													class="h-6 w-6 p-0 text-red-400 hover:text-red-300 hover:bg-red-500/20 opacity-0 group-hover:opacity-100 transition-opacity md:block hidden"
												>
													<Trash2 class="h-3 w-3" />
												</Button>
											</div>
										</ContextMenuTrigger>
										<ContextMenuContent>
											<ContextMenuItem onclick={() => deleteTask(task.id)} class="text-red-600">
												<Trash2 class="h-4 w-4 mr-2" />
												Delete Task
											</ContextMenuItem>
										</ContextMenuContent>
									</ContextMenu>
								{/each}
								{#if getTodayTasks().length === 0}
									<div class="text-center text-gray-400 py-8">
										No tasks found for today
									</div>
								{/if}
							</div>
						</CardContent>
					</Card>
				{:else if currentView === 'todayWeek'}
					<!-- Today's Week View - Tabbed Interface -->
					
					<Card class="bg-gray-800/50 shadow-md border-gray-700 p-1">
						<CardHeader class="mt-2">
							<div class="flex items-center justify-between">
								<CardTitle class="font-normal text-white font-[Megrim] uppercase tracking-wider">This Week's Tasks</CardTitle>
								<TaskProgress tasks={getTodayWeekTasksForProgress()} />
							</div>
						</CardHeader>
						<CardContent class="p-2">
							<!-- Tab Navigation -->
							<div class="flex gap-2 mb-6">
								<Button
									variant={todayWeekTab === 'all' ? 'default' : 'ghost'}
									size="sm"
									onclick={() => switchTodayWeekTab('all')}
									class="text-white {todayWeekTab === 'all' ? 'bg-white/20' : 'hover:bg-white/10'}"
								>
									All
								</Button>
								<Button
									variant={todayWeekTab === 'incomplete' ? 'default' : 'ghost'}
									size="sm"
									onclick={() => switchTodayWeekTab('incomplete')}
									class="text-white {todayWeekTab === 'incomplete' ? 'bg-white/20' : 'hover:bg-white/10'}"
								>
									Incomplete
								</Button>
								<Button
									variant={todayWeekTab === 'completed' ? 'default' : 'ghost'}
									size="sm"
									onclick={() => switchTodayWeekTab('completed')}
									class="text-white {todayWeekTab === 'completed' ? 'bg-white/20' : 'hover:bg-white/10'}"
								>
									Completed
								</Button>
							</div>

							<!-- Tasks List -->
							<div class="space-y-2">
								{#each getTodayWeekTasks() as task (task.id)}
									<ContextMenu>
										<ContextMenuTrigger>
											<div class="group flex items-center gap-3 p-3 rounded-lg bg-gray-700 hover:bg-gray-700/50 transition-colors">
												<Checkbox
													bind:checked={task.completed}
													onCheckedChange={async () => {
														console.log('Checkbox clicked for task:', task.id, 'New state:', task.completed);
														await updateTask(task);
													}}
													class="border-white data-[state=checked]:border-none data-[state=checked]:bg-blue-500 rounded-0 rounded-xs"
												/>
												<div class="flex-1 min-w-0">
													{#if editingTaskId === task.id}
														<Input
															bind:value={editingText}
															onkeypress={(e) => handleEditKeyPress(e, task)}
															onblur={() => saveTaskEdit(task)}
															class="w-full text-sm font-medium border-none bg-transparent p-0 text-white focus:border-none focus:ring-0 focus:outline-none"
															autofocus
														/>
													{:else}
														<div 
															class="cursor-pointer hover:text-gray-300 transition-colors"
															onclick={() => startEditing(task)}
														>
															<span class="text-sm font-medium {task.completed ? 'text-gray-400' : 'text-white'}">
																{task.title}
																<TagDisplay tags={task.tags} completed={task.completed} />
															</span>
														</div>
													{/if}
												</div>
												<div class="flex items-center gap-2">
													<span class="text-xs text-gray-400 bg-white/10 px-2 py-1 rounded">
														{task.dayOfWeek}
													</span>
													<!-- Desktop delete button (hidden on mobile) -->
													<Button
														variant="ghost"
														size="sm"
														onclick={() => deleteTask(task.id)}
														class="h-6 w-6 p-0 text-red-400 hover:text-red-300 hover:bg-red-500/20 opacity-0 group-hover:opacity-100 transition-opacity md:block hidden"
													>
														<Trash2 class="h-3 w-3" />
													</Button>
												</div>
											</div>
										</ContextMenuTrigger>
										<ContextMenuContent>
											<ContextMenuItem onclick={() => deleteTask(task.id)} class="text-red-600">
												<Trash2 class="h-4 w-4 mr-2" />
												Delete Task
											</ContextMenuItem>
										</ContextMenuContent>
									</ContextMenu>
								{/each}
								{#if getTodayWeekTasks().length === 0}
									<div class="text-center text-gray-400 py-8">
										No tasks found for this week
									</div>
								{/if}
							</div>
						</CardContent>
					</Card>
				{:else if currentView === 'allTasks'}
					<!-- All Tasks View - Tabbed Interface -->
					<Card class="bg-gray-800/50 border-gray-700 p-1">
						<CardHeader class="mt-2">
							<div class="flex items-center justify-between">
								<CardTitle class="text-white font-normal font-[Megrim] uppercase tracking-wider">All Tasks</CardTitle>
								<TaskProgress tasks={getAllTasks()} />
							</div>
						</CardHeader>
						<CardContent class="p-2">
							<!-- Tag Filter Input -->
							<div class="mb-6">
								<Input
									bind:value={tagFilterQuery}
									placeholder="Filter by tags (e.g., work urgent personal)"
									class="w-full border-white/20 bg-white/10 text-white placeholder:text-gray-400 focus:border-white/40 focus:ring-white/20"
								/>
							</div>

							<!-- Tasks List -->
							<div class="space-y-2">
								{#each getAllTasks() as task (task.id)}
									<ContextMenu>
										<ContextMenuTrigger>
											<div class="group flex items-center gap-3 p-3 rounded-lg bg-gray-700/30 hover:bg-gray-700/50 transition-colors">
												<Checkbox
													bind:checked={task.completed}
													onCheckedChange={async () => {
														console.log('Checkbox clicked for task:', task.id, 'New state:', task.completed);
														await updateTask(task);
													}}
													class="border-white data-[state=checked]:border-none data-[state=checked]:bg-blue-500 rounded-0 rounded-xs"
												/>
												<div class="flex-1 min-w-0">
													{#if editingTaskId === task.id}
														<Input
															bind:value={editingText}
															onkeypress={(e) => handleEditKeyPress(e, task)}
															onblur={() => saveTaskEdit(task)}
															class="w-full text-sm font-medium border-none bg-transparent p-0 text-white focus:border-none focus:ring-0 focus:outline-none"
															autofocus
														/>
													{:else}
														<div 
															class="cursor-pointer hover:text-gray-300 transition-colors"
															onclick={() => startEditing(task)}
														>
															<span class="text-sm font-medium {task.completed ? 'text-gray-400' : 'text-white'}">
																{task.title}
																<TagDisplay tags={task.tags} completed={task.completed} />
															</span>
														</div>
													{/if}
												</div>
												<div class="flex items-center gap-2">
													<span class="text-xs text-gray-400 bg-white/10 px-2 py-1 rounded">
														{formatTaskDate(task.weekDate, task.dayOfWeek)}
													</span>
													<!-- Desktop delete button (hidden on mobile) -->
													<Button
														variant="ghost"
														size="sm"
														onclick={() => deleteTask(task.id)}
														class="h-6 w-6 p-0 text-red-400 hover:text-red-300 hover:bg-red-500/20 opacity-0 group-hover:opacity-100 transition-opacity md:block hidden"
													>
														<Trash2 class="h-3 w-3" />
													</Button>
												</div>
											</div>
										</ContextMenuTrigger>
										<ContextMenuContent>
											<ContextMenuItem onclick={() => deleteTask(task.id)} class="text-red-600">
												<Trash2 class="h-4 w-4 mr-2" />
												Delete Task
											</ContextMenuItem>
										</ContextMenuContent>
									</ContextMenu>
								{/each}
								{#if getAllTasks().length === 0}
									<div class="text-center text-gray-400 py-8">
										No tasks found
									</div>
								{/if}
							</div>
						</CardContent>
					</Card>
				{/if}
			</div>
		{/if}
	</div>
	
	<!-- Offline Sync Toast Notifications -->
	
	<!-- Mobile Bottom Navigation -->
	<div class="md:hidden fixed bottom-0 left-0 right-0 bg-gray-800 border-t border-gray-700 z-50">
		<div class="flex justify-around items-center py-4 px-6">
			<button
				onclick={() => switchView('week')}
				class="w-3 h-3 rounded-full transition-all duration-200 {currentView === 'week' ? 'bg-yellow-100 shadow-lg' : 'bg-gray-500 hover:bg-gray-300'}"
				title="Week View"
			>
			</button>
			<button
				onclick={() => switchView('today')}
				class="w-3 h-3 rounded-full transition-all duration-200 {currentView === 'today' ? 'bg-yellow-100 shadow-lg' : 'bg-gray-500 hover:bg-gray-300'}"
				title="Today's Tasks"
			>
			</button>
			<button
				onclick={() => switchView('todayWeek')}
				class="w-3 h-3 rounded-full transition-all duration-200 {currentView === 'todayWeek' ? 'bg-yellow-100 shadow-lg' : 'bg-gray-500 hover:bg-gray-300'}"
				title="This Week's Tasks"
			>
			</button>
			<button
				onclick={() => switchView('allTasks')}
				class="w-3 h-3 rounded-full transition-all duration-200 {currentView === 'allTasks' ? 'bg-yellow-100 shadow-lg' : 'bg-gray-500 hover:bg-gray-300'}"
				title="All Tasks"
			>
			</button>
		</div>
	</div>
</div>
