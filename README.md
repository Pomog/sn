### install golang-migrate CLI
```
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate --version
```
```
migrate create -ext sql -dir database/migrations -seq create_users_table
migrate create -ext sql -dir database/migrations -seq create_session_table
migrate create -ext sql -dir database/migrations -seq create_post_table
migrate create -ext sql -dir database/migrations -seq create_file_table
migrate create -ext sql -dir database/migrations -seq create_group_table
migrate create -ext sql -dir database/migrations -seq create_groupInvite_table
migrate create -ext sql -dir database/migrations -seq create_groupMember_table
migrate create -ext sql -dir database/migrations -seq create_follow_table
migrate create -ext sql -dir database/migrations -seq create_followRequest_table
migrate create -ext sql -dir database/migrations -seq create_event_table
migrate create -ext sql -dir database/migrations -seq create_eventMember_table
migrate create -ext sql -dir database/migrations -seq create_event_table
migrate create -ext sql -dir database/migrations -seq create_postAllowedUser_table
migrate create -ext sql -dir database/migrations -seq create_groupRequest_table
migrate create -ext sql -dir database/migrations -seq create_messageUser_table
migrate create -ext sql -dir database/migrations -seq create_comment_table
```