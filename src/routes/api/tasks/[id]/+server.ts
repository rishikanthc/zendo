// src/routes/api/tasks/[id]/+server.ts
import type { RequestHandler } from '@sveltejs/kit';
import { json, error } from '@sveltejs/kit';
import { db } from '$lib/server/db';
import { task } from '$lib/server/db/schema';
import { eq } from 'drizzle-orm';

interface UpdateTaskBody {
	text?: string;
	day?: string;
	completed_time?: string; // ISO string or numeric timestamp
	tags?: string;
	completed?: boolean;
}

// PUT /api/tasks/:id
export const PUT: RequestHandler = async ({ params, request }) => {
	const idParam = params.id;
	if (!idParam) {
		throw error(400, { message: 'Missing task ID in URL.' });
	}

	const taskId = parseInt(idParam, 10);
	if (isNaN(taskId) || taskId <= 0) {
		throw error(400, { message: 'Invalid task ID.' });
	}

	const body: UpdateTaskBody = await request.json();
	const { text: newText, day: newDay, completed_time, tags, completed } = body;

	if (
		newText === undefined &&
		newDay === undefined &&
		completed_time === undefined &&
		tags === undefined &&
		completed === undefined
	) {
		throw error(400, {
			message:
				'At least one of `text`, `day`, `completed_time`, `tags`, or `completed` must be provided.'
		});
	}

	const updates: Record<string, any> = {};
	if (newText !== undefined) updates.text = newText;
	if (newDay !== undefined) updates.day = newDay;
	if (completed_time !== undefined) {
		// Handle null completed_time (when unchecking)
		updates.completed_time = completed_time ? new Date(completed_time) : null;
	}
	if (tags !== undefined) updates.tags = tags;
	if (completed !== undefined) updates.completed = completed;

	try {
		// 1) Ensure task exists
		const existing = await db.select().from(task).where(eq(task.id, taskId));
		if (existing.length === 0) {
			throw error(404, { message: `No task found with id = ${taskId}` });
		}

		// 2) Perform partial update
		await db.update(task).set(updates).where(eq(task.id, taskId));

		return json({ success: true });
	} catch (e: any) {
		if (e.status && e.message) {
			throw e;
		}
		console.error('Error updating task:', e);
		throw error(500, { message: 'Internal server error updating task.' });
	}
};

// DELETE /api/tasks/:id  — now just "archive"
export const DELETE: RequestHandler = async ({ params }) => {
	const idParam = params.id;
	if (!idParam) {
		throw error(400, { message: 'Missing task ID in URL.' });
	}

	const taskId = parseInt(idParam, 10);
	if (isNaN(taskId) || taskId <= 0) {
		throw error(400, { message: 'Invalid task ID.' });
	}

	try {
		// 1) Ensure task exists
		const existing = await db.select().from(task).where(eq(task.id, taskId));
		if (existing.length === 0) {
			throw error(404, { message: `No task found with id = ${taskId}` });
		}

		// 2) Archive instead of actual delete
		await db.update(task).set({ archived: true }).where(eq(task.id, taskId));

		return json({ success: true });
	} catch (e: any) {
		if (e.status && e.message) {
			throw e;
		}
		console.error('Error archiving task:', e);
		throw error(500, { message: 'Internal server error archiving task.' });
	}
};
