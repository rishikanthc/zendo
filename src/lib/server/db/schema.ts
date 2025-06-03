import { sqliteTable, integer, text } from 'drizzle-orm/sqlite-core';

export const user = sqliteTable('user', {
	id: text('id').primaryKey(),
	age: integer('age'),
	username: text('username').notNull().unique(),
	passwordHash: text('password_hash').notNull()
});

export const session = sqliteTable('session', {
	id: text('id').primaryKey(),
	userId: text('user_id')
		.notNull()
		.references(() => user.id),
	expiresAt: integer('expires_at', { mode: 'timestamp' }).notNull()
});

export const task = sqliteTable('task', {
	id: integer('id').primaryKey({ autoIncrement: true }),
	text: text('text').notNull(),
	day: text('day').notNull(), // ← This is your “day” column
	added_time: integer('added_time', { mode: 'timestamp' }).notNull(),
	completed_time: integer('completed_time', { mode: 'timestamp' }), // nullable
	tags: text('tags'),
	completed: integer('is_completed', { mode: 'boolean' }).notNull().default(false),
	archived: integer('archived', { mode: 'boolean' }).notNull().default(false)
});

export type Task = typeof task.$inferSelect;
export type Session = typeof session.$inferSelect;
export type User = typeof user.$inferSelect;
