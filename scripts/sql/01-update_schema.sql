



-- Step 1: Create a new table with the desired schema
CREATE TABLE Tasks_new(
    task_id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_name TEXT NOT NULL,
    estimated_duration_seconds INTEGER NOT NULL,
    actual_duration_seconds INTEGER,
    blocker_enabled INTEGER DEFAULT 0,
    screen_enabled INTEGER DEFAULT 0,
    screen_url TEXT,
    created_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP,
    completed INTEGER,
    completion_percent REAL
);

-- Step 2: Copy data from the old table to the new one
INSERT INTO Tasks_new
SELECT 
    id AS task_id,
    name AS task_name,
    planned_duration_minutes * 60 AS estimated_duration_seconds,
    actual_duration_minutes * 60 AS actual_duration_seconds, 
    blocker_enabled,
    screen_enabled,
    screen_url,
    created_at,
    finished_at,
    completed,
    completion_percent
FROM Tasks;

-- Step 3: Drop the old table
DROP TABLE Tasks;

-- Step 4: Rename the new table
ALTER TABLE Tasks_new RENAME TO Tasks;
