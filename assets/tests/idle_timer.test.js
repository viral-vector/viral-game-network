import { describe, it, expect, vi } from 'vitest';
import { IdleTimer } from '../src/utils/IdleTimer';

describe('IdleTimer', () => {
  it('becomes idle only after a full interval without activity', () => {
    vi.useFakeTimers();
    const idle = vi.fn();
    const reset = vi.fn();
    const timer = new IdleTimer(1000, idle, reset);
    vi.advanceTimersByTime(900);
    window.dispatchEvent(new MouseEvent('mousemove'));
    vi.advanceTimersByTime(900);
    expect(idle).not.toHaveBeenCalled();
    vi.advanceTimersByTime(100);
    expect(idle).toHaveBeenCalledTimes(1);
    expect(reset).toHaveBeenCalledTimes(2);
    timer.stop();
  });

  it('removes activity listeners and cancels its timer when stopped', () => {
    vi.useFakeTimers();
    const idle = vi.fn();
    const reset = vi.fn();
    const timer = new IdleTimer(1000, idle, reset);
    timer.stop();
    for (const event of ['mousemove', 'mousedown', 'keypress', 'touchstart', 'scroll']) {
      window.dispatchEvent(new Event(event));
    }
    vi.advanceTimersByTime(2000);
    expect(idle).not.toHaveBeenCalled();
    expect(reset).toHaveBeenCalledTimes(1);
  });
});
