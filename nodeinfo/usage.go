package nodeinfo

// Usage describes statistics for a server.
type Usage struct {
	Users         UserUsage `json:"users"`
	LocalPosts    int       `json:"localPosts"`    // The amount of posts that were made by users that are registered on this server.
	LocalComments int       `json:"localComments"` // The amount of comments that were made by users that are registered on this server.
}

// UserUsage describes statistics about the users of this server.
type UserUsage struct {
	Total          int `json:"total"`          // The total amount of on this server registered users.
	ActiveHalfyear int `json:"activeHalfyear"` // The amount of users that signed in at least once in the last 180 days
	ActiveMonth    int `json:"activeMonth"`    // The amount of users that signed in at least once in the last 30 days.
}
