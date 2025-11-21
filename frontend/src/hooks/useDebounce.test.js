import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useDebounce } from './useDebounce';

describe('useDebounce', () => {
    beforeEach(() => {
        vi.useFakeTimers();
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it('should return initial value immediately', () => {
        const { result } = renderHook(() => useDebounce('test', 500));
        expect(result.current).toBe('test');
    });

    it('should debounce value changes', () => {
        const { result, rerender } = renderHook(
            ({ value, delay }) => useDebounce(value, delay),
            {
                initialProps: { value: 'initial', delay: 500 }
            }
        );

        expect(result.current).toBe('initial');

        // Change value
        rerender({ value: 'changed', delay: 500 });

        // Value should not change immediately
        expect(result.current).toBe('initial');

        // Fast-forward time
        act(() => {
            vi.advanceTimersByTime(500);
        });

        // Value should now be updated
        expect(result.current).toBe('changed');
    });

    it('should cancel previous timeout on rapid changes', () => {
        const { result, rerender } = renderHook(
            ({ value }) => useDebounce(value, 500),
            {
                initialProps: { value: 'initial' }
            }
        );

        // Multiple rapid changes
        rerender({ value: 'change1' });
        act(() => vi.advanceTimersByTime(200));

        rerender({ value: 'change2' });
        act(() => vi.advanceTimersByTime(200));

        rerender({ value: 'final' });

        // Only the final value should be set after full delay
        act(() => {
            vi.advanceTimersByTime(500);
        });

        expect(result.current).toBe('final');
    });

    it('should use default delay if not provided', () => {
        const { result, rerender } = renderHook(
            ({ value }) => useDebounce(value),
            {
                initialProps: { value: 'initial' }
            }
        );

        rerender({ value: 'changed' });

        act(() => {
            vi.advanceTimersByTime(500); // Default delay from DEBOUNCE_DELAY.SEARCH
        });

        expect(result.current).toBe('changed');
    });
});
