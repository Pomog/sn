PRAGMA foreign_keys = OFF;

-- Create the `comment` table
CREATE TABLE `comment` (
                           `commentID` INTEGER PRIMARY KEY AUTOINCREMENT, -- Unique identifier for each comment
                           `postID`    INTEGER NOT NULL,                 -- Foreign key linking to post
                           `authorID`  INTEGER NOT NULL,                 -- Foreign key linking to user

                           `content`   TEXT NOT NULL,                    -- Content of the comment
                           `images`    TEXT NOT NULL DEFAULT '',         -- Associated images (optional)

                           `created`   DATE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Creation timestamp

    -- Foreign key constraints
                           FOREIGN KEY (postID) REFERENCES post (postID),
                           FOREIGN KEY (authorID) REFERENCES user (userID)
);

-- Create an index on the `postID` column
CREATE INDEX IF NOT EXISTS comment_index
    ON comment (postID);

-- Re-enable foreign key checks
PRAGMA foreign_keys = ON;
