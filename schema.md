# Database Schema

## Users Table

| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| id | INTEGER | PRIMARY KEY AUTOINCREMENT | Unique identifier for the user |
| username | TEXT | UNIQUE NOT NULL | User's login name |
| password_hash | TEXT | NOT NULL | Hashed password |
| display_name | TEXT | | Optional display name |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | Account creation time |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP | Last update time |
