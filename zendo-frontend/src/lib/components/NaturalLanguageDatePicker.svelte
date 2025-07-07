<script lang="ts">
	import { Button } from "$lib/components/ui/button/index.js";
	import { Calendar } from "$lib/components/ui/calendar/index.js";
	import { Input } from "$lib/components/ui/input/index.js";
	import { Dialog, DialogContent, DialogTrigger } from "$lib/components/ui/dialog/index.js";
	import CalendarIcon from "@lucide/svelte/icons/calendar";
	import { parseDate } from "chrono-node";
	import {
		CalendarDate,
		getLocalTimeZone,
		type DateValue
	} from "@internationalized/date";

	function formatDate(date: DateValue | undefined) {
		if (!date) return "";

		return date.toDate(getLocalTimeZone()).toLocaleDateString("en-US", {
			day: "2-digit",
			month: "long",
			year: "numeric"
		});
	}

	function getWeekDisplayString(date: Date): string {
		const start = new Date(date);
		const end = new Date(date.getTime() + 6 * 24 * 60 * 60 * 1000);
		return `${start.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })} - ${end.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}`;
	}

	const { value = undefined, onDateChange = () => {} } = $props<{
		value?: DateValue;
		onDateChange?: (date: DateValue) => void;
	}>();

	let open = $state(false);
	let inputValue = $state("");

	// Initialize input value based on current value
	$effect(() => {
		if (value) {
			inputValue = formatDate(value);
		}
	});

	function handleInputChange(newValue: string) {
		inputValue = newValue;
		const date = parseDate(newValue);
		if (date) {
			const calendarDate = new CalendarDate(
				date.getFullYear(),
				date.getMonth() + 1,
				date.getDate()
			);
			onDateChange(calendarDate);
		}
	}

	function handleCalendarChange(newValue: DateValue | undefined) {
		if (newValue) {
			inputValue = formatDate(newValue);
			onDateChange(newValue);
			open = false;
		}
	}
</script>

<div class="flex flex-col gap-2">
	<div class="relative flex gap-2">
		<Input
			value={inputValue}
			placeholder="Next week, tomorrow, or Dec 15"
			class="bg-gray-700 border-gray-600 text-white placeholder:text-gray-400 pr-10 focus:border-gray-500 focus:ring-gray-500"
			onkeydown={(e) => {
				if (e.key === "ArrowDown") {
					e.preventDefault();
					open = true;
				}
			}}
			oninput={(e) => handleInputChange(e.currentTarget.value)}
		/>
		<Dialog bind:open>
			<DialogTrigger>
				<Button
					variant="ghost"
					class="absolute right-2 top-1/2 size-6 -translate-y-1/2 text-gray-300 hover:text-white hover:bg-gray-600/50"
				>
					<CalendarIcon class="size-3.5" />
					<span class="sr-only">Select date</span>
				</Button>
			</DialogTrigger>
			<DialogContent class="w-auto overflow-hidden p-0">
				<Calendar
					type="single"
					value={value}
					captionLayout="dropdown"
					onValueChange={handleCalendarChange}
				/>
			</DialogContent>
		</Dialog>
	</div>
	{#if value}
		<div class="text-gray-300 px-1 text-sm">
			Showing week of <span class="font-medium">{getWeekDisplayString(value.toDate(getLocalTimeZone()))}</span>
		</div>
	{/if}
</div> 