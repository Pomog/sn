CREATE TABLE event (
                       `eventID` INTEGER PRIMARY KEY AUTOINCREMENT, -- Unique identifier for the event
                       `groupID` INTEGER NOT NULL,                 -- Group to which the event belongs
                       `authorID` INTEGER NOT NULL,                -- User who created the event

                       `title` TEXT NOT NULL,                      -- Title of the event
                       `about` TEXT NOT NULL,                      -- Description of the event
                       `time` DATE NOT NULL,                       -- Scheduled time of the event
                       `created` DATE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Creation timestamp

    -- Foreign key linking groupID to the group table
                       FOREIGN KEY (groupID) REFERENCES "group" (groupID),

    -- Foreign key linking authorID to the user table
                       FOREIGN KEY (authorID) REFERENCES user (userID)
);
