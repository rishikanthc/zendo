// src/routes/api/tasks/+server.ts
import type { RequestHandler } from '@sveltejs/kit';
import { json, error } from '@sveltejs/kit';
import { db } from '$lib/server/db';
import { task } from '$lib/server/db/schema';
import { eq, and } from 'drizzle-orm';

interface CreateTaskBody {
	text?: string;
	day?: string;
	tags?: string;
}

/**
 * GET /api/tasks?day=<day>
 * Returns all non-archived tasks for the given day.
 * Response: [ { text: string, completed: boolean } ]
 */
export const GET: RequestHandler = async ({ url }) => {
	const day = url.searchParams.get('day');
	if (!day) {
		throw error(400, { message: 'Query parameter `day` is required.' });
	}

	try {
		const rows = await db
			.select({
				id: task.id,
				text: task.text,
				completed: task.completed
			})
			.from(task)
			.where(and(eq(task.day, day), eq(task.archived, false)));

		// rows will be an array of objects { id: number; text: string; completed: number }
		// Drizzle returns boolean columns as 0/1, so we coerce to actual boolean:
		const result = rows.map((r) => ({
			id: r.id,
			text: r.text,
			completed: Boolean(r.completed)
		}));

		return json(result);
	} catch (e: any) {
		console.error('Error fetching tasks by day:', e);
		throw error(500, { message: 'Internal server error while getting tasks.' });
	}
};

/**
 * POST /api/tasks
 * Body: { text: string, day: string, tags?: string }
 */
export const POST: RequestHandler = async ({ request }) => {
	const body: CreateTaskBody = await request.json();

	if (!body.text || !body.day) {
		throw error(400, {
			message: '`text` and `day` are required to create a task.'
		});
	}

	const now = new Date();

	try {
		const result = await db
			.insert(task)
			.values({
				text: body.text,
				day: body.day,
				added_time: now,
				tags: body.tags ?? null,
				completed: false,
				archived: false
			})
			.returning({ id: task.id });

		if (result.length === 0) {
			throw new Error('Insert failed');
		}

		return json({ id: result[0].id }, { status: 201 });
	} catch (e: any) {
		console.error('Error inserting task:', e);
		throw error(500, { message: 'Internal server error creating task.' });
	}
};
