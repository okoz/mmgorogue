-- Procedures used by mmgorogue.

DELIMITER //

DROP PROCEDURE IF EXISTS create_user;
CREATE PROCEDURE create_user(user_name CHAR(16), password CHAR(64), email CHAR(254))
BEGIN
	DECLARE EXIT HANDLER FOR 1062 SELECT 'User exists' AS status;

	INSERT INTO users(user_name, password, email)
	VALUES(user_name, SHA2(password, 512), email);

	SELECT 'OK' AS status;
END//

DROP PROCEDURE IF EXISTS authenticate;
CREATE PROCEDURE authenticate(user_name CHAR(16), password CHAR(64))
BEGIN
	IF EXISTS(SELECT * FROM users
	   	WHERE users.user_name = user_name
		AND users.password = SHA2(password, 512))
	THEN
		UPDATE users SET last_login = NOW() WHERE users.user_name = user_name;
		SELECT 1 AS 'success';
	ELSE SELECT 0 AS 'success';
	END IF;
END//

DELIMITER ;
