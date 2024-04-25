-- cleanup 

DELETE 
FROM Tasks 
WHERE task_name = 'test' 
OR (task_name = '' AND actual_duration_seconds < 100);

SELECT changes();
