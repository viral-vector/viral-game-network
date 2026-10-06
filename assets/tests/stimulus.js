import { Application } from '@hotwired/stimulus';
import { afterEach } from 'vitest';

const applications = [];
afterEach(() => {
  for (const application of applications.splice(0)) application.stop();
});

export async function mount(identifier, Controller, markup) {
  document.body.innerHTML = markup;
  const application = Application.start();
  applications.push(application);
  application.register(identifier, Controller);
  for (let i = 0; i < 6; i++) await Promise.resolve();
  const element = document.querySelector(`[data-controller~="${identifier}"]`);
  const controller = application.getControllerForElementAndIdentifier(element, identifier);
  if (!controller) throw new Error(`Controller ${identifier} did not connect`);
  return controller;
}
