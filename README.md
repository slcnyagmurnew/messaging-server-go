# Messaging Server

Scheduler-driven messaging service written in Go
that automatically sends unsent messages from a PostgreSQL database at fixed intervals, records delivery details in Redis, and provides a start/stop control via HTTP API.

## Features

- **Scheduled Delivery:** Triggers every 2 minutes to fetch and send the first two unsent messages in sequence.
- **PostgreSQL Integration:** Marks messages as sent and updates their status in the database.
- **Redis Logging:** Stores each delivered message’s ID and timestamp in Redis, with the message ID as the key.
- **Start/Stop Control:** Start or stop the scheduler on demand using an HTTP endpoint.


- **API Endpoints:** (also see Swagger UI on ```localhost:8080/swagger/index.html```)
  - `/health` – Service health check

  - `/` Index page

  - `/swagger/*any` – Swagger UI

  - `/api/v1/messaging` – Control scheduler (POST)

  - `/api/v1/list` – List all sent messages (GET)

- ### Scheduler
    The scheduler loop runs in its own goroutine, isolated from the main HTTP server. It uses a ```context.WithCancel``` and 
    a ```done``` channel to manage graceful startup and shutdown without blocking the main process. A Ticker triggers the job function 
    at fixed intervals, with each invocation executed in its own goroutine (go job(ctx)), ensuring long-running tasks don't block subsequent ticks.

  - The Start method checks if the scheduler is already running by inspecting the cancel function, preventing concurrent starts.

  - The Stop method signals cancellation, waits on the done channel for the loop to exit, and resets internal state to allow future restarts.

- ### Graceful Shutdown
    Messaging server provides graceful shutdowns for the cases:

    - When Stop method hits to scheduler even some jobs are in running statement,
  application has ```WaitGroup``` to handle statuses of concurrent jobs and
  waits for go routines to finish their execution. Even server is not running under the pressure
  so that rapidly do fetch/update/insert operations, we can test graceful shutdown
  process with uncomment ```Sleep``` statement in ```SendMessages``` job.
  
    - HTTP server runs in background after setting up gin-gonic server and starts to listen
  ```SIGINT``` and ```SIGTERM``` signals to halt the program. It waits for 30 seconds
  when any HTTP requests still active in any endpoint before exit.

- ### Table Design
    Messages table takes place under ```mydb.public``` schema. Fifteen records will be appended to table when app initially runs. 
  ### `messages`

| Column  | Type            | Default                          | Description                                 |
| ------- | --------------- | -------------------------------- | ------------------------------------------- |
| `id`    | `integer`       | `nextval('messages_id_seq')`     | Unique message identifier (PK, auto-increment) |
| `content` | `varchar(200)` | —                                | The body/text of the message                |
| `phone` | `varchar(50)`   | —                                | Recipient phone number                      |
| `issent` | `boolean`      | `false`                          | Whether this message has been sent (flag)   |


## Prerequisites

- Go 1.24.3

- PostgreSQL 17

- Redis 8

## Configuration

The service relies on several environment variables (e.g. in your `docker-compose.yml`) to connect to external systems and control its behavior. Add the following under your service’s `environment:` section:


- **POSTGRES_URI**: PostgreSQL connection string (e.g., ```postgres://user:pass@host:5432/dbname?sslmode=disable```)
- **REDIS_HOST:** Redis address (e.g., ```localhost:6379```)
- **REDIS_TTL:** Default time-to-live for cache entries in Redis  (e.g., ```60 in seconds```)
- **SCHEDULER_INTERVAL:** Interval at which the internal scheduler polls for new work (e.g., ```120 in seconds```)
- **LOG_LEVEL:** Set logging level in uppercase. (e.g., ```INFO```)
- **ENVIRONMENT:** Set working environment. Set this variable as *DEVELOPMENT* to enable DEV mode. (e.g., ```PRODUCTION```)
- **WEBHOOK_URL:** Webhook endpoint for sending messages. Note that webhook.site URLs are limited to 100 requests total.
 If problem occurs with too many requests, set new webhook url by editing configurations like below:

  ![img.png](assets/webhook.png)

## Installation
1. Clone the repository
    ```
   git clone https://github.com/yourorg/messaging-server.git
    cd messaging-server
   ```
2. Start docker compose
    ```
   docker compose up -f deployments/docker-compose.yaml up -d --build
   ```
3. Runs on:
    ```
   http://localhost:8080/
   ```

## Project with Nested Directory

```
.
├── build
│   └── Dockerfile
├── cmd
│   └── main.go
├── deployments
│   └── docker-compose.yaml
├── docs
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── go.mod
├── go.sum
├── init
│   └── init.sql
└── internal
    ├── handlers
    │   ├── base.go
    │   └── messages.go
    ├── db
    │   ├── postgres_builder.go
    │   └── redis_builder.go
    ├── jobs
    │   └── send_message.go
    ├── model
    │   └── message.go
    ├── logger
    │   ├── logger.go
    ├── router
    │   ├── router.go
    └── scheduler
        └── scheduler.go
```


## Clean Up
To stop and remove all containers, networks, and associated volumes created by this application, run:
```
docker compose -f deployments/docker-compose.yaml down --volumes
```
