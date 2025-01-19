CREATE TABLE messageUser (
                             `messageID` INTEGER PRIMARY KEY AUTOINCREMENT, -- Unique identifier for the message
                             `sender`    INTEGER NOT NULL,                 -- User ID of the sender
                             `receiver`  INTEGER NOT NULL,                 -- User ID of the receiver
                             `content`   TEXT    NOT NULL,                 -- Message content
                             `created`   DATE    NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp of message creation

    -- Foreign key linking sender to the user table
                             FOREIGN KEY (sender) REFERENCES user (userID),

    -- Foreign key linking receiver to the user table
                             FOREIGN KEY (receiver) REFERENCES user (userID)
);

-- Create an index on (sender, receiver) for efficient query operations
CREATE INDEX messageUser_SR
    ON messageUser (sender, receiver);

-- Create an index on (receiver, sender) for efficient reverse query operations
CREATE INDEX messageUser_RS
    ON messageUser (receiver, sender);

-- Create the messageGroup table
CREATE TABLE messageGroup (
                              `messageID` INTEGER PRIMARY KEY AUTOINCREMENT, -- Unique identifier for the message
                              `sender`    INTEGER NOT NULL,                 -- User ID of the sender
                              `groupID`   INTEGER NOT NULL,                 -- Group ID where the message is sent
                              `content`   TEXT    NOT NULL,                 -- Message content
                              `created`   DATE    NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp of message creation

    -- Foreign key linking sender to the user table
                              FOREIGN KEY (sender) REFERENCES user (userID),

    -- Foreign key linking groupID to the group table
                              FOREIGN KEY (groupID) REFERENCES `group` (groupID)
);

-- Create an index on (sender, groupID) for efficient query operations
CREATE INDEX messageGroup_SR
    ON messageGroup (sender, groupID);

-- Create an index on (groupID, sender) for efficient reverse query operations
CREATE INDEX messageGroup_RS
    ON messageGroup (groupID, sender);
