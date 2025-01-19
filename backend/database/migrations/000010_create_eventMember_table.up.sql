CREATE TABLE eventMember (
                             `eventID` INTEGER NOT NULL,
                             `userID` INTEGER NOT NULL,
                             `status` TEXT NOT NULL CHECK (`status` IN ('GOING', 'NOT_GOING')),

    -- Ensure each user can only have one row per event; replace on conflict
                             UNIQUE (eventID, userID) ON CONFLICT REPLACE,

    -- Foreign key constraint linking eventID to the event table
                             FOREIGN KEY (eventID) REFERENCES event (eventID),

    -- Foreign key constraint linking userID to the user table
                             FOREIGN KEY (userID) REFERENCES user (userID)
);

-- Trigger to remove the row if the inserted status is NULL
CREATE TRIGGER eventMember_remove_if_null
    BEFORE INSERT ON eventMember
    WHEN NEW.status IS NULL
BEGIN
    -- Delete the existing row with the same eventID and userID
    DELETE FROM eventMember
    WHERE eventID = NEW.eventID AND userID = NEW.userID;

    -- Prevent the new row from being inserted
    SELECT RAISE(IGNORE);
END;

-- Trigger to clean up eventMember rows when a user leaves a group
CREATE TRIGGER group_leave_events
    AFTER DELETE ON groupMember
BEGIN
    -- Remove rows from eventMember for the events of the group the user left
    DELETE FROM eventMember
    WHERE (SELECT groupID FROM event e WHERE e.eventID = eventMember.eventID) = OLD.groupID
      AND eventMember.userID = OLD.userID;
END;
