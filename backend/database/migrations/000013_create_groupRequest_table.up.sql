CREATE TABLE groupRequest (
                              `groupID`  INTEGER NOT NULL, -- Identifier of the group
                              `senderID` INTEGER NOT NULL, -- Identifier of the user requesting to join the group

    -- Foreign key linking groupID to the group table
                              FOREIGN KEY (groupID) REFERENCES `group` (groupID),

    -- Foreign key linking senderID to the user table
                              FOREIGN KEY (senderID) REFERENCES user (userID),

    -- Ensure unique group-user pairs, replacing on conflict
                              UNIQUE (groupID, senderID) ON CONFLICT REPLACE
);

-- Drop and recreate the after_invite_accept trigger to include groupRequest cleanup
DROP TRIGGER IF EXISTS after_invite_accept;

CREATE TRIGGER after_invite_accept
    AFTER INSERT ON groupMember
BEGIN
    -- Remove the corresponding groupInvite after the member is added
    DELETE FROM groupInvite
    WHERE receiverID = NEW.userID AND groupID = NEW.groupID;

    -- Remove the corresponding groupRequest after the member is added
    DELETE FROM groupRequest
    WHERE groupID = NEW.groupID AND senderID = NEW.userID;
END;
