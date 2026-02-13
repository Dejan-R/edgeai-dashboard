// script.js - real-time dashboard for fault visualization
let faultChart = null;
let ws = null;
let pendingMessages = [];
let isChartReady = false;

document.addEventListener('DOMContentLoaded', () => {
    initEmptyChart();   // 1. Chart
    initWebSocket();    // 2. WebSocket  
    loadInitial();      // 3. History from database
});

// Chart initialization
function initEmptyChart() {
    const ctx = document.getElementById("faultChart");
    faultChart = new Chart(ctx, {
        type: 'scatter',
        data: {
            datasets: [{
                data: [],
                pointRadius: 4,
                backgroundColor: 'red',
                showLine: false
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: false,
            plugins: {
                legend: { display: false },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            const y = context.raw.y;
                            const x = new Date(context.raw.x);
                            const dateStr = x.toLocaleDateString('hr-HR');  
                            const timeStr = x.toLocaleTimeString('hr-HR', { hour12: false }); 
                            return `${y.toFixed(2)} at ${dateStr} ${timeStr}`;
                        }
                    }
                }
            },
            scales: {
                x: {
                    type: 'time',
                    time: {
                        unit: 'minute',
                        displayFormats: {
                            minute: 'dd.MM HH:mm',
                            hour: 'dd.MM HH:mm'
                        },
                        tooltipFormat: 'dd.MM.yyyy HH:mm:ss'
                    },
                    title: { display: true, text: 'Time' },
                    ticks: { autoSkip: true, maxTicksLimit: 10 },
                    bounds: 'data'
                },
                y: {
                    min: 0,
                    max: 400,
                    title: { display: true, text: 'Anomaly Score' }
                }
            }
        }
    });

    isChartReady = true;
    flushPending();
}


// WebSocket
function initWebSocket() {
    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
    ws = new WebSocket(`${protocol}//${location.host}/ws`);

    ws.onopen = () => console.log("WS connected");
    ws.onmessage = evt => {
        try { handleIncoming(JSON.parse(evt.data)); }
        catch(e) { console.error("Bad WS message", evt.data); }
    };
    ws.onclose = () => setTimeout(initWebSocket, 1000);
}

// WS message handling
function handleIncoming(msg) {
    if (msg.type === 'STATUS') { updateStatus(msg.data); return; }
    if (!isChartReady) { pendingMessages.push(msg); return; }
    if (msg.type === 'NEW_FAULT') {
        addFault(msg.data);
        appendToChart(msg.data);
    }
}

function flushPending() {
    while(pendingMessages.length) handleIncoming(pendingMessages.shift());
}
// Load initial data from database
async function loadInitial() {
    try {
        const res = await fetch('/api/faults');
        const data = await res.json();
        if (!Array.isArray(data)) return;
        data.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
        data.forEach(appendToChart);
        data.forEach(addFault);
    } catch (e) {
        console.error("Failed to load initial data", e);
    }
}

// Add point to chart
function appendToChart(e) {
    const t = new Date(e.timestamp);
    if (isNaN(t.getTime())) return;
    if (faultChart.data.datasets[0].data.some(p => p.x.getTime() === t.getTime() && p.y === e.score)) return;
    faultChart.data.datasets[0].data.push({ x: t, y: e.score });
    if (faultChart.data.datasets[0].data.length > 100)
        faultChart.data.datasets[0].data = faultChart.data.datasets[0].data.slice(-100);
    faultChart.update();
}
// Add entry to left panel
function addFault(e) {
    const container = document.getElementById("sensor-data");
    const template = document.getElementById("fault-template");
    const clone = template.content.cloneNode(true);
    const statusEl = clone.querySelector(".status");
    statusEl.textContent = e.status;
    statusEl.style.color = e.status === "FAULT" ? "red" : "green";
    const t = new Date(e.timestamp);
    const dateStr = t.toLocaleDateString('hr-HR'); 
     const timeStr = t.toLocaleTimeString('hr-HR', { hour12: false }); 
     clone.querySelector(".fault-time").textContent = `${dateStr} ${timeStr}`;
    clone.querySelector(".fault-score").textContent = e.score.toFixed(2);
    container.prepend(clone);
}



// MQTT / DB status
function updateStatus(d) {
    const mqtt = document.getElementById("mqtt-status");
    const db = document.getElementById("db-status");
    mqtt.textContent = d.mqtt;
    mqtt.style.color = d.mqtt === 'CONNECTED' ? 'green' : 'red';
    db.textContent = d.db;
    db.style.color = d.db === 'OK' ? 'green' : 'red';
}
