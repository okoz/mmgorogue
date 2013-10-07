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

DROP PROCEDURE IF EXISTS user_exists;
CREATE PROCEDURE user_exists(user_name CHAR(16))
BEGIN
	IF EXISTS(SELECT * FROM users WHERE users.user_name = user_name)
	THEN SELECT 1 AS 'exists';
	ELSE SELECT 0 AS 'exists';
	END IF;
END//

DROP PROCEDURE IF EXISTS user_inventory;
CREATE PROCEDURE user_inventory(user_name CHAR(16))
BEGIN
       SELECT COUNT(*) AS count, item_id, item_name
       FROM inventory
       INNER JOIN items ON inventory.item_id = items.id
       INNER JOIN users ON inventory.user_id = users.id
       WHERE users.user_name = user_name
       GROUP BY item_id;
END//

DROP PROCEDURE IF EXISTS user_give_item;
CREATE PROCEDURE user_give_item(user_name CHAR(16), item_name CHAR(64))
BEGIN
	INSERT INTO inventory VALUES(
	       (SELECT id FROM users WHERE users.user_name = user_name),
	       (SELECT id FROM items WHERE items.item_name = item_name)
	);
END//

DROP PROCEDURE IF EXISTS user_take_item;
CREATE PROCEDURE user_take_item(user_name CHAR(16), item_name CHAR(64))
BEGIN
       DELETE FROM inventory
       WHERE inventory.user_id = (SELECT id FROM users WHERE users.user_name = user_name)
       AND inventory.item_id = (SELECT id FROM items WHERE items.item_name = item_name)
       LIMIT 1;
END//

DELIMITER ;
