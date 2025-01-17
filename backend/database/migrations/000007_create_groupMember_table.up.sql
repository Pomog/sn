-- Create the `groupMember` table
CREATE TABLE groupMember (
                             groupID INTEGER NOT NULL,
                             userID INTEGER NOT NULL,

                             FOREIGN KEY (groupID) REFERENCES `group` (groupID) ON DELETE CASCADE,
                             FOREIGN KEY (userID) REFERENCES user (userID) ON DELETE CASCADE,

                             UNIQUE (groupID, userID)
);

-- Create an index to optimize queries involving userID and groupID
CREATE INDEX IF NOT EXISTS groupMember_reverse
    ON groupMember (userID, groupID);

-- Trigger to ensure users have an invite to private groups before joining
CREATE TRIGGER check_group_invite
    BEFORE INSERT ON groupMember
    WHEN EXISTS (
        SELECT 1
        FROM groupMember gm
        WHERE gm.groupID = NEW.groupID
    )
        AND NOT EXISTS (
            SELECT 1
            FROM `group` g
            WHERE g.groupID = NEW.groupID AND g.type = 'public'
        )
        AND NOT EXISTS (
            SELECT 1
            FROM groupInvite gi
            WHERE gi.groupID = NEW.groupID AND gi.receiverID = NEW.userID
        )
BEGIN
    SELECT RAISE(ROLLBACK, 'User does not have an invite to this private group');
END;

-- Trigger to remove group invites after the user accepts the invite
CREATE TRIGGER after_invite_accept
    AFTER INSERT ON groupMember
BEGIN
    DELETE FROM groupInvite
    WHERE receiverID = NEW.userID AND groupID = NEW.groupID;
END;

-- Trigger to prevent the group owner from leaving the group
CREATE TRIGGER group_owner_leave
    BEFORE DELETE ON groupMember
    WHEN (SELECT ownerID FROM `group` g WHERE g.groupID = OLD.groupID) = OLD.userID
BEGIN
    SELECT RAISE(ROLLBACK, 'Owner of the group cannot leave');
END;
