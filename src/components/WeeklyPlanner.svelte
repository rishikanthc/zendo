<script lang="ts">
	import { onMount } from 'svelte';
	import { Accordion, Checkbox, Label, Popover } from 'bits-ui';

	// Focus action for auto-focusing inputs
	function focus(node: HTMLElement) {
		node.focus();
		node.select();
	}

	type Task = {
		id: string;
		description: string;
		completed: boolean;
		popoverOpen?: boolean;
		isEditing?: boolean;
		editingText?: string;
	};

	type Day = {
		name: string;
		tasks: Task[];
		newTaskDescription: string;
		dayBgColor: string;
	};

	// Warm, earthy gradient - from sunrise to sunset
	const warmGradient = [
		'#FFF4E6', // Monday - Soft cream
		'#FFE8CC', // Tuesday - Warm peach
		'#FFD4A3', // Wednesday - Light apricot
		'#F4C2A1', // Thursday - Dusty rose
		'#E8B4A0', // Friday - Warm terracotta
		'#D4A29C', // Saturday - Muted coral
		'#C8958F', // Sunday - Deep rose
		'#B8857A' // Someday - Warm brown
	];

	// Initialize days array
	let days = $state<Day[]>([
		{ name: 'Monday', tasks: [], newTaskDescription: '', dayBgColor: warmGradient[0] },
		{ name: 'Tuesday', tasks: [], newTaskDescription: '', dayBgColor: warmGradient[1] },
		{ name: 'Wednesday', tasks: [], newTaskDescription: '', dayBgColor: warmGradient[2] },
		{ name: 'Thursday', tasks: [], newTaskDescription: '', dayBgColor: warmGradient[3] },
		{ name: 'Friday', tasks: [], newTaskDescription: '', dayBgColor: warmGradient[4] },
		{ name: 'Saturday', tasks: [], newTaskDescription: '', dayBgColor: warmGradient[5] },
		{ name: 'Sunday', tasks: [], newTaskDescription: '', dayBgColor: warmGradient[6] },
		{ name: 'Someday', tasks: [], newTaskDescription: '', dayBgColor: warmGradient[7] }
	]);

	let accordionValue = $state<string | undefined>(undefined);
	let taskHoverState = $state<{ [taskId: string]: boolean }>({});

	/**
	 * onMount: load existing tasks for each day from the backend.
	 * GET /api/tasks?day=<dayName> should return:
	 *   [ { id: number, text: string, completed: boolean }, ... ]
	 */
	onMount(async () => {
		for (let i = 0; i < days.length; i++) {
			const dayName = days[i].name;
			try {
				const res = await fetch(`/api/tasks?day=${encodeURIComponent(dayName)}`);
				if (!res.ok) {
					console.error(`Failed to load tasks for ${dayName}:`, res.statusText);
					continue;
				}
				const payload: Array<{ id?: number | string; text: string; completed: boolean | number }> =
					await res.json();

				// Safely handle the API response with proper null checking
				days[i].tasks = (payload || []).map((row) => ({
					id: row.id ? row.id.toString() : crypto.randomUUID(), // Safe conversion
					description: row.text || '',
					completed: Boolean(row.completed),
					popoverOpen: false,
					isEditing: false,
					editingText: ''
				}));
			} catch (err) {
				console.error(`Error fetching tasks for ${dayName}:`, err);
				// Ensure tasks array is always initialized
				days[i].tasks = [];
			}
		}
	});

	/**
	 * Create a new task via POST /api/tasks
	 * Body: { text: string, day: string }
	 * On success, backend returns { id: newId }.
	 * We append it into days[dayIndex].tasks.
	 */
	async function addTask(dayIndex: number) {
		const dayObj = days[dayIndex];
		const desc = dayObj.newTaskDescription.trim();
		if (!desc) return;

		try {
			const res = await fetch('/api/tasks', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					text: desc,
					day: dayObj.name
				})
			});

			if (res.ok) {
				const responseData = await res.json();
				const newId = responseData.id ? responseData.id.toString() : crypto.randomUUID();

				dayObj.tasks.push({
					id: newId,
					description: desc,
					completed: false,
					popoverOpen: false,
					isEditing: false,
					editingText: ''
				});
				dayObj.newTaskDescription = '';
			} else {
				console.error('Failed to add task:', await res.text());
			}
		} catch (err) {
			console.error('Error while adding task:', err);
		}
	}

	/**
	 * Archive a task via DELETE /api/tasks/:id
	 * On success, we remove it from the UI.
	 */
	async function deleteTask(dayIndex: number, taskId: string) {
		try {
			const res = await fetch(`/api/tasks/${taskId}`, { method: 'DELETE' });
			if (res.ok) {
				days[dayIndex].tasks = days[dayIndex].tasks.filter((t) => t.id !== taskId);
				// Clean up hover state
				delete taskHoverState[taskId];
			} else {
				console.error('Failed to archive task:', await res.text());
			}
		} catch (err) {
			console.error('Error while archiving task:', err);
		}
	}

	/**
	 * Toggle the "completed" flag via PUT /api/tasks/:id
	 * Body: { completed: boolean, completed_time: ISO string }
	 */
	async function toggleTask(dayIndex: number, taskId: string) {
		const taskObj = days[dayIndex].tasks.find((t) => t.id === taskId);
		if (!taskObj) return;

		const newCompleted = !taskObj.completed;
		try {
			const res = await fetch(`/api/tasks/${taskId}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					completed: newCompleted,
					completed_time: newCompleted ? new Date().toISOString() : null
				})
			});

			if (res.ok) {
				taskObj.completed = newCompleted;
			} else {
				console.error('Failed to toggle completed:', await res.text());
			}
		} catch (err) {
			console.error('Error while toggling completed:', err);
		}
	}

	/**
	 * Move a task from `fromDayIndex` into `toDayIndex`.
	 * Calls PUT /api/tasks/:id with { day: <newDayName> }.
	 */
	async function moveTask(fromDayIndex: number, task: Task, toDayIndex: number) {
		if (fromDayIndex === toDayIndex) return;

		try {
			const res = await fetch(`/api/tasks/${task.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					day: days[toDayIndex].name
				})
			});

			if (res.ok) {
				// 1) Remove from old day
				days[fromDayIndex].tasks = days[fromDayIndex].tasks.filter((t) => t.id !== task.id);

				// 2) Add to new day
				days[toDayIndex].tasks.push({
					id: task.id,
					description: task.description,
					completed: task.completed,
					popoverOpen: false,
					isEditing: false,
					editingText: ''
				});
			} else {
				console.error('Failed to move task:', await res.text());
			}
		} catch (err) {
			console.error('Error while moving task:', err);
		}
	}

	/**
	 * Start editing a task
	 */
	function startEditingTask(task: Task) {
		task.isEditing = true;
		task.editingText = task.description;
	}

	/**
	 * Save the edited task text via PUT /api/tasks/:id
	 * Body: { text: <newDescription> }
	 */
	async function saveTaskEdit(task: Task) {
		if (!task.editingText?.trim()) {
			task.isEditing = false;
			task.editingText = '';
			return;
		}

		const newText = task.editingText.trim();
		try {
			const res = await fetch(`/api/tasks/${task.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					text: newText
				})
			});

			if (res.ok) {
				task.description = newText;
			} else {
				console.error('Failed to update task text:', await res.text());
			}
		} catch (err) {
			console.error('Error while updating task text:', err);
		} finally {
			task.isEditing = false;
			task.editingText = '';
		}
	}

	/**
	 * Cancel editing a task
	 */
	function cancelTaskEdit(task: Task) {
		task.isEditing = false;
		task.editingText = '';
	}

	/**
	 * Handle keydown events in task editing mode
	 */
	function handleTaskEditKeydown(e: KeyboardEvent, task: Task) {
		if (e.key === 'Enter') {
			e.preventDefault();
			void saveTaskEdit(task);
		} else if (e.key === 'Escape') {
			cancelTaskEdit(task);
		}
	}

	/**
	 * If user presses Enter in the "new task" input
	 */
	function handleInputKeydown(e: KeyboardEvent, dayIndex: number) {
		if (e.key === 'Enter') {
			e.preventDefault();
			void addTask(dayIndex);
		}
	}
</script>

<div
	class="flex min-h-screen items-center justify-center bg-gradient-to-br from-[#FDF6E3] to-[#F4EAD5] lg:p-4"
>
	<div class="w-full max-w-2xl md:p-8 lg:p-6">
		<Accordion.Root class="w-full space-y-0" type="single" collapsible bind:value={accordionValue}>
			{#each days as day, dayIndex (day.name)}
				<Accordion.Item value={day.name} class="overflow-hidden rounded-sm">
					<Accordion.Header class="flex">
						<Accordion.Trigger
							class="flex flex-1 items-center justify-between p-10 font-[Megrim] text-4xl transition-all hover:brightness-95"
							style={`background-color: ${day.dayBgColor}; color: #333333;`}
						>
							{day.name}
						</Accordion.Trigger>
					</Accordion.Header>

					<Accordion.Content
						class="data-[state=closed]:animate-accordion-up data-[state=open]:animate-accordion-down overflow-hidden text-sm transition-all"
						style={`background-color: ${day.dayBgColor}; color: #333333;`}
					>
						<div class="p-4 pt-2">
							<!-- Input to add a new task -->
							<div class="mb-4 flex items-center space-x-2">
								<input
									type="text"
									placeholder="Add a new task for {day.name}… (Press Enter)"
									bind:value={day.newTaskDescription}
									on:keydown={(e) => handleInputKeydown(e, dayIndex)}
									class="h-10 flex-grow rounded-md px-3 py-2 text-sm text-[#333333] placeholder:text-[#777777]
                         focus-visible:ring-2 focus-visible:ring-[#555555] focus-visible:ring-offset-2 focus-visible:outline-none
                         disabled:cursor-not-allowed disabled:opacity-50"
								/>
							</div>

							{#if day.tasks.length === 0}
								<p class="text-[#555555] italic">No tasks for {day.name}.</p>
							{:else}
								<ul class="space-y-3">
									{#each day.tasks as task (task.id)}
										<li
											class="flex items-center justify-between p-1 transition-colors"
											on:mouseenter={() => (taskHoverState[task.id] = true)}
											on:mouseleave={() => (taskHoverState[task.id] = false)}
										>
											<!-- Left: checkbox + label / edit input -->
											<div class="flex flex-1 items-center space-x-3">
												<Checkbox.Root
													bind:checked={task.completed}
													onCheckedChange={() => void toggleTask(dayIndex, task.id)}
													id={`task-${day.name}-${task.id}`}
													class="peer ring-offset-background h-4 w-4 shrink-0 rounded-sm border border-[#555555]
                                 focus-visible:ring-2 focus-visible:ring-[#555555] focus-visible:ring-offset-2 focus-visible:outline-none
                                 disabled:cursor-not-allowed disabled:opacity-50
                                 data-[state=checked]:bg-[#555555] data-[state=checked]:text-white"
												>
													<Checkbox.Indicator class="flex items-center justify-center text-current">
														<svg viewBox="0 0 8 8" class="h-3 w-3">
															<path
																d="M1.5 3.5l2 2l3-3"
																stroke="currentColor"
																stroke-width="1.5"
																stroke-linecap="round"
																stroke-linejoin="round"
																fill="none"
															/>
														</svg>
													</Checkbox.Indicator>
												</Checkbox.Root>

												{#if task.isEditing}
													<!-- Editing mode: show input -->
													<input
														type="text"
														bind:value={task.editingText}
														on:keydown={(e) => handleTaskEditKeydown(e, task)}
														on:blur={() => saveTaskEdit(task)}
														class="flex-1 rounded border border-[#555555] bg-white px-2 py-1 text-sm text-[#333333]
                                   focus:border-[#777777] focus:ring-1 focus:ring-[#777777] focus:outline-none"
														use:focus
													/>
												{:else}
													<!-- Display mode: show clickable text -->
													<button
														type="button"
														on:click={() => startEditingTask(task)}
														class={`flex-1 cursor-text text-left text-sm leading-none font-medium
                                     ${task.completed ? 'text-[#777777] line-through' : 'text-[#333333]'}`}
													>
														{task.description}
													</button>
												{/if}
											</div>

											<!-- Right: Delete + Move popover (when hovered & not editing) -->
											<div class="flex flex-shrink-0 items-center space-x-2">
												{#if taskHoverState[task.id] && !task.isEditing}
													<!-- Delete Button -->
													<button
														type="button"
														on:click={() => void deleteTask(dayIndex, task.id)}
														class="inline-flex items-center justify-center rounded-md text-sm font-medium whitespace-nowrap
                                   text-red-600 transition-colors hover:bg-red-500/10
                                   focus-visible:ring-2 focus-visible:ring-red-500 focus-visible:ring-offset-2 focus-visible:outline-none
                                   disabled:pointer-events-none disabled:opacity-50"
													>
														<svg
															xmlns="http://www.w3.org/2000/svg"
															viewBox="0 0 24 24"
															fill="currentColor"
															class="h-4 w-4"
														>
															<path
																fill-rule="evenodd"
																d="M16.5 4.478v.227a48.816 48.816 0 0 1 3.878.512.75.75 0 1 1-.256 1.478l-.209-.035-1.005 13.07a3 3 0 0 1-2.991 2.77H8.084a3 3 0 0 1-2.991-2.77L4.087 6.66l-.209.035a.75.75 0 0 1-.256-1.478A48.567 48.567 0 0 1 7.5 4.705v-.227c0-1.564 1.213-2.9 2.816-2.951a52.662 52.662 0 0 1 3.369 0c1.603.051 2.815 1.387 2.815 2.951Zm-6.136-1.452a51.196 51.196 0 0 1 3.273 0C14.39 3.05 15 3.684 15 4.478v.113a49.488 49.488 0 0 0-6 0v-.113c0-.794.609-1.428 1.364-1.452Zm-.355 5.945a.75.75 0 1 0-1.5.058l.347 9a.75.75 0 1 0 1.499-.058l-.346-9Zm5.48.058a.75.75 0 1 0-1.498-.058l-.347 9a.75.75 0 0 0 1.5.058l.345-9Z"
																clip-rule="evenodd"
															/>
														</svg>
														<span class="sr-only">Delete task</span>
													</button>

													<!-- Move Popover -->
													<Popover.Root bind:open={task.popoverOpen}>
														<Popover.Trigger
															class="inline-flex items-center justify-center rounded-md text-sm font-medium whitespace-nowrap
                                     text-blue-600 transition-colors hover:bg-blue-500/10
                                     focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 focus-visible:outline-none
                                     disabled:pointer-events-none disabled:opacity-50"
														>
															<svg
																xmlns="http://www.w3.org/2000/svg"
																viewBox="0 0 24 24"
																fill="currentColor"
																class="h-4 w-4"
															>
																<path
																	fill-rule="evenodd"
																	d="M15.75 2.25H21a.75.75 0 01.75.75v5.25a.75.75 0 01-1.5 0V4.81L8.03 17.03a.75.75 0 01-1.06-1.06L19.19 3.5H15.75a.75.75 0 010-1.5zM1.5 5.25a3 3 0 013-3h5.25a.75.75 0 010 1.5H4.5a1.5 1.5 0 00-1.5 1.5v13.5a1.5 1.5 0 001.5 1.5h13.5a1.5 1.5 0 001.5-1.5V15a.75.75 0 011.5 0v3.75a3 3 0 01-3 3H4.5a3 3 0 01-3-3V5.25z"
																	clip-rule="evenodd"
																/>
															</svg>
															<span class="sr-only">Move task</span>
														</Popover.Trigger>

														<Popover.Content
															class="z-50 w-36 rounded-lg border border-gray-200 bg-white p-2 shadow-md"
															sideOffset={5}
														>
															<div class="mb-2 px-1 text-xs font-semibold text-gray-500">
																Move to:
															</div>
															<div class="space-y-1">
																{#each days as targetDay, targetIndex}
																	{#if targetIndex !== dayIndex}
																		<button
																			type="button"
																			on:click={() => {
																				void moveTask(dayIndex, task, targetIndex);
																				task.popoverOpen = false;
																			}}
																			class="w-full rounded px-2 py-1.5 text-left text-sm text-gray-700
                                             transition-colors hover:bg-gray-100 focus:bg-gray-100 focus:outline-none"
																		>
																			{targetDay.name}
																		</button>
																	{/if}
																{/each}
															</div>
														</Popover.Content>
													</Popover.Root>
												{/if}
											</div>
										</li>
									{/each}
								</ul>
							{/if}
						</div>
					</Accordion.Content>
				</Accordion.Item>
			{/each}
		</Accordion.Root>
	</div>
</div>

<style lang="postcss">
	@keyframes accordion-down {
		from {
			height: 0;
			opacity: 0;
		}
		to {
			height: var(--accordion-content-height, auto);
			opacity: 1;
		}
	}
	@keyframes accordion-up {
		from {
			height: var(--accordion-content-height, auto);
			opacity: 1;
		}
		to {
			height: 0;
			opacity: 0;
		}
	}
	.animate-accordion-down {
		animation: accordion-down 0.2s ease-out;
	}
	.animate-accordion-up {
		animation: accordion-up 0.2s ease-out;
	}
</style>
