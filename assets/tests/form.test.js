import { beforeEach, describe, it, expect, vi } from 'vitest';
import Form from '../src/admin/form_controller';
import { mount } from './stimulus';

const markup = '<div data-controller="form"><form action="/admin/user" method="POST"><input name="name" value="Alice"></form></div>';

describe('admin forms', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.stubGlobal('progress', { show: vi.fn(), hide: vi.fn() });
    vi.stubGlobal('pushNotification', vi.fn());
  });

  it('submits form data and emits the server result', async () => {
    const controller = await mount('form', Form, markup);
    const result = { status: 'success', message: 'Saved' };
    const request = vi.fn().mockResolvedValue({ ok: true, json: async () => result });
    vi.stubGlobal('fetch', request);
    const form = controller.element.querySelector('form');
    const processed = vi.fn();
    form.addEventListener('form-process', processed);
    const preventDefault = vi.fn();
    await controller.submitForm({ target: form, preventDefault });
    expect(preventDefault).toHaveBeenCalled();
    expect(request).toHaveBeenCalledWith('/admin/user', expect.objectContaining({ method: 'POST' }));
    expect(request.mock.calls[0][1].body.get('name')).toBe('Alice');
    expect(processed.mock.calls[0][0].detail).toEqual(result);
    expect(window.pushNotification).toHaveBeenCalledWith(expect.objectContaining({ message: 'Saved' }));
    vi.advanceTimersByTime(1000);
    expect(window.progress.hide).toHaveBeenCalled();
  });

  it('does not submit a rejected confirmation', async () => {
    const controller = await mount('form', Form, markup);
    const form = controller.element.querySelector('form');
    form.dataset.confirmMessage = 'Delete?';
    vi.stubGlobal('prompt', { show: vi.fn().mockRejectedValue(new Error('cancelled')) });
    const request = vi.fn();
    vi.stubGlobal('fetch', request);
    await controller.submitForm({ target: form, preventDefault: vi.fn() });
    expect(request).not.toHaveBeenCalled();
    expect(window.progress.show).not.toHaveBeenCalled();
  });

  it('reports HTTP and network failures to notifications and listeners', async () => {
    const controller = await mount('form', Form, markup);
    const form = controller.element.querySelector('form');
    const processed = vi.fn();
    form.addEventListener('form-process', processed);
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 400, json: async () => ({ message: 'Invalid' }) }));
    await controller.submitForm({ target: form, preventDefault: vi.fn() });
    expect(processed.mock.calls[0][0].detail.error).toContain('400');
    expect(window.pushNotification).toHaveBeenCalledWith(expect.objectContaining({ type: 'danger' }));
    fetch.mockRejectedValue(new Error('offline'));
    await controller.submitForm({ target: form, preventDefault: vi.fn() });
    expect(processed.mock.calls[1][0].detail.error).toBe('offline');
  });
});
