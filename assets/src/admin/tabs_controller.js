import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    connect() {
        const tabbs = document.querySelectorAll('.tabs-group .tabs ul li');
        const items = document.querySelectorAll('.tabs-group .item');

        for (let tab of tabbs) {
            tab.addEventListener('click', () => {
                const target = tab.getAttribute('data-tab');
                // Remove 'is-active' class from all tabs
                tabbs.forEach(t => t.classList.remove('is-active'));
                // Add 'is-active' class to the clicked tab
                tab.classList.add('is-active');
                // Hide all tab contents
                items.forEach(content => content.style.display = 'none');
                // Show the content corresponding to the clicked tab
                const targetContent = document.getElementById(`tab-${target}`);
                if (targetContent) {
                    targetContent.style.display = 'block';
                }
            });
        }
    }
}