# Pipeline 28 — PostgreSQL Source → MariaDB Destination

**Source streams:** 3 | **Destination:** MariaDB

> dbt SQL identical to `26-postgres-to-postgres.md`. DDL same as `27-postgres-to-mysql.md`.

---

## Connections
```json
{ "host":"src-host","port":5432,"database":"mydb","username":"reader","password":"..","ssl_mode":"disable" }
{ "host":"dest-host","port":3306,"database":"analytics","username":"writer","password":"..." }
```

---

## Destination DDLs (MariaDB)

```sql
CREATE TABLE analytics.dim_users (
    user_uuid      VARCHAR(36)  PRIMARY KEY,
    email_address  VARCHAR(255),
    full_name      VARCHAR(500),
    city           VARCHAR(255),
    country        VARCHAR(100),
    account_tier   VARCHAR(50),
    registered_on  DATE, last_active DATE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE analytics.fact_orders (
    order_id        VARCHAR(36) PRIMARY KEY,
    customer_ref    VARCHAR(36),
    order_amount    DECIMAL(10,2),
    order_status    VARCHAR(50),
    payment_method  VARCHAR(100),
    channel         VARCHAR(100),
    currency        VARCHAR(10),
    is_high_value   TINYINT(1),
    placed_on       DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.fact_payments (
    payment_id    VARCHAR(36) PRIMARY KEY,
    order_ref     VARCHAR(36),
    pay_method    VARCHAR(100),
    pay_amount    DECIMAL(10,2),
    pay_status    VARCHAR(50),
    is_successful TINYINT(1),
    paid_on       DATE
) ENGINE=InnoDB;
```
