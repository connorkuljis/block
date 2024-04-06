-- cleanup 

DELETE 
FROM Tasks 
WHERE name = 'test' 
OR (name = '' AND actual_duration_minutes < 5);

SELECT changes();
