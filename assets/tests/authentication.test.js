import { beforeEach, describe, it, expect, vi } from 'vitest';
import Authentication from '../src/admin/authentication_controller';
import { mount } from './stimulus';

describe('authentication', () => {
  beforeEach(() => vi.useFakeTimers());

  it('refreshes only when a token is close to expiration', async () => {
    const controller = await mount('authentication', Authentication, '<div data-controller="authentication"></div>');
    const now = Math.floor(Date.now() / 1000);
    for (const [expires, expected] of [[now + 60, true], [now + 600, false], [now - 60, false]]) {
      controller.expiresValue = expires;
      expect(controller.isJWTTokenExpire()).toBe(expected);
    }
  });

  it('handles missing and invalid expiry metadata without throwing', async () => {
    const controller = await mount('authentication', Authentication, '<div data-controller="authentication"></div>');
    expect(controller.isJWTTokenExpire()).toBeFalsy();
    controller.expiresValue = NaN;
    expect(controller.isJWTTokenExpire()).toBeFalsy();
  });

  it('suppresses refresh while the document is hidden or the user is idle', async () => {
    const controller = await mount('authentication', Authentication, '<div data-controller="authentication"></div>');
    const request = vi.fn();
    vi.stubGlobal('fetch', request);
    vi.stubGlobal('userIdle', true);
    await controller.getFreshJWTToken();
    expect(request).not.toHaveBeenCalled();
  });

  it('does not leave a refresh interval running after disconnect', async () => {
    const controller = await mount('authentication', Authentication, '<div data-controller="authentication"></div>');
    const check = vi.spyOn(controller, 'isJWTTokenExpire');
    controller.element.remove();
    for (let i = 0; i < 6; i++) await Promise.resolve();
    vi.advanceTimersByTime(90_000);
    expect(check).not.toHaveBeenCalled();
  });

  it('refreshes expiry metadata without reading or rewriting the session cookie', async () => {
    const controller = await mount('authentication', Authentication, '<div data-controller="authentication"></div>');
    const expires = Math.floor(Date.now() / 1000) + 3600;
    document.cookie = 'VNET_SESSION=opaque; path=/';
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => ({ expires_at: expires }) }));
    expect(await controller.getFreshJWTToken()).toBe(true);
    expect(controller.expiresValue).toBe(expires);
    expect(document.cookie).toContain('VNET_SESSION=opaque');
  });

  it('handles failed network requests and allows a later retry', async () => {
    const controller = await mount('authentication', Authentication, '<div data-controller="authentication"></div>');
    const request = vi.fn().mockRejectedValueOnce(new Error('offline')).mockResolvedValueOnce({ ok: true, json: async () => ({ expires_at: Math.floor(Date.now() / 1000) + 3600 }) });
    vi.stubGlobal('fetch', request);
    expect(await controller.getFreshJWTToken()).toBe(false);
    expect(await controller.getFreshJWTToken()).toBe(true);
    expect(request).toHaveBeenCalledTimes(2);
  });

  it('deduplicates overlapping refresh requests', async () => {
    const controller = await mount('authentication', Authentication, '<div data-controller="authentication"></div>');
    let resolve;
    const request = vi.fn().mockReturnValue(new Promise(done => { resolve = done; }));
    vi.stubGlobal('fetch', request);
    const first = controller.getFreshJWTToken();
    expect(await controller.getFreshJWTToken()).toBe(false);
    resolve({ ok: true, json: async () => ({ expires_at: Math.floor(Date.now() / 1000) + 3600 }) });
    expect(await first).toBe(true);
    expect(request).toHaveBeenCalledOnce();
  });
});
