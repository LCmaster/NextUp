import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { get } from 'svelte/store';
import { toast } from './toast';

describe('toast store', () => {
	beforeEach(() => {
		vi.useFakeTimers();
	});

	afterEach(() => {
		// Clear toasts after each test
		const currentToasts = get(toast);
		currentToasts.forEach((t) => toast.remove(t.id));
		vi.useRealTimers();
	});

	it('adds a success toast', () => {
		toast.success('It worked!');
		const current = get(toast);
		expect(current).toHaveLength(1);
		expect(current[0].message).toBe('It worked!');
		expect(current[0].type).toBe('success');
	});

	it('adds an error toast', () => {
		toast.error('Something failed');
		const current = get(toast);
		expect(current).toHaveLength(1);
		expect(current[0].type).toBe('error');
	});

	it('removes a toast by id', () => {
		toast.info('Info message');
		const current = get(toast);
		expect(current).toHaveLength(1);
		const id = current[0].id;

		toast.remove(id);
		expect(get(toast)).toHaveLength(0);
	});

	it('auto-removes toast after duration', () => {
		toast.success('Will disappear soon');
		expect(get(toast)).toHaveLength(1);

		// Advance timer by 3500ms (default duration)
		vi.advanceTimersByTime(3500);

		expect(get(toast)).toHaveLength(0);
	});

	it('multiple toasts are independent', () => {
		toast.success('First');
		toast.error('Second');
		toast.info('Third');

		expect(get(toast)).toHaveLength(3);

		const secondId = get(toast)[1].id;
		toast.remove(secondId);

		const remaining = get(toast);
		expect(remaining).toHaveLength(2);
		expect(remaining[0].message).toBe('First');
		expect(remaining[1].message).toBe('Third');
	});
});
