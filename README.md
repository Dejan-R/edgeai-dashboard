## EdgeAI_RPi: Real-Time Motor Vibration Dashboard

This project provides a Go backend + TimescaleDB + Web frontend to visualize and manage data from an embedded Edge AI node (nRF5340 + nRF7002) that detects industrial motor vibration anomalies in real-time.

The system collects MQTT messages from the Edge AI device, stores them in TimescaleDB, and presents a real-time dashboard in the browser with historical logs and live anomaly alerts.

---

## Features
 - **Go backend:**

     - MQTT subscriber receiving JSON messages from the Edge AI node
     - REST API with BASIC authentication
     - GWebSocket server for real-time frontend updates
     - TimescaleDB integration for historical data storage

-  **Frontend (HTML/JS/CSS):**
    - Real-time charts with Chart.js
    - WebSocket updates for live anomaly detection
    - Heartbeat monitoring and automatic reconnect
    - Simple, responsive dashboard for visualizing motor status

- **Database:**
    - TimescaleDB hypertables for efficient time-series storage
    - SQL setup script included (db/setup.sql)

---

## Project Structure
```text

EdgeAI_RPi/
│
├─ backend/
│   ├─ main.go           # Go backend: MQTT subscriber + HTTP server
│   ├─ config.go         # Configuration (MQTT, DB, ports, bcrypt hash)
│   ├─ mqtt.go           # MQTT logic + fault channel
│   ├─ db.go             # DB helper functions
│   ├─ api.go            # REST API endpoints + BASIC auth middleware
│   ├─ websocket.go      # WebSocket + heartbeat
│   ├─ go.mod
│   ├─ templates/
│   │   └─ index.html    # Web dashboard
│   └─ static/
│       ├─ style.css
│       └─ script.js     # WebSocket + Chart.js + heartbeat + reconnect
│
├─ db/
│   └─ setup.sql         # SQL script to create tables / hypertables
│
└─ README.md

```
> **How It Works**

Edge Node – Embedded nRF5340 + nRF7002 device monitors motor vibrations, classifies anomalies using TinyML, and publishes JSON messages via MQTT.

Backend – Go server subscribes to MQTT topics, stores data in TimescaleDB, and serves frontend via HTTP + WebSocket.

Frontend – Displays live charts and historical logs in a web browser, showing motor status and detected anomalies.

Database – TimescaleDB efficiently handles high-frequency time-series data for storage and queries.


> **Quick Start with Docker**
 - Clone the repository:

    git clone https://github.com/Dejan-R/EdgeAI_RPi.git
    cd EdgeAI_RPi

- Create a .env file with your configuration (MQTT broker, DB credentials, ports).

- Start the stack:

    docker-compose up -d

- Open a browser to http://localhost:8080
 to view the dashboard.

- Stop the stack:

    docker-compose down

All services (backend + DB + frontend) run in isolated containers, making setup fast and reproducible.


> **Requirements (without Docker)**

- Go 1.20+
- TimescaleDB / PostgreSQL
- MQTT broker accessible to Edge AI device
- Web browser (Chrome, Edge, Firefox)
- Optional: Docker for easier backend + DB setup


> **MQTT Testing / Example Messages**

- Default login for REST / WebSocket dashboard:

    - username: admin
    - password: admin123

- MQTT topic to subscribe for motor fault messages:

    - TOPIC: edgeai/fault

- Example JSON payloads sent from MCU (nRF5340DK) to backend:
```json
    {
    "score": 312.4,
    "threshold": 300,
    "status": "FAULT"
    }
  ```
  ```json
    {
    "score": 212.4,
    "threshold": 300,
    "status": "OK"
    }
```
Note: The dashboard expects the MCU to publish JSON messages using a TinyML model trained with Edge Impulse.

You can use MQTT Explorer or any MQTT client to monitor real-time messages.

## Author / Contact

**Dejan Rakijasic**  
LinkedIn: [https://www.linkedin.com/in/dejan-rakijasic/](https://www.linkedin.com/in/dejan-rakijasic/)
