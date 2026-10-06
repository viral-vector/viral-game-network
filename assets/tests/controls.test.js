import { it, expect, vi } from 'vitest';
import Pager from '../src/admin/pager_controller';
import Search from '../src/admin/search_controller';
import Tabs from '../src/admin/tabs_controller';
import Header from '../src/admin/header_footer_controller';
import Prompt from '../src/admin/modal_prompt_controller';
import Progress from '../src/admin/progress_controller';
import Notification from '../src/admin/notification_controller';
import ApiKey from '../src/admin/api_key_controller';
import Dashboard from '../src/admin/dashboard_controller';
import { mount } from './stimulus';

vi.mock('chart.js/auto', () => ({ default: vi.fn() }));

it('paginates while preserving search and continuation tokens', async () => {
  const controller = await mount('pager', Pager, '<nav data-controller="pager" data-pager-route-value="https://example.com/admin/users?search=alice" data-pager-pages-value="9" data-pager-paged-value="5" data-pager-page-token-value="cursor"><ul class="pagination-list"></ul></nav>');
  const url = new URL(controller.createUrl(6));
  expect(url.searchParams.get('search')).toBe('alice');
  expect(url.searchParams.get('page')).toBe('6');
  expect(url.searchParams.get('pageToken')).toBe('cursor');
  expect(controller.element.querySelector('.is-current').textContent).toBe('5');
  expect(controller.element.querySelectorAll('.pagination-ellipsis').length).toBe(2);
});

it('rejects search strings shorter than three characters', async () => {
  vi.stubGlobal('pushNotification', vi.fn());
  const controller = await mount('search', Search, '<div data-controller="search" data-search-action-value="https://example.com/admin/users"><input data-search-target="control" value="ab"><button data-search-target="button"></button></div>');
  controller.buttonTarget.click();
  expect(window.pushNotification).toHaveBeenCalledWith(expect.objectContaining({ type: 'warning' }));
  expect(controller.buttonTarget.disabled).toBe(false);
});

it('switches tabs and visible content', async () => {
  await mount('tabs', Tabs, '<div data-controller="tabs" class="tabs-group"><div class="tabs"><ul><li data-tab="one" class="is-active"></li><li data-tab="two"></li></ul></div><div id="tab-one" class="item"></div><div id="tab-two" class="item" style="display:none"></div></div>');
  document.querySelector('[data-tab="two"]').click();
  expect(document.getElementById('tab-one').style.display).toBe('none');
  expect(document.getElementById('tab-two').style.display).toBe('block');
});

it('highlights the current menu and toggles the navigation burger', async () => {
  await mount('header', Header, '<div data-controller="header" data-header-route-value="/admin/users"><a class="menu-item" href="/admin/users"></a><a class="menu-item" href="/admin/apps"></a><button class="navbar-burger" data-target="menu"></button><div id="menu"></div></div>');
  expect(document.querySelector('[href="/admin/users"]').classList.contains('is-active')).toBe(true);
  expect(document.querySelector('[href="/admin/apps"]').classList.contains('is-active')).toBe(false);
  document.querySelector('.navbar-burger').click();
  expect(document.getElementById('menu').classList.contains('is-active')).toBe(true);
});

it('resolves or rejects the confirmation prompt and hides it', async () => {
  const controller = await mount('prompt', Prompt, '<div data-controller="prompt"><span data-prompt-target="title"></span><span data-prompt-target="message"></span><button data-prompt-target="accept"></button><button data-prompt-target="reject"></button></div>');
  const accepted = window.prompt.show({ title: 'Delete?' });
  expect(controller.titleTarget.textContent).toBe('Delete?');
  expect(controller.element.classList.contains('is-active')).toBe(true);
  controller.acceptTarget.click();
  await expect(accepted).resolves.toBeUndefined();
  const rejected = window.prompt.show({});
  const assertion = expect(rejected).rejects.toBeUndefined();
  controller.rejectTarget.click();
  await assertion;
  expect(controller.element.classList.contains('is-active')).toBe(false);
});

it('shows progress and bounds its value', async () => {
  vi.useFakeTimers();
  const controller = await mount('progress', Progress, '<div data-controller="progress"><progress data-progress-target="progress"></progress></div>');
  window.progress.show();
  vi.advanceTimersByTime(100);
  expect(controller.progressTarget.style.display).toBe('block');
  controller.set(150);
  expect(controller.value).toBe(100);
  window.progress.hide();
  vi.advanceTimersByTime(100);
  expect(controller.progressTarget.style.display).toBe('none');
});

it('prioritizes notifications and converts SSE events into messages', async () => {
  vi.useFakeTimers();
  let stream;
  vi.stubGlobal('EventSource', class { constructor(url) { this.url = url; stream = this; } close() {} });
  const controller = await mount('notification', Notification, '<div class="is-hidden" data-controller="notification"><span data-notification-target="message"></span><span data-notification-target="pending"></span><button class="delete"></button></div>');
  window.pushNotification({ type: 'info', message: 'low', priority: 0 });
  window.pushNotification({ type: 'danger', message: 'high', priority: 2 });
  controller.Notifications();
  expect(controller.messageTarget.textContent).toBe('high');
  expect(controller.pendingTarget.textContent).toBe('1');
  stream.onmessage({ data: JSON.stringify({ severity: 'warning', message: 'cluster event' }) });
  expect(controller.messages).toContainEqual({ type: 'warning', message: 'cluster event' });
});

it('adds API keys from server-rendered HTML and removes deleted keys', async () => {
  vi.useFakeTimers();
  const controller = await mount('api-key', ApiKey, '<div data-controller="api-key"><button data-api-key-target="addkey"></button><form data-api-key-target="addkeyForm" style="display:none"></form><div data-api-key-target="keylist"></div></div>');
  controller.addkeyTarget.click();
  expect(controller.addkeyFormTarget.style.display).toBe('block');
  controller.addkeyFormTarget.dispatchEvent(new CustomEvent('form-process', { detail: { render: '<div class="list-item"><div class="list"><form></form></div></div>' } }));
  const form = controller.keylistTarget.querySelector('form');
  form.dispatchEvent(new CustomEvent('form-process', { detail: {} }));
  vi.advanceTimersByTime(500);
  expect(controller.keylistTarget.children.length).toBe(0);
});

it('constructs dashboard charts only when chart elements exist', async () => {
  const Chart = (await import('chart.js/auto')).default;
  vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue({});
  await mount('dashboard', Dashboard, '<div data-controller="dashboard"><canvas id="chart-sample-a"></canvas><canvas id="chart-sample-b"></canvas></div>');
  expect(Chart).toHaveBeenCalledTimes(2);
  expect(Chart.mock.calls.map(([, options]) => options.type)).toEqual(['bar', 'line']);
});
