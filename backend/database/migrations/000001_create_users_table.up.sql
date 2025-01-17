PRAGMA foreign_keys = OFF;

-- Create the user table
CREATE TABLE `user` (
                        `userID`    INTEGER PRIMARY KEY AUTOINCREMENT,  -- Unique user identifier
                        `email`     TEXT NOT NULL UNIQUE COLLATE NOCASE, -- User email, case-insensitive
                        `password`  TEXT NOT NULL,                      -- User password
                        `firstname` TEXT NOT NULL,                      -- First name of the user
                        `lastname`  TEXT NOT NULL,                      -- Last name of the user
                        `nickname`  TEXT NOT NULL DEFAULT '',           -- User's nickname (optional)
                        `created`   DATE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Account creation timestamp
                        `image`     TEXT,                               -- Reference to profile image file
                        `about`     TEXT NOT NULL DEFAULT '',           -- About section for the user
                        `birthday`  DATE NOT NULL,                      -- User's date of birth
                        `private`   BOOLEAN NOT NULL DEFAULT FALSE,     -- Privacy setting for the user

    -- Foreign key linking image to the file table
                        FOREIGN KEY (image) REFERENCES file (token)
);

-- Insert a system user with a default record
INSERT INTO user (userID, email, password, firstname, lastname, image, birthday)
VALUES (0, '', '', 'System', '', null, CURRENT_TIMESTAMP);

-- Re-enable foreign key checks
PRAGMA foreign_keys = ON;
