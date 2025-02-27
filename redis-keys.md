# Redis Keys Structure for Cloudflare TinyURL

## **1️⃣ URL Shortening Storage**

| **Key Pattern**             | **Purpose**                         | **Data Type** |
| --------------------------- | ----------------------------------- | ------------- |
| `shortURL:<shortURL>`       | Maps short URL to long URL          | `SET`         |
| `count:<shortURL>:all_time` | Stores total access count           | `INCR`        |
| `count:<shortURL>:24h`      | Stores 24-hour rolling access count | `INCR`        |
| `count:<shortURL>:week`     | Stores weekly rolling access count  | `INCR`        |

---

## **2️⃣ Click Event Tracking**

| **Key Pattern**       | **Purpose**                                | **Data Type**        |
| --------------------- | ------------------------------------------ | -------------------- |
| `click:<shortURL>:<snowflakeID>` | Tracks individual click event (24h) | `SET` (TTL: 24h) |
| `click:<shortURL>:<snowflakeID>:week` | Tracks individual click event (week) | `SET` (TTL: 7 days) |
| `expired_click_queue` | Stores expired click events for processing | `LIST (LPUSH/BRPOP)` |

---

## **3️⃣ Global Counters**

| **Key Pattern**        | **Purpose**                              | **Data Type** |
| ---------------------- | ---------------------------------------- | ------------- |
| `url_global_counter`   | Unique counter for generating short URLs | `INCR`        |
| `count:<shortURL>:all_time` | Tracks total clicks for a short URL | `INCR` |
| `count:<shortURL>:24h` | Tracks last 24-hour click count for a short URL | `INCR` |
| `count:<shortURL>:week` | Tracks last 7-day click count for a short URL | `INCR` |

---


