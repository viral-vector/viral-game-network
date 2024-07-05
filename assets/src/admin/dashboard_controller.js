import { Controller } from "@hotwired/stimulus"
import Chart from 'chart.js/auto';

export default class extends Controller {
    connect() {
        this.buildAForms()
        this.buildCharts()
    }

    buildCharts() {
        const chartA = new Chart(document.getElementById('chart-sample-a').getContext('2d'), {
            type: 'bar',
            data: {
                labels: ['Red', 'Blue', 'Yellow', 'Green', 'Purple', 'Orange'],
                datasets: [{
                    label: '# of Votes',
                    data: [12, 19, 3, 5, 2, 3],
                    backgroundColor: [
                        'rgba(255, 99, 132, 0.2)',
                        'rgba(54, 162, 235, 0.2)',
                        'rgba(255, 206, 86, 0.2)',
                        'rgba(75, 192, 192, 0.2)',
                        'rgba(153, 102, 255, 0.2)',
                        'rgba(255, 159, 64, 0.2)'
                    ],
                    borderColor: [
                        'rgba(255, 99, 132, 1)',
                        'rgba(54, 162, 235, 1)',
                        'rgba(255, 206, 86, 1)',
                        'rgba(75, 192, 192, 1)',
                        'rgba(153, 102, 255, 1)',
                        'rgba(255, 159, 64, 1)'
                    ],
                    borderWidth: 1
                }]
            },
            options: {
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });

        const chartB = new Chart(document.getElementById('chart-sample-b').getContext('2d'), {
            type: 'line',
            data: {
                labels: ['M', 'T', 'W', 'T', 'F', 'S', 'S'],
                datasets: [{
                    label: 'Orders Shipped',
                    data: [13, 19, 9, 13, 6, 3, 7],
                    backgroundColor: "rgba(26, 100, 156,0.6)"
                }, {
                    label: 'Orders Placed',
                    data: [5, 29, 5, 5, 2, 3, 10],
                    backgroundColor: "rgba(71, 175, 225,0.4)"
                }]
            }
        });
    }

    buildAForms() {
        document.querySelectorAll('form').forEach(element => {
            element.addEventListener('submit', async event => {
                event.preventDefault();

                let execute = true;
                if (event.target.dataset.confirmMessage) {
                    try {
                        let prompt = {
                            title: event.target.dataset.confirmTitle,
                            message: event.target.dataset.confirmMessage,
                            accept: event.target.dataset.confirmAccept,
                            reject: event.target.dataset.confirmReject
                        };
                        await window.showPrompt(prompt);
                    } catch (error) {
                        execute = false;
                    }
                }
                if (!execute) {
                    return;
                }

                fetch(event.target.getAttribute('action'), {
                    method: event.target.getAttribute('method'),
                    body: new FormData(event.target)
                }).then(response => {
                    if(!response.ok) {
                        throw new Error(`Error ${response.status} - ${response.statusText}`);
                    }
                    return response.json();
                }).then(data => {
                    console.info(data);
                }).catch(error => {
                    console.error(error);
                })
            })
        })
    }
}