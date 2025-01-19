CREATE TABLE postAllowedUser (
                                 `postID` INTEGER NOT NULL,  -- Identifier of the post
                                 `userID` INTEGER NOT NULL,  -- Identifier of the user allowed to access the post

                                 UNIQUE (postID, userID) ON CONFLICT REPLACE, -- Ensure unique post-user pairs, replacing on conflict

    -- Foreign key linking postID to the post table
                                 FOREIGN KEY (postID) REFERENCES post (postID),

    -- Foreign key linking userID to the user table
                                 FOREIGN KEY (userID) REFERENCES user (userID)
);
