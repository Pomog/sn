PRAGMA foreign_keys= off;

CREATE TABLE `user` (
                        `userID`    INTEGER PRIMARY KEY AUTOINCREMENT,
                        `email`     TEXT NOT NULL UNIQUE COLLATE NOCASE,
                        `password`  TEXT NOT NULL,
                        `firstname` TEXT NOT NULL,
                        `lastname`  TEXT NOT NULL,
                        `nickname`  TEXT NOT NULL DEFAULT '',
                        `created`   DATE NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `image`     TEXT,
                        `about`     TEXT NOT NULL DEFAULT '',
                        `birthday`  DATE NOT NULL,
                        `private`   BOOLEAN NOT NULL DEFAULT FALSE,

                        FOREIGN KEY (image) REFERENCES file (token)
);

PRAGMA foreign_keys= on;
