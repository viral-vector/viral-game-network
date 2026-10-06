import { beforeEach, describe, it, expect, vi } from 'vitest';
import Authentication from '../src/admin/authentication_controller';
import { mount } from './stimulus';

function token(expires) {
  return `header.${btoa(JSON.stringify({ exp: expires }))}.signature`;
}

describe('authentication', () => {
  beforeEach(() => vi.useFakeTimers());

  it('refreshes only when a token is close to expiration', async () => {
    const controller = await mount('authentication', Authentication, '<div data-controller="authentication"></div>');
    const now = Math.floor(Date.now() / 1000);
    for (const [expires, expected] of [[now + 60, true], [now + 600, false], [now - 60, false]]) {
      document.cookie = `VNET_SESSION=${token(expires)}; path=/`;
      expect(controller.isJWTTokenExpire()).toBe(expected);
    }
  });

  it('handles missing and malformed cookies without throwing', async () => {
    const controller = await mount('authentication', Authentication, '<div data-controller="authentication"></div>');
    expect(controller.isJWTTokenExpire()).toBeFalsy();
    document.cookie = 'VNET_SESSION=invalid; path=/';
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

  it('reads encoded cookies without matching similar names', async () => {
    const controller = await mount('authentication', Authentication, '<div data-controller="authentication"></div>');
    document.cookie = 'otherVNET_SESSION=wrong; path=/';
    document.cookie = 'VNET_SESSION=hello%20world; path=/';
    expect(controller.getCookie('VNET_SESSION')).toBe('hello world');
    expect(controller.getCookie('missing')).toBeUndefined();
  });
});
