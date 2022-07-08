package sargassum.pslive.web.policies.privatechat

import future.keywords

scope := {"/private-chats/**"}

default allow := false

# TODO: implement checks on these routes
routes := {
	# TODO: add checks to ensure the users exist and that the subject is one of the two users
	{"method": "SUB", "path": "/private-chats/*/*/chat/users"},
	{"method": "UNSUB", "path": "/private-chats/*/*/chat/users"},
	{"method": "MSG", "path": "/private-chats/*/*/chat/users"},
	{"method": "SUB", "path": "/private-chats/*/*/chat/messages"},
	{"method": "MSG", "path": "/private-chats/*/*/chat/messages"},
	{"method": "POST", "path": "/private-chats/*/*/chat/messages"},
}

errors contains "not implemented" if {
	not allow
}
