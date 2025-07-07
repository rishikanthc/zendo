import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// Tag parsing utilities
export function parseTagsFromTitle(title: string): { cleanTitle: string; tags: string[] } {
  const tagRegex = /#([a-zA-Z0-9_-]+)/g;
  const tags: string[] = [];
  let cleanTitle = title;
  
  // Extract all tags
  let match;
  while ((match = tagRegex.exec(title)) !== null) {
    tags.push(match[1]);
  }
  
  // Remove tags from title
  cleanTitle = title.replace(tagRegex, '').trim();
  
  return { cleanTitle, tags };
}

export function extractTagsFromTitle(title: string): string[] {
  const tagRegex = /#([a-zA-Z0-9_-]+)/g;
  const tags: string[] = [];
  
  let match;
  while ((match = tagRegex.exec(title)) !== null) {
    tags.push(match[1]);
  }
  
  return tags;
}

export function getCleanTitle(title: string): string {
  const tagRegex = /#([a-zA-Z0-9_-]+)/g;
  return title.replace(tagRegex, '').trim();
}

export function tagsToString(tags: string[]): string {
  return tags.join(',');
}

export function stringToTags(tagsString: string): string[] {
  if (!tagsString || tagsString.trim() === '') {
    return [];
  }
  return tagsString.split(',').map(tag => tag.trim()).filter(tag => tag.length > 0);
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChild<T> = T extends { child?: any } ? Omit<T, "child"> : T;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChildren<T> = T extends { children?: any } ? Omit<T, "children"> : T;
export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>;
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null };
