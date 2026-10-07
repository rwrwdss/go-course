package main

// mockTokens — «база» API-токенов. В реальности это JWT или сессии в Redis.
var mockTokens = map[string]User{
	"token-alice": {ID: 1, Name: "alice", Role: "admin"},
	"token-bob":   {ID: 2, Name: "bob", Role: "user"},
}
