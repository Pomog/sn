-- Users Table
CREATE TABLE IF NOT EXISTS users (
                                     id UUID PRIMARY KEY,
                                     email TEXT UNIQUE,
    -- pseudo TEXT UNIQUE,
                                     password TEXT,
                                     first_name TEXT,
                                     last_name TEXT,
                                     date_of_birth DATE,
                                     avatar_image TEXT,
                                     nickname TEXT,
                                     about_me TEXT,
                                     is_public BOOLEAN,
                                     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                     deleted_at TIMESTAMP
);

-- Followers Table
CREATE TABLE IF NOT EXISTS followers (
                                         id UUID PRIMARY KEY,
                                         follower_id INTEGER REFERENCES users(id),
                                         followee_id INTEGER REFERENCES users(id),
                                         status TEXT CHECK(status = 'requested' OR status = 'accepted' OR status = 'declined'),
                                         created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                         updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                         deleted_at TIMESTAMP
);

-- Posts Table
CREATE TABLE IF NOT EXISTS posts (
                                     id UUID PRIMARY KEY,
                                     group_id UUID,
                                     user_id INTEGER REFERENCES users(id),
                                     title TEXT,
                                     content TEXT,
                                     image_url TEXT,
                                     privacy TEXT CHECK(privacy = 'public' OR privacy = 'private' OR privacy = 'almost private' OR privacy = 'group'),
                                     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                     deleted_at TIMESTAMP
);


-- Groups Table
CREATE TABLE IF NOT EXISTS groups (
                                      id UUID PRIMARY KEY,
                                      title TEXT,
                                      description TEXT,
                                      banner_url TEXT,
                                      creator_id UUID REFERENCES users(id),
                                      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                      deleted_at TIMESTAMP
);

-- Group Members Table
CREATE TABLE IF NOT EXISTS group_members (
                                             id UUID PRIMARY KEY,
                                             group_id UUID REFERENCES groups(id),
                                             member_id UUID REFERENCES users(id),
                                             status TEXT CHECK(status = 'invited' OR status = 'requesting' OR status = 'accepted' OR status = 'declined'),
                                             role TEXT CHECK(role = 'admin' OR role = 'user'),
                                             created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                             updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                             deleted_at TIMESTAMP
);

-- Group Posts Table
CREATE TABLE groupPosts (
                            id UUID PRIMARY KEY,
                            group_id UUID REFERENCES Groups(id),
                            post_id UUID REFERENCES Posts(id),
                            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Events Table
CREATE TABLE IF NOT EXISTS events (
                                      id UUID PRIMARY KEY,
                                      group_id UUID REFERENCES groups(id),
                                      creator_id UUID REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
                                      title TEXT,
                                      description TEXT,
                                      date_time DATETIME,
                                      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                      deleted_at TIMESTAMP
);

-- Events Participant Table
CREATE TABLE IF NOT EXISTS events_participants (
                                                   id UUID PRIMARY KEY,
                                                   event_id UUID REFERENCES events(id),
                                                   member_id UUID REFERENCES users(id),
                                                   response TEXT CHECK(response = 'going' OR response = 'not_going'),
                                                   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                                   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                                   deleted_at TIMESTAMP
);

-- Private Messages Table
CREATE TABLE IF NOT EXISTS private_messages (
                                                id UUID PRIMARY KEY,
                                                sender_id UUID REFERENCES users(id),
                                                receiver_id UUID REFERENCES users(id),
                                                content TEXT,
                                                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                                deleted_at TIMESTAMP
);

-- Group Messages Table
CREATE TABLE IF NOT EXISTS group_messages (
                                              id UUID PRIMARY KEY,
                                              group_id UUID REFERENCES groups(id),
                                              sender_id UUID REFERENCES users(id),
                                              content TEXT,
                                              created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                              updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                              deleted_at TIMESTAMP
);

-- Notifications Table
CREATE TABLE IF NOT EXISTS notifications (
                                             id UUID PRIMARY KEY,
                                             user_id UUID REFERENCES users(id),
                                             group_id UUID,
                                             concern_id UUID,
                                             member_id UUID,
                                             is_invite BOOLEAN DEFAULT FALSE,
                                             type TEXT CHECK(type = 'follow_request'OR type = 'follow_accepted' OR type = 'follow_declined' OR type = 'unfollow' OR type = 'group_invitation' OR type = 'new_message' OR type = 'new_event'),
                                             message TEXT,
                                             created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                             deleted_at TIMESTAMP
);

-- Post Comments
CREATE TABLE comments (
                          id UUID PRIMARY KEY,
                          user_id UUID REFERENCES users(id),
                          post_id UUID REFERENCES posts(id),
                          content TEXT,
                          image_url TEXT,
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                          deleted_at TIMESTAMP
);

-- Posts Selected Users Table
CREATE TABLE selected_users (
                                id UUID PRIMARY KEY,
                                user_id UUID REFERENCES users(id),
                                post_id UUID REFERENCES posts(id),
                                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                deleted_at TIMESTAMP
);

-- Invitations Table
CREATE TABLE IF NOT EXISTS invitations (
                                           id UUID PRIMARY KEY,
                                           inviting_user_id UUID REFERENCES users(id),
                                           invited_user_id UUID REFERENCES users(id),
                                           group_member_id UUID REFERENCES group_members(id),
                                           created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                           updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                           deleted_at TIMESTAMP
);

-- Sessions Table
CREATE TABLE sessions (
                          id UUID PRIMARY KEY,
                          user_id UUID  REFERENCES Users(id),
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                          expiration_date TIMESTAMP,
                          deleted_at TIMESTAMP
);